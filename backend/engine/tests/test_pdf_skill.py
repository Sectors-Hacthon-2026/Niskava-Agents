"""Tests for the investigation-report-pdf skill and PDF renderer.

Covers:
- PDF file creation and existence at expected local path
- Filename contains ticker and today's date
- Non-advisory disclaimer is present on PDF output (Law 2)
- Output directory auto-creation if not already existing (Law 4)
- SkillResult validity, schema, and metric shape
- Tool definition compliance with function-calling schema
- Zero Sectors API calls during execution (Law 5 credit discipline)
- MOCK_SECTORS=1 offline compatibility
- PDF trigger prompts classified as 'deep' complexity in ReAct agent
- End-to-end integration dispatch via NiskavaToolRegistry.execute_skill
- Event emission for 'pdf_report_ready' when emitter is present in context
"""

import os
from datetime import date
from pathlib import Path
from typing import Any, Dict

import pytest


def _make_params() -> Dict[str, Any]:
    return {
        "ticker": "TESTX",
        "title": "TESTX Test Audit Trail",
        "summary": "This is an automated test summary. Volume anomaly detected.",
        "metrics": {
            "volume_z_score": 3.12,
            "foreign_flow_z_score": 1.80,
            "abnormal_return_pct": 5.5,
            "candles_analyzed": 30,
            "anomaly_detected": True,
            "latest_anomaly_date": "2026-09-29",
        },
        "evidence": [
            {
                "date": "2026-09-29",
                "headline": "TESTX test disclosure event",
                "verification_status": "SUPPORTED",
                "confidence_score": 0.95,
                "source": "Test Source",
            }
        ],
        "session_id": "TEST-SESSION-001",
    }


def test_render_investigation_pdf_returns_existing_file(tmp_path):
    """render_investigation_pdf must return an absolute path to an existing PDF file."""
    from engine.skills.investigation_report_pdf.renderer import render_investigation_pdf

    params = _make_params()
    output_path = render_investigation_pdf(params, output_dir=tmp_path)

    assert os.path.isabs(output_path), "Path must be absolute"
    assert os.path.exists(output_path), f"PDF file must exist at {output_path}"
    assert output_path.endswith(".pdf"), "Output must be a .pdf file"


def test_render_pdf_filename_contains_ticker_and_date(tmp_path):
    """Output filename must contain the ticker, today's date, and session suffix."""
    from engine.skills.investigation_report_pdf.renderer import render_investigation_pdf

    params = _make_params()
    output_path = render_investigation_pdf(params, output_dir=tmp_path)

    filename = Path(output_path).name
    assert "TESTX" in filename
    assert date.today().strftime("%Y%m%d") in filename
    assert filename.startswith("NISKAVA_")
    assert filename.endswith("_audit.pdf")


def test_render_pdf_contains_disclaimer(tmp_path):
    """The generated PDF binary must contain the non-advisory disclaimer text (Law 2)."""
    from engine.skills.investigation_report_pdf.renderer import (
        render_investigation_pdf,
        NON_ADVISORY_DISCLAIMER,
    )

    params = _make_params()
    output_path = render_investigation_pdf(params, output_dir=tmp_path)

    pdf_bytes = Path(output_path).read_bytes()
    # fpdf2 with set_compression(False) leaves text searchable; fallback to decompressed stream if any
    found = (
        b"non-advisory" in pdf_bytes.lower()
        or NON_ADVISORY_DISCLAIMER.encode("latin-1") in pdf_bytes
    )
    if not found:
        import zlib, re
        for m in re.finditer(rb"stream\r?\n(.*?)\r?\nendstream", pdf_bytes, re.DOTALL):
            try:
                if b"non-advisory" in zlib.decompress(m.group(1)).lower():
                    found = True
                    break
            except Exception:
                pass
    assert found, "PDF must contain non-advisory disclaimer text"


def test_render_pdf_output_dir_is_created(tmp_path):
    """render_investigation_pdf must create the output directory if it does not exist."""
    from engine.skills.investigation_report_pdf.renderer import render_investigation_pdf

    nested_dir = tmp_path / "new" / "nested" / "reports"
    assert not nested_dir.exists()

    params = _make_params()
    output_path = render_investigation_pdf(params, output_dir=nested_dir)

    assert nested_dir.exists()
    assert os.path.exists(output_path)


