# Niskava Memory Graph & Institutional Visualizer Refactor Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Upgrade Niskava Agent's memory graph into an institutional-grade knowledge engine with semantic alias resolution, PageRank/betweenness centrality, temporal `SUPERSEDES` logic, and a completely rebranded, clean financial research terminal visualizer (removing all neon cyber tropes).

**Architecture:** Refactor `LocalGraphMemory` in `backend/engine/memory/graph_memory.py` to support multi-faceted entity resolution, network centrality algorithms (`networkx.pagerank`, `betweenness_centrality`), and stateful conflict resolution. Re-architect `visualizer.py` into a modular template engine providing an institutional 3-column workspace (Entity Finder, Directed Entity Tree, Evidence Dossier) styled with refined obsidian slate palettes, tabular figures, and official financial taxonomy.

**Tech Stack:** Python 3.11+, NetworkX 3.2+, SQLite WAL (`sqlite3`), Vis.js Network, Pytest 8+, Go 1.24+ Core.

## Global Constraints

- **Compliance with AGENTS.md (Law 1, 4, 6):** All quant/metrics remain deterministic before LLM; persistence strictly in local SQLite (`~/.niskava/niskava.db`); graph traversal remains local-first without external paid graph/vector DBs.
- **Strict Anti-Cyber Clean Aesthetic:** Absolutely no neon cyan (`#00D2FF`), neon magenta (`#FF0055`), lightning emojis (`⚡`), or `[OSINT]` hacker badges. The UI must match institutional financial research terminals (Bloomberg, Koyfin, Linear).
- **Financial Taxonomy Standard:** Replace cyber jargon ("anomali z-score", "cyber catalyst", "bandarmology forensic") with standard financial market terms ("unusual volume outlier", "material disclosure", "institutional broker flow").
- **Clean Code & Quality Invariants:** Strict Python type hints (`typing`), immutable typed dicts or dataclasses, comprehensive docstrings, modular functions (<50 lines where feasible), and 100% test coverage with TDD.
- **Git Branch Protection:** All commits must be made on branch `dev`. Direct pushes to `main` are prohibited.

---

### Task 1: Scaffolding & Pytest Discovery Configuration

**Files:**
- Modify: `backend/engine/pyproject.toml`
- Test: `backend/engine/tests/test_quant.py`

**Interfaces:**
- Consumes: Existing pytest test suite
- Produces: Seamless execution of `pytest` from any directory without requiring manual `PYTHONPATH=backend`

- [x] **Step 1: Write a reproduction command to verify the collection error**

Run: `.venv/bin/pytest backend/engine/tests/test_evals.py` from repository root.
Expected: FAIL with `ModuleNotFoundError: No module named 'engine'`

- [x] **Step 2: Update `backend/engine/pyproject.toml` with pytest configuration**

Add `[tool.pytest.ini_options]` with `pythonpath = ["backend"]` and testpaths.
```toml
[tool.pytest.ini_options]
minversion = "8.0"
testpaths = ["backend/engine/tests"]
pythonpath = ["backend"]
addopts = "-v --strict-markers"
```

- [x] **Step 3: Run bare pytest to verify discovery works from root**

Run: `.venv/bin/pytest backend/engine/tests/test_evals.py`
Expected: PASS (1 passed)

- [x] **Step 4: Commit configuration**

```bash
git add backend/engine/pyproject.toml
git commit -m "chore(build): configure pytest pythonpath for root test discovery"
```

---

### Task 2: Advanced Entity Resolution & Alias Normalization

**Files:**
- Modify: `backend/engine/memory/graph_memory.py`
- Test: `backend/engine/tests/test_graph_memory.py`

**Interfaces:**
- Consumes: Entity queries (strings such as `"Antam"`, `"PT Aneka Tambang Tbk"`, `"BBRI"`, `"Bank Rakyat Indonesia"`)
- Produces: `LocalGraphMemory._resolve_target_nodes(G: nx.DiGraph, query: str) -> List[str]` returning canonical node IDs with alias resolution and token-level fuzzy overlap

