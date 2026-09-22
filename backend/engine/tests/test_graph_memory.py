"""Unit Tests for LocalGraphMemory (NetworkX + SQLite).

Verifies:
- Auto-DDL self healing
- Observation storage and edge weight boosting on repetition
- Deterministic investigation recording (Law 1, Law 6)
- NetworkX DiGraph loading
- Ego-Graph traversal (k <= 2) and exponential recency decay
- Prompt formatting into <investigative_memory> XML
- Shortest path tracing between market entities
- Topological graph statistics and central entities
- Selective and full memory reset
"""

from datetime import datetime, timedelta, timezone
import os
import sqlite3
import pytest

from engine.memory.graph_memory import LocalGraphMemory


@pytest.fixture
def temp_memory(tmp_path):
    db_file = str(tmp_path / "test_memory.db")
    return LocalGraphMemory(db_path=db_file, lambda_decay=0.05, max_radius=2)


def test_schema_creation(temp_memory):
    """Test that SQLite tables are created properly upon initialization."""
    with sqlite3.connect(temp_memory.db_path) as conn:
        cursor = conn.cursor()
        cursor.execute("SELECT name FROM sqlite_master WHERE type='table'")
        tables = {row[0] for row in cursor.fetchall()}
        assert "memory_nodes" in tables
        assert "memory_edges" in tables


def test_store_observation_and_weight_boost(temp_memory):
    """Test storing observations and ensuring repeat relations boost weight."""
    # First observation
    res1 = temp_memory.store_observation(
        source_label="ANTM",
        source_type="TICKER",
        relation="OPERATES",
        target_label="Smelter Haltim",
        target_type="FACILITY",
        context_snippet="Pabrik feronikel Halmahera Timur",
        session_id="INV-001",
    )
    assert res1["status"] == "STORED"
    assert res1["source"]["id"] == "ticker:antm"
    assert res1["target"]["id"] == "facility:smelter_haltim"

    # Second observation (same edge)
    res2 = temp_memory.store_observation(
        source_label="ANTM",
        source_type="TICKER",
        relation="OPERATES",
        target_label="Smelter Haltim",
        target_type="FACILITY",
        context_snippet="Uji coba komisioning",
        session_id="INV-002",
    )
    assert res2["status"] == "STORED"

    # Verify weight increment in SQLite
    with sqlite3.connect(temp_memory.db_path) as conn:
        cursor = conn.cursor()
        cursor.execute("SELECT weight, context_snippet FROM memory_edges WHERE source_id='ticker:antm' AND target_id='facility:smelter_haltim'")
        row = cursor.fetchone()
        assert row is not None
        assert row[0] == 1.5  # 1.0 + 0.5
        assert row[1] == "Uji coba komisioning"


def test_record_investigation_deterministic(temp_memory):
    """Test deterministic recording of investigation pipeline findings (Law 1, Law 6)."""
    class MockAnomaly:
        z_score = 3.84
        metric_type = "VOLUME"
        date = "2026-09-12"
        description = "Lonjakan volume 3.84 sigma di atas MA20"

    findings = [
        {
            "id": "FIND-01",
            "title": "Peresmian Pabrik Haltim",
            "claim_text": "Keterbukaan BEI mengenai operasi komersial",
            "verification_status": "SUPPORTED",
            "confidence_score": 0.95,
            "causality_status": "LIKELY_CATALYST",
        }
    ]

    rec = temp_memory.record_investigation(
        session_id="INV-2026-0042",
        ticker="ANTM",
        anomalies=[MockAnomaly()],
        findings=findings,
        sector="Basic Materials",
    )
    assert rec["status"] == "RECORDED"
    assert rec["ticker"] == "ANTM"
    assert rec["recorded_edges"] >= 4  # User->ANTM, ANTM->Sector, ANTM->Anomaly, ANTM->Finding

    # Verify nodes in NetworkX
    G = temp_memory.load_graph()
    assert "ticker:antm" in G
    assert "user:default" in G
    assert "sector:basic_materials" in G
    assert G.has_edge("user:default", "ticker:antm")
    assert G.has_edge("ticker:antm", "sector:basic_materials")