def test_skill_returns_valid_skill_result(tmp_path):
    """InvestigationReportPdfSkill.execute must return a SkillResult with pdf_path in metrics."""
    from engine.skills.investigation_report_pdf.logic import InvestigationReportPdfSkill
    from engine.skills.base import SkillResult

    skill = InvestigationReportPdfSkill()
    params = _make_params()

    result = skill.execute(arguments=params, context={"_output_dir": tmp_path})

    assert isinstance(result, SkillResult)
    assert result.skill_id == "investigation-report-pdf"
    assert result.verification_status == "SUPPORTED"
    assert result.confidence_score == 1.00
    assert "pdf_path" in result.metrics
    assert os.path.exists(result.metrics["pdf_path"])
    assert result.metrics["ticker"] == "TESTX"
    assert "filename" in result.metrics


def test_skill_tool_definition_shape():
    """get_tool_definition must return a dict with name, description, and parameters schema."""
    from engine.skills.investigation_report_pdf.logic import InvestigationReportPdfSkill

    skill = InvestigationReportPdfSkill()
    tool_def = skill.get_tool_definition()

    assert tool_def["name"] == "skill_investigation_report_pdf"
    assert "description" in tool_def
    assert "parameters" in tool_def
    required = tool_def["parameters"].get("required", [])
    assert "summary" in required
    props = tool_def["parameters"].get("properties", {})
    for expected_prop in ["ticker", "title", "summary", "metrics", "evidence", "session_id"]:
        assert expected_prop in props, f"Property '{expected_prop}' must be in tool schema"


def test_skill_does_not_call_sectors_api(tmp_path, monkeypatch):
    """execute must not call SectorsAPIClient — zero API credit spend (Law 5)."""
    from engine.skills.investigation_report_pdf.logic import InvestigationReportPdfSkill

    call_log = []

    class FakeSectorsClient:
        def __getattr__(self, name):
            def _spy(*args, **kwargs):
                call_log.append(name)
                raise AssertionError(f"SectorsAPIClient.{name} was unexpectedly called during PDF export!")
            return _spy

    skill = InvestigationReportPdfSkill()
    params = _make_params()

    skill.execute(arguments=params, context={"sectors_client": FakeSectorsClient(), "_output_dir": tmp_path})
    assert len(call_log) == 0, f"Unexpected Sectors API calls: {call_log}"


def test_skill_works_with_mock_sectors_env(tmp_path, monkeypatch):
    """Skill must work when MOCK_SECTORS=1 is set (offline mode)."""
    monkeypatch.setenv("MOCK_SECTORS", "1")
    from engine.skills.investigation_report_pdf.logic import InvestigationReportPdfSkill

    skill = InvestigationReportPdfSkill()
    params = _make_params()
    result = skill.execute(arguments=params, context={"_output_dir": tmp_path})

    assert result.metrics.get("pdf_path")
    assert os.path.exists(result.metrics["pdf_path"])


def test_pdf_keywords_classified_as_deep():
    """PDF export prompts must be classified as 'deep' complexity so the ReAct loop runs."""
    from engine.agent.react_agent import classify_prompt_complexity

    deep_prompts = [
        "Buatkan laporan PDF investigasi BBRI",
        "Export PDF analysis ANTM",
        "Generate a PDF report for TLKM investigation",
        "Simpan hasil investigasi ADRO ke PDF",
        "Tolong buat laporan pdf untuk saham BREN",
        "Cetak laporan investigasi BMRI",
        "download report audit trail",
        "simpan ke pdf sekarang",
    ]
    for prompt in deep_prompts:
        complexity = classify_prompt_complexity(prompt)
        assert complexity == "deep", (
            f"Prompt '{prompt}' should be 'deep' but got '{complexity}'."
        )


def test_skill_emits_pdf_report_ready_event(tmp_path):
    """InvestigationReportPdfSkill.execute must emit 'pdf_report_ready' if emitter is provided."""
    from engine.skills.investigation_report_pdf.logic import InvestigationReportPdfSkill

    events = []
    skill = InvestigationReportPdfSkill()
    params = _make_params()

    result = skill.execute(
        arguments=params,
        context={"_output_dir": tmp_path, "emitter": events.append},
    )

    assert len(events) == 1
    event = events[0]
    assert event["event"] == "pdf_report_ready"
    assert event["pdf_path"] == result.metrics["pdf_path"]
    assert event["ticker"] == "TESTX"
    assert event["filename"] == os.path.basename(result.metrics["pdf_path"])


