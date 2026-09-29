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

## 2. Input Schema
```json
{
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

## 3. Output
Returns a `SkillResult` where `metrics["pdf_path"]` contains the absolute local path
to the generated PDF file (e.g. `/home/user/.niskava/reports/NISKAVA_ANTM_20260929_ABCDEF_audit.pdf`).

## 4. Compliance
- Law 1: All quant metrics sourced from arguments or SQLite, never recomputed by LLM.
- Law 2: Non-advisory disclaimer banner is hard-coded on every page footer.
- Law 4: PDF saved to `~/.niskava/reports/`. No cloud upload.
- Law 5: Zero Sectors API calls. All data from arguments or local SQLite.