- [x] **Step 1: Write failing unit test for alias and synonym resolution**

Add `test_resolve_target_nodes_with_company_aliases` to `backend/engine/tests/test_graph_memory.py`:
```python
def test_resolve_target_nodes_with_company_aliases(mem):
    # Store with formal ticker
    mem.store_observation(
        source_label="User",
        source_type="USER",
        relation="INVESTIGATED",
        target_label="ANTM",
        target_type="TICKER",
        target_metadata={"company_name": "Aneka Tambang", "sector": "Basic Materials"},
    )
    # Search by Indonesian common name and lowercase company name
    nodes_common = mem.retrieve_ego_subgraph("Aneka Tambang")
    assert len(nodes_common["root_nodes"]) > 0
    assert "ticker:antm" in nodes_common["root_nodes"]

    nodes_fuzzy = mem.retrieve_ego_subgraph("PT Antam")
    assert "ticker:antm" in nodes_fuzzy["root_nodes"]
```

- [x] **Step 2: Run test to verify it fails**

Run: `PYTHONPATH=backend .venv/bin/pytest backend/engine/tests/test_graph_memory.py -k "test_resolve_target_nodes_with_company_aliases"`
Expected: FAIL with `AssertionError: assert 'ticker:antm' in []`

- [x] **Step 3: Implement alias dictionary and normalized token matcher in `_resolve_target_nodes`**

In `backend/engine/memory/graph_memory.py`:
- Strip corporate legal prefixes (`PT`, `Tbk`, `Persero`).
- Check `metadata.get("company_name")`, `metadata.get("aliases")`, and common IDX ticker stems.
- Score candidate matches (Exact Ticker: 1.0 > Exact Label: 0.9 > Token Jaccard > Substring).

```python
@staticmethod
def clean_corporate_tokens(text: str) -> str:
    cleaned = re.sub(r"\b(pt|tbk|persero|corp|corporation|inc)\b", "", text.lower())
    return re.sub(r"\s+", " ", cleaned).strip()
```

- [x] **Step 4: Run test to verify it passes**

Run: `PYTHONPATH=backend .venv/bin/pytest backend/engine/tests/test_graph_memory.py -k "test_resolve_target_nodes_with_company_aliases"`
Expected: PASS

- [x] **Step 5: Commit**

```bash
git add backend/engine/memory/graph_memory.py backend/engine/tests/test_graph_memory.py
git commit -m "feat(memory): implement semantic corporate alias resolution for graph nodes"
```

---

### Task 3: Graph Intelligence — PageRank & Betweenness Centrality

**Files:**
- Modify: `backend/engine/memory/graph_memory.py`
- Test: `backend/engine/tests/test_graph_memory.py`

**Interfaces:**
- Consumes: In-memory `NetworkX.DiGraph`
- Produces: `LocalGraphMemory.get_graph_stats() -> Dict[str, Any]` returning `top_central_entities` ranked by PageRank with betweenness bridge identification

- [x] **Step 1: Write failing unit test for PageRank and betweenness bridge detection**

Add `test_graph_centrality_pagerank_and_bridges` to `backend/engine/tests/test_graph_memory.py`:
```python
def test_graph_centrality_pagerank_and_bridges(mem):
    # Construct a hub & bridge topology: User -> Cluster 1 (ANTM, INCO) -> Holding (MIND ID) -> Cluster 2 (PTBA)
    mem.store_observation("User", "INVESTIGATED", "ANTM", source_type="USER", target_type="TICKER")
    mem.store_observation("User", "INVESTIGATED", "INCO", source_type="USER", target_type="TICKER")
    mem.store_observation("ANTM", "SUBSIDIARY_OF", "MIND ID", source_type="TICKER", target_type="ENTITY")
    mem.store_observation("INCO", "ASSOCIATE_OF", "MIND ID", source_type="TICKER", target_type="ENTITY")
    mem.store_observation("PTBA", "SUBSIDIARY_OF", "MIND ID", source_type="TICKER", target_type="ENTITY")

    stats = mem.get_graph_stats()
    assert "top_central_entities" in stats
    top_entity = stats["top_central_entities"][0]
    assert top_entity["label"] == "MIND ID"
    assert "pagerank" in top_entity
    assert "betweenness" in top_entity
```

