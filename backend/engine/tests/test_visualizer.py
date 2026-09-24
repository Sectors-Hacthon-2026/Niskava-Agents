"""Unit Tests for GraphVisualizer (HTML + JSON Export).

Verifies:
- Data export format (nodes, edges, styles, effective_weight)
- Node styling by financial entity type (TICKER, BROKER, ANOMALY, etc.)
- Generation of standalone HTML containing vis-network script and payload
- Writing visualization file to disk
"""

import os
import sqlite3
import pytest

from engine.memory.graph_memory import LocalGraphMemory
from engine.memory.visualizer import GraphVisualizer, format_node_label


@pytest.fixture
def populated_memory(tmp_path):
    db_file = str(tmp_path / "test_viz.db")
    memory = LocalGraphMemory(db_path=db_file)
    memory.store_observation(
        source_label="ANTM",
        source_type="TICKER",
        relation="OPERATES",
        target_label="Smelter Haltim",
        target_type="FACILITY",
        context_snippet="Fasilitas feronikel Halmahera",
        session_id="INV-01",
    )
    memory.store_observation(
        source_label="ANTM",
        source_type="TICKER",
        relation="TRIGGERED_ANOMALY",
        target_label="VOLUME_3.8s",
        target_type="ANOMALY_METRIC",
        context_snippet="Lonjakan volume 3.8 sigma",
        session_id="INV-01",
    )
    return memory


def test_export_graph_data(populated_memory):
    viz = GraphVisualizer(memory=populated_memory)
    data = viz.export_graph_data()

    assert "nodes" in data
    assert "edges" in data
    assert "stats" in data
    assert len(data["nodes"]) == 3
    assert len(data["edges"]) == 2

    # Check node properties
    antm_node = next(n for n in data["nodes"] if n["label"] == "ANTM")
    assert antm_node["group"] == "TICKER"
    assert antm_node["shape"] == "box"
    assert antm_node["color"]["background"] == "#1E3A8A"
    assert antm_node["raw_label"] == "ANTM"
    assert antm_node["widthConstraint"] == {"maximum": 150, "minimum": 80}
    assert antm_node["margin"] == 10

    # Check edge properties
    edge = data["edges"][0]
    assert "effective_weight" in edge
    assert "width" in edge
    assert "arrows" in edge


def test_generate_html_content(populated_memory):
    viz = GraphVisualizer(memory=populated_memory)
    html = viz.generate_html(title="Test Visualizer")

    assert "<!DOCTYPE html>" in html
    assert "vis-network" in html
    assert "ANTM" in html
    assert "Smelter Haltim" in html
    assert "NISKAVA AGENT" in html
    assert "Market Intelligence" in html


def test_export_to_file(populated_memory, tmp_path):
    viz = GraphVisualizer(memory=populated_memory)
    out_file = str(tmp_path / "niskava_graph_test.html")
    saved_path = viz.export_to_file(output_path=out_file)

    assert os.path.exists(saved_path)
    assert os.path.getsize(saved_path) > 1000
    with open(saved_path, "r", encoding="utf-8") as f:
        content = f.read()
        assert "Smelter Haltim" in content


def test_institutional_theme_no_cyber_slop(populated_memory):
    """Verify that visualizer generates clean institutional financial research UI without cyber/neon slop."""
    viz = GraphVisualizer(memory=populated_memory)
    html = viz.generate_html(title="Niskava Market Intelligence")

    # Assert absence of neon cyber tropes
    assert "⚡" not in html
    assert "CYBER" not in html.upper()
    assert "#00D2FF" not in html  # Neon cyan removed
    assert "#FF0055" not in html  # Neon hot pink removed

    # Assert presence of institutional financial palette & layout
    assert "Market Intelligence" in html
    assert "tnum" in html or "tabular-nums" in html
    assert "Dossier" in html


def test_html_template_is_english(populated_memory):
    """Regression: All user-facing strings in HTML visualizer must be English (Bug #2A)."""
    viz = GraphVisualizer(memory=populated_memory)
    html = viz.generate_html(title="Test English")

    for expected in [
        "Entity Search", "Node Classification", "Market Network Statistics",
        "Refit Canvas", "Research Session:", "Loading graph statistics...",
        "No active relations.",
    ]:
        assert expected in html, f"Expected English string '{expected}' not found in HTML"

    for marker in [
        "Pencarian Entitas", "Klasifikasi Simpul", "Statistik Jaringan Pasar",
        "Posisikan Ulang Kanvas", "Sesi Riset:", "Memuat statistik graf",
        "Pilih salah satu entitas", "Tidak ada relasi aktif", "Semua Sesi Aktif",
        "Emiten Saham", "Anggota Bursa", "Sektor Industri", "Profil Riset Pengguna",
        "Total Simpul", "Relasi Pasar", "Entitas Sentral",
        "Waktu Observasi", "Derajat Relasi", "Katalog Relasi Bukti",
        "Koneksi Kausalitas", "Kutipan Bukti", "Bobot Efektif", "Sesi Investigasi",
    ]:
        assert marker not in html, (
            f"Indonesian UI string '{marker}' found in visualizer HTML — must be English."
        )


