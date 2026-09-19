"""Cyber-OSINT Interactive Graph Visualizer for Niskava Agent.

Generates standalone, self-contained interactive HTML visualizations
with vis-network.js, dark cyber-forensic aesthetics, entity inspection,
search & focus, and ego-graph filtering.
"""

from datetime import datetime, timezone
import json
import math
import os
from typing import Any, Dict, List, Optional

from engine.memory.graph_memory import LocalGraphMemory


NODE_TYPE_STYLES = {
    "TICKER": {
        "color": {"background": "#00D2FF", "border": "#38BDF8", "highlight": {"background": "#38BDF8", "border": "#FFFFFF"}},
        "font": {"color": "#090D16", "face": "monospace", "size": 14, "bold": True},
        "shape": "dot",
        "size": 26,
    },
    "SECTOR": {
        "color": {"background": "#9D4EDD", "border": "#C77DFF", "highlight": {"background": "#C77DFF", "border": "#FFFFFF"}},
        "font": {"color": "#FFFFFF", "face": "sans-serif", "size": 13, "bold": True},
        "shape": "hexagon",
        "size": 22,
    },
    "BROKER": {
        "color": {"background": "#FFB703", "border": "#FCD34D", "highlight": {"background": "#FCD34D", "border": "#FFFFFF"}},
        "font": {"color": "#090D16", "face": "monospace", "size": 12, "bold": True},
        "shape": "triangle",
        "size": 22,
    },
    "PERSON": {
        "color": {"background": "#FB8500", "border": "#FFA726", "highlight": {"background": "#FFA726", "border": "#FFFFFF"}},
        "font": {"color": "#FFFFFF", "face": "sans-serif", "size": 12},
        "shape": "diamond",
        "size": 20,
    },
    "CATALYST_EVENT": {
        "color": {"background": "#06D6A0", "border": "#34D399", "highlight": {"background": "#34D399", "border": "#FFFFFF"}},
        "font": {"color": "#090D16", "face": "sans-serif", "size": 11, "bold": True},
        "shape": "box",
        "margin": 10,
        "size": 18,
    },
    "ANOMALY_METRIC": {
        "color": {"background": "#FF0055", "border": "#FB7185", "highlight": {"background": "#FB7185", "border": "#FFFFFF"}},
        "font": {"color": "#FFFFFF", "face": "monospace", "size": 11, "bold": True},
        "shape": "dot",
        "size": 22,
    },
    "USER": {
        "color": {"background": "#38BDF8", "border": "#BAE6FD", "highlight": {"background": "#BAE6FD", "border": "#FFFFFF"}},
        "font": {"color": "#090D16", "face": "sans-serif", "size": 14, "bold": True},
        "shape": "star",
        "size": 28,
    },
    "PRICE_LEVEL": {
        "color": {"background": "#475569", "border": "#94A3B8", "highlight": {"background": "#94A3B8", "border": "#FFFFFF"}},
        "font": {"color": "#F8FAFC", "face": "monospace", "size": 12},
        "shape": "ellipse",
        "size": 18,
    },
}

DEFAULT_NODE_STYLE = {
    "color": {"background": "#334155", "border": "#64748B", "highlight": {"background": "#64748B", "border": "#FFFFFF"}},
    "font": {"color": "#F1F5F9", "face": "sans-serif", "size": 12},
    "shape": "box",
    "margin": 8,
    "size": 18,
}

