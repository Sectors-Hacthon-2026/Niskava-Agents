"""Local Conversational & Cross-Session Graph Memory Engine (NetworkX + SQLite).

Complies strictly with:
- Law 1: Deterministic Before Generative (Deterministic post-investigation recording)
- Law 4: Local-First Data Sovereignty (Local SQLite persistence)
- Law 6: Local Conversational Graph Memory Engine (Ego-Graph traversal with exponential recency decay)
"""

from datetime import datetime, timezone
import json
import math
import os
import re
import sqlite3
from typing import Any, Dict, List, Optional

import networkx as nx


KNOWN_TICKER_ALIASES: Dict[str, str] = {
    "antam": "ANTM",
    "aneka tambang": "ANTM",
    "bca": "BBCA",
    "bank central asia": "BBCA",
    "bri": "BBRI",
    "bank rakyat indonesia": "BBRI",
    "mandiri": "BMRI",
    "bank mandiri": "BMRI",
    "bni": "BBNI",
    "bank negara indonesia": "BBNI",
    "telkom": "TLKM",
    "telkom indonesia": "TLKM",
    "adaro": "ADRO",
    "adaro energy": "ADRO",
    "vale": "INCO",
    "vale indonesia": "INCO",
    "bukit asam": "PTBA",
    "timah": "TINS",
    "medco": "MEDC",
    "bumi resources": "BUMI",
}


