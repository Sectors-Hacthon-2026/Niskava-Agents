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
                            f"3. Jika ada anomali, panggil `osint_harvest_market_news` untuk mencari pengumuman resmi atau berita katalis.\n"
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

    raise ValueError(f"Prompt not found: {name}")