def test_ego_subgraph_retrieval_and_decay(temp_memory):
    """Test 2-hop ego-graph traversal with recency decay."""
    # Seed chain: User -> ANTM -> Smelter Haltim -> Barito Group
    temp_memory.store_observation("User", "WATCHES", "ANTM", "USER", "TICKER")
    temp_memory.store_observation("ANTM", "OPERATES", "Smelter Haltim", "TICKER", "FACILITY")
    temp_memory.store_observation("Smelter Haltim", "SUPPLIES", "Barito Group", "FACILITY", "CONGLOMERATE")

    # Retrieve ego subgraph around ANTM (radius=1 should include User & Smelter Haltim)
    ego1 = temp_memory.retrieve_ego_subgraph("ANTM", radius=1)
    labels1 = {n["label"] for n in ego1["nodes"]}
    assert "ANTM" in labels1
    assert "User" in labels1
    assert "Smelter Haltim" in labels1
    assert "Barito Group" not in labels1  # 2 hops away

    # Retrieve ego subgraph around ANTM (radius=2 should include Barito Group)
    ego2 = temp_memory.retrieve_ego_subgraph("ANTM", radius=2)
    labels2 = {n["label"] for n in ego2["nodes"]}
    assert "Barito Group" in labels2

    # Check edges and recency decay properties
    assert len(ego2["edges"]) >= 3
    edge = ego2["edges"][0]
    assert "effective_weight" in edge
    assert "decay_factor" in edge
    assert "days_ago" in edge
    assert edge["effective_weight"] <= edge["weight"]


def test_format_investigative_prompt(temp_memory):
    """Test rendering of <investigative_memory> XML block."""
    temp_memory.store_observation(
        source_label="User",
        source_type="USER",
        relation="HOLDS_AT",
        target_label="Price: 1450",
        target_type="PRICE_LEVEL",
        context_snippet="Posisi modal di ANTM",
    )
    temp_memory.store_observation(
        source_label="Price: 1450",
        source_type="PRICE_LEVEL",
        relation="TICKER_REF",
        target_label="ANTM",
        target_type="TICKER",
    )

    prompt_xml = temp_memory.format_investigative_prompt("ANTM", radius=2)
    assert "<investigative_memory>" in prompt_xml
    assert "</investigative_memory>" in prompt_xml
    assert "ANTM" in prompt_xml
    assert "Price: 1450" in prompt_xml


def test_find_shortest_path(temp_memory):
    """Test shortest path discovery between two seemingly distant entities."""
    # Build a path: BBCA -> GOTO -> Telkomsel -> TLKM
    temp_memory.store_observation("BBCA", "PARTNERS_WITH", "GOTO", "TICKER", "TICKER")
    temp_memory.store_observation("GOTO", "CO_INVESTS", "Telkomsel", "TICKER", "ENTITY")
    temp_memory.store_observation("TLKM", "SUBSIDIARY", "Telkomsel", "TICKER", "ENTITY")

    path = temp_memory.find_shortest_path("BBCA", "TLKM")
    assert path is not None
    assert len(path) == 4
    labels = [step["label"] for step in path]
    assert labels[0] == "BBCA"
    assert labels[-1] == "TLKM"


def test_graph_stats_and_centrality(temp_memory):
    """Test topological stats and top central entity calculation."""
    # Create hub: ANTM connected to 3 entities
    temp_memory.store_observation("User", "INVESTIGATED", "ANTM", "USER", "TICKER")
    temp_memory.store_observation("ANTM", "OPERATES", "Smelter A", "TICKER", "FACILITY")
    temp_memory.store_observation("ANTM", "OPERATES", "Smelter B", "TICKER", "FACILITY")

    stats = temp_memory.get_graph_stats()
    assert stats["total_nodes"] == 4
    assert stats["total_edges"] == 3
    assert len(stats["top_central_entities"]) > 0
    # ANTM should have the highest connections (degree 3)
    assert stats["top_central_entities"][0]["label"] == "ANTM"
    assert stats["top_central_entities"][0]["connections"] == 3