HTML_TEMPLATE = """<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>__TITLE__</title>
    <script type="text/javascript" src="https://unpkg.com/vis-network/standalone/umd/vis-network.min.js"></script>
    <style>
        :root {
            --bg: #090D16;
            --surface: #0F172A;
            --surface-card: #1E293B;
            --border: #334155;
            --text-main: #F8FAFC;
            --text-muted: #94A3B8;
            --accent: #00D2FF;
            --success: #10B981;
            --warning: #F59E0B;
            --danger: #EF4444;
        }
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            background: var(--bg);
            color: var(--text-main);
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
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
            font-size: 16px;
            font-weight: 800;
            color: var(--accent);
            letter-spacing: 0.5px;
            display: flex;
            align-items: center;
            gap: 8px;
        }
        .badge-live {
            background: rgba(0, 210, 255, 0.15);
            color: var(--accent);
            font-size: 10px;
            padding: 2px 8px;
            border-radius: 99px;
            border: 1px solid var(--accent);
        }
        .section-title {
            font-size: 11px;
            font-weight: 700;
            color: var(--text-muted);
            text-transform: uppercase;
            letter-spacing: 0.5px;
            margin-top: 6px;
        }
        .search-box {
            width: 100%;
            background: var(--surface-card);
            border: 1px solid var(--border);
            border-radius: 8px;
            padding: 10px 14px;
            color: #FFF;
            font-size: 13px;
            outline: none;
        }
        .search-box:focus {
            border-color: var(--accent);
            box-shadow: 0 0 0 2px rgba(0, 210, 255, 0.2);
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
        }
        .legend-dot {
            width: 12px;
            height: 12px;
            border-radius: 50%;
        }
        .stats-card {
            background: var(--surface-card);
            border: 1px solid var(--border);
            border-radius: 8px;
            padding: 12px;
            font-size: 12px;
            line-height: 1.6;
        }
        .stats-val {
            font-weight: bold;
            color: var(--accent);
        }
        .canvas-area {
            flex: 1;
            position: relative;
            background: radial-gradient(circle at center, #0F172A 0%, #090D16 100%);
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
            background: rgba(15, 23, 42, 0.85);
            backdrop-filter: blur(8px);
            border: 1px solid var(--border);
            border-radius: 99px;
            padding: 8px 16px;
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
            font-size: 16px;
            font-weight: 700;
            color: var(--accent);
            word-break: break-word;
        }
        .inspector-badge {
            display: inline-block;
            font-size: 11px;
            padding: 3px 8px;
            border-radius: 4px;
            font-weight: 700;
            background: var(--border);
            color: #FFF;
            width: fit-content;
        }
        .detail-row {
            font-size: 12px;
            display: flex;
            flex-direction: column;
            gap: 4px;
            border-bottom: 1px solid rgba(51, 65, 85, 0.4);
            padding-bottom: 8px;
        }
        .detail-label {
            color: var(--text-muted);
            font-size: 11px;
            text-transform: uppercase;
        }
        .btn {
            background: var(--surface-card);
            color: #FFF;
            border: 1px solid var(--border);
            padding: 8px 14px;
            border-radius: 6px;
            cursor: pointer;
            font-size: 12px;
            transition: all 0.2s;
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
            ⚡ NISKAVA GRAPH
            <span class="badge-live">OSINT</span>
        </div>

        <div class="section-title">Pencarian & Fokus</div>
        <input type="text" id="searchInput" class="search-box" placeholder="Cari emiten, broker, entitas..." oninput="searchAndHighlight()">

        <div class="section-title">Ego-Graph Radius</div>
        <div style="display: flex; gap: 8px;">
            <button class="btn" style="flex:1" onclick="filterHops(1)">1-Hop</button>
            <button class="btn" style="flex:1" onclick="filterHops(2)">2-Hop</button>
            <button class="btn" style="flex:1" onclick="resetFilter()">Reset</button>
        </div>

        <div class="section-title">Legenda Entitas</div>
        <div class="legend">
            <div class="legend-item"><span class="legend-dot" style="background:#00D2FF"></span> Emiten Saham (TICKER)</div>
            <div class="legend-item"><span class="legend-dot" style="background:#FFB703"></span> Broker Saham (BROKER)</div>
            <div class="legend-item"><span class="legend-dot" style="background:#9D4EDD"></span> Sektor Industri (SECTOR)</div>
            <div class="legend-item"><span class="legend-dot" style="background:#06D6A0"></span> Keterbukaan & Berita (CATALYST)</div>
            <div class="legend-item"><span class="legend-dot" style="background:#FF0055"></span> Anomali Kuantitatif (Z-SCORE)</div>
            <div class="legend-item"><span class="legend-dot" style="background:#38BDF8"></span> Pengguna Riset (USER)</div>
        </div>

        <div class="section-title">Statistik Graf</div>
        <div class="stats-card" id="statsArea">
            Memuat statistik graf...
        </div>

        <button class="btn" style="margin-top: auto;" onclick="if(network) network.fit({animation: true})">Reset Posisi Kamera</button>
    </div>

    <div class="canvas-area">
        <div class="top-bar">
            <div class="glass-pill">
                <span>Investigative Session:</span>
                <strong style="color:var(--accent);">__SESSION_ID__</strong>
            </div>
            <div class="glass-pill">
                <span>Model:</span>
                <code>NetworkX DiGraph + SQLite</code>
            </div>
        </div>
        <div id="network"></div>
    </div>

    <div class="inspector" id="inspectorPanel">
        <div class="section-title">Inspektor Entitas</div>
        <div id="inspectorContent">
            <p style="color:var(--text-muted); font-size:13px; line-height:1.6;">
                Klik salah satu simpul (node) atau garis (edge) di kanvas untuk memeriksa bukti kausalitas, bobot kebaruan (recency decay), dan metadata.
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
                    borderWidth: 2,
                    shadow: true
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
                return '• <strong>' + e.label + '</strong> (' + e.connections + ' koneksi)';
            }).join('<br>') || '-';

            const html = 'Simpul Total: <span class="stats-val">' + (s.total_nodes || allNodes.length) + '</span><br>' +
                         'Relasi Aktif: <span class="stats-val">' + (s.total_edges || allEdges.length) + '</span><br>' +
                         'Entitas Sentral: <br>' + entities;
            document.getElementById('statsArea').innerHTML = html;
        }

        function inspectNode(nodeId) {
            const node = allNodes.find(function(n) { return n.id === nodeId; });
            if (!node) return;

            const connectedEdges = allEdges.filter(function(e) { return e.from === nodeId || e.to === nodeId; });
            const panel = document.getElementById('inspectorContent');

            const edgeHTML = connectedEdges.map(function(e) {
                const connLabel = (e.from === nodeId) ? ('➔ ' + e.to) : ('⬅ ' + e.from);
                return '<div style="background:var(--surface-card); padding:8px 10px; border-radius:6px; font-size:12px;">' +
                       '<strong>[' + e.relation + ']</strong> ' + connLabel + '<br>' +
                       '<small style="color:var(--text-muted);">' + (e.context_snippet || '-') + '</small><br>' +
                       '<span style="font-size:10px; color:var(--accent);">W_eff: ' + e.effective_weight + '</span>' +
                       '</div>';
            }).join('');

            panel.innerHTML = '<div class="inspector-title">' + node.label + '</div>' +
                '<div class="inspector-badge" style="background:' + (node.color && node.color.background ? node.color.background : '#334155') + '">' + node.group + '</div>' +
                '<div class="detail-row" style="margin-top:12px;">' +
                '<span class="detail-label">Node ID</span><code>' + node.id + '</code></div>' +
                '<div class="detail-row"><span class="detail-label">Observasi Terakhir</span><span>' + (node.last_observed_at || '-') + '</span></div>' +
                '<div class="detail-row"><span class="detail-label">Jumlah Relasi Terhubung</span><span class="stats-val">' + connectedEdges.length + '</span></div>' +
                '<div class="section-title" style="margin-top:10px;">Relasi Kausalitas</div>' +
                '<div style="display:flex; flex-direction:column; gap:8px;">' + edgeHTML + '</div>';
        }

        function inspectEdge(edgeId) {
            const edge = allEdges.find(function(e) { return e.id === edgeId; });
            if (!edge) return;
            const panel = document.getElementById('inspectorContent');
            panel.innerHTML = '<div class="inspector-title">[' + edge.relation + ']</div>' +
                '<div class="inspector-badge" style="background:#06D6A0">RELATION</div>' +
                '<div class="detail-row" style="margin-top:12px;"><span class="detail-label">Koneksi</span><span>' + edge.from + ' ➔ ' + edge.to + '</span></div>' +
                '<div class="detail-row"><span class="detail-label">Kutipan Bukti</span><span>' + (edge.context_snippet || '-') + '</span></div>' +
                '<div class="detail-row"><span class="detail-label">Bobot Efektif (Recency Decay)</span><span class="stats-val">' + edge.effective_weight + ' (Base: ' + edge.weight + ')</span></div>' +
                '<div class="detail-row"><span class="detail-label">Sesi Terkait</span><code>' + (edge.session_id || '-') + '</code></div>';
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
                alert("Pilih satu simpul (node) terlebih dahulu sebelum menerapkan filter Ego-Graph hop.");
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

        window.onload = initNetwork;
    </script>
</body>
</html>
"""


