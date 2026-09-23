"""Conversational Dialogue Triplestore Extractor for Niskava Agent.

Extracts user holdings, watchlist requests, and trading exits from unstructured
chat turns and maps them into associative memory graph triples.
"""

import re
from typing import Any, Dict, List
from engine.sectors.tickers import is_valid_idx_ticker


def extract_dialogue_observations(user_text: str) -> List[Dict[str, Any]]:
    """Extract structured portfolio and watchlist triples from free-form user messages."""
    observations: List[Dict[str, Any]] = []
    text = user_text.strip()
    if not text:
        return observations

    # 1. Exit / Take Profit / Cut Loss: e.g. "take profit ANTM di 1650", "tp BBCA 10000"
    exit_patterns = [
        r"(?:sudah\s+)?(?:take\s*profit|tp|jual|cut\s*loss|cl|exit)\s+([A-Za-z]{4})\s+(?:di|pada|level|harga)?\s*(\d+)",
        r"(?:posisi\s+)?([A-Za-z]{4})\s+(?:sudah\s+)?(?:di-?tp|di-?take\s*profit|di-?jual|di-?cl)\s+(?:di|pada|level|harga)?\s*(\d+)",
    ]
    for pat in exit_patterns:
        for match in re.finditer(pat, text, re.IGNORECASE):
            ticker = match.group(1).upper()
            price = match.group(2)
            if is_valid_idx_ticker(ticker):
                observations.append({
                    "ticker": ticker,
                    "source_label": "User",
                    "source_type": "USER",
                    "relation": "EXITED_AT",
                    "target_label": f"Price: {price}",
                    "target_type": "PRICE_LEVEL",
                    "context_snippet": f"Realized exit for {ticker} at price level {price}",
                })

    # 2. Buy / Entry: e.g. "beli BBCA di 9800", "entry ANTM @ 1450"
    buy_patterns = [
        r"(?:baru\s+)?(?:beli|buy|entry|masuk|koleksi|hold|posisi|punya)\s+([A-Za-z]{4})\s+(?:di|pada|level|harga|@)?\s*(\d+)",
        r"(?:harga\s+modal|entry\s+price)\s+([A-Za-z]{4})\s+(?:di|pada|level|harga|adalah)?\s*(\d+)",
    ]
    for pat in buy_patterns:
        for match in re.finditer(pat, text, re.IGNORECASE):
            ticker = match.group(1).upper()
            price = match.group(2)
            if is_valid_idx_ticker(ticker):
                if not any(o["ticker"] == ticker and o["relation"] == "EXITED_AT" for o in observations):
                    observations.append({
                        "ticker": ticker,
                        "source_label": "User",
                        "source_type": "USER",
                        "relation": "HOLDS_AT",
                        "target_label": f"Price: {price}",
                        "target_type": "PRICE_LEVEL",
                        "context_snippet": f"Entry position {ticker} at cost basis {price}",
                    })

    # 3. Watchlist / Monitor: e.g. "pantau pergerakan saham ANTM", "tolong pantau BBRI"
    watch_patterns = [
        r"(?:tolong\s+)?(?:pantau|awasi|monitor|watchlist)\s+(?:saham|pergerakan\s+saham)?\s*([A-Za-z]{4})",
        r"(?:masukkan|tambahkan)\s+([A-Za-z]{4})\s+(?:ke\s+)?(?:watchlist|pantauan)",
    ]
    for pat in watch_patterns:
        for match in re.finditer(pat, text, re.IGNORECASE):
            ticker = match.group(1).upper()
            if is_valid_idx_ticker(ticker):
                if not any(o["ticker"] == ticker for o in observations):
                    observations.append({
                        "ticker": ticker,
                        "source_label": "User",
                        "source_type": "USER",
                        "relation": "WATCHES",
                        "target_label": ticker,
                        "target_type": "TICKER",
                        "context_snippet": f"User monitoring stock movement: {ticker}",
                    })

    return observations
