"""Local Graph Memory MCP Tool Definitions and Dispatcher.

Complies strictly with:
- Law 4: Local-First Data Sovereignty (Local SQLite persistence)
- Law 6: Local Conversational Graph Memory Engine (Associative Ego-Graph with Recency Decay)
"""

from datetime import datetime, timezone
import json
import math
import sqlite3
import time
from typing import Any, Dict, List


def get_memory_tool_definitions() -> List[Dict[str, Any]]:
    """Return standardized MCP schemas for memory tools."""
    return [
        {
            "name": "memory_recall_context",
            "description": "Ambil relasi graf memori dan konteks masa lalu untuk suatu entitas (misal ticker atau topik) dari basis data SQLite lokal, diurutkan berdasarkan bobot peluruhan waktu (recency decay).",
            "inputSchema": {
                "type": "object",
                "properties": {
                    "query_entity": {
                        "type": "string",
                        "description": "Nama entitas atau ticker saham yang dicari (contoh: 'ANTM', 'smelter')",
                    },
                },
                "required": ["query_entity"],
            },
        },
        {
            "name": "memory_store_observation",
            "description": "Simpan observasi relasi baru (graf memori) antara dua entitas ke dalam SQLite lokal.",
            "inputSchema": {
                "type": "object",
                "properties": {
                    "source_label": {
                        "type": "string",
                        "description": "Label entitas asal (contoh: 'ANTM')",
                    },
                    "source_type": {
                        "type": "string",
                        "description": "Tipe entitas asal (contoh: 'TICKER', 'PERSON', 'FACILITY')",
                        "default": "TICKER",
                    },
                    "relation": {
                        "type": "string",
                        "description": "Relasi/predikat penghubung (contoh: 'OPERATES', 'AFFECTED_BY', 'SUSPENDED_BY')",
                    },
                    "target_label": {
                        "type": "string",
                        "description": "Label entitas tujuan (contoh: 'Smelter Haltim')",
                    },
                    "target_type": {
                        "type": "string",
                        "description": "Tipe entitas tujuan (contoh: 'FACILITY', 'EVENT')",
                        "default": "ENTITY",
                    },
                    "context_snippet": {
                        "type": "string",
                        "description": "Kutipan atau konteks bukti ringkas",
                        "default": "",
                    },
                },
                "required": ["source_label", "relation", "target_label"],
            },
        },
    ]


def _ensure_memory_schema(cursor: sqlite3.Cursor) -> None:
    """Ensure memory_nodes and memory_edges tables exist (Self-healing Auto-DDL)."""
    cursor.execute("""
        CREATE TABLE IF NOT EXISTS memory_nodes (
            id TEXT PRIMARY KEY,
            label TEXT NOT NULL,
            node_type TEXT NOT NULL,
            metadata_json TEXT,
            last_observed_at TEXT NOT NULL,
            created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
        );
    """)
    cursor.execute("""
        CREATE TABLE IF NOT EXISTS memory_edges (
            source_id TEXT NOT NULL REFERENCES memory_nodes(id) ON DELETE CASCADE,
            target_id TEXT NOT NULL REFERENCES memory_nodes(id) ON DELETE CASCADE,
            relation TEXT NOT NULL,
            context_snippet TEXT,
            session_id TEXT,
            weight REAL NOT NULL DEFAULT 1.0,
            last_observed_at TEXT NOT NULL,
            created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
            PRIMARY KEY (source_id, target_id, relation)
        );
    """)
    cursor.execute(
        "CREATE INDEX IF NOT EXISTS idx_mem_edges_src ON memory_edges(source_id);"
    )
    cursor.execute(
        "CREATE INDEX IF NOT EXISTS idx_mem_edges_tgt ON memory_edges(target_id);"
    )


