"""Deterministic Quantitative MCP Tool Definitions and Dispatcher.

Complies strictly with:
- Law 1: Deterministic Before Generative (NumPy calculations, zero LLM math)
"""

from typing import Any, Dict, List
from engine.mcp.tools.sectors import sanitize_ticker
from engine.quant.anomaly import detect_historical_anomalies
from engine.sectors.client import SectorsAPIClient


def get_quant_tool_definitions() -> List[Dict[str, Any]]:
    """Return standardized MCP schemas for quantitative analysis tools."""
    return [
        {
            "name": "quant_compute_anomalies",
            "description": "Compute deterministic statistical trading volume anomalies (MA20 Volume Z-Score) and abnormal price returns via NumPy (Deterministic Compute Gate).",
            "inputSchema": {
                "type": "object",
                "properties": {
                    "ticker": {
                        "type": "string",
                        "description": "4-letter IDX stock ticker symbol (e.g. ANTM, BBRI)",
                    },
                    "volume_z_threshold": {
                        "type": "number",
                        "description": "Anomaly threshold for volume Z-score (default: 2.5)",
                        "default": 2.5,
                    },
                    "return_threshold_pct": {
                        "type": "number",
                        "description": "Anomaly threshold for abnormal price return percentage (default: 5.0)",
                        "default": 5.0,
                    },
                },
                "required": ["ticker"],
            },
        },
    ]


def execute_quant_tool(
    client: SectorsAPIClient, name: str, arguments: Dict[str, Any]
) -> Any:
    """Execute a quantitative analysis tool call."""
    if name == "quant_compute_anomalies":
        ticker = sanitize_ticker(arguments.get("ticker", ""))

        try:
            volume_z = float(arguments.get("volume_z_threshold", 2.5))
        except (ValueError, TypeError):
            volume_z = 2.5

        try:
            return_pct = float(arguments.get("return_threshold_pct", 5.0))
        except (ValueError, TypeError):
            return_pct = 5.0

        candles = client.get_daily_candles(ticker)
        if len(candles) < 20:
            return {
                "status": "INSUFFICIENT_DATA",
                "message": (
                    f"Emiten {ticker} hanya memiliki {len(candles)} lilin perdagangan historis. "
                    "Minimal 20 hari bursa diperlukan untuk menghitung moving average MA20 dan "
                    "Z-score volume deterministik."
                ),
                "anomalies": [],
            }

        anomalies = detect_historical_anomalies(
            daily_candles=candles,
            volume_z_threshold=volume_z,
            return_threshold_pct=return_pct,
        )
        return [a.to_dict() for a in anomalies]

    raise ValueError(f"Unknown Quant tool: {name}")