def test_execute_skill_via_tool_registry_integration(tmp_path, monkeypatch):
    """Simulate the ReAct agent calling execute_skill('investigation_report_pdf') via NiskavaToolRegistry.

    Verifies the full dispatch chain: ToolRegistry.execute_skill -> SkillsRegistry.execute_skill
    -> InvestigationReportPdfSkill.execute -> render_investigation_pdf -> PDF on disk.
    """
    monkeypatch.setenv("MOCK_SECTORS", "1")

    from engine.agent.tools import NiskavaToolRegistry
    from engine.skills.registry import SkillsRegistry

    db_path = str(tmp_path / "test.db")
    events = []

    registry = SkillsRegistry()
    tool_registry = NiskavaToolRegistry(db_path=db_path, mock_mode=True, skills_registry=registry)
    tool_registry.emitter = events.append

    # Monkeypatch the skill's execute to redirect output to tmp_path
    pdf_skill = registry.get_skill("investigation-report-pdf")
    assert pdf_skill is not None, "investigation-report-pdf skill must be auto-discovered"

    original_execute = pdf_skill.execute

    def patched_execute(arguments, context):
        context["_output_dir"] = tmp_path
        return original_execute(arguments, context)

    pdf_skill.execute = patched_execute

    # Simulate what ReAct agent emits
    result_dict = tool_registry.execute_skill(
        "investigation_report_pdf",
        {
            "ticker": "BBCA",
            "summary": "Integration test summary for BBCA audit trail.",
            "metrics": {
                "volume_z_score": 2.9,
                "foreign_flow_z_score": 1.2,
                "abnormal_return_pct": 3.1,
                "candles_analyzed": 30,
                "anomaly_detected": True,
                "latest_anomaly_date": "2026-09-29",
            },
            "evidence": [],
            "session_id": "INT-TEST-001",
        },
    )

    assert result_dict["skill_id"] == "investigation-report-pdf"
    assert result_dict["verification_status"] == "SUPPORTED"
    assert "pdf_path" in result_dict["metrics"]
    assert os.path.exists(result_dict["metrics"]["pdf_path"])

    # Verify event was emitted
    assert len(events) == 1
    assert events[0]["event"] == "pdf_report_ready"
    assert events[0]["ticker"] == "BBCA"


def test_render_market_news_brief_hides_quant_matrix(tmp_path):
    """When report_type is MARKET_NEWS_BRIEF and metrics are empty/absent, quant anomaly table must NOT be rendered."""
    from engine.skills.investigation_report_pdf.renderer import render_investigation_pdf

    params = {
        "title": "IDX Daily Market News Brief",
        "ticker": "MARKET",
        "report_type": "MARKET_NEWS_BRIEF",
        "summary": "Bank Indonesia updates intervention strategy as Rupiah tests key psychological levels.",
        "news_items": [
            {
                "date": "2026-09-29",
                "headline": "Rupiah reaches Rp18.000 per USD amid macro shifts",
                "source": "Sectors News",
                "sentiment": "NEUTRAL",
            },
            {
                "date": "2026-09-29",
                "headline": "GoTo regaining liquidity following minimum price regulation",
                "source": "IDXnet Disclosures",
                "sentiment": "BULLISH",
            },
        ],
        "session_id": "TEST-NEWS-001",
    }

    output_path = render_investigation_pdf(params, output_dir=tmp_path)
    assert os.path.exists(output_path)

    # Read binary to verify content
    pdf_bytes = Path(output_path).read_bytes()
    # Should NOT have the quant table title
    assert b"QUANTITATIVE ANOMALY MATRIX" not in pdf_bytes
    # Should have the news table or executive summary
    assert b"EXECUTIVE SUMMARY" in pdf_bytes
    assert b"MARKET NEWS DIGEST" in pdf_bytes or b"Rupiah reaches Rp18.000" in pdf_bytes


