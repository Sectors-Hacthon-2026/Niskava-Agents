"""MCP Prompt Templates for Niskava Agent.

Standardizes analytical SOP workflows into ready-to-run prompt templates
accessible via prompts/list and prompts/get.
"""

from typing import Any, Dict, List


def get_prompt_definitions() -> List[Dict[str, Any]]:
    """Return list of standard MCP prompts available on Niskava."""
    return [
        {
            "name": "investigate_ticker_anomaly",
            "description": "Jalankan investigasi komprehensif 7-tahap terhadap anomali volume/return emiten IDX.",
            "arguments": [
                {
                    "name": "ticker",
                    "description": "Kode saham IDX (contoh: ANTM)",
                    "required": True,
                },
                {
                    "name": "days",
                    "description": "Rentang hari observasi (default: 30)",
                    "required": False,
                },
            ],
        },
        {
            "name": "bandarmology_insider_audit",
            "description": "Audit forensik akumulasi/distribusi broker dan transaksi orang dalam (insider trading).",
            "arguments": [
                {
                    "name": "ticker",
                    "description": "Kode saham IDX (contoh: BBRI)",
                    "required": True,
                }
            ],
        },
        {
            "name": "financial_health_stress_test",
            "description": "Stress-test balance sheet solvency/liquidity dan sanggahan rumor kepailitan/gagal bayar.",
            "arguments": [
                {"name": "ticker", "description": "Kode saham IDX (contoh: GOTO, ANTM)", "required": True},
                {"name": "rumor_claim", "description": "Klaim rumor yang ingin diverifikasi (opsional)", "required": False},
            ],
        },
        {
            "name": "mining_commodity_divergence",
            "description": "Korelasi Pearson antara pergerakan saham tambang IDX dan harga spot komoditas acuan global.",
            "arguments": [
                {"name": "ticker", "description": "Kode saham tambang IDX (contoh: ANTM, PTBA)", "required": True},
                {"name": "commodity", "description": "Komoditas (contoh: nickel, coal, gold)", "required": False},
            ],
        },
        {
            "name": "peer_valuation_benchmark",
            "description": "Benchmarking valuasi relatif emiten (P/E, P/B) terhadap median dan IQR subsektor.",
            "arguments": [
                {"name": "ticker", "description": "Kode saham IDX (contoh: BBCA)", "required": True},
                {"name": "subsector", "description": "Slug subsektor (opsional)", "required": False},
            ],
        },
        {
            "name": "event_causality_audit",
            "description": "Audit kausalitas temporal berita dan keterbukaan BEI terhadap lonjakan volume saham.",
            "arguments": [
                {"name": "ticker", "description": "Kode saham IDX (contoh: ANTM)", "required": True},
                {"name": "anomaly_date", "description": "Tanggal anomali lonjakan volume (YYYY-MM-DD)", "required": True},
            ],
        },
    ]


