"""Market Intelligence Interactive Graph Visualizer for Niskava Agent.

Generates standalone, self-contained interactive HTML visualizations
with vis-network.js, institutional financial aesthetics, entity inspection,
search & focus, and ego-graph filtering.
"""

from datetime import datetime, timezone
import json
import math
import os
from typing import Any, Dict, List, Optional

from engine.memory.graph_memory import LocalGraphMemory


# Institutional financial node styling (Clean, dignified, non-cyber)
NODE_TYPE_STYLES = {
    "TICKER": {
        "color": {"background": "#1E3A8A", "border": "#3B82F6", "highlight": {"background": "#2563EB", "border": "#93C5FD"}},
        "font": {"color": "#FFFFFF", "face": "monospace", "size": 13, "bold": True},
        "shape": "box",
        "margin": 10,
        "size": 22,
    },
    "SECTOR": {
        "color": {"background": "#312E81", "border": "#6366F1", "highlight": {"background": "#4F46E5", "border": "#C7D2FE"}},
        "font": {"color": "#FFFFFF", "face": "sans-serif", "size": 12, "bold": True},
        "shape": "box",
        "margin": 8,
        "size": 20,
    },
    "BROKER": {
        "color": {"background": "#78350F", "border": "#D97706", "highlight": {"background": "#B45309", "border": "#FDE68A"}},
        "font": {"color": "#FFFFFF", "face": "monospace", "size": 12, "bold": True},
        "shape": "box",
        "margin": 8,
        "size": 20,
    },
    "PERSON": {
        "color": {"background": "#1F2937", "border": "#4B5563", "highlight": {"background": "#374151", "border": "#E5E7EB"}},
        "font": {"color": "#F9FAFB", "face": "sans-serif", "size": 12},
        "shape": "box",
        "margin": 8,
        "size": 18,
    },
    "CATALYST_EVENT": {
        "color": {"background": "#064E3B", "border": "#059669", "highlight": {"background": "#10B981", "border": "#A7F3D0"}},
        "font": {"color": "#ECFDF5", "face": "sans-serif", "size": 11, "bold": True},
        "shape": "box",
        "margin": 10,
        "size": 18,
    },
    "ANOMALY_METRIC": {
        "color": {"background": "#7F1D1D", "border": "#DC2626", "highlight": {"background": "#EF4444", "border": "#FECACA"}},
        "font": {"color": "#FEF2F2", "face": "monospace", "size": 11, "bold": True},
        "shape": "box",
        "margin": 8,
        "size": 20,
    },
    "USER": {
        "color": {"background": "#0F172A", "border": "#38BDF8", "highlight": {"background": "#0284C7", "border": "#E0F2FE"}},
        "font": {"color": "#F8FAFC", "face": "sans-serif", "size": 13, "bold": True},
        "shape": "box",
        "margin": 10,
        "size": 24,
    },
    "PRICE_LEVEL": {
        "color": {"background": "#1E293B", "border": "#64748B", "highlight": {"background": "#334155", "border": "#CBD5E1"}},
        "font": {"color": "#F8FAFC", "face": "monospace", "size": 12},
        "shape": "box",
        "margin": 8,
        "size": 18,
    },
}

DEFAULT_NODE_STYLE = {
    "color": {"background": "#1E293B", "border": "#475569", "highlight": {"background": "#334155", "border": "#CBD5E1"}},
    "font": {"color": "#F1F5F9", "face": "sans-serif", "size": 12},
    "shape": "box",
    "margin": 8,
    "size": 18,
}

