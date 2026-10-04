"""Unit tests for inspect_document tool and document_audit skill."""

import os
import pytest

from engine.agent.tools import NiskavaToolRegistry


def test_inspect_document_tool_metadata(tmp_path):
    test_file = tmp_path / "fin_report.txt"
    test_file.write_text("Ringkasan Keuangan Q3 2026 PT Bukit Asam Tbk (PTBA).", encoding="utf-8")

    registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
    res = registry.inspect_document(str(test_file))

    assert res.get("filename") == "fin_report.txt"
    assert res.get("format") == "txt"
    assert res.get("total_pages") == 1
    assert "preview" in res
    assert len(res["preview"]) >= 1


def test_inspect_document_tool_page(tmp_path):
    test_file = tmp_path / "multipage.txt"
    content = "Halaman pertama info emiten.\n" * 100 + "Halaman kedua angka laba bersih Rp 2.1T.\n" * 100
    test_file.write_text(content, encoding="utf-8")

    registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
    page1 = registry.inspect_document(str(test_file), page=1)
    assert page1.get("page") == 1
    assert "Halaman pertama" in page1.get("content", "")

    page_out_of_bounds = registry.inspect_document(str(test_file), page=999)
    assert page_out_of_bounds.get("error") is True


def test_inspect_document_tool_search(tmp_path):
    test_file = tmp_path / "disclosure.txt"
    test_file.write_text(
        "Keterbukaan Informasi: Pembagian dividen interim sebesar Rp 120 per lembar saham.",
        encoding="utf-8",
    )

    registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
    search_res = registry.inspect_document(str(test_file), query="dividen")

    assert search_res.get("filename") == "disclosure.txt"
    assert search_res.get("total_matches", 0) >= 1
    assert "dividen" in search_res["matches"][0]["snippet"].lower()


def test_inspect_document_tool_missing_file(tmp_path):
    registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
    res = registry.inspect_document("/tmp/non_existent_file_98765.txt")
    assert res.get("error") is True
    assert "not found" in res.get("message", "").lower()


def test_inspect_document_in_get_all_tool_definitions(tmp_path):
    registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
    all_defs = registry.get_all_tool_definitions()
    names = [d["name"] for d in all_defs]
    assert "inspect_document" in names

    inspect_def = next(d for d in all_defs if d["name"] == "inspect_document")
    assert "parameters" in inspect_def
    assert "doc_path" in inspect_def["parameters"]["properties"]
    assert "doc_path" in inspect_def["parameters"]["required"]


def test_execute_tool_inspect_document(tmp_path):
    test_file = tmp_path / "doc.txt"
    test_file.write_text("Test content for execute_tool.", encoding="utf-8")

    registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
    res = registry.execute_tool("inspect_document", {"doc_path": str(test_file)})
    assert res.get("filename") == "doc.txt"
    assert res.get("total_pages") == 1


def test_execute_skill_document_audit(tmp_path):
    test_file = tmp_path / "annual_report.txt"
    test_file.write_text("Laporan Tahunan 2026: Laba usaha Rp 4.5 Triliun.", encoding="utf-8")

    registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
    res = registry.execute_tool("execute_skill", {
        "skill_id": "document_audit",
        "arguments": {"doc_path": str(test_file)},
    })
    assert res.get("verification_status") in ("SUPPORTED", "UNCERTAIN")
    assert res.get("confidence_score") == 1.00
    assert "Laporan Tahunan" in res.get("summary", "")