- [x] **Step 2: Run test to verify it fails**

Run: `PYTHONPATH=backend .venv/bin/pytest backend/engine/tests/test_graph_memory.py -k "test_graph_centrality_pagerank_and_bridges"`
Expected: FAIL (missing `pagerank` or wrong top entity)

- [x] **Step 3: Implement PageRank and Betweenness Centrality in `get_graph_stats`**

In `backend/engine/memory/graph_memory.py`:
- Use `nx.pagerank(G, alpha=0.85, weight="weight")` for authority flow.
- Use `nx.betweenness_centrality(G, normalized=True)` for bridge entities.
- Combine metrics into a composite centrality score: `composite = 0.6 * pr + 0.4 * bet`.

```python
pageranks = nx.pagerank(G, alpha=0.85, weight="weight") if num_nodes > 1 else {n: 1.0 for n in G.nodes()}
betweenness = nx.betweenness_centrality(G, normalized=True) if num_nodes > 2 else {n: 0.0 for n in G.nodes()}
```

- [x] **Step 4: Run test to verify it passes**

Run: `PYTHONPATH=backend .venv/bin/pytest backend/engine/tests/test_graph_memory.py -k "test_graph_centrality_pagerank_and_bridges"`
Expected: PASS

- [x] **Step 5: Commit**

```bash
git add backend/engine/memory/graph_memory.py backend/engine/tests/test_graph_memory.py
git commit -m "feat(memory): upgrade graph centrality to PageRank and betweenness bridge metrics"
```

---

### Task 4: Temporal `SUPERSEDES` Fact Resolution

**Files:**
- Modify: `backend/engine/memory/graph_memory.py`
- Test: `backend/engine/tests/test_graph_memory.py`

**Interfaces:**
- Consumes: Conflicting or evolved facts across sessions (`SUPERSEDES` relation)
- Produces: `retrieve_ego_subgraph()` and `format_investigative_prompt()` pruning or tagging superseded edges so agents only infer from active facts

- [x] **Step 1: Write failing unit test for `SUPERSEDES` edge invalidation**

Add `test_supersedes_relation_invalidates_prior_fact` to `backend/engine/tests/test_graph_memory.py`:
```python
def test_supersedes_relation_invalidates_prior_fact(mem):
    # Sesi 1: Buy ANTM 1450
    mem.store_observation("User", "HOLDS_AT", "Price: 1450", source_type="USER", target_type="PRICE_LEVEL", session_id="S1")
    # Sesi 2: Take Profit ANTM 1620 superseding 1450
    mem.store_observation("User", "HOLDS_AT", "Price: 1620", source_type="USER", target_type="PRICE_LEVEL", session_id="S2")
    mem.store_observation("Price: 1620", "SUPERSEDES", "Price: 1450", source_type="PRICE_LEVEL", target_type="PRICE_LEVEL", session_id="S2")

    ego = mem.retrieve_ego_subgraph("User", radius=2)
    active_edges = [e for e in ego["edges"] if not e.get("is_superseded")]
    superseded_edges = [e for e in ego["edges"] if e.get("is_superseded")]

    assert len(superseded_edges) >= 1
    assert any(e["target_label"] == "Price: 1450" for e in superseded_edges)
    assert any(e["target_label"] == "Price: 1620" for e in active_edges)
```

- [x] **Step 2: Run test to verify it fails**

Run: `PYTHONPATH=backend .venv/bin/pytest backend/engine/tests/test_graph_memory.py -k "test_supersedes_relation_invalidates_prior_fact"`
Expected: FAIL (`is_superseded` key missing or false)

- [x] **Step 3: Implement `SUPERSEDES` resolution in `retrieve_ego_subgraph`**