class LocalGraphMemory:
    """In-memory NetworkX graph backed by local SQLite persistence."""

    def __init__(
        self,
        db_path: str = "~/.niskava/niskava.db",
        lambda_decay: float = 0.05,
        max_radius: int = 2,
    ):
        self.db_path = os.path.expanduser(db_path)
        self.lambda_decay = lambda_decay
        self.max_radius = max_radius
        self._ensure_schema()

    def _get_connection(self) -> sqlite3.Connection:
        """Create a connection with WAL mode and foreign keys enabled."""
        os.makedirs(os.path.dirname(os.path.abspath(self.db_path)), exist_ok=True)
        conn = sqlite3.connect(self.db_path)
        conn.row_factory = sqlite3.Row
        conn.execute("PRAGMA journal_mode = WAL;")
        conn.execute("PRAGMA foreign_keys = ON;")
        return conn

    def _ensure_schema(self) -> None:
        """Self-healing Auto-DDL for memory_nodes and memory_edges."""
        with self._get_connection() as conn:
            conn.execute("""
                CREATE TABLE IF NOT EXISTS memory_nodes (
                    id TEXT PRIMARY KEY,
                    label TEXT NOT NULL,
                    node_type TEXT NOT NULL,
                    metadata_json TEXT DEFAULT '{}',
                    last_observed_at TEXT NOT NULL,
                    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
                );
            """)
            conn.execute("""
                CREATE TABLE IF NOT EXISTS memory_edges (
                    source_id TEXT NOT NULL REFERENCES memory_nodes(id) ON DELETE CASCADE,
                    target_id TEXT NOT NULL REFERENCES memory_nodes(id) ON DELETE CASCADE,
                    relation TEXT NOT NULL,
                    context_snippet TEXT,
                    session_id TEXT,
                    weight REAL NOT NULL DEFAULT 1.0,
                    confidence_score REAL NOT NULL DEFAULT 1.0,
                    last_observed_at TEXT NOT NULL,
                    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
                    PRIMARY KEY (source_id, target_id, relation)
                );
            """)
            conn.execute(
                "CREATE INDEX IF NOT EXISTS idx_mem_edges_src ON memory_edges(source_id);"
            )
            conn.execute(
                "CREATE INDEX IF NOT EXISTS idx_mem_edges_tgt ON memory_edges(target_id);"
            )
            conn.execute(
                "CREATE INDEX IF NOT EXISTS idx_mem_edges_session ON memory_edges(session_id);"
            )
            # Self-healing column migration for confidence_score
            try:
                conn.execute("ALTER TABLE memory_edges ADD COLUMN confidence_score REAL NOT NULL DEFAULT 1.0;")
            except sqlite3.OperationalError:
                pass

            # Self-healing table migration to decouple session_id from investigations FK
            try:
                cursor = conn.cursor()
                fks = cursor.execute("PRAGMA foreign_key_list(memory_edges)").fetchall()
                has_inv_fk = any(fk["table"] == "investigations" for fk in fks)
                if has_inv_fk:
                    conn.execute("PRAGMA foreign_keys = OFF;")
                    conn.execute("""
                        CREATE TABLE IF NOT EXISTS memory_edges_v2 (
                            source_id TEXT NOT NULL REFERENCES memory_nodes(id) ON DELETE CASCADE,
                            target_id TEXT NOT NULL REFERENCES memory_nodes(id) ON DELETE CASCADE,
                            relation TEXT NOT NULL,
                            context_snippet TEXT,
                            session_id TEXT,
                            weight REAL NOT NULL DEFAULT 1.0,
                            confidence_score REAL NOT NULL DEFAULT 1.0,
                            last_observed_at TEXT NOT NULL,
                            created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
                            PRIMARY KEY (source_id, target_id, relation)
                        );
                    """)
                    conn.execute("""
                        INSERT OR IGNORE INTO memory_edges_v2 
                        SELECT source_id, target_id, relation, context_snippet, session_id, weight, confidence_score, last_observed_at, created_at 
                        FROM memory_edges;
                    """)
                    conn.execute("DROP TABLE memory_edges;")
                    conn.execute("ALTER TABLE memory_edges_v2 RENAME TO memory_edges;")
                    conn.execute("CREATE INDEX IF NOT EXISTS idx_mem_edges_src ON memory_edges(source_id);")
                    conn.execute("CREATE INDEX IF NOT EXISTS idx_mem_edges_tgt ON memory_edges(target_id);")
                    conn.execute("CREATE INDEX IF NOT EXISTS idx_mem_edges_session ON memory_edges(session_id);")
                    conn.execute("PRAGMA foreign_keys = ON;")
            except Exception:
                pass

            conn.commit()

    def load_graph(self) -> nx.DiGraph:
        """Load active nodes and directed edges from SQLite into a NetworkX DiGraph."""
        G = nx.DiGraph()
        with self._get_connection() as conn:
            cursor = conn.cursor()
            cursor.execute("SELECT id, label, node_type, metadata_json, last_observed_at FROM memory_nodes")
            for row in cursor.fetchall():
                meta = {}
                if row["metadata_json"]:
                    try:
                        meta = json.loads(row["metadata_json"])
                    except Exception:
                        meta = {}
                G.add_node(
                    row["id"],
                    label=row["label"],
                    node_type=row["node_type"],
                    last_observed_at=row["last_observed_at"],
                    metadata=meta,
                )

            cursor.execute(
                """
                SELECT source_id, target_id, relation, context_snippet, session_id,
                       weight, confidence_score, last_observed_at
                FROM memory_edges
                """
            )
            for row in cursor.fetchall():
                G.add_edge(
                    row["source_id"],
                    row["target_id"],
                    relation=row["relation"],
                    context_snippet=row["context_snippet"] or "",
                    session_id=row["session_id"] or "",
                    weight=float(row["weight"]),
                    confidence_score=float(row["confidence_score"]),
                    last_observed_at=row["last_observed_at"],
                )
        return G

    @staticmethod
    def normalize_node_id(node_type: str, label: str) -> str:
        """Create consistent canonical node IDs (e.g. ticker:antm, person:budi_santoso)."""
        clean_type = node_type.strip().lower()
        if clean_type == "user":
            return "user:default"
        clean_label = re.sub(r"[^a-zA-Z0-9_-]", "_", label.strip().lower())
        clean_label = re.sub(r"_+", "_", clean_label).strip("_")
        return f"{clean_type}:{clean_label}"

    def store_observation(
        self,
        source_label: str,
        relation: str,
        target_label: str,
        source_type: str = "TICKER",
        target_type: str = "ENTITY",
        context_snippet: str = "",
        session_id: Optional[str] = None,
        confidence_score: float = 1.0,
        source_metadata: Optional[Dict[str, Any]] = None,
        target_metadata: Optional[Dict[str, Any]] = None,
    ) -> Dict[str, Any]:
        """Upsert nodes and an edge into SQLite memory graph, boosting weight on repeats."""
        src_id = self.normalize_node_id(source_type, source_label)
        tgt_id = self.normalize_node_id(target_type, target_label)
        rel = relation.strip().upper()
        now_iso = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")

        src_meta = json.dumps(source_metadata or {}, ensure_ascii=False)
        tgt_meta = json.dumps(target_metadata or {}, ensure_ascii=False)

        with self._get_connection() as conn:
            cursor = conn.cursor()
            # Upsert source node
            cursor.execute(
                """
                INSERT INTO memory_nodes (id, label, node_type, metadata_json, last_observed_at)
                VALUES (?, ?, ?, ?, ?)
                ON CONFLICT(id) DO UPDATE SET
                    label = excluded.label,
                    node_type = excluded.node_type,
                    last_observed_at = excluded.last_observed_at,
                    metadata_json = CASE
                        WHEN excluded.metadata_json != '{}' THEN excluded.metadata_json
                        ELSE memory_nodes.metadata_json
                    END
                """,
                (src_id, source_label.strip(), source_type.strip().upper(), src_meta, now_iso),
            )

            # Upsert target node
            cursor.execute(
                """
                INSERT INTO memory_nodes (id, label, node_type, metadata_json, last_observed_at)
                VALUES (?, ?, ?, ?, ?)
                ON CONFLICT(id) DO UPDATE SET
                    label = excluded.label,
                    node_type = excluded.node_type,
                    last_observed_at = excluded.last_observed_at,
                    metadata_json = CASE
                        WHEN excluded.metadata_json != '{}' THEN excluded.metadata_json
                        ELSE memory_nodes.metadata_json
                    END
                """,
                (tgt_id, target_label.strip(), target_type.strip().upper(), tgt_meta, now_iso),
            )

            # Upsert edge with weight increment
            cursor.execute(
                """
                INSERT INTO memory_edges 
                (source_id, target_id, relation, context_snippet, session_id, weight, confidence_score, last_observed_at)
                VALUES (?, ?, ?, ?, ?, 1.0, ?, ?)
                ON CONFLICT(source_id, target_id, relation) DO UPDATE SET
                    context_snippet = CASE
                        WHEN excluded.context_snippet != '' THEN excluded.context_snippet
                        ELSE memory_edges.context_snippet
                    END,
                    session_id = COALESCE(excluded.session_id, memory_edges.session_id),
                    weight = memory_edges.weight + 0.5,
                    confidence_score = excluded.confidence_score,
                    last_observed_at = excluded.last_observed_at
                """,
                (src_id, tgt_id, rel, context_snippet.strip(), session_id, confidence_score, now_iso),
            )
            conn.commit()

        return {
            "status": "STORED",
            "source": {"id": src_id, "label": source_label, "type": source_type},
            "relation": rel,
            "target": {"id": tgt_id, "label": target_label, "type": target_type},
            "context": context_snippet,
            "session_id": session_id,
        }

    def record_investigation(
        self,
        session_id: str,
        ticker: str,
        anomalies: Optional[List[Any]] = None,
        findings: Optional[List[Dict[str, Any]]] = None,
        sector: Optional[str] = None,
    ) -> Dict[str, Any]:
        """Deterministic post-investigation ingestion (Law 1, Law 6). Zero LLM cost."""
        ticker = ticker.upper().strip()
        recorded_edges = 0

        # 1. User investigated ticker
        self.store_observation(
            source_label="User",
            source_type="USER",
            relation="INVESTIGATED",
            target_label=ticker,
            target_type="TICKER",
            context_snippet=f"Investigasi otonom sesi {session_id}",
            session_id=session_id,
            confidence_score=1.0,
        )
        recorded_edges += 1

        # 2. Sector relation if available
        if sector:
            self.store_observation(
                source_label=ticker,
                source_type="TICKER",
                relation="BELONGS_TO_SECTOR",
                target_label=sector,
                target_type="SECTOR",
                context_snippet=f"Sektor industri {sector}",
                session_id=session_id,
                confidence_score=1.0,
            )
            recorded_edges += 1

        # 3. Anomalies
        if anomalies:
            for a in anomalies:
                z = getattr(a, "z_score", None) or (a.get("z_score") if isinstance(a, dict) else 0.0)
                metric = getattr(a, "metric_type", None) or (a.get("metric_type") if isinstance(a, dict) else "QUANT")
                date_str = getattr(a, "date", None) or (a.get("date") if isinstance(a, dict) else "")
                desc = getattr(a, "description", None) or (a.get("description") if isinstance(a, dict) else "")
                anom_label = f"{metric}_{date_str}_{round(float(z), 1)}s"

                self.store_observation(
                    source_label=ticker,
                    source_type="TICKER",
                    relation="TRIGGERED_ANOMALY",
                    target_label=anom_label,
                    target_type="ANOMALY_METRIC",
                    context_snippet=desc or f"Lonjakan {metric} sebesar {z:.2f} sigma pada {date_str}",
                    session_id=session_id,
                    confidence_score=1.0,
                )
                recorded_edges += 1

        # 4. Findings and catalysts
        if findings:
            for f in findings:
                title = f.get("title", "")
                stat = f.get("verification_status", "SUPPORTED")
                conf = float(f.get("confidence_score", 0.85))
                causality = f.get("causality_status", "")
                if not title:
                    continue

                rel = "CATALYZED_BY" if causality in ("LIKELY_CATALYST", "PRECEDED_ANNOUNCEMENT") else "ASSOCIATED_WITH"
                self.store_observation(
                    source_label=ticker,
                    source_type="TICKER",
                    relation=rel,
                    target_label=title,
                    target_type="CATALYST_EVENT",
                    context_snippet=f"[{stat}] {f.get('claim_text', title)}",
                    session_id=session_id,
                    confidence_score=conf,
                    target_metadata={"verification_status": stat, "causality": causality},
                )
                recorded_edges += 1

        return {
            "session_id": session_id,
            "ticker": ticker,
            "recorded_edges": recorded_edges,
            "status": "RECORDED",
        }

    @staticmethod
    def clean_corporate_tokens(text: str) -> str:
        """Strip common corporate suffixes, prefixes, and punctuation."""
        cleaned = re.sub(r"\b(pt|tbk|persero|corp|corporation|inc)\b", "", text.lower())
        cleaned = re.sub(r"[^a-z0-9\s]", " ", cleaned)
        return re.sub(r"\s+", " ", cleaned).strip()

    def _resolve_target_nodes(self, G: nx.DiGraph, query: str) -> List[str]:
        """Find matching node IDs in the graph using alias resolution and multi-strategy match."""
        q_raw = query.strip().lower()
        if not q_raw:
            return []

        q_clean = self.clean_corporate_tokens(q_raw)
        alias_ticker = KNOWN_TICKER_ALIASES.get(q_clean) or KNOWN_TICKER_ALIASES.get(q_raw)

        exact_matches = []
        alias_matches = []
        partial_matches = []

        # Check known alias first
        if alias_ticker:
            candidate_id = f"ticker:{alias_ticker.lower()}"
            if candidate_id in G:
                alias_matches.append(candidate_id)

        for node_id, data in G.nodes(data=True):
            label = str(data.get("label", "")).lower()
            clean_id = node_id.lower()
            label_clean = self.clean_corporate_tokens(label)

            # Direct exact match on label or node_id
            if label == q_raw or clean_id == f"ticker:{q_raw}" or clean_id == q_raw:
                exact_matches.append(node_id)
                continue

            if q_clean and (label_clean == q_clean or clean_id == f"ticker:{q_clean}"):
                exact_matches.append(node_id)
                continue

            # Check metadata fields (company_name, aliases)
            meta = data.get("metadata", {})
            if isinstance(meta, dict):
                company_name = str(meta.get("company_name", "")).lower()
                company_clean = self.clean_corporate_tokens(company_name)
                if q_clean and (q_clean == company_clean or (company_clean and (q_clean in company_clean or company_clean in q_clean))):
                    alias_matches.append(node_id)
                    continue

                for alias in meta.get("aliases", []):
                    alias_clean = self.clean_corporate_tokens(str(alias))
                    if q_clean and (q_clean == alias_clean or (alias_clean and (q_clean in alias_clean or alias_clean in q_clean))):
                        alias_matches.append(node_id)
                        break

            # Fallback partial matching
            if (q_raw and (q_raw in label or q_raw in clean_id)) or (q_clean and label_clean and q_clean in label_clean):
                partial_matches.append(node_id)

        # Prioritize exact matches > alias/metadata matches > partial matches
        seen = set()
        results = []
        for match in exact_matches + alias_matches + partial_matches:
            if match not in seen:
                seen.add(match)
                results.append(match)

        return results[:5]

    def retrieve_ego_subgraph(
        self,
        entity_query: str,
        radius: Optional[int] = None,
    ) -> Dict[str, Any]:
        """Extract an Ego-Graph (radius <= 2) around target entity with recency decay."""
        rad = radius if radius is not None else self.max_radius
        G = self.load_graph()

        if G.number_of_nodes() == 0:
            return {
                "query": entity_query,
                "root_nodes": [],
                "nodes": [],
                "edges": [],
            }

        matched_roots = self._resolve_target_nodes(G, entity_query)
        if not matched_roots:
            return {
                "query": entity_query,
                "root_nodes": [],
                "nodes": [],
                "edges": [],
            }

        # Build union of ego-graphs across matched roots
        subgraph_nodes = set()
        for root in matched_roots:
            ego = nx.ego_graph(G, root, radius=rad, undirected=True)
            subgraph_nodes.update(ego.nodes())

        H = G.subgraph(subgraph_nodes)
        now = datetime.now(timezone.utc)

        # Identify superseded entity nodes across the graph
        superseded_target_ids = set()
        for u, v, data in G.edges(data=True):
            if data.get("relation") == "SUPERSEDES":
                superseded_target_ids.add(v)

        edges_out = []
        for u, v, data in H.edges(data=True):
            observed_str = data.get("last_observed_at", "")
            delta_days = 0.0
            try:
                dt = datetime.fromisoformat(observed_str.replace("Z", "+00:00"))
                if dt.tzinfo is None:
                    dt = dt.replace(tzinfo=timezone.utc)
                delta_days = max(0.0, (now - dt).total_seconds() / 86400.0)
            except Exception:
                delta_days = 0.0

            base_weight = float(data.get("weight", 1.0))
            decay_factor = math.exp(-self.lambda_decay * delta_days)
            effective_weight = round(base_weight * decay_factor, 4)

            is_superseded = False
            if (u in superseded_target_ids or v in superseded_target_ids) and data.get("relation") != "SUPERSEDES":
                is_superseded = True
                effective_weight = 0.0

            u_data = G.nodes[u]
            v_data = G.nodes[v]

            edges_out.append({
                "source_id": u,
                "source_label": u_data.get("label", u),
                "source_type": u_data.get("node_type", "ENTITY"),
                "target_id": v,
                "target_label": v_data.get("label", v),
                "target_type": v_data.get("node_type", "ENTITY"),
                "relation": data.get("relation", "RELATES_TO"),
                "context_snippet": data.get("context_snippet", ""),
                "session_id": data.get("session_id", ""),
                "weight": base_weight,
                "effective_weight": effective_weight,
                "decay_factor": round(decay_factor, 4),
                "days_ago": round(delta_days, 1),
                "confidence_score": data.get("confidence_score", 1.0),
                "is_superseded": is_superseded,
            })

        # Sort edges by effective_weight descending (active facts first)
        edges_out.sort(key=lambda e: (not e["is_superseded"], e["effective_weight"]), reverse=True)

        nodes_out = []
        for n in H.nodes():
            n_data = G.nodes[n]
            nodes_out.append({
                "id": n,
                "label": n_data.get("label", n),
                "node_type": n_data.get("node_type", "ENTITY"),
                "last_observed_at": n_data.get("last_observed_at", ""),
                "metadata": n_data.get("metadata", {}),
            })

        return {
            "query": entity_query,
            "root_nodes": matched_roots,
            "nodes": nodes_out,
            "edges": edges_out[:25],
        }

    def format_investigative_prompt(
        self,
        entity_query: str,
        radius: int = 2,
        max_edges: int = 10,
    ) -> str:
        """Format recalled graph memories into compact XML block for agent system prompt."""
        ego = self.retrieve_ego_subgraph(entity_query, radius=radius)
        edges = ego.get("edges", [])
        if not edges:
            return ""

        lines = []
        for e in edges:
            if e.get("is_superseded") or e.get("relation") == "SUPERSEDES":
                continue
            src = e["source_label"]
            tgt = e["target_label"]
            rel = e["relation"]
            ctx = e["context_snippet"]
            ctx_part = f" — {ctx}" if ctx else ""
            days = e.get("days_ago", 0.0)
            recency = f" [{days:.0f}h lalu]" if days > 0 else " [hari ini]"
            lines.append(f"- ({src}) --[{rel}]--> ({tgt}){ctx_part}{recency}")
            if len(lines) >= max_edges:
                break

        if not lines:
            return ""

        body = "\n".join(lines)
        return (
            "<investigative_memory>\n"
            f"# Rekam memori graf lokal untuk '{entity_query}':\n"
            f"{body}\n"
            "</investigative_memory>"
        )

    def find_shortest_path(
        self,
        source_query: str,
        target_query: str,
    ) -> Optional[List[Dict[str, Any]]]:
        """Find the shortest connection path between two entities in the knowledge graph."""
        G = self.load_graph()
        src_nodes = self._resolve_target_nodes(G, source_query)
        tgt_nodes = self._resolve_target_nodes(G, target_query)

        if not src_nodes or not tgt_nodes:
            return None

        # Convert to undirected copy for path-finding
        U = G.to_undirected()

        best_path = None
        for s in src_nodes:
            for t in tgt_nodes:
                if s == t:
                    continue
                if nx.has_path(U, s, t):
                    p = nx.shortest_path(U, s, t)
                    if best_path is None or len(p) < len(best_path):
                        best_path = p

        if not best_path:
            return None

        path_elements = []
        for i in range(len(best_path)):
            curr_id = best_path[i]
            curr_node = G.nodes[curr_id]
            step: Dict[str, Any] = {
                "node_id": curr_id,
                "label": curr_node.get("label", curr_id),
                "type": curr_node.get("node_type", "ENTITY"),
            }
            if i < len(best_path) - 1:
                next_id = best_path[i + 1]
                # Check directed edge in either direction
                if G.has_edge(curr_id, next_id):
                    edge_data = G.edges[curr_id, next_id]
                    step["relation_to_next"] = edge_data.get("relation", "RELATES_TO")
                    step["direction"] = "FORWARD"
                elif G.has_edge(next_id, curr_id):
                    edge_data = G.edges[next_id, curr_id]
                    step["relation_to_next"] = edge_data.get("relation", "RELATES_TO")
                    step["direction"] = "REVERSE"
            path_elements.append(step)

        return path_elements

    def get_graph_stats(self) -> Dict[str, Any]:
        """Compute summary topological graph statistics and identify central nodes."""
        G = self.load_graph()
        num_nodes = G.number_of_nodes()
        num_edges = G.number_of_edges()

        if num_nodes == 0:
            return {
                "total_nodes": 0,
                "total_edges": 0,
                "node_types": {},
                "top_central_entities": [],
            }

        node_types: Dict[str, int] = {}
        for _, d in G.nodes(data=True):
            nt = d.get("node_type", "ENTITY")
            node_types[nt] = node_types.get(nt, 0) + 1

        # Centrality metrics: Degree, PageRank, and Betweenness Centrality
        degrees = dict(G.degree())
        try:
            pageranks = nx.pagerank(G, alpha=0.85, weight="weight") if num_nodes > 1 else {n: 1.0 for n in G.nodes()}
        except Exception:
            pageranks = {n: (degrees.get(n, 0) / max(1, num_nodes)) for n in G.nodes()}

        try:
            U = G.to_undirected()
            betweenness = nx.betweenness_centrality(U, normalized=True) if num_nodes > 2 else {n: 0.0 for n in G.nodes()}
        except Exception:
            betweenness = {n: 0.0 for n in G.nodes()}

        # Composite score combines authority (PageRank), bridge connectivity (Betweenness), and degree
        composite_scores: Dict[str, float] = {}
        max_deg = max(degrees.values()) if degrees else 1
        for n_id in G.nodes():
            pr = pageranks.get(n_id, 0.0)
            bet = betweenness.get(n_id, 0.0)
            deg_norm = degrees.get(n_id, 0) / max(1, max_deg)
            composite_scores[n_id] = round(0.4 * pr + 0.4 * bet + 0.2 * deg_norm, 4)

        sorted_nodes = sorted(G.nodes(), key=lambda x: composite_scores.get(x, 0.0), reverse=True)[:5]
        top_central = []
        for n_id in sorted_nodes:
            n_data = G.nodes[n_id]
            top_central.append({
                "id": n_id,
                "label": n_data.get("label", n_id),
                "type": n_data.get("node_type", "ENTITY"),
                "connections": degrees.get(n_id, 0),
                "pagerank": round(pageranks.get(n_id, 0.0), 4),
                "betweenness": round(betweenness.get(n_id, 0.0), 4),
                "composite_score": composite_scores.get(n_id, 0.0),
            })

        return {
            "total_nodes": num_nodes,
            "total_edges": num_edges,
            "node_types": node_types,
            "top_central_entities": top_central,
        }

    def clear_memory(self, session_id: Optional[str] = None) -> int:
        """Clear memory edges and orphaned nodes (optionally restricted to a single session)."""
        with self._get_connection() as conn:
            cursor = conn.cursor()
            if session_id:
                cursor.execute("DELETE FROM memory_edges WHERE session_id = ?", (session_id,))
                del_count = cursor.rowcount
            else:
                cursor.execute("DELETE FROM memory_edges")
                del_count = cursor.rowcount
                cursor.execute("DELETE FROM memory_nodes")

            # Clean orphaned nodes if only session was deleted
            if session_id:
                cursor.execute("""
                    DELETE FROM memory_nodes 
                    WHERE id NOT IN (SELECT source_id FROM memory_edges)
                      AND id NOT IN (SELECT target_id FROM memory_edges)
                """)

            conn.commit()
            return del_count