class GraphVisualizer:
    """Exports graph data and renders standalone cyber-OSINT HTML visualizations."""

    def __init__(self, memory: Optional[LocalGraphMemory] = None):
        self.memory = memory or LocalGraphMemory()

    def export_graph_data(
        self,
        session_id: Optional[str] = None,
    ) -> Dict[str, Any]:
        """Convert SQLite memory graph into vis-network compatible node/edge datasets."""
        G = self.memory.load_graph()
        now = datetime.now(timezone.utc)

        nodes_list: List[Dict[str, Any]] = []
        for n_id, data in G.nodes(data=True):
            node_type = str(data.get("node_type", "ENTITY")).upper()
            label = str(data.get("label", n_id))
            meta = data.get("metadata", {})
            style = NODE_TYPE_STYLES.get(node_type, DEFAULT_NODE_STYLE)

            hover_title = f"<b>{label}</b><br>Tipe: {node_type}<br>ID: {n_id}"
            if data.get("last_observed_at"):
                hover_title += f"<br>Observasi: {data['last_observed_at']}"
            if meta:
                hover_title += f"<br>Meta: {json.dumps(meta, ensure_ascii=False)}"

            node_entry = {
                "id": n_id,
                "label": label,
                "group": node_type,
                "title": hover_title,
                "last_observed_at": data.get("last_observed_at", ""),
                "metadata": meta,
                **style,
            }
            nodes_list.append(node_entry)

        edges_list: List[Dict[str, Any]] = []
        for u, v, data in G.edges(data=True):
            edge_session = data.get("session_id", "")
            if session_id and edge_session and edge_session != session_id:
                continue

            relation = str(data.get("relation", "RELATES_TO"))
            context = str(data.get("context_snippet", ""))
            weight = float(data.get("weight", 1.0))
            observed_str = data.get("last_observed_at", "")

            delta_days = 0.0
            try:
                dt = datetime.fromisoformat(observed_str.replace("Z", "+00:00"))
                if dt.tzinfo is None:
                    dt = dt.replace(tzinfo=timezone.utc)
                delta_days = max(0.0, (now - dt).total_seconds() / 86400.0)
            except Exception:
                delta_days = 0.0

            decay_factor = math.exp(-self.memory.lambda_decay * delta_days)
            effective_weight = round(weight * decay_factor, 3)

            is_contradicted = "CONTRADICTED" in context
            is_uncertain = "UNCERTAIN" in context
            edge_color = "#38BDF8"
            dashes = False

            if is_contradicted:
                edge_color = "#EF4444"
                dashes = [4, 4]
            elif is_uncertain:
                edge_color = "#F59E0B"
                dashes = [6, 4]
            elif "CATALYZED_BY" in relation:
                edge_color = "#10B981"
            elif "TRIGGERED_ANOMALY" in relation:
                edge_color = "#F43F5E"

            width = max(1.2, min(6.0, 1.2 + (effective_weight * 1.5)))

            hover_title = (
                f"<b>{relation}</b><br>"
                f"Konteks: {context or '-'}<br>"
                f"Bobot Efektif: {effective_weight} (W0: {weight}, decay: {round(decay_factor, 2)})<br>"
                f"Usia: {round(delta_days, 1)} hari lalu"
            )

            edge_entry = {
                "id": f"{u}->{v}:{relation}",
                "from": u,
                "to": v,
                "label": relation,
                "title": hover_title,
                "width": width,
                "relation": relation,
                "context_snippet": context,
                "weight": weight,
                "effective_weight": effective_weight,
                "session_id": edge_session,
                "arrows": {"to": {"enabled": True, "scaleFactor": 0.8}},
                "color": {"color": edge_color, "highlight": "#FFFFFF"},
                "dashes": dashes,
                "font": {"color": "#94A3B8", "size": 10, "align": "middle", "strokeWidth": 0},
            }
            edges_list.append(edge_entry)

        stats = self.memory.get_graph_stats()

        return {
            "nodes": nodes_list,
            "edges": edges_list,
            "stats": stats,
            "session_id": session_id or "ALL",
            "generated_at": datetime.now(timezone.utc).isoformat() + "Z",
        }

    def generate_html(
        self,
        session_id: Optional[str] = None,
        title: str = "Niskava Agent — Cyber-OSINT Market Knowledge Graph",
    ) -> str:
        """Render self-contained HTML page embedding vis-network and full interactive controls."""
        graph_data = self.export_graph_data(session_id=session_id)
        raw_json = json.dumps(graph_data, ensure_ascii=False)

        rendered = HTML_TEMPLATE.replace("__TITLE__", title)
        rendered = rendered.replace("__SESSION_ID__", session_id or "ALL_SESSIONS")
        rendered = rendered.replace("__RAW_JSON__", raw_json)
        return rendered

    def export_to_file(
        self,
        output_path: str = "~/.niskava/graph.html",
        session_id: Optional[str] = None,
    ) -> str:
        """Write self-contained HTML visualization to disk."""
        target_path = os.path.expanduser(output_path)
        os.makedirs(os.path.dirname(os.path.abspath(target_path)), exist_ok=True)
        content = self.generate_html(session_id=session_id)
        with open(target_path, "w", encoding="utf-8") as f:
            f.write(content)
        return target_path