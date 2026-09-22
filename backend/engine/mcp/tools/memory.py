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
            "description": "Retrieve graph memory relations and past context for an entity (e.g. ticker or topic) from local SQLite, ranked by temporal recency decay weight.",
            "inputSchema": {
                "type": "object",
                "properties": {
                    "query_entity": {
                        "type": "string",
                        "description": "Entity name or stock ticker to search (e.g. 'ANTM', 'smelter')",
                    },
                    "radius": {
                        "type": "integer",
                        "description": "Ego-graph traversal hop depth (default: 2, max: 2)",
                        "default": 2,
                    },
                },
                "required": ["query_entity"],
            },
        },
        {
            "name": "memory_store_observation",
            "description": "Persist a new relational observation (graph memory triple) between two entities into local SQLite.",
            "inputSchema": {
                "type": "object",
                "properties": {
                    "source_label": {
                        "type": "string",
                        "description": "Source entity label (e.g. 'ANTM')",
                    },
                    "source_type": {
                        "type": "string",
                        "description": "Source entity type (e.g. 'TICKER', 'PERSON', 'FACILITY')",
                        "default": "TICKER",
                    },
                    "relation": {
                        "type": "string",
                        "description": "Relation predicate (e.g. 'OPERATES', 'AFFECTED_BY', 'SUSPENDED_BY')",
                    },
                    "target_label": {
                        "type": "string",
                        "description": "Target entity label (e.g. 'Smelter Haltim')",
                    },
                    "target_type": {
                        "type": "string",
                        "description": "Target entity type (e.g. 'FACILITY', 'EVENT')",
                        "default": "ENTITY",
                    },
                    "context_snippet": {
                        "type": "string",
                        "description": "Brief evidence quote or context",
                        "default": "",
                    },
                    "session_id": {
                        "type": "string",
                        "description": "Associated investigation session ID (optional)",
                    },
                },
                "required": ["source_label", "relation", "target_label"],
            },
        },
        {
            "name": "memory_find_connection",
            "description": "Trace the shortest connection path between two market entities to uncover hidden affiliations or relationships.",
            "inputSchema": {
                "type": "object",
                "properties": {
                    "source_entity": {
                        "type": "string",
                        "description": "First entity name or label (e.g. 'ANTM')",
                    },
                    "target_entity": {
                        "type": "string",
                        "description": "Second entity name or label (e.g. 'BBCA')",
                    },
                },
                "required": ["source_entity", "target_entity"],
            },
        },
        {
            "name": "memory_get_graph_stats",
            "description": "Retrieve summary topological statistics of the memory graph and identify central hub entities.",
            "inputSchema": {
                "type": "object",
                "properties": {},
            },
        },
        {
            "name": "memory_export_graph_html",
            "description": "Export a standalone interactive graph visualization to a local HTML file (vis-network.js, institutional financial aesthetic).",
            "inputSchema": {
                "type": "object",
                "properties": {
                    "output_path": {
                        "type": "string",
                        "description": "Destination path for the HTML file (default: ~/.niskava/graph.html)",
                        "default": "~/.niskava/graph.html",
                    },
                    "session_id": {
                        "type": "string",
                        "description": "Filter nodes and edges by a specific session ID (optional)",
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
