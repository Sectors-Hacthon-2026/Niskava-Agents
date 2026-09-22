"""Unit Tests for GraphVisualizer (HTML + JSON Export).

Verifies:
- Data export format (nodes, edges, styles, effective_weight)
- Node styling by financial entity type (TICKER, BROKER, ANOMALY, etc.)
- Generation of standalone HTML containing vis-network script and payload
- Writing visualization file to disk
"""

import os
import pytest

from engine.memory.graph_memory import LocalGraphMemory
from engine.memory.visualizer import GraphVisualizer


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

