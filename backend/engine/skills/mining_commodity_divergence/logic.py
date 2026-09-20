"""Deterministic execution logic for mining-commodity-divergence skill."""

from pathlib import Path
from typing import Any, Dict, List, Optional

from engine.quant.correlation import align_and_correlate_series
from engine.sectors.client import SectorsAPIClient
from engine.skills.base import BaseSkill, SkillResult


class MiningCommodityDivergenceSkill(BaseSkill):
    """Calculates Pearson correlation between IDX mining stocks and global commodity prices."""

    def __init__(self, skill_dir: Optional[Path] = None):
        super().__init__(skill_dir or Path(__file__).parent)

    def get_tool_definition(self) -> Dict[str, Any]:
        return {
            "name": "skill_mining_commodity_divergence",
            "description": "Measure Pearson correlation between IDX mining stocks and global commodity spot prices (Nickel, Coal, Gold).",
            "parameters": {
                "type": "object",
                "properties": {
                    "ticker": {"type": "string", "description": "IDX mining stock ticker (e.g. ANTM, PTBA, TINS)"},
                    "commodity": {"type": "string", "description": "Commodity name (e.g. nickel, coal, gold - optional)", "default": None},
                },
                "required": ["ticker"],
            },
        }

    def execute(self, arguments: Dict[str, Any], context: Dict[str, Any]) -> SkillResult:
        ticker = arguments.get("ticker", "").upper().strip()
        commodity = arguments.get("commodity")

        client: Optional[SectorsAPIClient] = context.get("sectors_client")
        if not client:
            db_path = context.get("db_path", "~/.niskava/niskava.db")
            mock_mode = context.get("mock_mode", False)
            client = SectorsAPIClient(db_path=db_path, mock_mode=mock_mode)

        # 1. Determine commodity if not provided
        slug_map = {
            "ANTM": ("aneka-tambang", "NICKEL"),
            "INCO": ("vale-indonesia", "NICKEL"),
            "MBMA": ("merdeka-battery", "NICKEL"),
            "TINS": ("timah", "TIN"),
            "PTBA": ("bukit-asam", "COAL"),
            "ADRO": ("adaro-energy", "COAL"),
            "MDKA": ("merdeka-copper-gold", "GOLD"),
        }

        if not commodity:
            slug, default_comm = slug_map.get(ticker, (ticker.lower(), "NICKEL"))
            mining_info = client.get_mining_detail(slug)
            commodity = mining_info.get("commodity", default_comm)
        commodity = commodity.upper()

        # 2. Fetch candles and commodity prices
        candles = client.get_daily_candles(ticker)
        commodity_prices = client.get_commodity_price(commodity.lower())

        # 3. Deterministic Correlation & Return Gate
        r, stock_ret, comm_ret, divergence_class = align_and_correlate_series(candles, commodity_prices)

        evidence = [
            {
                "type": "COMMODITY_CORRELATION",
                "commodity": commodity,
                "pearson_r": r,
                "stock_30d_return_pct": stock_ret,
                "commodity_30d_return_pct": comm_ret,
                "divergence_class": divergence_class,
                "source": f"Sectors API /commodity-price/{commodity.lower()}/",
            }
        ]

        if divergence_class == "COMMODITY_DRIVEN":
            summary = (
                f"Movement in {ticker} (+{stock_ret}%) is COMMODITY_DRIVEN: High positive correlation (r = {r}) "
                f"with global {commodity} prices (+{comm_ret}%)."
            )
        else:
            summary = (
                f"Movement in {ticker} (+{stock_ret}%) reflects IDIOSYNCRATIC_COMPANY_ALPHA: Low correlation (r = {r}) "
                f"with global {commodity} prices (+{comm_ret}%), indicating company-specific catalysts."
            )

        return SkillResult(
            skill_id=self.skill_id,
            verification_status="SUPPORTED",
            confidence_score=0.85,
            metrics={
                "ticker": ticker,
                "commodity": commodity,
                "pearson_r": r,
                "stock_30d_return_pct": stock_ret,
                "commodity_30d_return_pct": comm_ret,
                "divergence_class": divergence_class,
            },
            evidence=evidence,
            summary=summary,
        )