def test_render_custom_sections_block(tmp_path):
    """Renderer should accept arbitrary custom sections and render them cleanly."""
    from engine.skills.investigation_report_pdf.renderer import render_investigation_pdf

    params = {
        "title": "Sector Divergence Analysis",
        "ticker": "ENERGY",
        "summary": "Comparative analysis across energy and basic materials sectors.",
        "sections": [
            {
                "heading": "Monetary Policy Impact",
                "content": "Interest rate stabilization expected to reduce debt servicing pressures.",
            },
            {
                "heading": "Commodity Supply Chain Outlook",
                "content": "Nickel and coal export margins remain resilient despite currency headwinds.",
            },
        ],
        "session_id": "TEST-SECTIONS-001",
    }

    output_path = render_investigation_pdf(params, output_dir=tmp_path)
    assert os.path.exists(output_path)
    pdf_bytes = Path(output_path).read_bytes()
    assert b"MONETARY POLICY IMPACT" in pdf_bytes
    assert b"COMMODITY SUPPLY CHAIN OUTLOOK" in pdf_bytes
    assert b"Interest rate stabilization" in pdf_bytes


def test_skill_execution_without_ticker_defaults_to_market(tmp_path):
    """When ticker is omitted from skill arguments, skill should succeed and default to MARKET."""
    from engine.skills.investigation_report_pdf.logic import InvestigationReportPdfSkill

    skill = InvestigationReportPdfSkill()
    res = skill.execute(
        arguments={
            "summary": "IDX Composite closed higher by 0.8% following positive sentiment from regional markets.",
            "report_type": "MARKET_NEWS_BRIEF",
            "news_items": [
                {
                    "date": "2026-09-29",
                    "headline": "IHSG Ditutup Menguat 0,8 Persen ke Level 7.750",
                    "source": "IDXNet",
                    "sentiment": "BULLISH",
                }
            ],
        },
        context={"_output_dir": tmp_path},
    )

    assert res.verification_status == "SUPPORTED"
    assert res.metrics["ticker"] == "MARKET"
    assert os.path.exists(res.metrics["pdf_path"])
    assert "NISKAVA_MARKET_" in os.path.basename(res.metrics["pdf_path"])


def test_renderer_clean_text_extended_symbols():
    """Verify that extended symbols (bullets, arrows, smart quotes) are safely sanitized."""
    from engine.skills.investigation_report_pdf.renderer import _clean_text

    raw = "Analisis: • Laba bersih ↑ 15% & dividen → Rp500/lembar — 'Target' “Realisasi”"
    cleaned = _clean_text(raw, single_line=True)
    assert "•" not in cleaned
    assert "-" in cleaned
    assert "+" in cleaned
    assert "->" in cleaned
    assert "--" in cleaned
    # Ensure latin-1 encodable without throwing
    cleaned.encode("latin-1")


def test_render_markdown_and_callout_box(tmp_path):
    """Test that markdown formatting and callout box render cleanly without fpdf exceptions."""
    from engine.skills.investigation_report_pdf.renderer import (
        _NiskavaReportPDF,
        _render_markdown_text,
        _render_callout_box,
    )

    pdf = _NiskavaReportPDF(ticker="ANTM", session_id="TEST01")
    pdf.add_page()

    # Callout box
    _render_callout_box(pdf, text="Key Takeaway: Nickel supply surplus expected in Q3.", title="EXECUTIVE TAKEAWAYS", style="info")

    # Markdown formatted text
    md_content = (
        "### 1. Market Overview\n"
        "Saham **ANTM** mencatatkan *abnormal return* sebesar **+6.2%**.\n\n"
        "- Volume spike mencapai 3.84x sigma\n"
        "- Net foreign inflow sebesar Rp 45 Miliar\n"
        "- IDX net disclosure terverifikasi resmi"
    )
    _render_markdown_text(pdf, md_content)

    out_file = tmp_path / "test_md_callout.pdf"
    pdf.output(str(out_file))

    assert out_file.exists()
    assert out_file.stat().st_size > 1000