In `backend/engine/memory/graph_memory.py`:
- Identify all edges with `relation == "SUPERSEDES"`.
- Traverse targets of `SUPERSEDES` and flag inbound/outbound prior fact edges with `is_superseded = True` and set their `effective_weight = 0.0`.
- In `format_investigative_prompt()`, exclude superseded edges from LLM prompt body.

- [x] **Step 4: Run test to verify it passes**

Run: `PYTHONPATH=backend .venv/bin/pytest backend/engine/tests/test_graph_memory.py -k "test_supersedes_relation_invalidates_prior_fact"`
Expected: PASS

- [x] **Step 5: Commit**

```bash
git add backend/engine/memory/graph_memory.py backend/engine/tests/test_graph_memory.py
git commit -m "feat(memory): implement temporal SUPERSEDES edge pruning and state resolution"
```

---

### Task 5: Conversational Dialogue Turn Ingestion

**Files:**
- Create: `backend/engine/memory/extractor.py`
- Modify: `backend/engine/agent/react_agent.py`
- Test: `backend/engine/tests/test_memory_extractor.py`

**Interfaces:**
- Consumes: Raw user message string in chat (e.g. `"Saya beli ANTM di 1450 untuk swing trading"`)
- Produces: `extract_dialogue_observations(user_text: str) -> List[Dict[str, Any]]` returning structured entity triples automatically saved to `LocalGraphMemory`

- [x] **Step 1: Write failing test for user dialogue extractor**

Create `backend/engine/tests/test_memory_extractor.py`:
```python
import pytest
from engine.memory.extractor import extract_dialogue_observations

def test_extract_user_buy_intent():
    text = "Saya baru beli BBCA di 9800 kemarin"
    observations = extract_dialogue_observations(text)
    assert len(observations) >= 1
    obs = observations[0]
    assert obs["ticker"] == "BBCA"
    assert obs["relation"] == "HOLDS_AT"
    assert "9800" in obs["target_label"]
```

- [x] **Step 2: Run test to verify it fails**

Run: `PYTHONPATH=backend .venv/bin/pytest backend/engine/tests/test_memory_extractor.py`
Expected: FAIL with `ModuleNotFoundError: No module named 'engine.memory.extractor'`

- [x] **Step 3: Implement deterministic dialogue extractor in `backend/engine/memory/extractor.py`**

Create `extractor.py`:
- Regex patterns for stock price entries: `(beli|masuk|buy|entry)\s+([A-Z]{4})\s+(di|pada|level)?\s*(\d+)`.
- Regex patterns for stock watchlists: `(pantau|watchlist|awasi)\s+([A-Z]{4})`.
- Return normalized triplestores `{"source_label": "User", "relation": "...", "target_label": "..."}`.

- [x] **Step 4: Integrate extractor into `react_agent.py` before ReAct loop**

In `backend/engine/agent/react_agent.py`:
- In `run_prompt_session()` / `chat()`, parse incoming user prompt with `extract_dialogue_observations()`.
- Automatically invoke `self.tools.memory.store_observation()` for any extracted items.

- [x] **Step 5: Run tests to verify it passes**

Run: `PYTHONPATH=backend .venv/bin/pytest backend/engine/tests/test_memory_extractor.py`
Expected: PASS

- [x] **Step 6: Commit**

```bash
git add backend/engine/memory/extractor.py backend/engine/agent/react_agent.py backend/engine/tests/test_memory_extractor.py
git commit -m "feat(memory): add automatic conversational dialogue triple extractor"
```

---

### Task 6: Visualizer Rebrand — Institutional Financial Dossier

**Files:**
- Modify: `backend/engine/memory/visualizer.py`
- Test: `backend/engine/tests/test_visualizer.py`

**Interfaces:**
- Consumes: Graph data from `LocalGraphMemory`
- Produces: Clean, modern, institutional 3-column HTML dashboard (Entity Finder, Directed Entity Tree, Evidence Dossier) styled with Obsidian Slate and official financial taxonomy.

- [x] **Step 1: Write test for institutional styling and taxonomy in `test_visualizer.py`**

