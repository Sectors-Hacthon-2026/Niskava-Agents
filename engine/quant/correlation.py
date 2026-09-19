"""Deterministic time-series Pearson correlation calculation using pure NumPy.

Abides by Law 1 (Deterministic Before Generative):
Aligns daily stock price returns with global commodity benchmark price deltas,
computing correlation coefficients deterministically before LLM evaluation.
"""

from typing import Any, Dict, List, Tuple
import numpy as np


def compute_pearson_correlation(
    series_x: List[float],
    series_y: List[float],
) -> float:
    """Compute Pearson correlation coefficient r between two numeric series."""
    if len(series_x) != len(series_y) or len(series_x) < 3:
        return 0.0

    arr_x = np.array(series_x, dtype=np.float64)
    arr_y = np.array(series_y, dtype=np.float64)

    std_x = np.std(arr_x)
    std_y = np.std(arr_y)

    if std_x == 0.0 or std_y == 0.0:
        return 0.0

    corr_matrix = np.corrcoef(arr_x, arr_y)
    r = float(corr_matrix[0, 1])
    if np.isnan(r):
        return 0.0
    return round(float(np.clip(r, -1.0, 1.0)), 4)


def align_and_correlate_series(
    stock_candles: List[Dict[str, Any]],
    commodity_prices: List[Dict[str, Any]],
) -> Tuple[float, float, float, str]:
    """Align stock daily returns with commodity price percentage changes.

    Args:
        stock_candles: List of candlestick dicts with 'date' and 'close'.
        commodity_prices: List of dicts with 'date' and 'price'.

    Returns:
        tuple[pearson_r, stock_return_pct, commodity_return_pct, divergence_class]
    """
    if len(stock_candles) < 3 or len(commodity_prices) < 3:
        return 0.0, 0.0, 0.0, "INSUFFICIENT_DATA"

    # Map commodity date -> price
    comm_map = {item["date"]: float(item["price"]) for item in commodity_prices if "date" in item and "price" in item}

    aligned_stock_returns: List[float] = []
    aligned_comm_returns: List[float] = []

    for i in range(1, len(stock_candles)):
        curr_d = stock_candles[i].get("date")
        prev_d = stock_candles[i - 1].get("date")

        curr_close = float(stock_candles[i].get("close", 0.0))
        prev_close = float(stock_candles[i - 1].get("close", 0.0))

        if prev_close > 0 and curr_d in comm_map and prev_d in comm_map:
            stock_ret = (curr_close - prev_close) / prev_close
            comm_ret = (comm_map[curr_d] - comm_map[prev_d]) / comm_map[prev_d] if comm_map[prev_d] > 0 else 0.0
            aligned_stock_returns.append(stock_ret)
            aligned_comm_returns.append(comm_ret)

    # Fallback to index-based alignment if exact dates don't match (e.g. trading days vs calendar days)
    if len(aligned_stock_returns) < 3:
        min_len = min(len(stock_candles), len(commodity_prices))
        aligned_stock_returns = []
        aligned_comm_returns = []
        for i in range(1, min_len):
            p1 = float(stock_candles[i]["close"])
            p0 = float(stock_candles[i - 1]["close"])
            c1 = float(commodity_prices[i]["price"])
            c0 = float(commodity_prices[i - 1]["price"])
            if p0 > 0 and c0 > 0:
                aligned_stock_returns.append((p1 - p0) / p0)
                aligned_comm_returns.append((c1 - c0) / c0)

    r = compute_pearson_correlation(aligned_stock_returns, aligned_comm_returns)

    # Calculate overall cumulative return over window
    first_stock = float(stock_candles[0].get("close", 1.0))
    last_stock = float(stock_candles[-1].get("close", 1.0))
    stock_cum_ret = round(((last_stock - first_stock) / first_stock) * 100.0, 2) if first_stock > 0 else 0.0

    first_comm = float(commodity_prices[0].get("price", 1.0))
    last_comm = float(commodity_prices[-1].get("price", 1.0))
    comm_cum_ret = round(((last_comm - first_comm) / first_comm) * 100.0, 2) if first_comm > 0 else 0.0

    if r >= 0.60:
        divergence_class = "COMMODITY_DRIVEN"
    elif r <= -0.40:
        divergence_class = "INVERSE_DIVERGENCE"
    else:
        divergence_class = "IDIOSYNCRATIC_COMPANY_ALPHA"

    return r, stock_cum_ret, comm_cum_ret, divergence_class