def test_clear_memory(temp_memory):
    """Test clearing memory selectively by session and completely."""
    temp_memory.store_observation("User", "INVESTIGATED", "ANTM", source_type="USER", target_type="TICKER", session_id="S1")
    temp_memory.store_observation("User", "INVESTIGATED", "BBCA", source_type="USER", target_type="TICKER", session_id="S2")

    # Clear S1 only
    del1 = temp_memory.clear_memory(session_id="S1")
    assert del1 == 1

    G1 = temp_memory.load_graph()
    assert "ticker:antm" not in G1
    assert "ticker:bbca" in G1

    # Clear all
    del2 = temp_memory.clear_memory()
    assert del2 == 1
    G2 = temp_memory.load_graph()
    assert G2.number_of_nodes() == 0


def test_resolve_target_nodes_with_company_aliases(temp_memory):
    """Test resolving graph nodes via company name aliases and normalized tokens."""
    temp_memory.store_observation(
        source_label="User",
        source_type="USER",
        relation="INVESTIGATED",
        target_label="ANTM",
        target_type="TICKER",
        target_metadata={"company_name": "Aneka Tambang", "sector": "Basic Materials"},
    )
    # Search by Indonesian common name and lowercase company name
    nodes_common = temp_memory.retrieve_ego_subgraph("Aneka Tambang")
    assert len(nodes_common["root_nodes"]) > 0
    assert "ticker:antm" in nodes_common["root_nodes"]

    nodes_fuzzy = temp_memory.retrieve_ego_subgraph("PT Antam")
    assert "ticker:antm" in nodes_fuzzy["root_nodes"]


def test_graph_centrality_pagerank_and_bridges(temp_memory):
    """Test PageRank and betweenness centrality to detect true market bridges."""
    # Topology: User -> ANTM, INCO -> MIND ID -> PTBA -> Coal Fleet
    temp_memory.store_observation("User", "INVESTIGATED", "ANTM", source_type="USER", target_type="TICKER")
    temp_memory.store_observation("User", "INVESTIGATED", "INCO", source_type="USER", target_type="TICKER")
    temp_memory.store_observation("ANTM", "SUBSIDIARY_OF", "MIND ID", source_type="TICKER", target_type="ENTITY")
    temp_memory.store_observation("INCO", "ASSOCIATE_OF", "MIND ID", source_type="TICKER", target_type="ENTITY")
    temp_memory.store_observation("MIND ID", "CONTROLS", "PTBA", source_type="ENTITY", target_type="TICKER")
    temp_memory.store_observation("PTBA", "OPERATES", "Coal Fleet", source_type="TICKER", target_type="FACILITY")

    stats = temp_memory.get_graph_stats()
    assert "top_central_entities" in stats
    top_entity = stats["top_central_entities"][0]
    assert "pagerank" in top_entity
    assert "betweenness" in top_entity
    assert "composite_score" in top_entity
    assert top_entity["pagerank"] > 0.0
    labels = [e["label"] for e in stats["top_central_entities"]]
    assert "MIND ID" in labels[:2]


def test_supersedes_relation_invalidates_prior_fact(temp_memory):
    """Test SUPERSEDES relationship invalidating prior facts during retrieval."""
    # Sesi 1: Buy ANTM 1450
    temp_memory.store_observation("User", "HOLDS_AT", "Price: 1450", source_type="USER", target_type="PRICE_LEVEL", session_id="S1")
    # Sesi 2: Take Profit ANTM 1620 superseding 1450
    temp_memory.store_observation("User", "HOLDS_AT", "Price: 1620", source_type="USER", target_type="PRICE_LEVEL", session_id="S2")
    temp_memory.store_observation("Price: 1620", "SUPERSEDES", "Price: 1450", source_type="PRICE_LEVEL", target_type="PRICE_LEVEL", session_id="S2")

    ego = temp_memory.retrieve_ego_subgraph("User", radius=2)
    active_edges = [e for e in ego["edges"] if not e.get("is_superseded")]
    superseded_edges = [e for e in ego["edges"] if e.get("is_superseded")]

    assert len(superseded_edges) >= 1
    assert any(e["target_label"] == "Price: 1450" for e in superseded_edges)
    assert any(e["target_label"] == "Price: 1620" for e in active_edges)

    # In formatted prompt, superseded edges must be excluded from active facts
    prompt = temp_memory.format_investigative_prompt("User")
    assert "Price: 1620" in prompt
    assert "Price: 1450" not in prompt



