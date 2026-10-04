"""Unit tests for deterministic document parser in document_audit skill."""

import os
import pytest
from pypdf import PdfWriter

from engine.skills.document_audit.parser import (
    parse_document,
    search_document,
    get_document_page,
)


def test_parse_text_document(tmp_path):
    sample_txt = tmp_path / "disclosure.txt"
    sample_txt.write_text(
        "PT Aneka Tambang Tbk mencatat kenaikan laba bersih sebesar 25% YoY pada Q2 2026.",
        encoding="utf-8",
    )

    doc = parse_document(str(sample_txt))
    assert doc["filename"] == "disclosure.txt"
    assert doc["format"] == "txt"
    assert doc["page_count"] == 1
    assert "PT Aneka Tambang Tbk" in doc["text"]
    assert doc["file_size"] > 0
    assert len(doc["pages"]) == 1
    assert doc["pages"][0]["page_number"] == 1


def test_search_document(tmp_path):
    sample_txt = tmp_path / "report.txt"
    content = (
        "Halaman 1: Pendahuluan dan Ringkasan Bisnis.\n"
        "Halaman 2: Laporan Keuangan Konsolidasi Laba Bersih tercatat Rp 1.5 Triliun.\n"
        "Halaman 3: Catatan Atas Laporan Keuangan dan Liabilitas Jangka Pendek."
    )
    sample_txt.write_text(content, encoding="utf-8")

    doc = parse_document(str(sample_txt))
    matches = search_document(doc, "Laba Bersih")
    assert len(matches) >= 1
    assert "Rp 1.5 Triliun" in matches[0]["snippet"]
    assert matches[0]["page_number"] == 1


def test_get_document_page(tmp_path):
    sample_txt = tmp_path / "multipage.txt"
    # Create content that spans multiple virtual pages (>2000 chars)
    page1_text = "Page 1 content " * 150  # ~2250 chars
    page2_text = "Page 2 content with revenue figures: Rp 50 Triliun " * 100
    sample_txt.write_text(page1_text + page2_text, encoding="utf-8")

    doc = parse_document(str(sample_txt))
    assert doc["page_count"] >= 2

    p1 = get_document_page(doc, 1)
    assert p1 is not None
    assert "Page 1 content" in p1

    p2 = get_document_page(doc, 2)
    assert p2 is not None
    assert "Rp 50 Triliun" in p2

    # Out of range page
    assert get_document_page(doc, 999) is None
    assert get_document_page(doc, 0) is None


def test_parse_missing_file_raises_error():
    with pytest.raises(FileNotFoundError):
        parse_document("/path/to/non_existent_file_12345.pdf")


def test_parse_unsupported_binary_format(tmp_path):
    bin_file = tmp_path / "image.xyz"
    bin_file.write_bytes(b"\x00\x01\x02\x03\x04")

    doc = parse_document(str(bin_file))
    assert doc["format"] == "xyz"
    assert doc["page_count"] == 1
    assert "Binary or unsupported format" in doc["text"]


def test_parse_pdf_document(tmp_path):
    # Create a real, minimal PDF using pypdf
    pdf_path = tmp_path / "test_report.pdf"
    writer = PdfWriter()
    writer.add_blank_page(width=200, height=200)
    with open(pdf_path, "wb") as f:
        writer.write(f)

    doc = parse_document(str(pdf_path))
    assert doc["filename"] == "test_report.pdf"
    assert doc["format"] == "pdf"
    assert doc["page_count"] == 1
    assert len(doc["pages"]) == 1