def test_render_generic_table_and_key_value_grid(tmp_path):
    """Test deterministic rendering of generic tables and key-value grids."""
    from engine.skills.investigation_report_pdf.renderer import (
        _NiskavaReportPDF,
        _render_generic_table,
        _render_key_value_grid,
    )

    pdf = _NiskavaReportPDF(ticker="BBCA", session_id="TEST-GRID-001")
    pdf.add_page()

    # 1. Render Generic Table with custom widths and rows
    table_data = {
        "title": "PEER VALUATION COMPARISON",
        "headers": ["Ticker", "PER", "PBV", "ROE (%)"],
        "col_widths": [40, 45, 45, 60],
        "rows": [
            ["BBCA", "18.5x", "4.2x", "22.1%"],
            ["BBRI", "12.8x", "2.1x", "17.4%"],
            ["BMRI", "11.0x", "1.9x", "16.8%"],
        ],
    }
    _render_generic_table(pdf, table_data)

    # 2. Render Key-Value Grid (dict form)
    kv_dict = {
        "title": "FINANCIAL HIGHLIGHTS",
        "Market Capitalization": "Rp 1,250 Trillion",
        "Dividend Yield": "3.8%",
        "Free Float": "45.2%",
        "Sector": "Financial Services",
    }
    _render_key_value_grid(pdf, kv_dict)

    # 3. Render Key-Value Grid (explicit items and cols=1 form)
    kv_items = {
        "title": "CORPORATE METRICS",
        "cols": 1,
        "items": [
            {"key": "Auditor", "value": "Big Four Accounting Firm"},
            {"key": "Listing Date", "value": "2000-05-31"},
        ],
    }
    _render_key_value_grid(pdf, kv_items)

    # 4. Graceful handling of empty table & empty kv_data
    _render_generic_table(pdf, {})
    _render_generic_table(pdf, {"headers": [], "rows": []})
    _render_key_value_grid(pdf, {})
    _render_key_value_grid(pdf, {"items": []})

    out_file = tmp_path / "test_table_kv_grid.pdf"
    pdf.output(str(out_file))

    assert out_file.exists()
    assert out_file.stat().st_size > 1500

    pdf_bytes = out_file.read_bytes()
    assert b"PEER VALUATION COMPARISON" in pdf_bytes
    assert b"BBCA" in pdf_bytes
    assert b"18.5x" in pdf_bytes
    assert b"FINANCIAL HIGHLIGHTS" in pdf_bytes
    assert b"Market Capitalization" in pdf_bytes
    assert b"CORPORATE METRICS" in pdf_bytes
    assert b"Big Four" in pdf_bytes