HTML_TEMPLATE = """<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>__TITLE__</title>
    <script type="text/javascript" src="https://unpkg.com/vis-network/standalone/umd/vis-network.min.js"></script>
    <style>
        :root {
            --bg: #0B0E14;
            --surface: #121620;
            --surface-card: #181E2C;
            --border: #242D40;
            --border-subtle: #1A2130;
            --text-main: #F1F5F9;
            --text-muted: #8E9BAE;
            --accent: #3B82F6;
            --accent-emerald: #10B981;
            --accent-amber: #D97706;
            --accent-rose: #E11D48;
        }
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            background: var(--bg);
            color: var(--text-main);
            font-family: -apple-system, BlinkMacSystemFont, 'Inter', 'Geist', 'Segoe UI', Roboto, sans-serif;
            font-feature-settings: 'tnum' 1, 'cv02' 1, 'cv03' 1, 'cv04' 1;
            font-variant-numeric: tabular-nums;
            display: flex;
            height: 100vh;
            overflow: hidden;
        }
        .sidebar {
            width: 320px;
            background: var(--surface);
            border-right: 1px solid var(--border);
            display: flex;
            flex-direction: column;
            padding: 20px;
            gap: 16px;
            overflow-y: auto;
            z-index: 10;
        }
        .brand {
            display: flex;
            flex-direction: column;
            gap: 4px;
        }
        .brand-title {
            font-size: 15px;
            font-weight: 700;
            color: var(--text-main);
            letter-spacing: 0.5px;
            display: flex;
            align-items: center;
            justify-content: space-between;
        }
        .brand-subtitle {
            font-size: 11px;
            color: var(--text-muted);
            font-weight: 500;
        }
        .badge-live {
            background: rgba(59, 130, 246, 0.12);
            color: var(--accent);
            font-size: 10px;
            font-weight: 600;
            padding: 2px 8px;
            border-radius: 4px;
            border: 1px solid rgba(59, 130, 246, 0.3);
        }
        .section-title {
            font-size: 11px;
            font-weight: 600;
            color: var(--text-muted);
            text-transform: uppercase;
            letter-spacing: 0.5px;
            margin-top: 6px;
        }
        .search-box {
            width: 100%;
            background: var(--surface-card);
            border: 1px solid var(--border);
            border-radius: 6px;
            padding: 9px 12px;
            color: #FFF;
            font-size: 13px;
            outline: none;
            transition: border-color 0.15s;
        }
        .search-box:focus {
            border-color: var(--accent);
        }
        .legend {
            display: flex;
            flex-direction: column;
            gap: 8px;
            font-size: 12px;
        }
        .legend-item {
            display: flex;
            align-items: center;
            gap: 10px;
            color: var(--text-main);
        }
        .legend-dot {
            width: 10px;
            height: 10px;
            border-radius: 2px;
        }
        .stats-card {
            background: var(--surface-card);
            border: 1px solid var(--border);
            border-radius: 6px;
            padding: 12px;
            font-size: 12px;
            line-height: 1.6;
        }
        .stats-val {
            font-weight: 600;
            color: var(--accent);
            font-family: monospace;
        }
        .canvas-area {
            flex: 1;
            position: relative;
            background: var(--bg);
        }
        #network {
            width: 100%;
            height: 100%;
        }
        .top-bar {
            position: absolute;
            top: 16px;
            left: 20px;
            right: 360px;
            display: flex;
            align-items: center;
            gap: 12px;
            pointer-events: none;
        }
        .glass-pill {
            background: rgba(18, 22, 32, 0.92);
            border: 1px solid var(--border);
            border-radius: 6px;
            padding: 6px 14px;
            font-size: 12px;
            color: var(--text-main);
            pointer-events: auto;
            display: flex;
            align-items: center;
            gap: 8px;
        }
        .inspector {
            width: 340px;
            background: var(--surface);
            border-left: 1px solid var(--border);
            display: flex;
            flex-direction: column;
            padding: 20px;
            gap: 14px;
            overflow-y: auto;
            z-index: 10;
        }
        .inspector-title {
            font-size: 15px;
            font-weight: 700;
            color: var(--text-main);
            word-break: break-word;
        }
        .inspector-badge {
            display: inline-block;
            font-size: 10px;
            padding: 2px 8px;
            border-radius: 4px;
            font-weight: 600;
            background: var(--border);
            color: #FFF;
            width: fit-content;
        }
        .detail-row {
            font-size: 12px;
            display: flex;
            flex-direction: column;
            gap: 4px;
            border-bottom: 1px solid var(--border-subtle);
            padding-bottom: 8px;
        }
        .detail-label {
            color: var(--text-muted);
            font-size: 11px;
            text-transform: uppercase;
        }
        .btn {
            background: var(--surface-card);
            color: var(--text-main);
            border: 1px solid var(--border);
            padding: 7px 12px;
            border-radius: 6px;
            cursor: pointer;
            font-size: 12px;
            font-weight: 500;
            transition: all 0.15s;
        }
        .btn:hover {
            background: var(--border);
            border-color: var(--accent);
        }
    </style>
</head>
<body>
    <div class="sidebar">
        <div class="brand">
            <div class="brand-title">
                <span>NISKAVA AGENT</span>
                <span class="badge-live">Market Intelligence</span>
            </div>
            <div class="brand-subtitle">Entity Network & Research Graph</div>
        </div>

        <div class="section-title">Entity Search</div>
        <input type="text" id="searchInput" class="search-box" placeholder="Search ticker, broker, disclosure..." oninput="searchAndHighlight()">

        <div class="section-title">Ego-Graph Radius</div>
        <div style="display: flex; gap: 8px;">
            <button class="btn" style="flex:1" onclick="filterHops(1)">1-Hop</button>
            <button class="btn" style="flex:1" onclick="filterHops(2)">2-Hop</button>
            <button class="btn" style="flex:1" onclick="resetFilter()">Reset</button>
        </div>

        <div class="section-title">Node Classification</div>
        <div class="legend">
            <div class="legend-item"><span class="legend-dot" style="background:#3B82F6"></span> Stock Issuer (TICKER)</div>
            <div class="legend-item"><span class="legend-dot" style="background:#D97706"></span> Exchange Member (BROKER)</div>
            <div class="legend-item"><span class="legend-dot" style="background:#6366F1"></span> Industry Sector (SECTOR)</div>
            <div class="legend-item"><span class="legend-dot" style="background:#059669"></span> Disclosures & Corporate Actions</div>
            <div class="legend-item"><span class="legend-dot" style="background:#DC2626"></span> Volume Outlier & Fund Flow</div>
            <div class="legend-item"><span class="legend-dot" style="background:#38BDF8"></span> User Research Profile (USER)</div>
        </div>

        <div class="section-title">Market Network Statistics</div>
        <div class="stats-card" id="statsArea">
            Loading graph statistics...
        </div>

        <button class="btn" style="margin-top: auto;" onclick="if(network) network.fit({animation: true})">Refit Canvas</button>
    </div>

    <div class="canvas-area">
        <div class="top-bar">
            <div class="glass-pill">
                <span>Research Session:</span>
                <strong style="color:var(--accent);">__SESSION_ID__</strong>
            </div>
            <div class="glass-pill">
                <span>Engine:</span>
                <code>NetworkX DiGraph + Local SQLite WAL</code>
            </div>
        </div>
        <div id="network"></div>
    </div>

    <div class="inspector" id="inspectorPanel">
        <div class="section-title">Intelligence Dossier</div>
        <div id="inspectorContent">
            <p style="color:var(--text-muted); font-size:13px; line-height:1.6;">
                Select an entity or relation on canvas to inspect official IDX disclosure quotes, temporal recency weight, and metadata.
            </p>
        </div>
    </div>

    <script type="text/javascript">
        const rawData = __RAW_JSON__;
        let network = null;
        let allNodes = rawData.nodes || [];
        let allEdges = rawData.edges || [];
        let nodesDataSet = new vis.DataSet(allNodes);
        let edgesDataSet = new vis.DataSet(allEdges);

        function initNetwork() {
            const container = document.getElementById('network');
            const data = { nodes: nodesDataSet, edges: edgesDataSet };
            const options = {
                nodes: {
                    borderWidth: 1,
                    shadow: false
                },
                edges: {
                    smooth: { type: 'cubicBezier', forceDirection: 'none', roundness: 0.15 },
                    shadow: false
                },
                physics: {
                    barnesHut: {
                        gravitationalConstant: -3500,
                        centralGravity: 0.3,
                        springLength: 95,
                        springConstant: 0.04,
                        damping: 0.09
                    },
                    stabilization: { iterations: 150 }
                },
                interaction: {
                    hover: true,
                    tooltipDelay: 200,
                    navigationButtons: true,
                    keyboard: true
                }
            };

            network = new vis.Network(container, data, options);

            network.on("click", function (params) {
                if (params.nodes.length > 0) {
                    inspectNode(params.nodes[0]);
                } else if (params.edges.length > 0) {
                    inspectEdge(params.edges[0]);
                }
            });

            renderStats();
        }

        function renderStats() {
            const s = rawData.stats || {};
            const entities = (s.top_central_entities || []).map(function(e) {
                const pr = e.pagerank ? ' [PR: ' + e.pagerank + ']' : '';
                return '• <strong>' + e.label + '</strong> (' + e.connections + ' connections)' + pr;
            }).join('<br>') || '-';

            const html = 'Total Nodes: <span class="stats-val">' + (s.total_nodes || allNodes.length) + '</span><br>' +
                         'Market Relations: <span class="stats-val">' + (s.total_edges || allEdges.length) + '</span><br>' +
                         'Central Entities & Hubs: <br>' + entities;
            document.getElementById('statsArea').innerHTML = html;
        }

        function inspectNode(nodeId) {
            const node = allNodes.find(function(n) { return n.id === nodeId; });
            if (!node) return;

            const connectedEdges = allEdges.filter(function(e) { return e.from === nodeId || e.to === nodeId; });
            const panel = document.getElementById('inspectorContent');

            const edgeHTML = connectedEdges.map(function(e) {
                const connLabel = (e.from === nodeId) ? ('➔ ' + e.to) : ('⬅ ' + e.from);
                const supersededBadge = e.is_superseded ? ' <span style="color:var(--accent-rose); font-size:10px;">[SUPERSEDED]</span>' : '';
                return '<div style="background:var(--surface-card); border:1px solid var(--border); padding:8px 10px; border-radius:6px; font-size:12px;">' +
                       '<strong>[' + e.relation + ']</strong> ' + connLabel + supersededBadge + '<br>' +
                       '<small style="color:var(--text-muted);">' + (e.context_snippet || '-') + '</small><br>' +
                       '<span style="font-size:10px; color:var(--accent);">Effective Weight: ' + e.effective_weight + '</span>' +
                       '</div>';
            }).join('');

            panel.innerHTML = '<div class="inspector-title">' + node.label + '</div>' +
                '<div class="inspector-badge" style="background:' + (node.color && node.color.background ? node.color.background : '#1E293B') + '">' + node.group + '</div>' +
                '<div class="detail-row" style="margin-top:12px;">' +
                '<span class="detail-label">Identifier</span><code>' + node.id + '</code></div>' +
                '<div class="detail-row"><span class="detail-label">Observation Time</span><span>' + (node.last_observed_at || '-') + '</span></div>' +
                '<div class="detail-row"><span class="detail-label">Relation Degree</span><span class="stats-val">' + connectedEdges.length + '</span></div>' +
                '<div class="section-title" style="margin-top:10px;">Evidence Relation Catalog</div>' +
                '<div style="display:flex; flex-direction:column; gap:8px;">' + (edgeHTML || '<p style="font-size:12px; color:var(--text-muted);">No active relations.</p>') + '</div>';
        }

        function inspectEdge(edgeId) {
            const edge = allEdges.find(function(e) { return e.id === edgeId; });
            if (!edge) return;
            const panel = document.getElementById('inspectorContent');
            const supersededInfo = edge.is_superseded ? '<div class="detail-row"><span class="detail-label" style="color:var(--accent-rose);">Validity Status</span><span style="color:var(--accent-rose); font-weight:600;">SUPERSEDED (Superseded by new transaction)</span></div>' : '';
            panel.innerHTML = '<div class="inspector-title">[' + edge.relation + ']</div>' +
                '<div class="inspector-badge" style="background:#059669">RELATION</div>' +
                '<div class="detail-row" style="margin-top:12px;"><span class="detail-label">Causality Connection</span><span>' + edge.from + ' ➔ ' + edge.to + '</span></div>' +
                '<div class="detail-row"><span class="detail-label">Evidence Snippet / Document</span><span>' + (edge.context_snippet || '-') + '</span></div>' +
                '<div class="detail-row"><span class="detail-label">Temporal Effective Weight</span><span class="stats-val">' + edge.effective_weight + ' (Base: ' + edge.weight + ')</span></div>' +
                supersededInfo +
                '<div class="detail-row"><span class="detail-label">Investigation Session</span><code>' + (edge.session_id || '-') + '</code></div>';
        }

        function searchAndHighlight() {
            const q = document.getElementById('searchInput').value.trim().toLowerCase();
            if (!q) {
                nodesDataSet.update(allNodes);
                return;
            }
            const matched = allNodes.filter(function(n) {
                return n.label.toLowerCase().includes(q) || n.id.toLowerCase().includes(q);
            });
            if (matched.length > 0) {
                const target = matched[0];
                network.focus(target.id, { scale: 1.2, animation: true });
                inspectNode(target.id);
            }
        }

        function filterHops(radius) {
            const selected = network.getSelectedNodes();
            if (selected.length === 0) {
                alert("Please select one node before applying Ego-Graph hop filter.");
                return;
            }
            const root = selected[0];
            let activeNodes = new Set([root]);
            let frontier = [root];

            for (let r = 0; r < radius; r++) {
                let nextFrontier = [];
                for (const curr of frontier) {
                    const conn = network.getConnectedNodes(curr);
                    for (const neighbor of conn) {
                        if (!activeNodes.has(neighbor)) {
                            activeNodes.add(neighbor);
                            nextFrontier.push(neighbor);
                        }
                    }
                }
                frontier = nextFrontier;
            }

            const filteredNodes = allNodes.filter(function(n) { return activeNodes.has(n.id); });
            const filteredEdges = allEdges.filter(function(e) { return activeNodes.has(e.from) && activeNodes.has(e.to); });
            nodesDataSet.clear();
            nodesDataSet.add(filteredNodes);
            edgesDataSet.clear();
            edgesDataSet.add(filteredEdges);
            network.fit({ animation: true });
        }

        function resetFilter() {
            nodesDataSet.clear();
            nodesDataSet.add(allNodes);
            edgesDataSet.clear();
            edgesDataSet.add(allEdges);
            network.fit({ animation: true });
        }

        window.addEventListener('load', initNetwork);
    </script>
</body>
</html>
"""


