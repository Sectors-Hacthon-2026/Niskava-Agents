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
                "Generate and export an institutional PDF audit trail report for an IDX stock investigation. "
                "ONLY invoke when the user explicitly requests a PDF, downloadable report, or printed document. "
                "Do NOT invoke for standard analysis questions. Requires ticker and summary at minimum."
            ),
            "parameters": {
                "type": "object",
                "properties": {
                    "ticker": {
                        "type": "string",
                        "description": "IDX stock ticker (e.g. ANTM, BBCA). Case-insensitive.",
                    },
                    "title": {
                        "type": "string",
                        "description": "Report title (e.g. 'ANTM Investigation Audit Trail — 30 Days').",
                    },
                    "summary": {
                        "type": "string",
                        "description": "Investigative narrative synthesized from prior tool observations. Plain text.",
                    },
                    "metrics": {
                        "type": "object",
                        "description": (
                            "Pre-computed quantitative metrics dict. Keys: volume_z_score (float), "
                            "foreign_flow_z_score (float), abnormal_return_pct (float), "
                            "candles_analyzed (int), anomaly_detected (bool), latest_anomaly_date (str)."
                        ),
                    },
                    "evidence": {
                        "type": "array",
                        "description": (
                            "List of evidence items. Each item: date (str), headline (str), "
                            "verification_status (SUPPORTED|UNCERTAIN|CONTRADICTED), "
                            "confidence_score (float), source (str)."
                        ),
                        "items": {"type": "object"},
                    },
                    "session_id": {
                        "type": "string",
                        "description": "Current investigation session ID for traceability.",
                    },
                },
                "required": ["ticker", "summary"],
            },
        }

    def execute(self, arguments: Dict[str, Any], context: Dict[str, Any]) -> SkillResult:
        """Execute PDF generation from pre-computed investigation data.

        Args:
            arguments: Must contain 'ticker' (str) and 'summary' (str).
                       Optional: 'title', 'metrics' (dict), 'evidence' (list), 'session_id'.
            context: Agent execution context. Reads '_output_dir' (Path) for test overrides.
                     May contain 'emitter' for IPC event notifications.
                     Does NOT read sectors_client — zero API calls by design (Law 5).

        Returns:
            SkillResult with metrics['pdf_path'] as the absolute path to the written PDF.
        """
        ticker = str(arguments.get("ticker", "")).upper().strip()
        if not ticker:
            return SkillResult(
                skill_id=self.skill_id,
                verification_status="CONTRADICTED",
                confidence_score=0.55,
                metrics={"error": "ticker is required"},
                summary="PDF generation failed: 'ticker' argument is missing.",
            )

        output_dir: Optional[Path] = context.get("_output_dir")
        session_id = (
            arguments.get("session_id")
            or context.get("session_id")
            or "000000"
        )

        params = {
            "ticker": ticker,
            "title": arguments.get("title", f"{ticker} Investigation Audit Trail"),
            "summary": arguments.get("summary", ""),
            "metrics": arguments.get("metrics") or {},
            "evidence": arguments.get("evidence") or [],
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
                f"Investigation audit trail PDF generated successfully for {ticker}. "
                f"File saved to: {pdf_path}"
            ),
        )