def get_prompt_messages(name: str, arguments: Dict[str, Any]) -> Dict[str, Any]:
    """Render prompt messages for a given prompt name and arguments."""
    ticker = arguments.get("ticker", "").upper()
    days = arguments.get("days", "30")

    if name == "investigate_ticker_anomaly":
        return {
            "description": f"7-Stage Investigation Workflow for {ticker}",
            "messages": [
                {
                    "role": "user",
                    "content": {
                        "type": "text",
                        "text": (
                            f"Lakukan investigasi terstruktur terhadap saham {ticker} untuk rentang {days} hari terakhir.\n"
                            f"Langkah-langkah yang harus dilakukan:\n"
                            f"1. Panggil `sectors_get_daily_candles` untuk mendapatkan data candlestick.\n"
                            f"2. Panggil `quant_compute_anomalies` untuk mendeteksi lonjakan volume Z-score dan return abnormal.\n"
                            f"3. Jika ada anomali, panggil `news_harvest_market_news` untuk mencari pengumuman resmi atau berita katalis.\n"
                            f"4. Panggil `sectors_get_foreign_flow` dan `sectors_get_suspensions` untuk memverifikasi data pendukung.\n"
                            f"5. Susun sintesis investigasi dengan Taksonomi Tiga Tingkat (SUPPORTED / UNCERTAIN / CONTRADICTED) "
                            f"dan sertakan disclaimer non-advisori finansial."
                        ),
                    },
                }
            ],
        }

    if name == "bandarmology_insider_audit":
        return {
            "description": f"Bandarmology & Insider Audit for {ticker}",
            "messages": [
                {
                    "role": "user",
                    "content": {
                        "type": "text",
                        "text": (
                            f"Lakukan audit kepemilikan dan aliran modal pada saham {ticker}.\n"
                            f"1. Panggil `sectors_get_broker_summary` untuk melihat top broker buyer dan seller.\n"
                            f"2. Panggil `sectors_get_foreign_flow` untuk menganalisis akumulasi/distribusi asing.\n"
                            f"3. Panggil `sectors_get_filings` untuk memeriksa transaksi insider (direksi/komisaris).\n"
                            f"4. Evaluasi apakah pergerakan harga didorong oleh investor institusi atau ritel."
                        ),
                    },
                }
            ],
        }

    if name == "financial_health_stress_test":
        rumor = arguments.get("rumor_claim", "Isu kesulitan likuiditas / gagal bayar kupon")
        return {
            "description": f"Financial Health & Rumor Refutation for {ticker}",
            "messages": [
                {
                    "role": "user",
                    "content": {
                        "type": "text",
                        "text": (
                            f"Lakukan stress test kesehatan neraca pada saham {ticker} dan uji rumor: '{rumor}'.\n"
                            f"1. Panggil `skill_financial_health_stress_test` dengan ticker '{ticker}' dan rumor_claim '{rumor}'.\n"
                            f"2. Evaluasi rasio likuiditas (Current & Quick Ratio) dan solvabilitas (DER & Interest Coverage).\n"
                            f"3. Tetapkan status verifikasi CONTRADICTED jika kas melimpah dan tidak ada indikasi gagal bayar."
                        ),
                    },
                }
            ],
        }

    if name == "mining_commodity_divergence":
        commodity = arguments.get("commodity", "nickel")
        return {
            "description": f"Mining Commodity Divergence Analysis for {ticker}",
            "messages": [
                {
                    "role": "user",
                    "content": {
                        "type": "text",
                        "text": (
                            f"Lakukan audit korelasi harga saham tambang {ticker} terhadap komoditas {commodity}.\n"
                            f"1. Panggil `skill_mining_commodity_divergence` dengan ticker '{ticker}' dan commodity '{commodity}'.\n"
                            f"2. Evaluasi nilai koefisien korelasi Pearson r.\n"
                            f"3. Simpulkan apakah kenaikan harga merupakan COMMODITY_DRIVEN atau IDIOSYNCRATIC_COMPANY_ALPHA."
                        ),
                    },
                }
            ],
        }

    if name == "peer_valuation_benchmark":
        subsector = arguments.get("subsector", "")
        return {
            "description": f"Peer Valuation Benchmark for {ticker}",
            "messages": [
                {
                    "role": "user",
                    "content": {
                        "type": "text",
                        "text": (
                            f"Lakukan benchmarking valuasi relatif pada saham {ticker} terhadap peers industri.\n"
                            f"1. Panggil `skill_peer_valuation_benchmark` dengan ticker '{ticker}'.\n"
                            f"2. Bandingkan P/E dan P/B saham terhadap median dan IQR subsektor.\n"
                            f"3. Tetapkan valuasi posture (DISCOUNT / PREMIUM) tanpa memberikan target harga spekulatif."
                        ),
                    },
                }
            ],
        }

    if name == "event_causality_audit":
        anomaly_date = arguments.get("anomaly_date", "2026-09-12")
        return {
            "description": f"Event Causality Audit for {ticker} at {anomaly_date}",
            "messages": [
                {
                    "role": "user",
                    "content": {
                        "type": "text",
                        "text": (
                            f"Lakukan audit kausalitas berita dan keterbukaan BEI untuk lonjakan saham {ticker} pada {anomaly_date}.\n"
                            f"1. Panggil `skill_event_causality_audit` dengan ticker '{ticker}' dan anomaly_date '{anomaly_date}'.\n"
                            f"2. Cek stempel waktu rilis keterbukaan resmi BEI vs waktu lonjakan volume transaksi.\n"
                            f"3. Tentukan klasifikasi kausalitas: LIKELY_CATALYST, PRECEDED_ANNOUNCEMENT, atau UNEXPLAINED_BY_NEWS."
                        ),
                    },
                }
            ],
        }

    raise ValueError(f"Prompt not found: {name}")