Add tests ensuring cyber terms and neon colors are absent:
```python
def test_institutional_theme_no_cyber_slop(memory_with_data):
    viz = CyberOSINTVisualizer(memory=memory_with_data) # will be aliased/renamed to MarketIntelligenceVisualizer
    html = viz.generate_html(session_id="S1")
    # Assert absence of cyber slop
    assert "⚡" not in html
    assert "CYBER" not in html.upper()
    assert "#00D2FF" not in html # No neon cyan
    assert "#FF0055" not in html # No neon hot pink
    # Assert presence of institutional terms and palette
    assert "Market Intelligence" in html
    assert "Unusual Volume Outliers" in html or "Volume Outlier" in html
    assert "Material Disclosures" in html or "Corporate Actions" in html
    assert "font-feature-settings: 'tnum'" in html or "tabular-nums" in html
```

- [x] **Step 2: Run test to verify it fails**

Run: `PYTHONPATH=backend .venv/bin/pytest backend/engine/tests/test_visualizer.py -k "test_institutional_theme_no_cyber_slop"`
Expected: FAIL with assertion errors on cyber tropes.

- [x] **Step 3: Redesign `visualizer.py` with Institutional Architecture**

In `backend/engine/memory/visualizer.py`:
1. **Palette Update:**
   - Background: Deep Obsidian `#0B0E14`, Surface `#121620`, Border `#1F2637`.
   - Primary Accent: Slate Sapphire `#2563EB` (for Tickers), Emerald `#059669` (for Material Disclosures), Soft Amber `#D97706` (for Volume Outliers), Slate Indigo `#6366F1` (for Sectors).
2. **Typography & Layout:**
   - Inter / Geist font family, monospace tabular figures `tnum` for prices and dates.
   - Clean 3-column research terminal layout:
     - Left: **Entity Navigator** (Search, Sector filters, 1-Hop / 2-Hop toggle).
     - Center: **Network Canvas** (Vis.js with hierarchical or stabilized BarnesHut physics, rectangular micro-card nodes, clean directed arrows).
     - Right: **Evidence Dossier** (Comprehensive entity metadata, IDX disclosure citations, timestamp recency, verification status badge).
3. **Taxonomy & Rebranding:**
   - Rebrand title from "NISKAVA GRAPH OSINT" to **"NISKAVA AGENT — Market Intelligence & Entity Network"**.
   - Node styling: replace star/triangle shapes with rounded rectangle boxes (`shape: 'box'`, `margin: 10`, `borderWidth: 1`).

- [x] **Step 4: Run tests to verify it passes**

Run: `PYTHONPATH=backend .venv/bin/pytest backend/engine/tests/test_visualizer.py`
Expected: PASS (all tests in suite pass)

- [x] **Step 5: Commit**

```bash
git add backend/engine/memory/visualizer.py backend/engine/tests/test_visualizer.py
git commit -m "feat(visualizer): rebrand graph visualizer into institutional market research terminal"
```

---

### Task 7: Quality Gate, Build & End-to-End Verification

**Files:**
- Verify: Full codebase (`backend/core/...`, `backend/engine/...`, `cmd/niskava/...`)

- [x] **Step 1: Execute complete Go test suite with race detector**

Run: `go test -v -race ./backend/core/... ./clients/cli/...`
Expected: PASS with 0 race conditions.

- [x] **Step 2: Execute full Python test suite**

Run: `PYTHONPATH=backend .venv/bin/pytest backend/engine/tests/ -v`
Expected: PASS (all 95+ tests passing).

- [x] **Step 3: Verify code formatting and linting**

Run: `make lint`
Expected: PASS (clean `gofmt`, `go vet`, and python compilation).

- [x] **Step 4: Build binary and test export command**

Run:
```bash
make build
./bin/niskava graph --help
./bin/niskava mcp <<< '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
```
Expected: Clean binary build and successful MCP tools listing.

- [x] **Step 5: Final review commit**

```bash
git commit --allow-empty -m "chore(release): verify quality gates for memory graph & institutional visualizer"
```
