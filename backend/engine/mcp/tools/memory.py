"""Local Graph Memory MCP Tool Definitions and Dispatcher.

Complies strictly with:
- Law 4: Local-First Data Sovereignty (Local SQLite persistence)
- Law 6: Local Conversational Graph Memory Engine (Associative Ego-Graph with Recency Decay)
"""

from typing import Any, Dict, List

from engine.memory.graph_memory import LocalGraphMemory
from engine.memory.visualizer import GraphVisualizer


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
                    "radius": {
                        "type": "integer",
                        "description": "Kedalaman hop penelusuran Ego-Graph (default: 2, max: 2)",
                        "default": 2,
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
                    "session_id": {
                        "type": "string",
                        "description": "ID sesi investigasi terkait (opsional)",
                    },
                },
                "required": ["source_label", "relation", "target_label"],
            },
        },
        {
            "name": "memory_find_connection",
            "description": "Lacak jalur koneksi terpendek (shortest path) antara dua entitas pasar untuk membongkar afiliasi atau keterkaitan tersembunyi.",
            "inputSchema": {
                "type": "object",
                "properties": {
                    "source_entity": {
                        "type": "string",
                        "description": "Nama/label entitas pertama (contoh: 'ANTM')",
                    },
                    "target_entity": {
                        "type": "string",
                        "description": "Nama/label entitas kedua (contoh: 'BBCA')",
                    },
                },
                "required": ["source_entity", "target_entity"],
            },
        },
        {
            "name": "memory_get_graph_stats",
            "description": "Ambil ringkasan statistik topologi graf memori dan entitas sentral (hub/god nodes).",
            "inputSchema": {
                "type": "object",
                "properties": {},
            },
        },
        {
            "name": "memory_export_graph_html",
            "description": "Ekspor visualisasi graf interaktif mandiri ke file HTML lokal (gaya Cyber-OSINT Graphify).",
            "inputSchema": {
                "type": "object",
                "properties": {
                    "output_path": {
                        "type": "string",
                        "description": "Path tujuan penyimpanan file HTML (default: ~/.niskava/graph.html)",
                        "default": "~/.niskava/graph.html",
                    },
                    "session_id": {
                        "type": "string",
                        "description": "Filter berdasarkan ID sesi tertentu (opsional)",
                    },
                },
            },
        },
    ]


def execute_memory_tool(
    db_path: str, name: str, arguments: Dict[str, Any]
) -> Any:
    """Execute a memory tool call against local SQLite database using LocalGraphMemory."""
    mem = LocalGraphMemory(db_path=db_path)

    if name == "memory_recall_context":
        query = arguments.get("query_entity", "").strip()
        radius = int(arguments.get("radius", 2))
        res = mem.retrieve_ego_subgraph(entity_query=query, radius=radius)
        return {
            "query": res["query"],
            "nodes_found": len(res["nodes"]),
            "nodes": res["nodes"],
            "edges": res["edges"],
        }

    if name == "memory_store_observation":
        src_label = arguments.get("source_label", "").strip()
        src_type = arguments.get("source_type", "TICKER").strip()
        relation = arguments.get("relation", "").strip().upper()
        tgt_label = arguments.get("target_label", "").strip()
        tgt_type = arguments.get("target_type", "ENTITY").strip()
        context = arguments.get("context_snippet", "").strip()
        session_id = arguments.get("session_id")

        return mem.store_observation(
            source_label=src_label,
            source_type=src_type,
            relation=relation,
            target_label=tgt_label,
            target_type=tgt_type,
            context_snippet=context,
            session_id=session_id,
        )

    if name == "memory_find_connection":
        src = arguments.get("source_entity", "").strip()
        tgt = arguments.get("target_entity", "").strip()
        path = mem.find_shortest_path(src, tgt)
        return {
            "source": src,
            "target": tgt,
            "path_found": path is not None,
            "hops": len(path) - 1 if path else 0,
            "path": path or [],
        }

    if name == "memory_get_graph_stats":
        return mem.get_graph_stats()

    if name == "memory_export_graph_html":
        out_path = arguments.get("output_path", "~/.niskava/graph.html")
        sess_id = arguments.get("session_id")
        viz = GraphVisualizer(memory=mem)
        saved = viz.export_to_file(output_path=out_path, session_id=sess_id)
        return {
            "status": "EXPORTED",
            "file_path": saved,
            "session_id": sess_id or "ALL",
        }

    raise ValueError(f"Unknown Memory tool: {name}")
