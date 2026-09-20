"""Deterministic execution logic for event-causality-audit skill."""

from datetime import datetime, timedelta
from pathlib import Path
from typing import Any, Dict, List, Optional

from engine.osint.harvester import DualEngineOSINTHarvester
from engine.sectors.client import SectorsAPIClient
from engine.skills.base import BaseSkill, SkillResult


class EventCausalityAuditSkill(BaseSkill):
    """Correlates quantitative anomaly timestamps with disclosures, corporate actions, and news."""

    def __init__(self, skill_dir: Optional[Path] = None):
        super().__init__(skill_dir or Path(__file__).parent)

    def get_tool_definition(self) -> Dict[str, Any]:
        return {
            "name": "skill_event_causality_audit",
            "description": "Audit temporal causality between price/volume spikes and external events (IDX disclosures, suspensions, news).",
            "parameters": {
                "type": "object",
                "properties": {
                    "ticker": {"type": "string", "description": "IDX stock ticker (e.g. ANTM)"},
                    "anomaly_date": {"type": "string", "description": "ISO date of the volume/price anomaly (YYYY-MM-DD)"},
                    "window_days_before": {"type": "integer", "description": "Days before anomaly to inspect (default: 2)", "default": 2},
                    "window_days_after": {"type": "integer", "description": "Days after anomaly to inspect (default: 1)", "default": 1},
                },
                "required": ["ticker"],
            },
        }

    def execute(self, arguments: Dict[str, Any], context: Dict[str, Any]) -> SkillResult:
        ticker = arguments.get("ticker", "").upper().strip()
        anomaly_date = arguments.get("anomaly_date")
        days_before = int(arguments.get("window_days_before", 2))
        days_after = int(arguments.get("window_days_after", 1))

        client: Optional[SectorsAPIClient] = context.get("sectors_client")
        if not client:
            db_path = context.get("db_path", "~/.niskava/niskava.db")
            mock_mode = context.get("mock_mode", False)
            client = SectorsAPIClient(db_path=db_path, mock_mode=mock_mode)

        harvester: Optional[DualEngineOSINTHarvester] = context.get("osint_harvester")
        if not harvester:
            mock_mode = context.get("mock_mode", False)
            harvester = DualEngineOSINTHarvester(mock_mode=mock_mode)

        # 1. Fetch Suspensions & UMA notices
        suspensions = client.get_suspensions(ticker)

        # 2. Fetch Corporate Actions
        corp_actions = client.get_corporate_actions(ticker)

        # 3. Harvest News & Filings
        report = client.get_company_report(ticker)
        company_name = report.get("company_name", ticker)
        sectors_news = client.get_news(ticker)
        osint_items = harvester.harvest(ticker=ticker, company_name=company_name, sectors_news_items=sectors_news)

        # 4. Temporal Precedence Evaluation
        evidence = []
        causality_label = "UNEXPLAINED_BY_NEWS"
        confidence_score = 0.65

        # Check official suspensions first (Tier 1)
        for s in suspensions:
            s_date = s.get("suspension_date", "")
            evidence.append({
                "type": "EXCHANGE_SUSPENSION_OR_UMA",
                "title": f"IDX Notice: {s.get('reason', 'Trading Suspension')}",
                "date": s_date,
                "url": s.get("pdf_url", ""),
                "source": "Indonesia Stock Exchange (IDXnet)",
                "tier": "Tier 1",
            })
            if anomaly_date and s_date:
                if s_date >= anomaly_date:
                    causality_label = "TRIGGERED_REGULATORY_INQUIRY"
                    confidence_score = 0.95

        # Check corporate actions
        for ca in corp_actions:
            evidence.append({
                "type": "CORPORATE_ACTION",
                "title": f"Corporate Action: {ca.get('action_type', 'ACTION')} (Cum-Date: {ca.get('cum_date', 'N/A')})",
                "date": ca.get("cum_date", ""),
                "source": "IDX Corporate Actions Feed",
                "tier": "Tier 1",
            })
            if anomaly_date and ca.get("cum_date") == anomaly_date:
                causality_label = "LIKELY_CATALYST"
                confidence_score = 1.00

        # Check OSINT News items
        for item in osint_items:
            evidence.append({
                "type": "FINANCIAL_MEDIA_NEWS",
                "title": item.title,
                "date": item.publication_date,
                "url": item.source_url,
                "source": item.source_name,
                "tier": "Tier 2",
            })
            if causality_label == "UNEXPLAINED_BY_NEWS":
                causality_label = "LIKELY_CATALYST"
                confidence_score = 0.85

        summary = (
            f"Event causality audit for {ticker}: Status classified as '{causality_label}' "
            f"backed by {len(evidence)} regulatory and financial media evidence items."
        )

        verification_status = "SUPPORTED" if evidence else "UNCERTAIN"

        return SkillResult(
            skill_id=self.skill_id,
            verification_status=verification_status,
            confidence_score=confidence_score,
            metrics={
                "ticker": ticker,
                "anomaly_date": anomaly_date or "RECENT",
                "causality_label": causality_label,
                "total_evidence_items": len(evidence),
                "suspension_count": len(suspensions),
                "corporate_action_count": len(corp_actions),
            },
            evidence=evidence,
            summary=summary,
        )