class GraphVisualizer:
    """Cyber-OSINT and Market Intelligence Graph Visualizer."""

    def __init__(self, memory: Optional[LocalGraphMemory] = None):
        self.memory = memory or LocalGraphMemory()

    def export_graph_data(self, session_id: Optional[str] = None) -> Dict[str, Any]:
        """Convert in-memory graph to Vis.js DataSet format with colors and weights."""
        G = self.memory.load_graph()
        stats = self.memory.get_graph_stats()

        nodes = []
        for n, data in G.nodes(data=True):
            ntype = data.get("node_type", "ENTITY").upper()
            style = NODE_TYPE_STYLES.get(ntype, DEFAULT_NODE_STYLE)

            nodes.append({
                "id": n,
                "label": data.get("label", n),
                "title": f"[{ntype}] {data.get('label', n)}",
                "group": ntype,
                "color": style["color"],
                "font": style["font"],
                "shape": style["shape"],
                "size": style.get("size", 18),
                "margin": style.get("margin", 8),
                "last_observed_at": data.get("last_observed_at", ""),
                "metadata": data.get("metadata", {}),
            })

        edges = []
        now = datetime.now(timezone.utc)
        for u, v, data in G.edges(data=True):
            if session_id and data.get("session_id") != session_id:
                continue

            observed_str = data.get("last_observed_at", "")
            delta_days = 0.0
            try:
                dt = datetime.fromisoformat(observed_str.replace("Z", "+00:00"))
                if dt.tzinfo is None:
                    dt = dt.replace(tzinfo=timezone.utc)
                delta_days = max(0.0, (now - dt).total_seconds() / 86400.0)
            except Exception:
                delta_days = 0.0

            base_w = float(data.get("weight", 1.0))
            decay_factor = math.exp(-0.05 * delta_days)
            eff_w = round(base_w * decay_factor, 2)
            rel = data.get("relation", "RELATES_TO")

            edge_width = min(6, max(1, int(eff_w * 2)))

            edges.append({
                "id": f"{u}_{rel}_{v}",
                "from": u,
                "to": v,
                "label": rel,
                "relation": rel,
                "font": {"size": 10, "color": "#94A3B8", "strokeWidth": 0, "align": "horizontal"},
                "arrows": "to",
                "color": {"color": "#334155", "highlight": "#38BDF8"},
                "width": edge_width,
                "weight": base_w,
                "effective_weight": eff_w,
                "context_snippet": data.get("context_snippet", ""),
                "session_id": data.get("session_id", ""),
                "last_observed_at": observed_str,
            })

        return {
            "nodes": nodes,
            "edges": edges,
            "stats": stats,
            "generated_at": datetime.now(timezone.utc).isoformat() + "Z",
        }

    def generate_html(self, session_id: Optional[str] = None, title: str = "Niskava Agent Market Intelligence") -> str:
        """Generate a complete, self-contained HTML page string."""
        data = self.export_graph_data(session_id=session_id)
        raw_json = json.dumps(data, ensure_ascii=False)
        sess_str = session_id or "All Active Sessions"

        html = HTML_TEMPLATE.replace("__TITLE__", title)
        html = html.replace("__SESSION_ID__", sess_str)
        html = html.replace("__RAW_JSON__", raw_json)
        return html

    def export_to_file(
        self,
        output_path: str = "~/.niskava/graph.html",
        session_id: Optional[str] = None,
        title: str = "Niskava Agent Market Intelligence",
    ) -> str:
        """Generate and save interactive HTML to disk."""
        path = os.path.expanduser(output_path)
        os.makedirs(os.path.dirname(os.path.abspath(path)), exist_ok=True)

        html_content = self.generate_html(session_id=session_id, title=title)
        with open(path, "w", encoding="utf-8") as f:
            f.write(html_content)

        return path