def test_render_investigation_pdf_dynamic_blocks(tmp_path):
    """Test dynamic block-based rendering in render_investigation_pdf and backward compatibility."""
    from engine.skills.investigation_report_pdf.renderer import (
        render_investigation_pdf,
        NON_ADVISORY_DISCLAIMER,
    )

    # 1. Sequential dynamic blocks test
    blocks_params = {
        "ticker": "BBRI",
        "title": "BBRI Dynamic Multi-Block Forensic Audit Trail",
        "session_id": "DYN-BLOCKS-001",
        "blocks": [
            {
                "type": "callout",
                "title": "ANOMALY ALERT",
                "text": "Abnormal volume spike detected concurrently with IDX corporate disclosure.",
                "style": "warning",
            },
            {
                "type": "markdown",
                "title": "FORENSIC NARRATIVE",
                "content": (
                    "### Investigation Overview\n"
                    "Forensic audit reveals **substantial foreign accumulation** prior to news.\n"
                    "- Statistical anomaly verified by NumPy gate\n"
                    "- IDXnet disclosure confirms dividend distribution plan"
                ),
            },
            {
                "type": "key_value",
                "title": "VALUATION MULTIPLES",
                "data": {
                    "P/E Ratio": "12.8x",
                    "P/BV Ratio": "2.1x",
                    "ROE": "17.4%",
                    "Dividend Payout": "80%",
                },
            },
            {
                "type": "table",
                "title": "PEER BENCHMARK MATRIX",
                "headers": ["Bank", "Ticker", "NPL (%)", "NIM (%)"],
                "rows": [
                    ["Bank Rakyat Indonesia", "BBRI", "2.8%", "7.6%"],
                    ["Bank Mandiri", "BMRI", "1.2%", "5.3%"],
                    ["Bank Central Asia", "BBCA", "1.8%", "5.6%"],
                ],
            },
            {
                "type": "quant_metrics",
                "metrics": {
                    "volume_z_score": 3.45,
                    "foreign_flow_z_score": 2.80,
                    "abnormal_return_pct": 5.8,
                    "candles_analyzed": 60,
                    "anomaly_detected": True,
                    "latest_anomaly_date": "2026-09-28",
                },
            },
            {
                "type": "evidence_matrix",
                "evidence": [
                    {
                        "date": "2026-09-28",
                        "headline": "BBRI distributes interim cash dividend of Rp 120/share",
                        "verification_status": "SUPPORTED",
                        "confidence_score": 1.00,
                    }
                ],
            },
            {
                "type": "news_matrix",
                "news_items": [
                    {
                        "date": "2026-09-28",
                        "source": "IDXnet",
                        "headline": "Keterbukaan Informasi Rencana Dividen Interim BBRI",
                        "status": "SUPPORTED",
                    }
                ],
            },
        ],
    }

    dyn_pdf_path = render_investigation_pdf(blocks_params, output_dir=tmp_path)
    assert os.path.isabs(dyn_pdf_path)
    assert os.path.exists(dyn_pdf_path)

    dyn_bytes = Path(dyn_pdf_path).read_bytes()
    assert b"ANOMALY ALERT" in dyn_bytes
    assert b"FORENSIC NARRATIVE" in dyn_bytes
    assert b"VALUATION MULTIPLES" in dyn_bytes
    assert b"PEER BENCHMARK MATRIX" in dyn_bytes
    assert b"QUANTITATIVE ANOMALY MATRIX" in dyn_bytes
    assert b"EVIDENCE & CAUSALITY MATRIX" in dyn_bytes
    assert b"MARKET NEWS DIGEST" in dyn_bytes
    assert NON_ADVISORY_DISCLAIMER.encode("latin-1") in dyn_bytes or b"non-advisory" in dyn_bytes.lower()

    # 2. Backward compatibility test: legacy keys including custom_tables
    legacy_params = {
        "ticker": "BMRI",
        "title": "BMRI Legacy Investigation Report",
        "session_id": "LEGACY-002",
        "summary": "Mandiri demonstrates solid net interest margin performance.",
        "metrics": {
            "volume_z_score": 2.75,
            "candles_analyzed": 30,
            "anomaly_detected": True,
        },
        "custom_tables": [
            {
                "title": "CREDIT QUALITY BREAKDOWN",
                "headers": ["Segment", "Loan Share", "NPL"],
                "rows": [
                    ["Corporate", "48%", "0.9%"],
                    ["Commercial", "22%", "1.5%"],
                    ["Micro & Retail", "30%", "2.1%"],
                ],
            }
        ],
    }

    legacy_pdf_path = render_investigation_pdf(legacy_params, output_dir=tmp_path)
    assert os.path.exists(legacy_pdf_path)
    legacy_bytes = Path(legacy_pdf_path).read_bytes()
    assert b"EXECUTIVE SUMMARY" in legacy_bytes
    assert b"Mandiri demonstrates solid" in legacy_bytes
    assert b"QUANTITATIVE ANOMALY MATRIX" in legacy_bytes
    assert b"CREDIT QUALITY BREAKDOWN" in legacy_bytes
    assert b"Corporate" in legacy_bytes


def test_skill_execute_with_blocks(tmp_path):
    """Test InvestigationReportPdfSkill.execute when passed dynamic blocks."""
    from engine.skills.investigation_report_pdf.logic import InvestigationReportPdfSkill

    skill = InvestigationReportPdfSkill()
    events = []
    result = skill.execute(
        arguments={
            "ticker": "ASII",
            "blocks": [
                {
                    "type": "callout",
                    "title": "SECTOR OVERVIEW",
                    "text": "Automotive market share expanding across EV and hybrid categories.",
                },
                {
                    "type": "table",
                    "title": "SALES VOLUME",
                    "headers": ["Quarter", "4W Units", "2W Units"],
                    "rows": [
                        {"Quarter": "Q1 2026", "4W Units": "125,000", "2W Units": "1,100,000"},
                        {"Quarter": "Q2 2026", "4W Units": "132,000", "2W Units": "1,150,000"},
                    ],
                },
            ],
            "session_id": "ASII-TEST-BLOCKS",
        },
        context={"_output_dir": tmp_path, "emitter": events.append},
    )

    assert result.verification_status == "SUPPORTED"
    assert result.confidence_score == 1.00
    assert os.path.exists(result.metrics["pdf_path"])
    assert len(events) == 1
    assert events[0]["event"] == "pdf_report_ready"
    assert events[0]["ticker"] == "ASII"