def execute_memory_tool(
    db_path: str, name: str, arguments: Dict[str, Any]
) -> Any:
    """Execute a memory tool call against local SQLite database."""
    conn = sqlite3.connect(db_path)
    conn.row_factory = sqlite3.Row
    cursor = conn.cursor()

    try:
        # Auto-DDL self-healing
        _ensure_memory_schema(cursor)
        conn.commit()

        if name == "memory_recall_context":
            query = arguments.get("query_entity", "").strip()
            # 1. Search matching nodes
            cursor.execute(
                """
                SELECT id, label, node_type, metadata_json, last_observed_at
                FROM memory_nodes
                WHERE label LIKE ? OR id LIKE ?
                LIMIT 5
                """,
                (f"%{query}%", f"%{query.lower()}%"),
            )
            nodes = [dict(r) for r in cursor.fetchall()]

            edges = []
            if nodes:
                node_ids = [n["id"] for n in nodes]
                placeholders = ",".join("?" for _ in node_ids)
                cursor.execute(
                    f"""
                    SELECT e.source_id, e.target_id, e.relation, e.context_snippet, e.weight,
                           e.last_observed_at, sn.label AS source_label, tn.label AS target_label
                    FROM memory_edges e
                    JOIN memory_nodes sn ON e.source_id = sn.id
                    JOIN memory_nodes tn ON e.target_id = tn.id
                    WHERE e.source_id IN ({placeholders}) OR e.target_id IN ({placeholders})
                    LIMIT 50
                    """,
                    node_ids + node_ids,
                )
                raw_edges = cursor.fetchall()
                now = datetime.now(timezone.utc)

                for r in raw_edges:
                    edge = dict(r)
                    observed_str = edge.get("last_observed_at", "")
                    delta_days = 0.0
                    try:
                        cleaned_ts = observed_str.replace("Z", "+00:00")
                        dt = datetime.fromisoformat(cleaned_ts)
                        if dt.tzinfo is None:
                            dt = dt.replace(tzinfo=timezone.utc)
                        delta_days = max(
                            0.0, (now - dt).total_seconds() / 86400.0
                        )
                    except Exception:
                        delta_days = 0.0

                    base_weight = float(edge.get("weight", 1.0))
                    # Law 6: Recency decay e^(-lambda * delta_t) with lambda = 0.05 (~14 day half-life)
                    decay_factor = math.exp(-0.05 * delta_days)
                    effective_weight = round(base_weight * decay_factor, 4)

                    edge["days_ago"] = round(delta_days, 1)
                    edge["decay_factor"] = round(decay_factor, 4)
                    edge["effective_weight"] = effective_weight
                    edges.append(edge)

                # Sort by effective_weight descending (Law 6)
                edges.sort(key=lambda e: e["effective_weight"], reverse=True)

            return {
                "query": query,
                "nodes_found": len(nodes),
                "nodes": nodes,
                "edges": edges[:20],
            }

        if name == "memory_store_observation":
            src_label = arguments.get("source_label", "").strip()
            src_type = arguments.get("source_type", "TICKER").strip()
            relation = arguments.get("relation", "").strip().upper()
            tgt_label = arguments.get("target_label", "").strip()
            tgt_type = arguments.get("target_type", "ENTITY").strip()
            context = arguments.get("context_snippet", "").strip()

            src_id = f"{src_type.lower()}:{src_label.lower().replace(' ', '_')}"
            tgt_id = f"{tgt_type.lower()}:{tgt_label.lower().replace(' ', '_')}"
            now_iso = time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())

            # Upsert source node
            cursor.execute(
                """
                INSERT INTO memory_nodes (id, label, node_type, metadata_json, last_observed_at)
                VALUES (?, ?, ?, ?, ?)
                ON CONFLICT(id) DO UPDATE SET
                    last_observed_at = excluded.last_observed_at
                """,
                (src_id, src_label, src_type, "{}", now_iso),
            )

            # Upsert target node
            cursor.execute(
                """
                INSERT INTO memory_nodes (id, label, node_type, metadata_json, last_observed_at)
                VALUES (?, ?, ?, ?, ?)
                ON CONFLICT(id) DO UPDATE SET
                    last_observed_at = excluded.last_observed_at
                """,
                (tgt_id, tgt_label, tgt_type, "{}", now_iso),
            )

            # Upsert edge
            cursor.execute(
                """
                INSERT INTO memory_edges (source_id, target_id, relation, context_snippet, weight, last_observed_at)
                VALUES (?, ?, ?, ?, 1.0, ?)
                ON CONFLICT(source_id, target_id, relation) DO UPDATE SET
                    context_snippet = excluded.context_snippet,
                    weight = weight + 0.5,
                    last_observed_at = excluded.last_observed_at
                """,
                (src_id, tgt_id, relation, context, now_iso),
            )

            conn.commit()
            return {
                "status": "STORED",
                "source": {"id": src_id, "label": src_label},
                "relation": relation,
                "target": {"id": tgt_id, "label": tgt_label},
                "context": context,
            }

        raise ValueError(f"Unknown Memory tool: {name}")
    finally:
        conn.close()
