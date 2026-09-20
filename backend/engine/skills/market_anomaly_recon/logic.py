"""Deterministic execution logic for market-anomaly-recon skill."""

from pathlib import Path
from typing import Any, Dict, List, Optional
import numpy as np

from engine.quant.anomaly import compute_foreign_flow_z_score, detect_historical_anomalies
from engine.sectors.client import SectorsAPIClient
from engine.skills.base import BaseSkill, SkillResult


class MarketAnomalyReconSkill(BaseSkill):
    """Executes quantitative anomaly reconnaissance across volume, price, and foreign flow."""

    def __init__(self, skill_dir: Optional[Path] = None):
        super().__init__(skill_dir or Path(__file__).parent)

    def get_tool_definition(self) -> Dict[str, Any]:
        return {
            "name": "skill_market_anomaly_recon",
            "description": "Execute quantitative trading anomaly reconnaissance (MA20 volume Z-score, abnormal price returns, foreign flow spikes).",
            "parameters": {
                "type": "object",
                "properties": {
                    "ticker": {"type": "string", "description": "IDX stock ticker (e.g. ANTM, BBCA)"},
                    "days": {"type": "integer", "description": "Observation window in trading days (default: 30)", "default": 30},
                    "volume_z_threshold": {"type": "number", "description": "Volume surge Z-score threshold (default: 2.5)", "default": 2.5},
                },
                "required": ["ticker"],
            },
        }

    def execute(self, arguments: Dict[str, Any], context: Dict[str, Any]) -> SkillResult:
        ticker = arguments.get("ticker", "").upper().strip()
        days = int(arguments.get("days", 30))
        volume_z_threshold = float(arguments.get("volume_z_threshold", 2.5))

        client: Optional[SectorsAPIClient] = context.get("sectors_client")
        if not client:
            db_path = context.get("db_path", "~/.niskava/niskava.db")
            mock_mode = context.get("mock_mode", False)
            client = SectorsAPIClient(db_path=db_path, mock_mode=mock_mode)

        # 1. Fetch candles
        candles = client.get_daily_candles(ticker)

        # 2. Deterministic Anomaly Gate
        anomalies = detect_historical_anomalies(
            daily_candles=candles,
            volume_z_threshold=volume_z_threshold,
            return_threshold_pct=5.0,
        )

        # 3. Fetch Foreign Flow & Compute Fz
        foreign_flows = client.get_foreign_flow(ticker)
        f_z = 0.0
        latest_flow = 0.0
        if len(foreign_flows) >= 2:
            flow_values = np.array([float(f.get("net_foreign_buy", 0.0)) for f in foreign_flows[:-1]])
            latest_flow = float(foreign_flows[-1].get("net_foreign_buy", 0.0))
            _, _, f_z = compute_foreign_flow_z_score(flow_values, latest_flow)

        anomaly_detected = len(anomalies) > 0 or abs(f_z) >= 2.5
        top_anomaly = anomalies[-1] if anomalies else None

        metrics = {
            "ticker": ticker,
            "candles_analyzed": len(candles),
            "anomaly_detected": anomaly_detected,
            "total_anomalies": len(anomalies),
            "latest_foreign_flow_net": latest_flow,
            "foreign_flow_z_score": round(f_z, 2),
        }

        evidence = []
        if top_anomaly:
            metrics.update({
                "latest_anomaly_date": top_anomaly.date,
                "latest_volume_z_score": round(top_anomaly.z_score, 2),
                "latest_price_change_pct": round(top_anomaly.price_change_pct, 2),
                "anomaly_classification": top_anomaly.classification,
            })
            evidence.append({
                "type": "QUANT_ANOMALY",
                "date": top_anomaly.date,
                "description": top_anomaly.description,
                "z_score": round(top_anomaly.z_score, 2),
                "price_change_pct": round(top_anomaly.price_change_pct, 2),
                "source": "Sectors Financial API v2 /daily/",
            })

        if abs(f_z) >= 2.5:
            evidence.append({
                "type": "FOREIGN_FLOW_ANOMALY",
                "foreign_flow_z_score": round(f_z, 2),
                "latest_net_idr": latest_flow,
                "source": "Sectors Financial API v2 /foreign-flow/",
            })

        summary = (
            f"Quantitative anomaly detected for {ticker}: {len(anomalies)} volume/price anomaly event(s) "
            f"identified over {len(candles)} trading sessions. Foreign flow Z-score: {round(f_z, 2)}."
            if anomaly_detected
            else f"No statistical volume or price anomalies detected for {ticker} across {len(candles)} trading sessions."
        )

        return SkillResult(
            skill_id=self.skill_id,
            verification_status="SUPPORTED" if anomaly_detected else "SUPPORTED",
            confidence_score=1.00,
            metrics=metrics,
            evidence=evidence,
            summary=summary,
        )
