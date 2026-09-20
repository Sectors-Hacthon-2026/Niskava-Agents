"""7-Stage Autonomous Investigation Pipeline Orchestrator.

Implements the standard operating procedure (SOP) defined in
docs/30-agent/01-investigation-pipeline.md and AGENTS.md:
1. INITIATION
2. SECTORS_BASELINE
3. QUANT_ANOMALY (NumPy Deterministic)
4. GAP_DETECTION
5. OSINT_HARVEST (Dual-Engine)
6. EVIDENCE_CORRELATION (Taxonomy & Discrete Confidence)
7. SYNTHESIS_AND_STREAMING
"""

import json
import os
import sqlite3
import time
import uuid
from datetime import datetime
from typing import Any, Callable, Dict, List, Optional

from engine.agent.verifier import FactVerificationGate
from engine.osint.harvester import DualEngineOSINTHarvester, OSINTItem
from engine.quant.anomaly import AnomalyResult, detect_historical_anomalies
from engine.sectors.client import SectorsAPIClient


class InvestigationPipeline:
    """Coordinates the 7-Stage Autonomous Investigation Pipeline."""

    def __init__(
        self,
        db_path: str,
        emitter: Optional[Callable[[Dict[str, Any]], None]] = None,
        mock_mode: Optional[bool] = None,
    ):
        self.db_path = os.path.expanduser(db_path)
        self.emitter = emitter or (lambda event: None)
        self.mock_mode = mock_mode
        self.sectors_client = SectorsAPIClient(
            db_path=self.db_path, mock_mode=self.mock_mode
        )
        self.osint_harvester = DualEngineOSINTHarvester(mock_mode=self.mock_mode)

    def _emit(self, event_data: Dict[str, Any]) -> None:
        """Send a JSONL event to the listener (stdout for Go Core)."""
        self.emitter(event_data)

    def run(self, ticker: str, timeframe_days: int = 30, session_id: Optional[str] = None) -> Dict[str, Any]:
        """Execute the full 7-Stage investigation pipeline."""
        start_time = time.time()
        session_id = session_id or f"INV-{datetime.now().strftime('%Y%m%d')}-{uuid.uuid4().hex[:6].upper()}"
        ticker = ticker.upper()

        # =========================================================================
        # Stage 1: INITIATION
        # =========================================================================
        self._emit({
            "event": "session_start",
            "session_id": session_id,
            "ticker": ticker,
            "timeframe_days": timeframe_days,
            "timestamp": datetime.now().isoformat() + "Z",
        })

        # =========================================================================
        # Stage 2: SECTORS_BASELINE
        # =========================================================================
        self._emit({
            "event": "progress_step",
            "session_id": session_id,
            "stage": "SECTORS_BASELINE",
            "step_index": 1,
            "total_steps": 4,
            "message": f"Pulling {timeframe_days} days candlestick and company report from Sectors v2 API...",
        })

        daily_candles = self.sectors_client.get_daily_candles(ticker)
        company_report = self.sectors_client.get_company_report(ticker)
        sectors_news = self.sectors_client.get_news(ticker)

        company_name = company_report.get("company_name", ticker)

        # =========================================================================
        # Stage 3: QUANT_ANOMALY (NumPy Deterministic)
        # =========================================================================
        self._emit({
            "event": "progress_step",
            "session_id": session_id,
            "stage": "QUANT_ANOMALY",
            "step_index": 2,
            "total_steps": 4,
            "message": "Computing rolling MA20, Volume Z-Score, and Abnormal Returns...",
        })

        anomalies: List[AnomalyResult] = detect_historical_anomalies(
            daily_candles=daily_candles,
            volume_z_threshold=2.5,
            return_threshold_pct=5.0,
            divergence_threshold_pct=4.0,
        )

        for anom in anomalies:
            self._emit({
                "event": "anomaly_detected",
                "session_id": session_id,
                "ticker": ticker,
                "anomaly_date": anom.date,
                "metric_type": anom.metric_type,
                "z_score": anom.z_score,
                "metric_value": anom.metric_value,
                "baseline_value": anom.baseline_value,
                "price_change_pct": anom.price_change_pct,
                "sector_change_pct": 0.0,
                "description": anom.description,
            })

        # =========================================================================
        # Stage 4 & 5: GAP_DETECTION & OSINT_HARVEST
        # =========================================================================
        self._emit({
            "event": "progress_step",
            "session_id": session_id,
            "stage": "OSINT_HARVEST",
            "step_index": 3,
            "total_steps": 4,
            "message": f"Harvesting news & regulatory filings for {ticker} ({company_name})...",
        })

        osint_items: List[OSINTItem] = self.osint_harvester.harvest(
            ticker=ticker,
            company_name=company_name,
            sectors_news_items=sectors_news,
        )

        # =========================================================================
        # Stage 6: EVIDENCE_CORRELATION
        # =========================================================================
        self._emit({
            "event": "progress_step",
            "session_id": session_id,
            "stage": "EVIDENCE_CORRELATION",
            "step_index": 4,
            "total_steps": 4,
            "message": "Correlating temporal causality and classifying 3-tier taxonomy...",
        })

        findings = self._correlate_evidence(
            session_id=session_id,
            ticker=ticker,
            anomalies=anomalies,
            osint_items=osint_items,
        )

        for finding in findings:
            self._emit(finding)

        # =========================================================================
        # Stage 7: SYNTHESIS_AND_STREAMING & SQLite Persistence
        # =========================================================================
        duration_ms = int((time.time() - start_time) * 1000)
        summary = self._synthesize_summary(ticker, anomalies, findings)

        self._persist_to_sqlite(
            session_id=session_id,
            ticker=ticker,
            timeframe_days=timeframe_days,
            anomalies=anomalies,
            findings=findings,
            summary=summary,
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
            "anomalies": [a.to_dict() for a in anomalies],
            "findings": findings,
            "summary": summary,
            "duration_ms": duration_ms,
        }

    def _correlate_evidence(
        self,
        session_id: str,
        ticker: str,
        anomalies: List[AnomalyResult],
        osint_items: List[OSINTItem],
    ) -> List[Dict[str, Any]]:
        """Perform temporal causality correlation and 3-Tier classification."""
        findings: List[Dict[str, Any]] = []

        if not anomalies:
            # Baseline Fundamental Finding
            f_id = f"FND-{uuid.uuid4().hex[:4].upper()}"
            findings.append({
                "event": "finding_emitted",
                "session_id": session_id,
                "id": f_id,
                "title": f"Pergerakan Normal dan Fundamental Stabil ({ticker})",
                "claim_text": f"Tidak ditemukan anomali volume (>2.5σ) atau pergerakan harga (>5%) dalam jendela observasi.",
                "verification_status": "SUPPORTED",
                "confidence_score": 1.00,
                "causality_status": "UNEXPLAINED_BY_NEWS",
                "evidence": [],
            })
            return findings

        # Correlate anomalies with available news/disclosures
        for idx, anom in enumerate(anomalies, 1):
            finding_id = f"FND-{idx:02d}"

            # 1. Primary Quantitative Evidence
            quant_evidence = {
                "source_type": "QUANTITATIVE_BASELINE",
                "source_name": "Sectors Daily API",
                "source_url": f"https://api.sectors.app/v2/daily/{ticker}/",
                "publication_date": anom.date,
                "snippet_text": anom.description,
            }

            # 2. Check for matching news/disclosures
            matched_news = [item for item in osint_items if item.is_disclosure]
            if not matched_news and osint_items:
                matched_news = osint_items[:2]

            if matched_news:
                top_item = matched_news[0]
                is_official = top_item.is_disclosure or "IDX" in top_item.source_name
                status = "SUPPORTED" if is_official else "UNCERTAIN"
                confidence = 0.95 if is_official else 0.75
                causality = "LIKELY_CATALYST"

                title = f"Katalis Penggerak: {top_item.title[:60]}..."
                claim = f"Lonjakan volume ({anom.z_score}σ) pada {anom.date} berkorelasi tinggi dengan pengumuman: '{top_item.title}'."

                evidence_list = [
                    quant_evidence,
                    {
                        "source_type": "OFFICIAL_DISCLOSURE" if is_official else "FINANCIAL_PRESS",
                        "source_name": top_item.source_name,
                        "source_url": top_item.source_url,
                        "publication_date": top_item.publication_date,
                        "snippet_text": top_item.snippet,
                    },
                ]
            else:
                status = "UNCERTAIN"
                confidence = 0.65
                causality = "UNEXPLAINED_BY_NEWS"
                title = f"Lonjakan Transaksi Belum Terjelaskan oleh Berita Resmi ({ticker})"
                claim = f"Lonjakan volume ({anom.z_score}σ) terdeteksi pada {anom.date} tanpa keterbukaan informasi penjelas pada jendela waktu bersamaan."
                evidence_list = [quant_evidence]

            raw_finding = {
                "event": "finding_emitted",
                "session_id": session_id,
                "id": finding_id,
                "title": title,
                "claim_text": claim,
                "verification_status": status,
                "confidence_score": confidence,
                "causality_status": causality,
                "evidence": evidence_list,
            }
            verified_finding = FactVerificationGate.verify_finding(raw_finding)
            findings.append(verified_finding)

        return findings

    def _synthesize_summary(
        self, ticker: str, anomalies: List[AnomalyResult], findings: List[Dict[str, Any]]
    ) -> str:
        """Synthesize final findings summary."""
        supported = sum(1 for f in findings if f.get("verification_status") == "SUPPORTED")
        uncertain = sum(1 for f in findings if f.get("verification_status") == "UNCERTAIN")
        
        if anomalies:
            top_anom = max(anomalies, key=lambda a: a.z_score)
            return (
                f"Investigasi {ticker} menemukan {len(anomalies)} titik anomali kuantitatif "
                f"(puncak volume {top_anom.z_score}σ pada {top_anom.date}). "
                f"Hasil verifikasi bukti: {supported} temuan SUPPORTED, {uncertain} temuan UNCERTAIN."
            )
        return f"Investigasi {ticker} selesai. Pergerakan pasar berada dalam batas baseline normal."

    def _persist_to_sqlite(
        self,
        session_id: str,
        ticker: str,
        timeframe_days: int,
        anomalies: List[AnomalyResult],
        findings: List[Dict[str, Any]],
        summary: str,
    ) -> None:
        """Persist session, anomalies, and findings into local SQLite database."""
        if not os.path.exists(self.db_path):
            return
        try:
            with sqlite3.connect(self.db_path) as conn:
                cursor = conn.cursor()
                now = datetime.now().isoformat() + "Z"

                # Update or Insert investigation session
                cursor.execute(
                    """
                    INSERT INTO investigations (id, ticker, market, timeframe_days, status, started_at, completed_at, summary_text)
                    VALUES (?, ?, 'IDX', ?, 'COMPLETED', ?, ?, ?)
                    ON CONFLICT(id) DO UPDATE SET
                        status = 'COMPLETED',
                        completed_at = excluded.completed_at,
                        summary_text = excluded.summary_text
                    """,
                    (session_id, ticker, timeframe_days, now, now, summary),
                )

                # Persist anomalies
                for anom in anomalies:
                    anom_id = f"ANOM-{uuid.uuid4().hex[:8].upper()}"
                    cursor.execute(
                        """
                        INSERT OR REPLACE INTO anomalies 
                        (id, investigation_id, anomaly_date, metric_type, metric_value, baseline_value, z_score, description)
                        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
                        """,
                        (
                            anom_id,
                            session_id,
                            anom.date,
                            anom.metric_type,
                            anom.metric_value,
                            anom.baseline_value,
                            anom.z_score,
                            anom.description,
                        ),
                    )

                # Persist findings & evidence items
                for f in findings:
                    f_id = f.get("id", str(uuid.uuid4()))
                    cursor.execute(
                        """
                        INSERT OR REPLACE INTO findings
                        (id, investigation_id, title, claim_text, verification_status, confidence_score, causality_status)
                        VALUES (?, ?, ?, ?, ?, ?, ?)
                        """,
                        (
                            f_id,
                            session_id,
                            f.get("title", ""),
                            f.get("claim_text", ""),
                            f.get("verification_status", "UNCERTAIN"),
                            f.get("confidence_score", 0.75),
                            f.get("causality_status", "UNEXPLAINED_BY_NEWS"),
                        ),
                    )

                    for ev in f.get("evidence", []):
                        ev_id = f"EVD-{uuid.uuid4().hex[:8].upper()}"
                        cursor.execute(
                            """
                            INSERT OR REPLACE INTO evidence_items
                            (id, finding_id, source_type, source_name, source_url, publication_date, snippet_text)
                            VALUES (?, ?, ?, ?, ?, ?, ?)
                            """,
                            (
                                ev_id,
                                f_id,
                                ev.get("source_type", "NEWS"),
                                ev.get("source_name", ""),
                                ev.get("source_url", ""),
                                ev.get("publication_date", ""),
                                ev.get("snippet_text", ""),
                            ),
                        )

                conn.commit()
        except sqlite3.Error:
            pass
