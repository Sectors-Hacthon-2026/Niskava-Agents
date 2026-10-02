---
name: investigation-report-pdf
description: Generate and export an institutional PDF research note and audit trail for an IDX stock investigation. ONLY invoke this skill when the user explicitly requests a PDF, downloadable report, printed document, or file export. Do NOT invoke for standard analytical questions.
triggers:
  - User explicitly asks to export, download, or print a PDF report
  - User says "buatkan laporan PDF", "export PDF", "generate report PDF", "simpan ke PDF", "cetak laporan"
  - User requests a downloadable or shareable investigation document
  - Follow-up command to convert current investigation findings into a PDF file
tools:
  - execute_skill
---

# Investigation Report PDF Skill

## 1. Overview
The `investigation-report-pdf` skill generates a deterministic, institutional-grade PDF
audit trail for a completed IDX stock investigation. It reads pre-computed metrics and
evidence from local SQLite or the calling agent's `arguments`, and renders them into a
formatted document using pure-Python `fpdf2`.

**This skill produces a file, not a market analysis.** It must be preceded by analytical
skills such as `market_anomaly_recon` or `event_causality_audit`.

## 2. Dynamic Input Modes
The skill supports three distinct report layouts automatically adapted to user intent:

### Mode A: Ticker Quantitative Investigation (`TICKER_INVESTIGATION`)
Use when analyzing a specific stock with quantitative anomalies:
```json
{
  "report_type": "TICKER_INVESTIGATION",
  "ticker": "ANTM",
  "title": "ANTM Investigation Audit Trail — 30 Days",
  "summary": "Agent-generated narrative synthesis from prior investigation steps.",
  "metrics": {
    "volume_z_score": 3.82,
    "foreign_flow_z_score": 2.61,
    "abnormal_return_pct": 6.4,
    "candles_analyzed": 30,
    "anomaly_detected": true,
    "latest_anomaly_date": "2026-09-27"
  },
  "evidence": [
    {
      "date": "2026-09-27",
      "headline": "ANTM signs joint venture for nickel smelter",
      "verification_status": "SUPPORTED",
      "confidence_score": 0.95,
      "source": "IDXnet Disclosure"
    }
  ],
  "session_id": "INV-20260929-ABCDEF"
}
```

### Mode B: Macro Market News Digest (`MARKET_NEWS_BRIEF`)
Use when summarizing general market news, macro sentiment, or index overview (no N/A tables generated):
```json
{
  "report_type": "MARKET_NEWS_BRIEF",
  "ticker": "MARKET",
  "title": "IDX Daily Market Intelligence Brief — 2026-09-29",
  "summary": "Executive digest of top macroeconomic and corporate developments.",
  "news_items": [
    {
      "date": "2026-09-29",
      "headline": "Bank Indonesia adjusts monetary intervention strategy",
      "source": "Sectors News",
      "status": "VERIFIED"
    }
  ],
  "session_id": "WEB-20260929-001"
}
```

### Mode C: Custom Research Note (`CUSTOM_RESEARCH`)
Use for custom research notes with structured headings:
```json
{
  "report_type": "CUSTOM_RESEARCH",
  "ticker": "SECTOR",
  "title": "Mining vs Banking Sector Divergence Study",
  "summary": "Comparative structural overview.",
  "sections": [
    {
      "heading": "Commodity Price Outlook",
      "content": "Global nickel benchmark movements continue to dictate margins."
    }
  ]
}
```

## 3. Output
Returns a `SkillResult` where `metrics["pdf_path"]` contains the absolute local path
to the generated PDF file (e.g. `/home/user/.niskava/reports/NISKAVA_ANTM_20260929_ABCDEF_audit.pdf`).

## 4. Compliance
- Law 1: All quant metrics sourced from arguments or SQLite, never recomputed by LLM. Empty metrics are cleanly suppressed rather than displaying empty N/A matrices.
- Law 2: Non-advisory disclaimer banner is hard-coded on every page footer.
- Law 4: PDF saved to `~/.niskava/reports/`. No cloud upload.
- Law 5: Zero Sectors API calls. All data from arguments or local SQLite.