def test_format_node_label_wraps_long_text():
    """Verify that format_node_label wraps long headlines into balanced multi-line cards."""
    # Empty or None string handling
    assert format_node_label("") == ""
    assert format_node_label(None) == ""

    # Short label stays unchanged on one line
    short = format_node_label("ANTM")
    assert short == "ANTM"
    assert "\n" not in short

    # Long headline wraps into balanced lines (default max_chars_per_line=22, max_lines=3)
    long_headline = "PT Aneka Tambang Tbk Resmikan Smelter Haltim dengan Kapasitas Produksi Feronikel 13.500 TNi per Tahun"
    formatted = format_node_label(long_headline, max_chars_per_line=22, max_lines=3)
    lines = formatted.split("\n")
    assert 1 < len(lines) <= 3
    for line in lines:
        assert len(line) <= 22
    assert "..." in lines[-1]

    # Custom constraints: max_lines=2, max_chars_per_line=15
    two_lines = format_node_label("Alpha Beta Gamma Delta Epsilon Zeta Eta Theta", max_chars_per_line=15, max_lines=2)
    two_lines_list = two_lines.split("\n")
    assert len(two_lines_list) <= 2
    for line in two_lines_list:
        assert len(line) <= 15


def test_export_graph_data_excludes_orphan_nodes_when_filtered(tmp_path):
    """Verify scoped export excludes orphan nodes with 0 active edges."""
    db_file = str(tmp_path / "test_orphan_filter.db")
    memory = LocalGraphMemory(db_path=db_file)

    # Session S1 observation: ANTM -> Smelter Haltim
    memory.store_observation(
        source_label="ANTM",
        source_type="TICKER",
        relation="OPERATES",
        target_label="Smelter Haltim",
        target_type="FACILITY",
        session_id="S1",
    )
    # Session S2 observation: BBCA -> BCA Finance
    memory.store_observation(
        source_label="BBCA",
        source_type="TICKER",
        relation="OWNS",
        target_label="BCA Finance",
        target_type="FACILITY",
        session_id="S2",
    )

    # Directly insert an orphan node with 0 edges into SQLite memory_nodes
    with sqlite3.connect(memory.db_path) as conn:
        conn.execute(
            """
            INSERT INTO memory_nodes (id, label, node_type, metadata_json, last_observed_at)
            VALUES (?, ?, ?, ?, datetime('now'))
            """,
            ("isolated:orphan", "Isolated Orphan Entity", "ENTITY", "{}"),
        )

    viz = GraphVisualizer(memory=memory)

    # Unscoped export includes all nodes in the graph including the orphan
    unscoped_data = viz.export_graph_data()
    unscoped_ids = {n["id"] for n in unscoped_data["nodes"]}
    assert "isolated:orphan" in unscoped_ids
    assert "ticker:antm" in unscoped_ids
    assert "ticker:bbca" in unscoped_ids
    assert len(unscoped_data["nodes"]) == 5

    # Scoped by session_id='S1': ONLY nodes in active edges for S1 are included (no orphan, no S2 nodes)
    s1_data = viz.export_graph_data(session_id="S1")
    s1_node_ids = {n["id"] for n in s1_data["nodes"]}
    assert s1_node_ids == {"ticker:antm", "facility:smelter_haltim"}
    assert "isolated:orphan" not in s1_node_ids
    assert "ticker:bbca" not in s1_node_ids
    assert len(s1_data["edges"]) == 1

    # Scoped by ticker='ANTM': centers ego-subgraph on ANTM and excludes orphan & S2 nodes
    ticker_data = viz.export_graph_data(ticker="ANTM", depth=1)
    ticker_node_ids = {n["id"] for n in ticker_data["nodes"]}
    assert "ticker:antm" in ticker_node_ids
    assert "facility:smelter_haltim" in ticker_node_ids
    assert "isolated:orphan" not in ticker_node_ids
    assert "ticker:bbca" not in ticker_node_ids

    # Node types filtering
    ticker_only = viz.export_graph_data(node_types=["TICKER"])
    assert all(n["group"] == "TICKER" for n in ticker_only["nodes"])
    assert "ticker:antm" in {n["id"] for n in ticker_only["nodes"]}
    assert "facility:smelter_haltim" not in {n["id"] for n in ticker_only["nodes"]}

    # Verify node payload formatting
    for node in s1_data["nodes"]:
        assert "label" in node
        assert "raw_label" in node
        assert "widthConstraint" in node
        assert node["widthConstraint"] == {"maximum": 150, "minimum": 80}
        assert "color" in node
        assert node["margin"] == 10


def test_generate_html_embed_mode(populated_memory, tmp_path):
    """Verify embed mode CSS injection and parameter propagation."""
    viz = GraphVisualizer(memory=populated_memory)

    # Standard mode (embed=False): sidebar and top-bar are visible, no override style
    html_standard = viz.generate_html(embed=False)
    assert "<!DOCTYPE html>" in html_standard
    assert ".sidebar { display: none !important; }" not in html_standard

    # Embed mode (embed=True): injects CSS to hide sidebar & top-bar and expand canvas
    html_embed = viz.generate_html(embed=True)
    expected_css = "<style>.sidebar { display: none !important; } .top-bar { display: none !important; } .canvas-area { width: 100vw; height: 100vh; }</style>"
    assert expected_css in html_embed

    # Verify parameters propagation to export_to_file
    out_file = str(tmp_path / "embed_test.html")
    saved = viz.export_to_file(
        output_path=out_file,
        ticker="ANTM",
        depth=1,
        embed=True,
    )
    assert os.path.exists(saved)
    with open(saved, "r", encoding="utf-8") as f:
        saved_content = f.read()
    assert expected_css in saved_content
    assert "ANTM" in saved_content


