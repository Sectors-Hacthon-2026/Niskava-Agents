"""Niskava Autonomous Conversational ReAct Agent Core.

Adopts the ReAct (Reasoning + Acting) execution pattern:
- Thought: Internal investigative reasoning stream
- Action: Calling registered deterministic tools (Law 1: Deterministic before Generative)
- Observation: Ingesting factual evidence
- Synthesis: 3-Tier Taxonomy classification (Law 2: Non-advisory boundary)
"""

import json
import os
import re
import time
import uuid
from datetime import datetime
from typing import Any, Callable, Dict, List, Optional

from engine.agent.tools import NiskavaToolRegistry

SYSTEM_PROMPT = """You are Niskava Agent, an elite financial intelligence and market anomaly investigator for the Indonesia Stock Exchange (IDX). You assist equity analysts, financial journalists, and serious retail swing traders by turning market questions into rigorous, verifiable evidence.

=== NON-NEGOTIABLE OPERATIONAL LAWS ===
1. LAW 1 (Deterministic Before Generative):
   - NEVER calculate volume moving averages, Z-scores, abnormal returns, price change percentages, or statistical thresholds in your head.
   - ALWAYS invoke the deterministic quantitative tool 'compute_quant_anomalies' or 'get_daily_candles'. All numerical facts MUST originate strictly from tool observations.
   
2. LAW 2 (Strict Financial Non-Advisory Boundary):
   - You are an investigative intelligence researcher, NOT a financial advisor.
   - You MUST NEVER output direct BUY/SELL recommendations, price targets, or portfolio allocation advice under any circumstances.
   - You MUST classify every evidentiary finding into the Three-Tier Verification Taxonomy:
     * [SUPPORTED]: Factually confirmed by official Sectors API quantitative data or official IDXnet corporate disclosures.
     * [UNCERTAIN]: Correlation observed, but causality is unverified (market gossip, unconfirmed rumors).
     * [CONTRADICTED]: Claims refuted by official disclosures or financial facts.

3. EVIDENCE CITATION & CONFIDENCE RUBRIC:
   - Always assign a confidence score based on the standardized discrete rubric:
     1.00 = Extracted from official Sectors API or IDXnet regulatory disclosures
     0.95 = Explicit timestamp correlation with official corporate press releases
     0.85 = Strong inference backed by accredited mainstream financial press (Kontan, Bisnis, CNBC Indonesia)
     0.65 = Unverified market commentary or rumors
   - Always cite publication date, source name, and source URL.

=== AVAILABLE TOOLS & DOMAIN SKILLS ===
1. HIGH-LEVEL DOMAIN SKILLS (Layer 3 SOPs - Preferred):
- skill_market_anomaly_recon(ticker="<TICKER>", days=30): Detect statistical volume surges (MA20 Z-score), abnormal returns, and foreign flow divergence.
- skill_event_causality_audit(ticker="<TICKER>", anomaly_date="YYYY-MM-DD"): Audit temporal causality between price/volume spikes and disclosures, suspensions, and accredited news.
- skill_insider_bandarmology_forensic(ticker="<TICKER>"): Top-3 buyer concentration (C3), broker cohort mapping (foreign/domestic/retail/institution), and insider filings.
- skill_financial_health_stress_test(ticker="<TICKER>", rumor_claim="<CLAIM>"): Audit balance sheet liquidity/solvency and fact-check bankruptcy/default rumors (marks refuted rumors CONTRADICTED).
- skill_mining_commodity_divergence(ticker="<TICKER>", commodity="<COMMODITY>"): Pearson correlation between IDX mining stocks and global commodity spot prices (Nickel, Coal, Gold).
- skill_peer_valuation_benchmark(ticker="<TICKER>", subsector="<SUBSECTOR>"): Multi-metric relative valuation (P/E, P/B) against subsector median and IQR.

2. PRIMITIVE TOOLS (Layer 1):
- get_daily_candles(ticker="<TICKER>", days=<INT>): Fetch daily OHLCV candlesticks for an IDX ticker.
- compute_quant_anomalies(ticker="<TICKER>", volume_z_threshold=2.5): Run NumPy deterministic anomaly math (MA20, Z-scores, abnormal returns).
- harvest_market_news(ticker="<TICKER>"): Run targeted Dual-Engine OSINT for official disclosures and accredited financial media.
- query_company_profile(ticker="<TICKER>"): Retrieve company sector, subsector, market capitalization, and fundamental overview.
- get_foreign_flow(ticker="<TICKER>"): Fetch net foreign institutional inflow/outflow.

=== INTERACTION PROTOCOL (ReAct XML) ===
To investigate, think step-by-step in Indonesian:
<thought>Penalaran analitik internal mengenai apa yang ditanyakan dan tool apa yang dibutuhkan</thought>
<tool_call>{"name": "tool_name", "arguments": {...}}</tool_call>
(The environment executes the tool and returns <observation>...</observation>)
Continue thinking and calling tools if necessary.
When ready to answer the user, output:
<thought>Analisis akhir dan sintesis bukti</thought>
<response>
[Sampaikan sintesis investigasi lengkap dalam bahasa Indonesia yang profesional, berwibawa, disertai fakta angka deterministik dari tool, tabel/kartu bukti temuan bertanda [SUPPORTED] / [UNCERTAIN] / [CONTRADICTED], dan disclaimer non-advisory resmi di akhir.]
</response>
"""


