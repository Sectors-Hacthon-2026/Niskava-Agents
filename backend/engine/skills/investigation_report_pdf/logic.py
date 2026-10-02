"""InvestigationReportPdfSkill — Layer 3 Domain Skill.

Generates an institutional PDF audit trail from pre-computed investigation data.
Invoked ONLY when the user explicitly requests a PDF export or downloadable report.

Compliance:
- Law 1: Deterministic before Generative (reads pre-computed metrics, zero LLM math).
- Law 2: Non-advisory disclaimer is hard-coded in the renderer on every page.
- Law 4: Output written to ~/.niskava/reports/ (local-first, never cloud).
- Law 5: Zero Sectors API calls. All data from arguments or local SQLite context.
"""

from __future__ import annotations

import os
from pathlib import Path
from typing import Any, Dict, Optional

from engine.skills.base import BaseSkill, SkillResult
from engine.skills.investigation_report_pdf.renderer import render_investigation_pdf


class InvestigationReportPdfSkill(BaseSkill):
    """Render and export an IDX investigation audit trail to an institutional PDF file.

    This skill is OPTIONAL and must only be invoked when the user explicitly requests
    a PDF, downloadable report, or printable audit document.
    """

    def __init__(self, skill_dir: Optional[Path] = None) -> None:
        super().__init__(skill_dir or Path(__file__).parent)

    def get_tool_definition(self) -> Dict[str, Any]:
        """Return JSON-schema compatible tool definition for ReAct LLM function calling."""
        return {
            "name": "skill_investigation_report_pdf",
            "description": (
                "Generate and export an institutional PDF research report or audit trail for IDX stocks or general market news. "
                "ONLY invoke when the user explicitly requests a PDF, downloadable report, or printed document. "
                "Supports single stock investigations, macro market news digests, and sector/custom research notes. "
                "Requires summary at minimum. If reporting on general market news without a single ticker, set ticker='MARKET'."
            ),
            "parameters": {
                "type": "object",
                "properties": {
                    "ticker": {
                        "type": "string",
                        "description": "IDX stock ticker (e.g. ANTM, BBCA) or 'MARKET' / 'IHSG' for general market overview.",
                    },
                    "title": {
                        "type": "string",
                        "description": "Report title (e.g. 'ANTM Investigation Audit Trail' or 'IDX Daily Market Intelligence Brief').",
                    },
                    "summary": {
                        "type": "string",
                        "description": "Synthesized executive summary or investigative narrative. Plain text.",
                    },
                    "report_type": {
                        "type": "string",
                        "enum": ["TICKER_INVESTIGATION", "MARKET_NEWS_BRIEF", "CUSTOM_RESEARCH"],
                        "description": "Report layout type. Auto-detected if omitted.",
                    },
                    "metrics": {
                        "type": "object",
                        "description": (
                            "Pre-computed quantitative metrics dict (for single-stock investigations). "
                            "Keys: volume_z_score, foreign_flow_z_score, abnormal_return_pct, candles_analyzed, anomaly_detected."
                        ),
                    },
                    "evidence": {
                        "type": "array",
                        "description": (
                            "List of causal evidence items. Each item: date (str), headline (str), "
                            "verification_status (SUPPORTED|UNCERTAIN|CONTRADICTED), confidence_score (float), source (str)."
                        ),
                        "items": {"type": "object"},
                    },
                    "news_items": {
                        "type": "array",
                        "description": (
                            "List of curated market news or disclosures (for market news digests). "
                            "Each item: date (str), headline (str), source (str), sentiment/status (str)."
                        ),
                        "items": {"type": "object"},
                    },
                    "sections": {
                        "type": "array",
                        "description": (
                            "List of custom analytical narrative sections. "
                            "Each item: heading (str), content (str)."
                        ),
                        "items": {"type": "object"},
                    },
                    "blocks": {
                        "type": "array",
                        "description": (
                            "Ordered list of dynamic blocks ('callout', 'markdown', 'table', 'key_value', 'quant_metrics', 'evidence_matrix', 'news_matrix')."
                        ),
                        "items": {"type": "object"},
                    },
                    "custom_tables": {
                        "type": "array",
                        "description": "List of custom data tables (headers, rows, title).",
                        "items": {"type": "object"},
                    },
                    "session_id": {
                        "type": "string",
                        "description": "Current investigation session ID for traceability.",
                    },
                },
                "required": ["summary"],
            },
        }

    def execute(self, arguments: Dict[str, Any], context: Dict[str, Any]) -> SkillResult:
        """Execute PDF generation from pre-computed investigation or market news data.

        Args:
            arguments: Contains 'summary' (or 'blocks') and optional 'ticker' (defaults to 'MARKET'),
                       'title', 'report_type', 'metrics', 'evidence', 'news_items', 'sections',
                       'blocks', 'custom_tables', 'session_id'.
            context: Agent execution context. Reads '_output_dir' (Path) for test overrides.
                     May contain 'emitter' for IPC event notifications.
                     Does NOT read sectors_client — zero API calls by design (Law 5).

        Returns:
            SkillResult with metrics['pdf_path'] as the absolute path to the written PDF.
        """
        raw_ticker = str(arguments.get("ticker", "")).strip()
        ticker = raw_ticker.upper() if raw_ticker else "MARKET"

        summary = arguments.get("summary", "")
        blocks = arguments.get("blocks")
        if not summary and not blocks:
            return SkillResult(
                skill_id=self.skill_id,
                verification_status="CONTRADICTED",
                confidence_score=0.55,
                metrics={"error": "summary or blocks is required"},
                summary="PDF generation failed: 'summary' or 'blocks' argument is missing.",
            )

        if not summary and blocks:
            summary = "Investigation report generated with dynamic blocks."

        output_dir_arg = arguments.get("output_dir") or arguments.get("_output_dir")
        output_dir: Optional[Path] = context.get("_output_dir") or (
            Path(str(output_dir_arg)) if output_dir_arg else None
        )
        session_id = (
            arguments.get("session_id")
            or context.get("session_id")
            or "000000"
        )

        params = {
            "ticker": ticker,
            "title": arguments.get("title"),
            "summary": summary,
            "report_type": arguments.get("report_type"),
            "metrics": arguments.get("metrics") or {},
            "evidence": arguments.get("evidence") or [],
            "news_items": arguments.get("news_items") or [],
            "sections": arguments.get("sections") or [],
            "blocks": blocks or [],
            "custom_tables": arguments.get("custom_tables") or [],
            "session_id": str(session_id),
        }

        pdf_path = render_investigation_pdf(params, output_dir=output_dir)
        filename = os.path.basename(pdf_path)

        # Emit live IPC event if emitter is present in context
        emitter = context.get("emitter")
        if callable(emitter):
            try:
                emitter({
                    "event": "pdf_report_ready",
                    "pdf_path": pdf_path,
                    "filename": filename,
                    "ticker": ticker,
                })
            except Exception:
                pass

        return SkillResult(
            skill_id=self.skill_id,
            verification_status="SUPPORTED",
            confidence_score=1.00,
            metrics={"pdf_path": pdf_path, "filename": filename, "ticker": ticker},
            summary=(
                f"Report PDF generated successfully for {ticker}. "
                f"File saved to: {pdf_path}"
            ),
        )