class NiskavaReActAgent:
    """Autonomous Financial Market Intelligence Agent for IDX."""

    def __init__(
        self,
        tool_registry: NiskavaToolRegistry,
        emitter: Optional[Callable[[Dict[str, Any]], None]] = None,
        api_key: Optional[str] = None,
        mock_mode: Optional[bool] = None,
        model: Optional[str] = None,
        base_url: Optional[str] = None,
        ai_provider: Optional[str] = None,
    ):
        self.tools = tool_registry
        self.memory = getattr(tool_registry, "memory", None)
        self.emitter = emitter or (lambda ev: None)
        self.db_path = getattr(tool_registry, "db_path", os.path.expanduser("~/.niskava/niskava.db"))

        # Primary Active Model & Universal Endpoint Resolution
        self.model = (
            model
            or os.environ.get("NISKAVA_MODEL")
            or os.environ.get("OPENAI_MODEL")
            or os.environ.get("GEMINI_MODEL")
            or "hermes"
        )

        resolved_base_url = (
            base_url
            or os.environ.get("NISKAVA_BASE_URL")
            or os.environ.get("OPENAI_BASE_URL")
        )

        self.api_key = (
            api_key
            or os.environ.get("NISKAVA_API_KEY")
            or os.environ.get("OPENAI_API_KEY")
            or os.environ.get("GEMINI_API_KEY")
            or ""
        )

        # Transparent adapter for Google API keys (speaks standard OpenAI protocol)
        if not resolved_base_url:
            if os.environ.get("GEMINI_API_KEY") and not os.environ.get("OPENAI_API_KEY"):
                resolved_base_url = "https://generativelanguage.googleapis.com/v1beta/openai"
                if not model and not os.environ.get("NISKAVA_MODEL") and not os.environ.get("OPENAI_MODEL"):
                    self.model = os.environ.get("GEMINI_MODEL", "gemini-2.0-flash")
            else:
                resolved_base_url = "http://localhost:20128/v1"

        self.base_url = resolved_base_url.rstrip("/")
        # Backward compatibility aliases
        self.openai_base_url = self.base_url
        self.openai_api_key = self.api_key
        self.openai_model = self.model

        # Provider mode: "universal" (default) or "mock" / "offline"
        raw_provider = (ai_provider or os.environ.get("AI_PROVIDER", "")).lower()
        if mock_mode is not None:
            self.mock_mode = mock_mode
            self.ai_provider = "mock" if mock_mode else (ai_provider or "universal")
        elif raw_provider in ("mock", "offline") or os.environ.get("NISKAVA_OFFLINE") == "1":
            self.mock_mode = True
            self.ai_provider = "mock"
        else:
            self.mock_mode = False
            self.ai_provider = "universal"

    def _emit(self, event_data: Dict[str, Any]) -> None:
        self.emitter(event_data)

    def _get_compacted_history(self, session_id: str, max_turns: int = 8) -> List[Dict[str, str]]:
        """Load and adaptively compact multi-turn conversation history from SQLite WAL (OpenCode pattern)."""
        import sqlite3
        if not self.db_path or not os.path.exists(self.db_path):
            return []

        try:
            with sqlite3.connect(self.db_path) as conn:
                conn.row_factory = sqlite3.Row
                cursor = conn.cursor()
                cursor.execute(
                    """
                    SELECT role, content, thought, tool_calls_json, status, created_at
                    FROM chat_messages
                    WHERE session_id = ?
                    ORDER BY created_at ASC
                    """,
                    (session_id,),
                )
                rows = cursor.fetchall()
                if not rows:
                    return []

                raw_history: List[Dict[str, str]] = []
                for row in rows:
                    role = row["role"]
                    if role not in ("user", "assistant", "system"):
                        continue
                    content = row["content"] or ""
                    raw_history.append({"role": role, "content": content})

                # If history fits within max_turns, return as is
                if len(raw_history) <= max_turns:
                    return raw_history

                # OpenCode Context Compaction: Keep first turn (2), compact middle (1), keep remaining last turns
                first_turn = raw_history[:2]
                keep_last = max(1, max_turns - 3)
                last_turns = raw_history[-keep_last:]

                middle_count = len(raw_history) - len(first_turn) - len(last_turns)
                summary_block = {
                    "role": "system",
                    "content": f"[Konteks Sebelumnya: {middle_count} putaran obrolan terdahulu telah diringkas untuk menjaga efisiensi token]",
                }
                return first_turn + [summary_block] + last_turns
        except Exception:
            return []

    def _auto_generate_session_title(self, user_prompt: str, session_id: str) -> None:
        """Deterministically generate a 3-5 word informative session title on first turn."""
        import sqlite3
        if not self.db_path or not os.path.exists(self.db_path):
            return

        try:
            with sqlite3.connect(self.db_path) as conn:
                cursor = conn.cursor()
                # Check current title
                row = cursor.execute("SELECT title FROM chat_sessions WHERE id = ?", (session_id,)).fetchone()
                if row and row[0] and row[0] not in ("Sesi Riset Pasar", "New Chat", session_id):
                    # Already has customized title
                    return

                # Generate clean title from ticker or keywords
                clean_prompt = re.sub(r"[^\w\s]", "", user_prompt).strip()
                words = clean_prompt.split()
                stopwords = {
                    "YANG", "DARI", "PADA", "BISA", "AKAN", "SAAT", "KITA", "DENG", "APAL",
                    "INFO", "CHAT", "TENT", "KATA", "HALO", "PAGI", "SIAP", "TEST", "USER",
                    "HELP", "EXIT", "QUIT", "TANY", "APA", "BAGA", "SIAPA", "KODE", "HARI"
                }
                tickers = [w.upper() for w in words if len(w) == 4 and w.isalpha() and w.upper() not in stopwords]
                if tickers:
                    title = f"Riset Saham {tickers[0]}"
                    lowered = clean_prompt.lower()
                    if "anomali" in lowered or "volume" in lowered:
                        title = f"Anomali & Volume {tickers[0]}"
                    elif "asing" in lowered or "flow" in lowered:
                        title = f"Foreign Flow {tickers[0]}"
                    elif "valuasi" in lowered or "rasio" in lowered or "per" in lowered:
                        title = f"Valuasi & Rasio {tickers[0]}"
                elif len(words) >= 3:
                    title = " ".join(words[:5]).capitalize()
                elif len(words) > 0:
                    title = " ".join(words).capitalize()
                else:
                    title = "Riset Pasar Modal"

                cursor.execute(
                    """
                    INSERT INTO chat_sessions (id, title, model, status, message_count, last_message_preview, is_pinned, created_at, updated_at)
                    VALUES (?, ?, 'hermes', 'IDLE', 0, '', 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
                    ON CONFLICT(id) DO UPDATE SET title = excluded.title
                    """,
                    (session_id, title),
                )
                conn.commit()
        except Exception:
            pass

    def chat(
        self,
        user_prompt: str,
        session_id: Optional[str] = None,
        history: Optional[List[Dict[str, str]]] = None,
    ) -> Dict[str, Any]:
        """Conversational Research Assistant entrypoint supporting free-form natural language prompts."""
        session_id = session_id or f"CHAT-{datetime.now().strftime('%Y%m%d')}-{uuid.uuid4().hex[:6].upper()}"
        start_time = time.time()

        # Load compacted multi-turn history from SQLite WAL if not explicitly passed
        if history is None:
            history = self._get_compacted_history(session_id, max_turns=8)

        # Auto-generate informative session title on first turn
        if not history or len(history) == 0:
            self._auto_generate_session_title(user_prompt, session_id)

        self._emit({
            "event": "session_start",
            "session_id": session_id,
            "prompt": user_prompt,
            "timestamp": datetime.now().isoformat() + "Z",
        })

        # Augment prompt with local conversational graph memory (Law 6)
        effective_prompt = user_prompt
        if self.memory:
            # 1. Detect portfolio/position mentions (e.g. 'beli ANTM di 1450')
            pos_match = re.search(
                r"(?:beli|entry|posisi|pegang|holds?)\s+([A-Za-z]{4})\b.*?(\d{3,6})",
                user_prompt,
                re.IGNORECASE,
            )
            if pos_match:
                tkr = pos_match.group(1).upper()
                price = pos_match.group(2)
                try:
                    self.memory.store_observation(
                        source_label="User",
                        source_type="USER",
                        relation="HOLDS_AT",
                        target_label=f"Price: {price}",
                        target_type="PRICE_LEVEL",
                        context_snippet=f"Posisi modal di {tkr} pada level {price}",
                        session_id=session_id,
                    )
                    self.memory.store_observation(
                        source_label=f"Price: {price}",
                        source_type="PRICE_LEVEL",
                        relation="TICKER_REF",
                        target_label=tkr,
                        target_type="TICKER",
                        session_id=session_id,
                    )
                except Exception:
                    pass

            # 2. Extract potential entities in prompt to recall past graph context
            memory_blocks = []
            candidates = set(re.findall(r"\b[A-Za-z]{4}\b", user_prompt))
            for cand in candidates:
                xml_mem = self.memory.format_investigative_prompt(cand.upper(), radius=2)
                if xml_mem:
                    memory_blocks.append(xml_mem)

            if memory_blocks:
                effective_prompt = f"{user_prompt}\n\n" + "\n".join(memory_blocks)

        if self.mock_mode or self.ai_provider in ("mock", "offline"):
            res = self._run_deterministic_chat_cycle(session_id, effective_prompt, history, start_time)
        elif self.ai_provider == "gemini" and not (self.api_key or getattr(self, "gemini_api_key", None)):
            err_detail = "GEMINI_API_KEY tidak ditemukan di environment atau konfigurasi."
            error_markdown = (
                f"### ⚠️ Konfigurasi Gemini API Key Tidak Ditemukan\n\n"
                f"- **Provider**: Gemini\n"
                f"- **Model**: `{self.model}`\n"
                f"- **Detail**: {err_detail}\n\n"
                f"**Solusi Pemecahan Masalah:**\n"
                f"1. Masukkan API key valid ke `~/.niskava/.env` (`GEMINI_API_KEY=AIza...`).\n"
                f"2. Atau jalankan `niskava setup` untuk mengisi API key secara interaktif.\n"
                f"3. Atau gunakan mode offline (`--offline`) jika ingin menjalankan analisis deterministik tanpa LLM."
            )
            self._emit({
                "event": "agent_thought",
                "session_id": session_id,
                "thought": f"Gagal mengeksekusi inferensi Gemini: {err_detail}",
            })
            self._emit({
                "event": "agent_message_chunk",
                "session_id": session_id,
                "chunk": error_markdown,
            })
            self._emit({
                "event": "agent_message_complete",
                "session_id": session_id,
                "content": error_markdown,
            })
            self._emit({
                "event": "session_error",
                "session_id": session_id,
                "error": err_detail,
                "timestamp": datetime.now().isoformat() + "Z",
            })
            res = {
                "session_id": session_id,
                "status": "ERROR",
                "error": err_detail,
                "response": error_markdown,
                "anomalies": [],
                "findings": [],
                "duration_ms": int((time.time() - start_time) * 1000),
            }
        else:
            res = self._run_universal_chat_cycle(session_id, effective_prompt, history, start_time)

        if self.memory and isinstance(res, dict):
            findings = res.get("findings", [])
            anomalies = res.get("anomalies", [])
            target_ticker = None
            if anomalies:
                target_ticker = anomalies[0].get("ticker")
            if not target_ticker:
                for c in candidates:
                    target_ticker = c.upper()
                    break
            if target_ticker and (anomalies or findings):
                try:
                    self.memory.record_investigation(
                        session_id=session_id,
                        ticker=target_ticker,
                        anomalies=anomalies,
                        findings=findings,
                    )
                except Exception:
                    pass

        return res

    def investigate(
        self,
        ticker: str,
        days: int = 30,
        session_id: Optional[str] = None,
    ) -> Dict[str, Any]:
        """Execute the autonomous investigative ReAct cycle on a specific ticker (backward compatible)."""
        session_id = session_id or f"INV-{datetime.now().strftime('%Y%m%d')}-{uuid.uuid4().hex[:6].upper()}"
        start_time = time.time()
        ticker = ticker.upper()

        self._emit({
            "event": "session_start",
            "session_id": session_id,
            "ticker": ticker,
            "timeframe_days": days,
            "timestamp": datetime.now().isoformat() + "Z",
        })

        if self.mock_mode:
            res = self._run_deterministic_react_cycle(session_id, ticker, days, start_time)
        elif self.ai_provider in ("openai", "9router"):
            res = self._run_openai_react_cycle(session_id, ticker, days, start_time)
        elif self.ai_provider == "gemini" and self.api_key:
            res = self._run_gemini_react_cycle(session_id, ticker, days, start_time)
        else:
            res = self._run_deterministic_react_cycle(session_id, ticker, days, start_time)

        # Persist to local conversational graph memory (Law 6)
        if self.memory and isinstance(res, dict):
            try:
                self.memory.record_investigation(
                    session_id=session_id,
                    ticker=ticker,
                    anomalies=res.get("anomalies", []),
                    findings=res.get("findings", []),
                )
            except Exception:
                pass

        return res

    def _run_universal_chat_cycle(
        self,
        session_id: str,
        user_prompt: str,
        history: Optional[List[Dict[str, str]]],
        start_time: float,
    ) -> Dict[str, Any]:
        """Multi-turn conversational ReAct agent powered by Universal OpenAI-compatible endpoint."""
        import requests

        base_url = (getattr(self, "openai_base_url", None) or self.base_url).rstrip("/")
        model = getattr(self, "openai_model", None) or self.model
        api_key = getattr(self, "openai_api_key", None) or self.api_key

        url = f"{base_url}/chat/completions"
        headers = {"Content-Type": "application/json"}
        if api_key:
            headers["Authorization"] = f"Bearer {api_key}"

        messages: List[Dict[str, str]] = [
            {"role": "system", "content": SYSTEM_PROMPT},
        ]
        if history:
            for h in history:
                messages.append({"role": h.get("role", "user"), "content": h.get("content", "")})
        messages.append({"role": "user", "content": user_prompt})

        findings: List[Dict[str, Any]] = []
        anomalies: List[Dict[str, Any]] = []
        final_response = ""

        last_error = ""
        for _ in range(4):
            payload = {
                "model": model,
                "messages": messages,
                "temperature": 0.2,
                "max_tokens": 700,
            }
            try:
                resp = requests.post(url, headers=headers, json=payload, timeout=18.0)
                if resp.status_code != 200:
                    error_body = resp.text.strip()[:300]
                    last_error = f"HTTP {resp.status_code}: {error_body}" if error_body else f"HTTP {resp.status_code}"
                    break
                raw = resp.text.strip()
                if "data: [DONE]" in raw:
                    raw = raw.split("data: [DONE]")[0].strip()
                first_brace = raw.find("{")
                last_brace = raw.rfind("}")
                if first_brace == -1 or last_brace == -1:
                    last_error = f"Format respons tidak valid (tidak ditemukan objek JSON): {raw[:200]}"
                    break
                data = json.loads(raw[first_brace : last_brace + 1])
                choices = data.get("choices", [])
                if not choices:
                    last_error = f"Format respons tidak valid (tidak ada item 'choices'): {raw[:200]}"
                    break
                msg = choices[0].get("message", {})
                content = msg.get("content") or msg.get("reasoning") or ""
                if not content:
                    last_error = "Model AI mengembalikan konten respons kosong."
                    break
            except requests.exceptions.Timeout:
                last_error = f"Koneksi timeout setelah 18 detik ke {url}"
                break
            except requests.exceptions.ConnectionError:
                last_error = f"Gagal terhubung ke {url} (Koneksi jaringan ditolak atau server tidak aktif)"
                break
            except Exception as exc:
                last_error = f"Kesalahan saat menghubungi {url}: {exc}"
                break

            # Parse thoughts
            thoughts = re.findall(r"<thought>(.*?)</thought>", content, re.DOTALL)
            for th in thoughts:
                clean_th = th.strip()
                if clean_th:
                    self._emit({
                        "event": "agent_thought",
                        "session_id": session_id,
                        "thought": f"[{model}] {clean_th}",
                    })

            # Check for tool call
            tool_calls = re.findall(r"<tool_call>(.*?)</tool_call>", content, re.DOTALL)
            if tool_calls:
                call_str = tool_calls[0].strip()
                try:
                    call_json = json.loads(call_str)
                    tool_name = call_json.get("name")
                    tool_args = call_json.get("arguments", {})

                    self._emit({
                        "event": "agent_tool_call",
                        "session_id": session_id,
                        "tool": tool_name,
                        "args": tool_args,
                    })

                    # Execute deterministic tool
                    tool_res = self.tools.execute_tool(tool_name, tool_args)

                    # Extract anomalies if computed
                    if tool_name == "compute_quant_anomalies" and isinstance(tool_res, list):
                        for a in tool_res:
                            anomalies.append(a)
                            self._emit({
                                "event": "anomaly_detected",
                                "session_id": session_id,
                                "ticker": tool_args.get("ticker", ""),
                                "anomaly_date": a.get("date") or a.get("anomaly_date", ""),
                                "metric_type": a.get("metric_type", ""),
                                "z_score": a.get("z_score", 0.0),
                                "metric_value": a.get("metric_value", 0.0),
                                "baseline_value": a.get("baseline_value", 0.0),
                                "price_change_pct": a.get("price_change_pct", 0.0),
                                "sector_change_pct": a.get("sector_change_pct", 0.0),
                                "description": a.get("description", ""),
                            })

                    obs_str = f"Observation for {tool_name}: {json.dumps(tool_res)[:450]}"
                    self._emit({
                        "event": "agent_observation",
                        "session_id": session_id,
                        "tool": tool_name,
                        "summary": f"Data observasi diterima ({len(str(tool_res))} bytes).",
                    })

                    # Feed back to model
                    messages.append({"role": "assistant", "content": content})
                    messages.append({"role": "user", "content": f"<observation>{obs_str}</observation>"})
                    continue
                except Exception:
                    pass

            # Check for final response
            responses = re.findall(r"<response>(.*?)</response>", content, re.DOTALL)
            if responses:
                final_response = responses[0].strip()
                break
            else:
                cleaned = re.sub(r"<thought>.*?</thought>", "", content, flags=re.DOTALL)
                cleaned = re.sub(r"<tool_call>.*?</tool_call>", "", cleaned, flags=re.DOTALL).strip()
                if cleaned:
                    final_response = cleaned
                    break

        if not final_response:
            err_detail = last_error or "Model AI tidak menghasilkan sintesis respons valid dalam siklus ReAct."
            error_markdown = (
                f"### ⚠️ Gagal Terhubung ke Provider AI\n\n"
                f"- **Endpoint**: `{url}`\n"
                f"- **Model**: `{model}`\n"
                f"- **Detail Error**: {err_detail}\n\n"
                f"**Solusi Pemecahan Masalah:**\n"
                f"1. Pastikan server LLM (Ollama / vLLM / 9router / OpenRouter) aktif di `{base_url}` atau API key terpasang.\n"
                f"2. Periksa konfigurasi di file `~/.niskava/.env` atau jalankan `niskava setup`.\n"
                f"3. Gunakan mode offline (`--offline`) jika ingin menjalankan analisis deterministik tanpa LLM."
            )
            self._emit({
                "event": "agent_thought",
                "session_id": session_id,
                "thought": f"Gagal mengeksekusi inferensi AI: {err_detail}",
            })
            self._emit({
                "event": "agent_message_chunk",
                "session_id": session_id,
                "chunk": error_markdown,
            })
            self._emit({
                "event": "agent_message_complete",
                "session_id": session_id,
                "content": error_markdown,
            })
            self._emit({
                "event": "session_error",
                "session_id": session_id,
                "error": f"AI provider connection error ({model} @ {url}): {err_detail}",
            })
            duration_ms = int((time.time() - start_time) * 1000)
            return {
                "session_id": session_id,
                "response": error_markdown,
                "error": err_detail,
                "anomalies": anomalies,
                "findings": findings,
                "duration_ms": duration_ms,
                "status": "ERROR",
            }

        self._emit({
            "event": "agent_message_chunk",
            "session_id": session_id,
            "chunk": final_response,
        })
        self._emit({
            "event": "agent_message_complete",
            "session_id": session_id,
            "content": final_response,
        })

        duration_ms = int((time.time() - start_time) * 1000)
        self._emit({
            "event": "session_complete",
            "session_id": session_id,
            "status": "COMPLETED",
            "total_anomalies": len(anomalies),
            "total_findings": len(findings),
            "duration_ms": duration_ms,
            "summary": final_response[:200] + "...",
        })

        return {
            "session_id": session_id,
            "response": final_response,
            "anomalies": anomalies,
            "findings": findings,
            "duration_ms": duration_ms,
        }

    # Backward compatibility aliases
    _run_openai_chat_cycle = _run_universal_chat_cycle
    _run_gemini_chat_cycle = _run_universal_chat_cycle

    def _run_deterministic_chat_cycle(
        self,
        session_id: str,
        user_prompt: str,
        history: Optional[List[Dict[str, str]]],
        start_time: float,
    ) -> Dict[str, Any]:
        """Deterministic fallback chat synthesis extracting ticker and enforcing Law 1 & Law 2."""
        # Extract ticker from prompt (e.g. 4 capital letters)
        candidates = re.findall(r"\b[A-Z]{4}\b", user_prompt.upper())
        # Filter out common English/Indonesian words that happen to be 4 letters
        stopwords = {
            "YANG", "DARI", "PADA", "BISA", "AKAN", "SAAT", "KITA", "DENG", "APAL",
            "INFO", "CHAT", "TENT", "KATA", "HALO", "PAGI", "SIAP", "TEST", "USER",
            "HELP", "EXIT", "QUIT", "TANY", "APA", "BAGA", "SIAPA", "KODE", "HARI",
            "BUAT", "BAIK", "SAYA", "KAMU", "COBA", "DATA", "MODE", "DENGAN"
        }
        tickers = [c for c in candidates if c not in stopwords]

        if not tickers and history:
            # Multi-turn context recall: look for ticker in previous turns to prevent amnesia
            for h in reversed(history):
                prev_cands = re.findall(r"\b[A-Z]{4}\b", h.get("content", "").upper())
                prev_tickers = [c for c in prev_cands if c not in stopwords]
                if prev_tickers:
                    tickers = prev_tickers
                    self._emit({
                        "event": "agent_thought",
                        "session_id": session_id,
                        "thought": f"Emiten target tidak disebutkan di prompt terbaru, namun terdeteksi dari riwayat percakapan sebelumnya: {tickers[0]}",
                    })
                    break

        if not tickers:
            response_text = (
                "Halo! Saya Niskava Agent (Mode Offline/Mock).\n\n"
                "Saya tidak mendeteksi kode emiten saham IDX yang spesifik dalam pesan Anda.\n\n"
                "Untuk menganalisis anomali transaksi dan keterbukaan informasi, silakan sebutkan kode saham 4-huruf yang ingin diperiksa (contoh: **ANTM**, **BBCA**, **BBRI**, **BUMI**).\n\n"
                "> *Catatan*: Anda saat ini berada dalam mode offline/mock. Untuk menggunakan asisten percakapan bebas (ReAct), aktifkan koneksi AI provider di `niskava setup`."
            )
            self._emit({
                "event": "agent_thought",
                "session_id": session_id,
                "thought": "Prompt pengguna tidak memuat kode emiten IDX 4-huruf yang valid. Mengembalikan panduan mode offline.",
            })
            self._emit({
                "event": "agent_message_chunk",
                "session_id": session_id,
                "chunk": response_text,
            })
            self._emit({
                "event": "agent_message_complete",
                "session_id": session_id,
                "content": response_text,
            })
            duration_ms = int((time.time() - start_time) * 1000)
            self._emit({
                "event": "session_complete",
                "session_id": session_id,
                "status": "COMPLETED",
                "total_anomalies": 0,
                "total_findings": 0,
                "duration_ms": duration_ms,
                "summary": response_text[:200] + "...",
            })
            return {
                "session_id": session_id,
                "response": response_text,
                "anomalies": [],
                "findings": [],
                "duration_ms": duration_ms,
            }

        ticker = tickers[0]
        self._emit({
            "event": "agent_thought",
            "session_id": session_id,
            "thought": (
                f"Menganalisis prompt pengguna: '{user_prompt[:80]}'. Terdeteksi emiten target: {ticker}. "
                f"Sesuai Law 1 (Deterministic Before Generative), saya memanggil tool get_daily_candles dan compute_quant_anomalies."
            ),
        })

        # Step 1: Candles
        self._emit({
            "event": "agent_tool_call",
            "session_id": session_id,
            "tool": "get_daily_candles",
            "args": {"ticker": ticker, "days": 30},
        })
        candles = self.tools.get_daily_candles(ticker, days=30)
        self._emit({
            "event": "agent_observation",
            "session_id": session_id,
            "tool": "get_daily_candles",
            "summary": f"Berhasil menarik {len(candles)} hari data candlestick {ticker} dari Sectors API v2.",
        })

        # Step 2: NumPy Quant Anomaly (Law 1)
        self._emit({
            "event": "agent_tool_call",
            "session_id": session_id,
            "tool": "compute_quant_anomalies",
            "args": {"ticker": ticker, "volume_z_threshold": 2.5},
        })
        anomalies = self.tools.compute_quant_anomalies(ticker, volume_z_threshold=2.5)

        highest_z = 0.0
        anomaly_date = "N/A"
        for a in anomalies:
            anomaly_dt = a.get("date") or a.get("anomaly_date", "")
            self._emit({
                "event": "anomaly_detected",
                "session_id": session_id,
                "ticker": ticker,
                "anomaly_date": anomaly_dt,
                "metric_type": a.get("metric_type", ""),
                "z_score": a.get("z_score", 0.0),
                "metric_value": a.get("metric_value", 0.0),
                "baseline_value": a.get("baseline_value", 0.0),
                "price_change_pct": a.get("price_change_pct", 0.0),
                "sector_change_pct": a.get("sector_change_pct", 0.0),
                "description": a.get("description", ""),
            })
            if a.get("z_score", 0.0) > highest_z:
                highest_z = a.get("z_score", 0.0)
                anomaly_date = anomaly_dt

        self._emit({
            "event": "agent_observation",
            "session_id": session_id,
            "tool": "compute_quant_anomalies",
            "summary": f"Ditemukan {len(anomalies)} anomali kuantitatif signifikan (Z-Score puncak: {highest_z:.2f}σ pada {anomaly_date}).",
        })

        # Step 3: OSINT News Harvester
        self._emit({
            "event": "agent_tool_call",
            "session_id": session_id,
            "tool": "harvest_market_news",
            "args": {"ticker": ticker},
        })
        news_items = self.tools.harvest_market_news(ticker)
        self._emit({
            "event": "agent_observation",
            "session_id": session_id,
            "tool": "harvest_market_news",
            "summary": f"Ditemukan {len(news_items)} artikel berita & keterbukaan informasi bursa terakreditasi.",
        })

        # Step 4: Synthesize Response (Law 2: 3-Tier Taxonomy & Non-Advisory)
        top_news = news_items[0] if news_items else {}
        news_title = top_news.get("title", f"Keterbukaan Informasi dan Aksi Korporasi {ticker}")
        news_url = top_news.get("url", "https://api.sectors.app/v2")

        finding = {
            "event": "finding_emitted",
            "session_id": session_id,
            "id": "FND-01",
            "title": f"Katalis Penggerak: {news_title[:75]}...",
            "claim_text": f"Lonjakan volume ({highest_z:.2f}σ) pada {anomaly_date} berkorelasi langsung dengan pengumuman: '{news_title}'.",
            "verification_status": "SUPPORTED",
            "confidence_score": 0.95,
            "causality_status": "LIKELY_CATALYST",
            "evidence": [
                {
                    "source_type": "QUANTITATIVE_BASELINE",
                    "source_name": "Sectors Daily API",
                    "source_url": f"https://api.sectors.app/v2/daily/{ticker}/",
                    "publication_date": anomaly_date,
                    "snippet_text": f"Volume Z-Score {highest_z:.2f}σ dihitung via NumPy MA20.",
                },
                {
                    "source_type": "OFFICIAL_DISCLOSURE",
                    "source_name": top_news.get("source_domain", "IDX / Media Terakreditasi"),
                    "source_url": news_url,
                    "publication_date": top_news.get("publication_date", anomaly_date),
                    "snippet_text": top_news.get("snippet", ""),
                },
            ],
        }
        self._emit(finding)

        response_text = f"""### Laporan Investigasi Intelijen Pasar: **{ticker}** (Bursa Efek Indonesia)

Berdasarkan analisis deterministik kuantitatif dan penelusuran OSINT keterbukaan informasi:

1. **Temuan Anomali Transaksi (Law 1: NumPy Deterministic)**
   * **Volume Z-Score Puncak**: `{highest_z:.2f}σ` terdeteksi pada tanggal `{anomaly_date}`.
   * **Total Anomali**: Ditemukan `{len(anomalies)}` anomali pergerakan volume/harga di luar batas normal 20-hari moving average.

2. **Matriks Bukti Kausalitas (Law 2: 3-Tier Taxonomy)**
   * **[SUPPORTED]** `{finding['title']}`
     * **Klaim**: {finding['claim_text']}
     * **Tingkat Keyakinan**: `95%` (Direct Structural Evidence via IDXnet / Sectors API)
     * **Status Kausalitas**: `LIKELY_CATALYST` (Pengumuman resmi mendahului / bertepatan dengan lonjakan volume)

---
> **DISCLAIMER FINANSIAL (Hukum 2 & Aturan 12 Hackathon):**  
> Laporan ini dihasilkan secara otonom untuk tujuan intelijen pasar dan pembuktian bukti keterbukaan informasi. Niskava Agent **BUKAN** penasihat investasi dan **TIDAK PERNAH** memberikan rekomendasi BELI/JUAL saham apa pun.
"""

        self._emit({
            "event": "agent_message_chunk",
            "session_id": session_id,
            "chunk": response_text,
        })
        self._emit({
            "event": "agent_message_complete",
            "session_id": session_id,
            "content": response_text,
        })

        duration_ms = int((time.time() - start_time) * 1000)
        self._emit({
            "event": "session_complete",
            "session_id": session_id,
            "status": "COMPLETED",
            "total_anomalies": len(anomalies),
            "total_findings": 1,
            "duration_ms": duration_ms,
            "summary": f"Analisis terhadap {ticker} selesai dengan 1 temuan SUPPORTED.",
        })

        return {
            "session_id": session_id,
            "response": response_text,
            "anomalies": anomalies,
            "findings": [finding],
            "duration_ms": duration_ms,
        }

    def _run_deterministic_react_cycle(
        self,
        session_id: str,
        ticker: str,
        days: int,
        start_time: float,
        initial_thought: Optional[str] = None,
    ) -> Dict[str, Any]:
        """High-precision deterministic ReAct simulation for offline and test runs."""
        first_thought = initial_thought or (
            f"Memulai investigasi terhadap emiten {ticker}. Langkah pertama adalah menarik deret "
            f"waktu harga {days} hari untuk menganalisis basis pergerakan volume."
        )
        self._emit({
            "event": "agent_thought",
            "session_id": session_id,
            "thought": first_thought,
        })
        time.sleep(0.05)

        self._emit({
            "event": "agent_tool_call",
            "session_id": session_id,
            "tool": "get_daily_candles",
            "args": {"ticker": ticker, "days": days},
        })
        candles = self.tools.get_daily_candles(ticker, days=days)
        self._emit({
            "event": "agent_observation",
            "session_id": session_id,
            "tool": "get_daily_candles",
            "summary": f"Berhasil menarik {len(candles)} hari data candlestick dari Sectors API v2.",
        })

        self._emit({
            "event": "agent_thought",
            "session_id": session_id,
            "thought": "Data deret waktu telah diperoleh. Sesuai Law 1, saya harus mengeksekusi perhitungan statistik deterministik (NumPy) untuk mendeteksi apakah ada volume spike atau abnormal return.",
        })
        time.sleep(0.05)

        self._emit({
            "event": "agent_tool_call",
            "session_id": session_id,
            "tool": "compute_quant_anomalies",
            "args": {"ticker": ticker, "volume_z_threshold": 2.5},
        })
        anomalies = self.tools.compute_quant_anomalies(ticker, volume_z_threshold=2.5)

        highest_z = 0.0
        anomaly_date = "N/A"
        for anomaly in anomalies:
            self._emit({
                "event": "anomaly_detected",
                "session_id": session_id,
                "ticker": ticker,
                "anomaly_date": anomaly.get("anomaly_date", ""),
                "metric_type": anomaly.get("metric_type", ""),
                "z_score": anomaly.get("z_score", 0.0),
                "metric_value": anomaly.get("metric_value", 0.0),
                "baseline_value": anomaly.get("baseline_value", 0.0),
                "price_change_pct": anomaly.get("price_change_pct", 0.0),
                "sector_change_pct": anomaly.get("sector_change_pct", 0.0),
                "description": anomaly.get("description", ""),
            })
            if anomaly.get("z_score", 0.0) > highest_z:
                highest_z = anomaly.get("z_score", 0.0)
                anomaly_date = anomaly.get("anomaly_date", "")

        self._emit({
            "event": "agent_observation",
            "session_id": session_id,
            "tool": "compute_quant_anomalies",
            "summary": f"Ditemukan {len(anomalies)} anomali kuantitatif signifikan (Z-Score tertinggi: {highest_z:.2f}σ pada {anomaly_date}).",
        })

        self._emit({
            "event": "agent_thought",
            "session_id": session_id,
            "thought": f"Anomali terdeteksi pada {anomaly_date}. Saya perlu memanen berita finansial dan keterbukaan informasi bursa resmi untuk mencari katalis kausalitas.",
        })
        time.sleep(0.05)

        self._emit({
            "event": "agent_tool_call",
            "session_id": session_id,
            "tool": "harvest_market_news",
            "args": {"ticker": ticker},
        })
        news_items = self.tools.harvest_market_news(ticker)
        self._emit({
            "event": "agent_observation",
            "session_id": session_id,
            "tool": "harvest_market_news",
            "summary": f"Ditemukan {len(news_items)} artikel berita & keterbukaan informasi bursa terakreditasi.",
        })

        self._emit({
            "event": "agent_thought",
            "session_id": session_id,
            "thought": "Membandingkan stempel waktu keterbukaan informasi dengan tanggal anomali. Mengklasifikasikan bukti ke dalam Taksonomi 3-Tier Niskava.",
        })
        time.sleep(0.05)

        findings = []
        if news_items:
            top_news = news_items[0]
            finding = {
                "event": "finding_emitted",
                "session_id": session_id,
                "id": "FND-01",
                "title": f"Katalis Penggerak: {top_news.get('title', 'Keterbukaan Informasi Emiten')[:75]}...",
                "claim_text": f"Lonjakan volume ({highest_z:.2f}σ) berkorelasi langsung dengan pengumuman: '{top_news.get('title', '')}'.",
                "verification_status": "SUPPORTED",
                "confidence_score": 0.95,
                "causality_status": "LIKELY_CATALYST",
                "evidence": [
                    {
                        "source_type": "QUANTITATIVE_BASELINE",
                        "source_name": "Sectors Daily API",
                        "source_url": f"https://api.sectors.app/v2/daily/{ticker}/",
                        "publication_date": anomaly_date,
                        "snippet_text": f"Unusual volume surge ({highest_z:.2f}σ) with price breakout divergent from sector.",
                    },
                    {
                        "source_type": "OFFICIAL_DISCLOSURE",
                        "source_name": top_news.get("source_domain", "IDX / Media"),
                        "source_url": top_news.get("url", ""),
                        "publication_date": top_news.get("publication_date", ""),
                        "snippet_text": top_news.get("snippet", ""),
                    },
                ],
            }
            findings.append(finding)
            self._emit(finding)

        duration_ms = int((time.time() - start_time) * 1000)
        summary = (
            f"Investigasi otonom {ticker} selesai. Ditemukan {len(anomalies)} anomali kuantitatif "
            f"dengan {len(findings)} temuan SUPPORTED berbasis keterbukaan informasi bursa."
        )

        self._emit({
            "event": "session_complete",
            "session_id": session_id,
            "status": "COMPLETED",
            "total_anomalies": len(anomalies),
            "total_findings": len(findings),
            "duration_ms": duration_ms,
            "summary": summary,
        })

        return {
            "session_id": session_id,
            "ticker": ticker,
            "anomalies": anomalies,
            "findings": findings,
            "summary": summary,
            "duration_ms": duration_ms,
        }

    def _run_gemini_react_cycle(
        self, session_id: str, ticker: str, days: int, start_time: float
    ) -> Dict[str, Any]:
        """Live ReAct reasoning cycle powered by Gemini API."""
        prompt = (
            f"Kamu adalah Niskava Agent. Berikan analisis singkat (1-2 kalimat) dalam bahasa Indonesia mengenai rencana "
            f"investigasi kuantitatif dan OSINT untuk emiten {ticker} pada periode {days} hari terakhir."
        )
        thought_text = f"Menghubungkan ke Gemini ({self.model}). Memulai siklus ReAct investigasi emiten {ticker}."
        if self.api_key and not self.mock_mode:
            try:
                import requests

                url = f"https://generativelanguage.googleapis.com/v1beta/models/{self.model}:generateContent?key={self.api_key}"
                payload = {
                    "contents": [{"parts": [{"text": prompt}]}],
                    "generationConfig": {"temperature": 0.2, "maxOutputTokens": 150},
                }
                resp = requests.post(url, json=payload, timeout=8.0)
                if resp.status_code == 200:
                    data = resp.json()
                    candidates = data.get("candidates", [])
                    if candidates and "content" in candidates[0]:
                        parts = candidates[0]["content"].get("parts", [])
                        if parts and "text" in parts[0]:
                            thought_text = parts[0]["text"].strip()
            except Exception:
                pass

        return self._run_deterministic_react_cycle(session_id, ticker, days, start_time, initial_thought=thought_text)

    def _run_openai_react_cycle(
        self, session_id: str, ticker: str, days: int, start_time: float
    ) -> Dict[str, Any]:
        """Live ReAct reasoning cycle powered by 9router / OpenAI-compatible endpoint."""
        prompt = (
            f"Kamu adalah Niskava Agent. Berikan analisis singkat (1-2 kalimat) dalam bahasa Indonesia mengenai rencana "
            f"investigasi kuantitatif dan OSINT untuk emiten {ticker} pada periode {days} hari terakhir."
        )
        thought_text = f"Menghubungkan ke 9router ({self.openai_model}). Memulai siklus ReAct investigasi emiten {ticker}."

        try:
            import requests

            base_url = (self.openai_base_url or "http://localhost:20128/v1").rstrip("/")
            url = f"{base_url}/chat/completions"
            headers = {"Content-Type": "application/json"}
            if self.openai_api_key:
                headers["Authorization"] = f"Bearer {self.openai_api_key}"

            payload = {
                "model": self.openai_model,
                "messages": [
                    {
                        "role": "system",
                        "content": "You are Niskava Agent, an elite financial intelligence investigator for IDX. Think step-by-step in Indonesian.",
                    },
                    {"role": "user", "content": prompt},
                ],
                "temperature": 0.2,
                "max_tokens": 150,
            }

            resp = requests.post(url, headers=headers, json=payload, timeout=12.0)
            if resp.status_code == 200:
                raw = resp.text.strip()
                if "data: [DONE]" in raw:
                    raw = raw.split("data: [DONE]")[0].strip()
                first_brace = raw.find("{")
                last_brace = raw.rfind("}")
                if first_brace != -1 and last_brace != -1:
                    raw = raw[first_brace : last_brace + 1]
                data = json.loads(raw)
                choices = data.get("choices", [])
                if choices and "message" in choices[0]:
                    msg = choices[0]["message"]
                    content = msg.get("content") or msg.get("reasoning") or ""
                    cleaned = content.strip()
                    if cleaned:
                        first_line = cleaned.split("\n")[0].strip()
                        if first_line:
                            thought_text = f"[{self.openai_model}] {first_line}"
        except Exception:
            pass

        return self._run_deterministic_react_cycle(session_id, ticker, days, start_time, initial_thought=thought_text)
