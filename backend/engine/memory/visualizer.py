"""Market Intelligence Interactive Graph Visualizer for Niskava Agent.

Generates standalone, self-contained interactive HTML visualizations
with vis-network.js, institutional financial aesthetics, entity inspection,
search & focus, and ego-graph filtering.
"""

from datetime import datetime, timezone
import json
import math
import os
import textwrap
from typing import Any, Dict, List, Optional

from engine.memory.graph_memory import LocalGraphMemory


def format_node_label(text: str, max_chars_per_line: int = 22, max_lines: int = 3) -> str:
    """Break long financial headlines or entity anomalies into balanced multi-line cards."""
    if not text:
        return ""
    clean_text = " ".join(str(text).split())
    if not clean_text:
        return ""
    lines = textwrap.wrap(
        clean_text,
        width=max_chars_per_line,
        max_lines=max_lines,
        placeholder="...",
    )
    return "\n".join(lines)


# Institutional financial market network styling - High-contrast knowledge graph
# Circular dot nodes with dynamic degree mass, crisp borders, and institutional dark void palette
NODE_TYPE_STYLES = {
    "TICKER": {
        "color": {
            "background": "#1E2329",
            "border": "#FCD535",
            "highlight": {"background": "#2B313A", "border": "#FFE270"},
            "hover": {"background": "#2B313A", "border": "#FFF0A0"},
        },
        "font": {"color": "#FCD535", "face": "Inter, -apple-system, sans-serif", "size": 11, "bold": True, "strokeWidth": 2, "strokeColor": "#0B0E14", "vadjust": 2},
        "shape": "dot",
        "margin": 10,
        "widthConstraint": {"maximum": 150, "minimum": 80},
        "size": 18,
    },
    "SECTOR": {
        "color": {
            "background": "#1E1B4B",
            "border": "#6366F1",
            "highlight": {"background": "#312E81", "border": "#A5B4FC"},
            "hover": {"background": "#312E81", "border": "#C7D2FE"},
        },
        "font": {"color": "#A5B4FC", "face": "Inter, -apple-system, sans-serif", "size": 11, "bold": True, "strokeWidth": 2, "strokeColor": "#0B0E14", "vadjust": 2},
        "shape": "dot",
        "margin": 10,
        "widthConstraint": {"maximum": 150, "minimum": 80},
        "size": 20,
    },
    "BROKER": {
        "color": {
            "background": "#2E1065",
            "border": "#8B5CF6",
            "highlight": {"background": "#4C1D95", "border": "#C4B5FD"},
            "hover": {"background": "#4C1D95", "border": "#DDD6FE"},
        },
        "font": {"color": "#C4B5FD", "face": "monospace", "size": 11, "bold": True, "strokeWidth": 2, "strokeColor": "#0B0E14", "vadjust": 2},
        "shape": "dot",
        "margin": 10,
        "widthConstraint": {"maximum": 150, "minimum": 80},
        "size": 16,
    },
    "PERSON": {
        "color": {
            "background": "#1F2937",
            "border": "#64748B",
            "highlight": {"background": "#374151", "border": "#CBD5E1"},
            "hover": {"background": "#374151", "border": "#E2E8F0"},
        },
        "font": {"color": "#E2E8F0", "face": "Inter, -apple-system, sans-serif", "size": 11, "strokeWidth": 2, "strokeColor": "#0B0E14", "vadjust": 2},
        "shape": "dot",
        "margin": 10,
        "widthConstraint": {"maximum": 150, "minimum": 80},
        "size": 14,
    },
    "CATALYST_EVENT": {
        "color": {
            "background": "#064E3B",
            "border": "#10B981",
            "highlight": {"background": "#065F46", "border": "#6EE7B7"},
            "hover": {"background": "#065F46", "border": "#A7F3D0"},
        },
        "font": {"color": "#6EE7B7", "face": "Inter, -apple-system, sans-serif", "size": 11, "bold": True, "strokeWidth": 2, "strokeColor": "#0B0E14", "vadjust": 2},
        "shape": "dot",
        "margin": 10,
        "widthConstraint": {"maximum": 150, "minimum": 80},
        "size": 16,
    },
    "ANOMALY_METRIC": {
        "color": {
            "background": "#450A0A",
            "border": "#EF4444",
            "highlight": {"background": "#7F1D1D", "border": "#FCA5A5"},
            "hover": {"background": "#7F1D1D", "border": "#FECACA"},
        },
        "font": {"color": "#FCA5A5", "face": "monospace", "size": 11, "bold": True, "strokeWidth": 2, "strokeColor": "#0B0E14", "vadjust": 2},
        "shape": "dot",
        "margin": 10,
        "widthConstraint": {"maximum": 150, "minimum": 80},
        "size": 17,
    },
    "USER": {
        "color": {
            "background": "#0F172A",
            "border": "#38BDF8",
            "highlight": {"background": "#1E293B", "border": "#7DD3FC"},
            "hover": {"background": "#1E293B", "border": "#BAE6FD"},
        },
        "font": {"color": "#38BDF8", "face": "Inter, -apple-system, sans-serif", "size": 12, "bold": True, "strokeWidth": 2, "strokeColor": "#0B0E14", "vadjust": 2},
        "shape": "dot",
        "margin": 10,
        "widthConstraint": {"maximum": 150, "minimum": 80},
        "size": 24,
    },
    "PRICE_LEVEL": {
        "color": {
            "background": "#1E293B",
            "border": "#94A3B8",
            "highlight": {"background": "#334155", "border": "#E2E8F0"},
            "hover": {"background": "#334155", "border": "#F1F5F9"},
        },
        "font": {"color": "#CBD5E1", "face": "monospace", "size": 11, "strokeWidth": 2, "strokeColor": "#0B0E14", "vadjust": 2},
        "shape": "dot",
        "margin": 10,
        "widthConstraint": {"maximum": 150, "minimum": 80},
        "size": 14,
    },
}

DEFAULT_NODE_STYLE = {
    "color": {
        "background": "#1E293B",
        "border": "#64748B",
        "highlight": {"background": "#334155", "border": "#CBD5E1"},
        "hover": {"background": "#334155", "border": "#E2E8F0"},
    },
    "font": {"color": "#E2E8F0", "face": "Inter, -apple-system, sans-serif", "size": 11, "strokeWidth": 2, "strokeColor": "#0B0E14", "vadjust": 2},
    "shape": "dot",
    "margin": 10,
    "widthConstraint": {"maximum": 150, "minimum": 80},
    "size": 15,
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
            --bg: #080B11;
            --surface: #0F131D;
            --surface-card: #151A26;
            --surface-hover: #1C2333;
            --border: #232C3E;
            --border-subtle: #18202F;
            --border-focus: #FCD535;
            --text-main: #F1F5F9;
            --text-secondary: #CBD5E1;
            --text-muted: #7E8B9E;
            --accent-gold: #FCD535;
            --accent-blue: #38BDF8;
            --accent-indigo: #6366F1;
            --accent-emerald: #10B981;
            --accent-rose: #EF4444;
            --accent-purple: #8B5CF6;
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

        /* Custom refined scrollbars */
        ::-webkit-scrollbar { width: 5px; height: 5px; }
        ::-webkit-scrollbar-track { background: transparent; }
        ::-webkit-scrollbar-thumb { background: #232C3E; border-radius: 4px; }
        ::-webkit-scrollbar-thumb:hover { background: #334155; }

        /* Left Sidebar */
        .sidebar {
            width: 320px;
            background: var(--surface);
            border-right: 1px solid var(--border);
            display: flex;
            flex-direction: column;
            padding: 18px;
            gap: 14px;
            overflow-y: auto;
            z-index: 20;
            box-shadow: 4px 0 24px rgba(0, 0, 0, 0.4);
            flex-shrink: 0;
            transition: transform 0.28s cubic-bezier(0.16, 1, 0.3, 1);
        }
        .brand {
            display: flex;
            align-items: center;
            justify-content: space-between;
            padding-bottom: 12px;
            border-bottom: 1px solid var(--border-subtle);
        }
        .brand-left-group {
            display: flex;
            align-items: center;
            gap: 10px;
        }
        .brand-logo-badge {
            width: 34px;
            height: 34px;
            border-radius: 9px;
            background: #FFFFFF;
            border: 1.5px solid rgba(252, 213, 53, 0.5);
            display: flex;
            align-items: center;
            justify-content: center;
            flex-shrink: 0;
            overflow: hidden;
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.25);
            transition: transform 0.2s cubic-bezier(0.16, 1, 0.3, 1), box-shadow 0.2s ease;
        }
        .brand-logo-badge:hover {
            transform: scale(1.05);
            box-shadow: 0 6px 16px rgba(252, 213, 53, 0.35);
        }
        .brand-logo-img {
            width: 100%;
            height: 100%;
            object-fit: cover;
            display: block;
        }
        .brand-title-wrap {
            display: flex;
            align-items: baseline;
            gap: 5px;
            line-height: 1.1;
        }
        .brand-name {
            font-size: 15px;
            font-weight: 800;
            color: #FFFFFF;
            letter-spacing: 0.5px;
        }
        .brand-tag {
            font-size: 11px;
            font-weight: 700;
            color: var(--accent-gold);
            letter-spacing: 0.4px;
        }
        .badge-live {
            background: rgba(252, 213, 53, 0.1);
            color: var(--accent-gold);
            font-size: 10px;
            font-weight: 600;
            padding: 2px 7px;
            border-radius: 4px;
            border: 1px solid rgba(252, 213, 53, 0.25);
            letter-spacing: 0.4px;
        }

        .section-title {
            font-size: 11px;
            font-weight: 700;
            color: var(--text-muted);
            text-transform: uppercase;
            letter-spacing: 0.6px;
            margin-top: 2px;
        }

        /* Search wrapper */
        .search-wrapper {
            position: relative;
            display: flex;
            align-items: center;
            width: 100%;
        }
        .search-icon {
            position: absolute;
            left: 10px;
            color: var(--text-muted);
            pointer-events: none;
        }
        .search-box {
            width: 100%;
            background: var(--surface-card);
            border: 1px solid var(--border);
            border-radius: 7px;
            padding: 8px 30px 8px 32px;
            color: #FFF;
            font-size: 12px;
            outline: none;
            transition: all 0.15s ease;
        }
        .search-box:focus {
            border-color: var(--border-focus);
            background: #19202E;
            box-shadow: 0 0 0 1px rgba(252, 213, 53, 0.2);
        }
        .search-clear {
            position: absolute;
            right: 8px;
            background: transparent;
            border: none;
            color: var(--text-muted);
            cursor: pointer;
            font-size: 14px;
            display: none;
            padding: 2px;
        }
        .search-clear:hover { color: #FFF; }

        /* Legend Classification */
        .legend {
            display: flex;
            flex-direction: column;
            gap: 4px;
        }
        .legend-item {
            display: flex;
            align-items: center;
            justify-content: space-between;
            padding: 5px 8px;
            border-radius: 6px;
            cursor: pointer;
            transition: background 0.15s ease;
        }
        .legend-item:hover {
            background: var(--surface-hover);
        }
        .legend-item.active {
            background: rgba(252, 213, 53, 0.12);
            border: 1px solid rgba(252, 213, 53, 0.35);
        }
        .legend-item.active .legend-item-left span:last-child {
            color: var(--accent-gold);
            font-weight: 700;
        }
        .legend-item-left {
            display: flex;
            align-items: center;
            gap: 9px;
            font-size: 11.5px;
            color: var(--text-secondary);
        }
        .legend-dot {
            width: 9px;
            height: 9px;
            border-radius: 50%;
            flex-shrink: 0;
        }
        .legend-count {
            font-family: monospace;
            font-size: 10.5px;
            color: var(--text-muted);
            background: #10141E;
            border: 1px solid var(--border-subtle);
            padding: 1px 6px;
            border-radius: 4px;
        }

        /* Stats Card & Leaderboard */
        .stats-card {
            background: var(--surface-card);
            border: 1px solid var(--border);
            border-radius: 8px;
            padding: 12px;
            font-size: 12px;
            line-height: 1.5;
        }
        .stats-kpi-grid {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 8px;
            margin-bottom: 12px;
        }
        .stats-kpi-card {
            background: #10141F;
            border: 1px solid var(--border-subtle);
            border-radius: 6px;
            padding: 8px 10px;
        }
        .stats-kpi-label {
            font-size: 10px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            color: var(--text-muted);
            font-weight: 600;
        }
        .stats-kpi-val {
            font-size: 18px;
            font-weight: 800;
            font-family: monospace;
            color: var(--text-main);
            margin-top: 2px;
        }
        .hub-title {
            font-size: 10.5px;
            font-weight: 700;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            color: var(--text-muted);
            margin-bottom: 8px;
        }
        .hub-list {
            display: flex;
            flex-direction: column;
            gap: 8px;
        }
        .hub-item {
            padding: 8px 10px;
            border-radius: 8px;
            background: linear-gradient(135deg, rgba(21, 26, 38, 0.7) 0%, rgba(16, 21, 32, 0.9) 100%);
            border: 1px solid var(--border-subtle);
            cursor: pointer;
            transition: all 0.18s cubic-bezier(0.16, 1, 0.3, 1);
            display: flex;
            flex-direction: column;
            gap: 5px;
        }
        .hub-item:hover {
            border-color: rgba(252, 213, 53, 0.4);
            background: #171E2D;
            transform: translateY(-1px);
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
        }
        .hub-item-top {
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 8px;
            width: 100%;
        }
        .hub-item-left {
            display: flex;
            align-items: center;
            gap: 7px;
            min-width: 0;
            flex: 1;
        }
        .hub-rank-badge {
            font-size: 9px;
            font-weight: 800;
            color: var(--accent-gold);
            background: rgba(252, 213, 53, 0.12);
            border: 1px solid rgba(252, 213, 53, 0.25);
            padding: 2px 5px;
            border-radius: 4px;
            flex-shrink: 0;
            letter-spacing: 0.3px;
        }
        .hub-label {
            color: var(--text-main);
            font-weight: 600;
            font-size: 11px;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
            display: inline-block;
        }
        .hub-conn {
            font-size: 10px;
            font-family: monospace;
            color: var(--accent-blue);
            background: rgba(56, 189, 248, 0.08);
            border: 1px solid rgba(56, 189, 248, 0.2);
            padding: 1px 6px;
            border-radius: 4px;
            flex-shrink: 0;
            white-space: nowrap;
        }
        .hub-meta-row {
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 6px;
        }
        .hub-bar-bg {
            flex: 1;
            height: 4px;
            background: #18202F;
            border-radius: 2px;
            overflow: hidden;
        }
        .hub-bar-fill {
            height: 100%;
            background: linear-gradient(90deg, var(--accent-blue), var(--accent-gold));
            border-radius: 2px;
        }
        .pr-pill {
            font-size: 9px;
            color: var(--text-muted);
            font-family: monospace;
            flex-shrink: 0;
        }

        /* Buttons */
        .btn {
            background: var(--surface-card);
            color: var(--text-main);
            border: 1px solid var(--border);
            padding: 8px 12px;
            border-radius: 6px;
            cursor: pointer;
            font-size: 12px;
            font-weight: 600;
            transition: all 0.15s ease;
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 6px;
        }
        .btn:hover {
            background: var(--border);
            border-color: var(--accent-gold);
        }
        .btn-refit {
            margin-top: auto;
            background: #171E2D;
            border-color: #2D3950;
        }

        /* Canvas Area */
        .canvas-area {
            flex: 1;
            position: relative;
            background-color: #07090E;
            background-image:
                radial-gradient(circle at 50% 50%, rgba(26, 36, 56, 0.45) 0%, rgba(7, 9, 14, 0.98) 100%),
                radial-gradient(rgba(148, 163, 184, 0.08) 1px, transparent 1px);
            background-size: 100% 100%, 28px 28px;
            overflow: hidden;
        }
        #network {
            width: 100%;
            height: 100%;
        }

        /* Top HUD Bar */
        .top-bar {
            position: absolute;
            top: 14px;
            left: 16px;
            right: 16px;
            display: flex;
            align-items: center;
            justify-content: space-between;
            pointer-events: none;
            z-index: 15;
            gap: 12px;
        }
        .top-bar-left, .top-bar-center, .top-bar-right {
            display: flex;
            align-items: center;
            gap: 8px;
        }
        .hud-pill {
            background: rgba(15, 19, 29, 0.88);
            border: 1px solid var(--border);
            backdrop-filter: blur(14px);
            border-radius: 7px;
            padding: 6px 12px;
            font-size: 12px;
            color: var(--text-main);
            pointer-events: auto;
            display: flex;
            align-items: center;
            gap: 8px;
            box-shadow: 0 4px 16px rgba(0, 0, 0, 0.4);
        }
        .hud-dot {
            width: 7px;
            height: 7px;
            border-radius: 50%;
            background: var(--accent-emerald);
            box-shadow: 0 0 8px rgba(16, 185, 129, 0.6);
            animation: pulse-dot 2.5s infinite ease-in-out;
        }
        @keyframes pulse-dot {
            0%, 100% { opacity: 1; transform: scale(1); }
            50% { opacity: 0.35; transform: scale(0.85); }
        }
        .hud-label {
            color: var(--text-muted);
            font-size: 11px;
            font-weight: 500;
        }
        .hud-val {
            font-weight: 700;
            color: var(--accent-gold);
            font-size: 11.5px;
        }
        .hud-code {
            color: var(--accent-blue);
            font-size: 11px;
            font-family: monospace;
            background: rgba(56, 189, 248, 0.08);
            padding: 1px 6px;
            border-radius: 4px;
            border: 1px solid rgba(56, 189, 248, 0.2);
        }
        .telemetry-item {
            display: flex;
            align-items: center;
            gap: 5px;
            font-size: 11px;
        }
        .telemetry-label { color: var(--text-muted); }
        .telemetry-val { font-weight: 700; font-family: monospace; color: var(--text-main); }
        .telemetry-sep { width: 1px; height: 12px; background: var(--border); }

        .hud-btn {
            background: rgba(15, 19, 29, 0.88);
            border: 1px solid var(--border);
            backdrop-filter: blur(14px);
            border-radius: 7px;
            padding: 6px 12px;
            font-size: 11.5px;
            font-weight: 600;
            color: var(--text-main);
            pointer-events: auto;
            cursor: pointer;
            display: flex;
            align-items: center;
            gap: 6px;
            box-shadow: 0 4px 16px rgba(0, 0, 0, 0.4);
            transition: all 0.15s ease;
        }
        .hud-btn:hover {
            background: var(--surface-hover);
            border-color: var(--accent-gold);
            color: var(--accent-gold);
        }
        .hud-btn.active {
            background: #1B2433;
            border-color: var(--accent-blue);
            color: var(--accent-blue);
        }

        /* Floating canvas zoom controls */
        .canvas-controls {
            position: absolute;
            bottom: 20px;
            right: 20px;
            display: flex;
            flex-direction: column;
            gap: 6px;
            z-index: 15;
        }
        .control-btn {
            width: 32px;
            height: 32px;
            background: rgba(15, 19, 29, 0.88);
            border: 1px solid var(--border);
            backdrop-filter: blur(12px);
            border-radius: 7px;
            color: var(--text-main);
            cursor: pointer;
            display: flex;
            align-items: center;
            justify-content: center;
            box-shadow: 0 4px 12px rgba(0,0,0,0.35);
            transition: all 0.15s ease;
        }
        .control-btn:hover {
            border-color: var(--accent-gold);
            color: var(--accent-gold);
            background: var(--surface-hover);
        }

        /* Toast notification */
        .toast-msg {
            position: fixed;
            bottom: 24px;
            left: 50%;
            transform: translateX(-50%) translateY(20px);
            background: rgba(15, 23, 42, 0.95);
            border: 1px solid var(--accent-gold);
            color: var(--text-main);
            padding: 8px 16px;
            border-radius: 8px;
            font-size: 11.5px;
            font-weight: 600;
            box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5);
            backdrop-filter: blur(12px);
            opacity: 0;
            transition: all 0.25s cubic-bezier(0.16, 1, 0.3, 1);
            pointer-events: none;
            z-index: 100;
        }
        .toast-msg.show {
            opacity: 1;
            transform: translateX(-50%) translateY(0);
        }

        /* Right Inspector Panel - Knowledge Inspector */
        .inspector {
            width: 340px;
            background: var(--surface);
            border-left: 1px solid var(--border);
            display: flex;
            flex-direction: column;
            padding: 18px;
            gap: 14px;
            overflow-y: auto;
            z-index: 20;
            box-shadow: -4px 0 24px rgba(0, 0, 0, 0.4);
            flex-shrink: 0;
            transition: transform 0.28s cubic-bezier(0.16, 1, 0.3, 1);
        }
        .inspector-header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            padding-bottom: 10px;
            border-bottom: 1px solid var(--border-subtle);
        }
        .inspector-status {
            font-size: 10px;
            font-weight: 600;
            color: var(--accent-emerald);
            background: rgba(16, 185, 129, 0.1);
            padding: 2px 7px;
            border-radius: 4px;
            border: 1px solid rgba(16, 185, 129, 0.25);
        }
        .inspector-empty {
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            text-align: center;
            padding: 40px 10px;
            gap: 10px;
            color: var(--text-muted);
        }
        .inspector-title {
            font-size: 16px;
            font-weight: 800;
            color: var(--text-main);
            word-break: break-word;
            line-height: 1.3;
        }
        .inspector-badge {
            display: inline-block;
            font-size: 10.5px;
            padding: 2px 9px;
            border-radius: 4px;
            font-weight: 700;
            letter-spacing: 0.4px;
            width: fit-content;
            margin-top: 4px;
        }
        .detail-grid {
            display: flex;
            flex-direction: column;
            gap: 8px;
            margin-top: 10px;
        }
        .detail-row {
            font-size: 11.5px;
            display: flex;
            flex-direction: column;
            gap: 3px;
            background: var(--surface-card);
            border: 1px solid var(--border-subtle);
            border-radius: 6px;
            padding: 7px 10px;
        }
        .detail-label {
            color: var(--text-muted);
            font-size: 10px;
            text-transform: uppercase;
            font-weight: 600;
            letter-spacing: 0.4px;
        }
        .relation-card {
            background: var(--surface-card);
            border: 1px solid var(--border-subtle);
            border-radius: 6px;
            padding: 9px 11px;
            font-size: 11.5px;
            display: flex;
            flex-direction: column;
            gap: 4px;
            cursor: pointer;
            transition: all 0.15s ease;
        }
        .relation-card:hover {
            border-color: rgba(252, 213, 53, 0.45);
            background: #171E2E;
            transform: translateY(-1px);
        }
        .relation-header {
            display: flex;
            align-items: center;
            justify-content: space-between;
        }
        .relation-type-tag {
            font-family: monospace;
            font-size: 10.5px;
            font-weight: 700;
            color: var(--accent-gold);
        }
        .weight-meter {
            display: flex;
            align-items: center;
            gap: 6px;
            font-size: 10px;
            color: var(--accent-blue);
            font-family: monospace;
        }
        .weight-bar-bg {
            flex: 1;
            height: 3px;
            background: #1C2333;
            border-radius: 2px;
            overflow: hidden;
        }
        .weight-bar-fill {
            height: 100%;
            background: var(--accent-blue);
            border-radius: 2px;
        }
        .causality-card {
            background: var(--surface-card);
            border: 1px solid var(--border);
            border-radius: 7px;
            padding: 10px 12px;
            margin-top: 10px;
            display: flex;
            flex-direction: column;
            gap: 6px;
        }
        .causality-title {
            font-size: 10px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            color: var(--text-muted);
            font-weight: 700;
        }
        .causality-flow-row {
            display: flex;
            align-items: center;
            gap: 8px;
            flex-wrap: wrap;
        }
        .causality-badge {
            background: #141B28;
            border: 1px solid var(--border);
            color: var(--accent-gold);
            font-size: 11px;
            font-weight: 700;
            padding: 4px 8px;
            border-radius: 5px;
            cursor: pointer;
            transition: all 0.15s ease;
            max-width: 140px;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }
        .causality-badge:hover {
            border-color: var(--accent-gold);
            background: #1E273A;
            transform: translateY(-1px);
        }
        .causality-arrow {
            color: var(--accent-blue);
            font-weight: 800;
            font-size: 13px;
        }
        .evidence-quote {
            color: var(--text-secondary);
            font-size: 11.5px;
            line-height: 1.5;
            background: #0B0E16;
            border-left: 3px solid var(--accent-gold);
            padding: 6px 10px;
            border-radius: 0 4px 4px 0;
            margin-top: 3px;
            word-break: break-word;
        }

        /* Drawer Backdrop Overlay & Controls */
        .drawer-backdrop {
            display: none;
            position: fixed;
            top: 0;
            left: 0;
            width: 100vw;
            height: 100vh;
            background: rgba(8, 11, 17, 0.72);
            backdrop-filter: blur(4px);
            z-index: 25;
            opacity: 0;
            transition: opacity 0.25s ease;
            pointer-events: none;
        }
        .drawer-backdrop.active {
            display: block;
            opacity: 1;
            pointer-events: auto;
        }

        .drawer-close-btn {
            background: transparent;
            border: 1px solid var(--border-subtle);
            color: var(--text-muted);
            font-size: 13px;
            cursor: pointer;
            padding: 2px 7px;
            border-radius: 4px;
            display: none;
            line-height: 1;
            transition: all 0.15s ease;
        }
        .drawer-close-btn:hover {
            color: var(--text-main);
            border-color: var(--border);
            background: var(--surface-card);
        }

        .drawer-toggle-btn {
            display: none;
        }

        /* Responsive Media Queries */
        @media (max-width: 1024px) {
            .inspector {
                position: fixed;
                top: 0;
                right: 0;
                bottom: 0;
                height: 100vh;
                width: 360px;
                max-width: 85vw;
                z-index: 30;
                transform: translateX(100%);
                box-shadow: -8px 0 32px rgba(0, 0, 0, 0.6);
            }
            .inspector.open, .inspector.drawer-open {
                transform: translateX(0);
            }
            .inspector .drawer-close-btn {
                display: inline-flex;
                align-items: center;
                justify-content: center;
            }
            #inspectorToggleBtn {
                display: flex;
            }
            .top-bar-left .hud-pill:nth-child(2) {
                display: none;
            }
        }

        @media (max-width: 768px) {
            .sidebar {
                position: fixed;
                top: 0;
                left: 0;
                bottom: 0;
                height: 100vh;
                width: 320px;
                max-width: 85vw;
                z-index: 30;
                transform: translateX(-100%);
                box-shadow: 8px 0 32px rgba(0, 0, 0, 0.6);
            }
            .sidebar.open, .sidebar.drawer-open {
                transform: translateX(0);
            }
            .sidebar .drawer-close-btn {
                display: inline-flex;
                align-items: center;
                justify-content: center;
            }
            #sidebarToggleBtn {
                display: flex;
            }
            .canvas-area {
                width: 100vw;
                height: 100vh;
                flex: 1 1 100%;
            }
            .top-bar {
                top: 8px;
                left: 8px;
                right: 8px;
                gap: 6px;
                flex-wrap: wrap;
            }
            .top-bar-center {
                display: none;
            }
            .hud-pill {
                padding: 4px 8px;
                font-size: 11px;
                gap: 5px;
            }
            .hud-btn {
                padding: 4px 8px;
                font-size: 11px;
                gap: 5px;
            }
            .hud-btn span.btn-label-optional {
                display: none;
            }
        }
    </style>
</head>
<body>
    <div class="drawer-backdrop" id="drawerBackdrop" onclick="closeDrawers()"></div>
    <div class="sidebar">
        <div class="brand">
            <div class="brand-left-group">
                <div class="brand-logo-badge" title="Niskava Agent Market Intelligence">
                    <img src="data:image/png;base64,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" alt="Niskava" class="brand-logo-img">
                </div>
                <div class="brand-title-wrap">
                    <span class="brand-name">Niskava</span>
                    <span class="brand-tag">Agent</span>
                </div>
            </div>
            <div style="display:flex; align-items:center; gap:8px;">
                <span class="badge-live">Market Intelligence</span>
                <button class="drawer-close-btn" onclick="closeDrawers()" title="Close Sidebar">✕</button>
            </div>
        </div>

        <div class="section-title">Entity Search</div>
        <div class="search-wrapper">
            <svg class="search-icon" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
            <input type="text" id="searchInput" class="search-box" placeholder="Search ticker, broker, disclosure..." oninput="searchAndHighlight()">
            <button class="search-clear" id="searchClearBtn" onclick="clearSearch()" title="Clear">✕</button>
        </div>

        <div class="section-title">Node Classification</div>
        <div class="legend" id="legendContainer">
            <div class="legend-item" onclick="filterByGroup('TICKER')">
                <div class="legend-item-left">
                    <span class="legend-dot" style="background:#FCD535; box-shadow:0 0 8px rgba(252,213,53,0.45)"></span>
                    <span>Stock Issuer (TICKER)</span>
                </div>
                <span class="legend-count" id="count-TICKER">-</span>
            </div>
            <div class="legend-item" onclick="filterByGroup('BROKER')">
                <div class="legend-item-left">
                    <span class="legend-dot" style="background:#8B5CF6; box-shadow:0 0 8px rgba(139,92,246,0.45)"></span>
                    <span>Exchange Member (BROKER)</span>
                </div>
                <span class="legend-count" id="count-BROKER">-</span>
            </div>
            <div class="legend-item" onclick="filterByGroup('SECTOR')">
                <div class="legend-item-left">
                    <span class="legend-dot" style="background:#6366F1; box-shadow:0 0 8px rgba(99,102,241,0.45)"></span>
                    <span>Industry Sector (SECTOR)</span>
                </div>
                <span class="legend-count" id="count-SECTOR">-</span>
            </div>
            <div class="legend-item" onclick="filterByGroup('CATALYST_EVENT')">
                <div class="legend-item-left">
                    <span class="legend-dot" style="background:#10B981; box-shadow:0 0 8px rgba(16,185,129,0.45)"></span>
                    <span>Disclosures & Corporate Actions</span>
                </div>
                <span class="legend-count" id="count-CATALYST_EVENT">-</span>
            </div>
            <div class="legend-item" onclick="filterByGroup('ANOMALY_METRIC')">
                <div class="legend-item-left">
                    <span class="legend-dot" style="background:#EF4444; box-shadow:0 0 8px rgba(239,68,68,0.45)"></span>
                    <span>Volume Outlier & Fund Flow</span>
                </div>
                <span class="legend-count" id="count-ANOMALY_METRIC">-</span>
            </div>
            <div class="legend-item" onclick="filterByGroup('USER')">
                <div class="legend-item-left">
                    <span class="legend-dot" style="background:#38BDF8; box-shadow:0 0 8px rgba(56,189,248,0.45)"></span>
                    <span>User Research Profile (USER)</span>
                </div>
                <span class="legend-count" id="count-USER">-</span>
            </div>
        </div>

        <div class="section-title">Market Network Statistics</div>
        <div class="stats-card" id="statsArea">
            Loading graph statistics...
        </div>

        <button class="btn btn-refit" onclick="fitView()">
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3"></path></svg>
            <span>Refit Canvas</span>
        </button>
    </div>

    <div class="canvas-area">
        <div class="top-bar">
            <div class="top-bar-left">
                <button class="hud-btn drawer-toggle-btn" id="sidebarToggleBtn" onclick="toggleSidebar()" title="Toggle Sidebar">
                    <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><line x1="3" y1="12" x2="21" y2="12"></line><line x1="3" y1="6" x2="21" y2="6"></line><line x1="3" y1="18" x2="21" y2="18"></line></svg>
                    <span>Menu</span>
                </button>
                <div class="hud-pill">
                    <span class="hud-dot"></span>
                    <span class="hud-label">Research Session:</span>
                    <strong class="hud-val">__SESSION_ID__</strong>
                </div>
                <div class="hud-pill">
                    <span class="hud-label">Engine:</span>
                    <code class="hud-code">NetworkX DiGraph + Local SQLite WAL</code>
                </div>
            </div>
            <div class="top-bar-center">
                <div class="hud-pill">
                    <div class="telemetry-item">
                        <span class="telemetry-label">Entities:</span>
                        <span class="telemetry-val" id="hudTotalNodes">0</span>
                    </div>
                    <div class="telemetry-sep"></div>
                    <div class="telemetry-item">
                        <span class="telemetry-label">Relations:</span>
                        <span class="telemetry-val" id="hudTotalEdges">0</span>
                    </div>
                    <div class="telemetry-sep"></div>
                    <div class="telemetry-item">
                        <span class="telemetry-label">Central Hub:</span>
                        <span class="telemetry-val" id="hudTopHub" style="color:var(--accent-gold);">-</span>
                    </div>
                </div>
            </div>
            <div class="top-bar-right">
                <button class="hud-btn" id="physicsToggleBtn" onclick="togglePhysics()" title="Toggle physics drift simulation">
                    <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"></circle><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path></svg>
                    <span id="physicsStatusText">Physics: Settled</span>
                </button>
                <button class="hud-btn" onclick="fitView()" title="Fit view to all entities">
                    <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 3h6v6M9 21H3v-6M21 3l-7 7M3 21l7-7"></path></svg>
                    <span class="btn-label-optional">Fit Canvas</span>
                </button>
                <button class="hud-btn drawer-toggle-btn" id="inspectorToggleBtn" onclick="toggleInspector()" title="Toggle Inspector Panel">
                    <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="16" y1="13" x2="8" y2="13"></line><line x1="16" y1="17" x2="8" y2="17"></line><polyline points="10 9 9 9 8 9"></polyline></svg>
                    <span>Inspector</span>
                </button>
            </div>
        </div>

        <div id="network"></div>

        <div class="canvas-controls">
            <button class="control-btn" onclick="zoomIn()" title="Zoom In">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
            </button>
            <button class="control-btn" onclick="zoomOut()" title="Zoom Out">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2"><line x1="5" y1="12" x2="19" y2="12"></line></svg>
            </button>
            <button class="control-btn" onclick="fitView()" title="Center & Fit">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3"></path></svg>
            </button>
        </div>
    </div>

    <div class="inspector" id="inspectorPanel">
        <div class="inspector-header">
            <div class="section-title">Knowledge Inspector</div>
            <div style="display:flex; align-items:center; gap:8px;">
                <span class="inspector-status">IDXnet Verified</span>
                <button class="drawer-close-btn" onclick="closeDrawers()" title="Close Inspector">✕</button>
            </div>
        </div>
        <div id="inspectorContent">
            <div class="inspector-empty">
                <svg width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round" style="color:var(--text-muted); opacity:0.5;"><circle cx="12" cy="12" r="3"></circle><circle cx="19" cy="5" r="2"></circle><circle cx="5" cy="19" r="2"></circle><path d="M10.4 10.4 6.4 17.6"></path><path d="M13.6 13.6 17.6 6.4"></path></svg>
                <p style="color:var(--text-muted); font-size:12px; line-height:1.6;">
                    Select an entity or relation on the canvas to inspect verified IDX disclosures, market causality, and metadata.
                </p>
            </div>
        </div>
    </div>

    <script type="text/javascript">
        const rawData = __RAW_JSON__;
        let network = null;
        let allNodes = rawData.nodes || [];
        let allEdges = rawData.edges || [];
        let nodesDataSet = new vis.DataSet(allNodes);
        let edgesDataSet = new vis.DataSet(allEdges);
        let physicsEnabled = true;

        function initNetwork() {
            const container = document.getElementById('network');
            const data = { nodes: nodesDataSet, edges: edgesDataSet };

            // Dynamic neural physics: scale repulsion and spring length with node density
            const nodeCount = allNodes.length;
            const dynamicSpring = Math.max(220, Math.min(480, 160 + Math.sqrt(nodeCount) * 36));
            const dynamicGrav = Math.min(-180, -90 - nodeCount * 3.6);

            const options = {
                nodes: {
                    shape: 'dot',
                    borderWidth: 2,
                    borderWidthSelected: 3.5,
                    shadow: {
                        enabled: true,
                        color: 'rgba(0, 0, 0, 0.55)',
                        size: 10,
                        x: 0,
                        y: 3
                    }
                },
                edges: {
                    smooth: { type: 'continuous', roundness: 0.18 },
                    selectionWidth: 4,
                    hoverWidth: 3
                },
                physics: {
                    enabled: true,
                    solver: 'forceAtlas2Based',
                    forceAtlas2Based: {
                        gravitationalConstant: dynamicGrav,
                        centralGravity: 0.007,
                        springLength: dynamicSpring,
                        springConstant: 0.045,
                        damping: 0.42,
                        avoidOverlap: 0.85
                    },
                    stabilization: {
                        enabled: true,
                        iterations: 180,
                        updateInterval: 25,
                        fit: true
                    }
                },
                interaction: {
                    hover: true,
                    hoverConnectedEdges: false,
                    edgeThreshold: 20,
                    tooltipDelay: 90,
                    hideEdgesOnDrag: false,
                    navigationButtons: false,
                    keyboard: true
                }
            };

            network = new vis.Network(container, data, options);

            network.once('stabilizationIterationsDone', function () {
                network.setOptions({ physics: { enabled: false } });
                physicsEnabled = false;
                const statusText = document.getElementById('physicsStatusText');
                if (statusText) statusText.textContent = 'Physics: Settled';
                const toggleBtn = document.getElementById('physicsToggleBtn');
                if (toggleBtn) toggleBtn.classList.remove('active');
                fitView();
            });

            window.addEventListener('resize', function() {
                if (window.innerWidth > 1024) {
                    closeDrawers();
                }
                fitView();
            });

            let isFocused = false;

            network.on("hoverNode", function (params) {
                if (isFocused) return;
                highlightNeighbors(params.node);
            });

            network.on("blurNode", function () {
                if (isFocused) return;
                restoreGraphView();
            });

            network.on("click", function (params) {
                if (params.nodes && params.nodes.length > 0) {
                    isFocused = true;
                    highlightNeighbors(params.nodes[0]);
                    inspectNode(params.nodes[0]);
                } else if (params.edges && params.edges.length > 0) {
                    isFocused = true;
                    const clickedEdgeId = params.edges[0];
                    highlightEdge(clickedEdgeId);
                    inspectEdge(clickedEdgeId);
                } else {
                    isFocused = false;
                    restoreGraphView();
                    showDefaultInspector();
                }
            });

            // Delegated click handler for Inspector elements (XSS-safe & quote-escaping safe)
            const inspectorPanelEl = document.getElementById('inspectorPanel');
            if (inspectorPanelEl) {
                inspectorPanelEl.addEventListener('click', function(e) {
                    const focusBtn = e.target.closest('[data-focus-entity]');
                    if (focusBtn) {
                        const ent = focusBtn.getAttribute('data-focus-entity');
                        if (ent) focusOnEntity(ent);
                        return;
                    }
                    const relCard = e.target.closest('[data-edge-id]');
                    if (relCard) {
                        const eid = relCard.getAttribute('data-edge-id');
                        if (eid) {
                            highlightEdge(eid);
                            inspectEdge(eid);
                        }
                        return;
                    }
                });
            }

            renderStats();
        }

        function highlightEdge(edgeId) {
            const edge = allEdges.find(function(e) { return e.id === edgeId; });
            if (!edge) return;

            const endpointIds = new Set([edge.from, edge.to]);

            const nodeUpdates = allNodes.map(function (n) {
                if (endpointIds.has(n.id)) {
                    return { id: n.id, opacity: 1.0 };
                } else {
                    return { id: n.id, opacity: 0.14 };
                }
            });

            const edgeUpdates = allEdges.map(function (e) {
                if (e.id === edgeId) {
                    return {
                        id: e.id,
                        color: { color: '#FCD535', highlight: '#FCD535' },
                        width: (e.width || 2) + 2.5,
                        opacity: 1.0
                    };
                } else {
                    return {
                        id: e.id,
                        color: { color: 'rgba(148, 163, 184, 0.05)' },
                        opacity: 0.05
                    };
                }
            });

            nodesDataSet.update(nodeUpdates);
            edgesDataSet.update(edgeUpdates);
        }

        function highlightNeighbors(nodeId) {
            const connectedNodeIds = new Set(network.getConnectedNodes(nodeId));
            connectedNodeIds.add(nodeId);
            const connectedEdgeIds = new Set(network.getConnectedEdges(nodeId));

            const nodeUpdates = allNodes.map(function (n) {
                if (connectedNodeIds.has(n.id)) {
                    return { id: n.id, opacity: 1.0 };
                } else {
                    return { id: n.id, opacity: 0.14 };
                }
            });

            const edgeUpdates = allEdges.map(function (e) {
                if (connectedEdgeIds.has(e.id)) {
                    return {
                        id: e.id,
                        color: { color: '#FCD535', highlight: '#FCD535' },
                        width: (e.width || 2) + 2.0,
                        opacity: 1.0
                    };
                } else {
                    return {
                        id: e.id,
                        color: { color: 'rgba(148, 163, 184, 0.05)' },
                        opacity: 0.05
                    };
                }
            });

            nodesDataSet.update(nodeUpdates);
            edgesDataSet.update(edgeUpdates);
        }

        function restoreGraphView() {
            const nodeResets = allNodes.map(function (n) {
                return { id: n.id, opacity: 1.0 };
            });
            const edgeResets = allEdges.map(function (e) {
                return { id: e.id, color: e.color, width: e.width, opacity: 1.0 };
            });
            nodesDataSet.update(nodeResets);
            edgesDataSet.update(edgeResets);
        }

        function showDefaultInspector() {
            const panel = document.getElementById('inspectorContent');
            if (!panel) return;
            panel.innerHTML = '<div class="inspector-empty">' +
                '  <svg width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round" style="color:var(--text-muted); opacity:0.5;"><circle cx="12" cy="12" r="3"></circle><circle cx="19" cy="5" r="2"></circle><circle cx="5" cy="19" r="2"></circle><path d="M10.4 10.4 6.4 17.6"></path><path d="M13.6 13.6 17.6 6.4"></path></svg>' +
                '  <p style="color:var(--text-muted); font-size:12px; line-height:1.6;">' +
                '    Select an entity or relation on the canvas to inspect verified IDX disclosures, market causality, and metadata.' +
                '  </p>' +
                '</div>';
        }

        function renderStats() {
            const s = rawData.stats || {};
            const totalN = s.total_nodes || allNodes.length;
            const totalE = s.total_edges || allEdges.length;

            // Update top bar telemetry
            const hudN = document.getElementById('hudTotalNodes');
            if (hudN) hudN.textContent = totalN;
            const hudE = document.getElementById('hudTotalEdges');
            if (hudE) hudE.textContent = totalE;

            // Update Legend counts
            const counts = {};
            allNodes.forEach(function (n) {
                counts[n.group] = (counts[n.group] || 0) + 1;
            });
            ['TICKER', 'BROKER', 'SECTOR', 'CATALYST_EVENT', 'ANOMALY_METRIC', 'USER'].forEach(function (g) {
                const el = document.getElementById('count-' + g);
                if (el) el.textContent = counts[g] || 0;
            });

            const topEntities = (s.top_central_entities || []);
            const hudHub = document.getElementById('hudTopHub');
            if (hudHub && topEntities.length > 0) {
                hudHub.textContent = topEntities[0].label;
            }

            const maxConn = topEntities.length > 0 ? (topEntities[0].connections || 1) : 1;

            let hubListHTML = '';
            if (topEntities.length > 0) {
                hubListHTML = topEntities.slice(0, 5).map(function(e, idx) {
                    const pct = Math.min(100, Math.round((e.connections / maxConn) * 100));
                    const pr = e.pagerank ? '<span class="pr-pill">PR: ' + e.pagerank + '</span>' : '';
                    const rawLabel = e.label || '';
                    const safeLabel = rawLabel.replace(/"/g, '&quot;');
                    // Clean long anomaly or corporate disclosure titles for concise display
                    let displayLabel = rawLabel;
                    if (displayLabel.startsWith('VOLUME_AND_PRICE_SURGE_')) {
                        const parts = displayLabel.split('_');
                        displayLabel = 'SURGE ' + (parts[4] || '');
                    } else if (displayLabel.length > 24) {
                        displayLabel = displayLabel.slice(0, 22) + '...';
                    }
                    const safeDisplay = displayLabel.replace(/</g, '&lt;').replace(/>/g, '&gt;');

                    return '<div class="hub-item" data-entity="' + safeLabel + '" title="' + safeLabel + '">' +
                           '  <div class="hub-item-top">' +
                           '    <div class="hub-item-left">' +
                           '      <span class="hub-rank-badge">#' + (idx + 1) + '</span>' +
                           '      <strong class="hub-label">' + safeDisplay + '</strong>' +
                           '    </div>' +
                           '    <span class="hub-conn">' + e.connections + ' links</span>' +
                           '  </div>' +
                           '  <div class="hub-meta-row">' +
                           '    <div class="hub-bar-bg"><div class="hub-bar-fill" style="width:' + pct + '%;"></div></div>' +
                           '    ' + pr +
                           '  </div>' +
                           '</div>';
                }).join('');
            } else {
                hubListHTML = '<div style="font-size:11px; color:var(--text-muted);">-</div>';
            }

            const html = '<div class="stats-kpi-grid">' +
                         '  <div class="stats-kpi-card">' +
                         '    <div class="stats-kpi-label">Total Nodes</div>' +
                         '    <div class="stats-kpi-val stats-val">' + totalN + '</div>' +
                         '  </div>' +
                         '  <div class="stats-kpi-card">' +
                         '    <div class="stats-kpi-label">Market Relations</div>' +
                         '    <div class="stats-kpi-val stats-val">' + totalE + '</div>' +
                         '  </div>' +
                         '</div>' +
                         '<div class="hub-title">Central Entities & Hubs</div>' +
                         '<div class="hub-list">' + hubListHTML + '</div>';

            const statsAreaEl = document.getElementById('statsArea');
            if (statsAreaEl) {
                statsAreaEl.innerHTML = html;
                statsAreaEl.querySelectorAll('.hub-item[data-entity]').forEach(function(item) {
                    item.addEventListener('click', function() {
                        focusOnEntity(this.getAttribute('data-entity'));
                    });
                });
            }
        }

        function inspectNode(nodeId) {
            const node = allNodes.find(function(n) { return n.id === nodeId; });
            if (!node) return;
            if (window.innerWidth <= 1024) {
                toggleInspector(true);
            }
            const panel = document.getElementById('inspectorContent');
            if (!panel) return;

            const connectedEdges = allEdges.filter(function(e) { return e.from === nodeId || e.to === nodeId; });
            const maxEdgeWeight = allEdges.reduce(function(m, e) {
                return Math.max(m, e.effective_weight || 1.0);
            }, 3.0);

            const edgeHTML = connectedEdges.map(function(e) {
                const isOutgoing = (e.from === nodeId);
                const targetNode = allNodes.find(function(n) { return n.id === (isOutgoing ? e.to : e.from); });
                const targetLabel = targetNode ? (targetNode.raw_label || targetNode.label) : (isOutgoing ? e.to : e.from);
                const connLabel = isOutgoing ? ('➔ ' + targetLabel) : ('⬅ ' + targetLabel);
                const supersededBadge = e.is_superseded ? ' <span style="color:var(--accent-rose); font-size:10px;">[SUPERSEDED]</span>' : '';
                const pct = Math.min(100, Math.max(8, Math.round((e.effective_weight / maxEdgeWeight) * 100)));
                const safeEdgeId = (e.id || '').replace(/"/g, '&quot;');

                return '<div class="relation-card" data-edge-id="' + safeEdgeId + '" title="Click to inspect this relation">' +
                       '  <div class="relation-header">' +
                       '    <span class="relation-type-tag">[' + e.relation + ']</span>' +
                       '    <span style="font-size:10.5px; color:var(--text-secondary);">' + connLabel + supersededBadge + '</span>' +
                       '  </div>' +
                       '  <div style="font-size:11px; color:var(--text-muted); line-height:1.4;">' + (e.context_snippet || '-') + '</div>' +
                       '  <div class="weight-meter">' +
                       '    <span>Weight: ' + e.effective_weight + '</span>' +
                       '    <div class="weight-bar-bg"><div class="weight-bar-fill" style="width:' + pct + '%;"></div></div>' +
                       '  </div>' +
                       '</div>';
            }).join('');

            panel.innerHTML = '<div class="inspector-title">' + (node.raw_label || node.label) + '</div>' +
                '<div class="inspector-badge" style="background:' + (node.color && node.color.background ? node.color.background : '#1E293B') + '; border:1px solid ' + (node.color && node.color.border ? node.color.border : '#3B82F6') + '; color:' + (node.font && node.font.color ? node.font.color : '#FFF') + '">' + node.group + '</div>' +
                '<div class="detail-grid">' +
                '  <div class="detail-row"><span class="detail-label">Identifier</span><code>' + node.id + '</code></div>' +
                '  <div class="detail-row"><span class="detail-label">Observation Time</span><span>' + (node.last_observed_at || '-') + '</span></div>' +
                '  <div class="detail-row"><span class="detail-label">Relation Degree</span><span class="stats-val">' + connectedEdges.length + ' connections</span></div>' +
                '</div>' +
                '<div class="section-title" style="margin-top:14px; margin-bottom:6px;">Evidence Relation Catalog</div>' +
                '<div style="display:flex; flex-direction:column; gap:8px;">' + (edgeHTML || '<p style="font-size:11.5px; color:var(--text-muted);">No active relations.</p>') + '</div>';
        }

        function inspectEdge(edgeId) {
            const edge = allEdges.find(function(e) { return e.id === edgeId; });
            if (!edge) return;
            if (window.innerWidth <= 1024) {
                toggleInspector(true);
            }
            const panel = document.getElementById('inspectorContent');
            if (!panel) return;

            const fromNode = allNodes.find(function(n) { return n.id === edge.from; });
            const toNode = allNodes.find(function(n) { return n.id === edge.to; });
            const fromLabel = fromNode ? (fromNode.raw_label || fromNode.label) : edge.from;
            const toLabel = toNode ? (toNode.raw_label || toNode.label) : edge.to;
            const safeFrom = (edge.from || '').replace(/"/g, '&quot;');
            const safeTo = (edge.to || '').replace(/"/g, '&quot;');

            const maxEdgeWeight = allEdges.reduce(function(m, e) {
                return Math.max(m, e.effective_weight || 1.0);
            }, 3.0);
            const pct = Math.min(100, Math.max(10, Math.round((edge.effective_weight / maxEdgeWeight) * 100)));

            const supersededInfo = edge.is_superseded
                ? '<div class="detail-row"><span class="detail-label" style="color:var(--accent-rose);">Validity Status</span><span style="color:var(--accent-rose); font-weight:600;">SUPERSEDED (Superseded by recent transaction)</span></div>'
                : '<div class="detail-row"><span class="detail-label" style="color:var(--accent-emerald);">Validity Status</span><span style="color:var(--accent-emerald); font-weight:600;">ACTIVE (Valid Market Relation)</span></div>';

            panel.innerHTML =
                '<div class="inspector-title">[' + edge.relation + ']</div>' +
                '<div class="inspector-badge" style="background:#0F291E; border:1px solid #10B981; color:#34D399;">MARKET RELATION</div>' +
                '<div class="causality-card">' +
                '  <div class="causality-title">Causality Flow</div>' +
                '  <div class="causality-flow-row">' +
                '    <button class="causality-badge" data-focus-entity="' + safeFrom + '" title="Focus on source entity">' + fromLabel + '</button>' +
                '    <span class="causality-arrow">➔</span>' +
                '    <button class="causality-badge" data-focus-entity="' + safeTo + '" title="Focus on target entity">' + toLabel + '</button>' +
                '  </div>' +
                '</div>' +
                '<div class="detail-grid">' +
                '  <div class="detail-row">' +
                '    <span class="detail-label">Evidence Snippet / Disclosure Quote</span>' +
                '    <div class="evidence-quote">' + (edge.context_snippet || 'No specific disclosure quote recorded.') + '</div>' +
                '  </div>' +
                '  <div class="detail-row">' +
                '    <span class="detail-label">Temporal Effective Weight</span>' +
                '    <div class="weight-meter" style="margin-top:4px;">' +
                '      <span class="stats-val">' + edge.effective_weight + '</span>' +
                '      <span style="color:var(--text-muted); font-size:10px;">(Base: ' + edge.weight + ')</span>' +
                '      <div class="weight-bar-bg"><div class="weight-bar-fill" style="width:' + pct + '%;"></div></div>' +
                '    </div>' +
                '  </div>' +
                supersededInfo +
                '  <div class="detail-row">' +
                '    <span class="detail-label">Last Observation</span>' +
                '    <span>' + (edge.last_observed_at || '-') + '</span>' +
                '  </div>' +
                '  <div class="detail-row">' +
                '    <span class="detail-label">Investigation Session</span>' +
                '    <code>' + (edge.session_id || '-') + '</code>' +
                '  </div>' +
                '</div>';
        }

        function searchAndHighlight() {
            const q = document.getElementById('searchInput').value.trim().toLowerCase();
            const clearBtn = document.getElementById('searchClearBtn');
            if (clearBtn) clearBtn.style.display = q ? 'block' : 'none';

            if (!q) {
                restoreGraphView();
                return;
            }
            const matched = allNodes.filter(function(n) {
                return n.label.toLowerCase().includes(q) || n.id.toLowerCase().includes(q);
            });
            if (matched.length > 0) {
                const target = matched[0];
                highlightNeighbors(target.id);
                network.focus(target.id, { scale: 1.15, animation: { duration: 400, easingFunction: 'easeInOutQuad' } });
                inspectNode(target.id);
            }
        }

        function clearSearch() {
            document.getElementById('searchInput').value = '';
            document.getElementById('searchClearBtn').style.display = 'none';
            restoreGraphView();
        }

        function focusOnEntity(entityLabel) {
            const matched = allNodes.filter(function(n) {
                return (n.raw_label || n.label).toLowerCase() === entityLabel.toLowerCase() || n.id.toLowerCase() === entityLabel.toLowerCase();
            });
            if (matched.length > 0) {
                const target = matched[0];
                network.selectNodes([target.id]);
                highlightNeighbors(target.id);
                network.focus(target.id, { scale: 1.25, animation: { duration: 500, easingFunction: 'easeInOutQuad' } });
                inspectNode(target.id);
            }
        }

        function showToast(message) {
            let toast = document.getElementById('appToast');
            if (!toast) {
                toast = document.createElement('div');
                toast.id = 'appToast';
                toast.className = 'toast-msg';
                document.body.appendChild(toast);
            }
            toast.textContent = message;
            toast.classList.add('show');
            setTimeout(function() {
                toast.classList.remove('show');
            }, 2800);
        }

        let activeGroupFilter = null;

        function filterByGroup(groupName) {
            const legendItems = document.querySelectorAll('.legend-item');
            
            // Toggle off if clicking same active group
            if (activeGroupFilter === groupName) {
                activeGroupFilter = null;
                legendItems.forEach(function(el) { el.classList.remove('active'); });
                restoreGraphView();
                showToast("Classification filter cleared");
                return;
            }

            const groupNodes = allNodes.filter(function(n) { return n.group === groupName; });
            if (groupNodes.length === 0) {
                showToast("No entities found for group: " + groupName);
                return;
            }

            activeGroupFilter = groupName;
            legendItems.forEach(function(el) {
                const groupAttr = el.getAttribute('onclick') || '';
                if (groupAttr.includes("'" + groupName + "'")) {
                    el.classList.add('active');
                } else {
                    el.classList.remove('active');
                }
            });

            const groupNodeIds = new Set(groupNodes.map(function(n) { return n.id; }));
            const incidentEdges = allEdges.filter(function(e) {
                return groupNodeIds.has(e.from) || groupNodeIds.has(e.to);
            });
            const incidentEdgeIds = new Set(incidentEdges.map(function(e) { return e.id; }));

            // High-contrast highlighting without destroying node positions
            const nodeUpdates = allNodes.map(function(n) {
                if (groupNodeIds.has(n.id)) {
                    return { id: n.id, opacity: 1.0 };
                } else {
                    return { id: n.id, opacity: 0.12 };
                }
            });

            const edgeUpdates = allEdges.map(function(e) {
                if (incidentEdgeIds.has(e.id)) {
                    return {
                        id: e.id,
                        color: { color: '#FCD535', highlight: '#FCD535' },
                        width: (e.width || 1) + 1,
                        opacity: 0.85
                    };
                } else {
                    return {
                        id: e.id,
                        color: { color: 'rgba(148, 163, 184, 0.04)' },
                        opacity: 0.03
                    };
                }
            });

            nodesDataSet.update(nodeUpdates);
            edgesDataSet.update(edgeUpdates);

            if (groupNodes.length > 0) {
                const groupIds = groupNodes.map(function(n) { return n.id; });
                network.fit({
                    nodes: groupIds,
                    animation: { duration: 450, easingFunction: 'easeInOutQuad' }
                });
            }

            showToast("Focused on classification: " + groupName);
        }

        function resetFilter() {
            activeGroupFilter = null;
            document.querySelectorAll('.legend-item').forEach(function(el) { el.classList.remove('active'); });
            restoreGraphView();
            fitView();
            showToast("Graph reset to complete market network.");
        }

        function togglePhysics() {
            physicsEnabled = !physicsEnabled;
            network.setOptions({ physics: { enabled: physicsEnabled } });
            const btn = document.getElementById('physicsToggleBtn');
            const txt = document.getElementById('physicsStatusText');
            if (physicsEnabled) {
                if (btn) btn.classList.add('active');
                if (txt) txt.textContent = 'Physics: Active';
            } else {
                if (btn) btn.classList.remove('active');
                if (txt) txt.textContent = 'Physics: Settled';
            }
        }

        function fitView() {
            if (network) {
                network.fit({
                    animation: {
                        duration: 500,
                        easingFunction: 'easeInOutQuad'
                    }
                });
            }
        }

        function zoomIn() {
            if (network) {
                const scale = network.getScale();
                network.moveTo({ scale: scale * 1.3, animation: { duration: 250, easingFunction: 'easeInOutQuad' } });
            }
        }

        function zoomOut() {
            if (network) {
                const scale = network.getScale();
                network.moveTo({ scale: scale * 0.75, animation: { duration: 250, easingFunction: 'easeInOutQuad' } });
            }
        }

        function toggleSidebar(forceState) {
            const sidebar = document.querySelector('.sidebar');
            const backdrop = document.getElementById('drawerBackdrop');
            if (!sidebar) return;
            const shouldOpen = typeof forceState === 'boolean' ? forceState : !sidebar.classList.contains('open');
            if (shouldOpen) {
                sidebar.classList.add('open');
                if (window.innerWidth <= 768) {
                    const inspector = document.querySelector('.inspector');
                    if (inspector) inspector.classList.remove('open');
                }
                if (backdrop && window.innerWidth <= 768) {
                    backdrop.classList.add('active');
                }
            } else {
                sidebar.classList.remove('open');
                const inspector = document.querySelector('.inspector');
                if (backdrop && (!inspector || !inspector.classList.contains('open'))) {
                    backdrop.classList.remove('active');
                }
            }
        }

        function toggleInspector(forceState) {
            const inspector = document.querySelector('.inspector');
            const backdrop = document.getElementById('drawerBackdrop');
            if (!inspector) return;
            const shouldOpen = typeof forceState === 'boolean' ? forceState : !inspector.classList.contains('open');
            if (shouldOpen) {
                inspector.classList.add('open');
                if (window.innerWidth <= 768) {
                    const sidebar = document.querySelector('.sidebar');
                    if (sidebar) sidebar.classList.remove('open');
                }
                if (backdrop && window.innerWidth <= 1024) {
                    backdrop.classList.add('active');
                }
            } else {
                inspector.classList.remove('open');
                const sidebar = document.querySelector('.sidebar');
                if (backdrop && (!sidebar || !sidebar.classList.contains('open'))) {
                    backdrop.classList.remove('active');
                }
            }
        }

        function closeDrawers() {
            const sidebar = document.querySelector('.sidebar');
            const inspector = document.querySelector('.inspector');
            const backdrop = document.getElementById('drawerBackdrop');
            if (sidebar) sidebar.classList.remove('open');
            if (inspector) inspector.classList.remove('open');
            if (backdrop) backdrop.classList.remove('active');
        }

        document.addEventListener('click', function(e) {
            const sidebar = document.querySelector('.sidebar');
            const inspector = document.querySelector('.inspector');
            const sidebarBtn = document.getElementById('sidebarToggleBtn');
            const inspectorBtn = document.getElementById('inspectorToggleBtn');

            if (window.innerWidth <= 768 && sidebar && sidebar.classList.contains('open')) {
                if (!sidebar.contains(e.target) && (!sidebarBtn || !sidebarBtn.contains(e.target))) {
                    toggleSidebar(false);
                }
            }
            if (window.innerWidth <= 1024 && inspector && inspector.classList.contains('open')) {
                if (!inspector.contains(e.target) && (!inspectorBtn || !inspectorBtn.contains(e.target)) && !e.target.closest('#network')) {
                    toggleInspector(false);
                }
            }
        });

        document.addEventListener('keydown', function(e) {
            if (e.key === 'Escape') {
                closeDrawers();
            }
        });

        window.addEventListener('load', initNetwork);
    </script>
</body>
</html>
"""


class GraphVisualizer:
    """Market Intelligence Graph Visualizer."""

    def __init__(self, memory: Optional[LocalGraphMemory] = None):
        self.memory = memory or LocalGraphMemory()

    def export_graph_data(
        self,
        session_id: Optional[str] = None,
        ticker: Optional[str] = None,
        depth: int = 1,
        node_types: Optional[List[str]] = None,
    ) -> Dict[str, Any]:
        """Convert in-memory graph to Vis.js DataSet format with colors, width constraints, and weights."""
        stats = self.memory.get_graph_stats()

        candidate_nodes: List[Dict[str, Any]] = []
        candidate_edges: List[Dict[str, Any]] = []

        if ticker:
            ego = self.memory.retrieve_ego_subgraph(ticker, radius=depth)
            candidate_nodes = ego.get("nodes", [])
            candidate_edges = ego.get("edges", [])
        else:
            G = self.memory.load_graph()
            for n, data in G.nodes(data=True):
                candidate_nodes.append({
                    "id": n,
                    "label": data.get("label", n),
                    "node_type": data.get("node_type", "ENTITY"),
                    "last_observed_at": data.get("last_observed_at", ""),
                    "metadata": data.get("metadata", {}),
                })

            now = datetime.now(timezone.utc)
            for u, v, data in G.edges(data=True):
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
                candidate_edges.append({
                    "source_id": u,
                    "target_id": v,
                    "relation": data.get("relation", "RELATES_TO"),
                    "context_snippet": data.get("context_snippet", ""),
                    "session_id": data.get("session_id", ""),
                    "weight": base_w,
                    "effective_weight": eff_w,
                    "last_observed_at": observed_str,
                    "is_superseded": data.get("is_superseded", False),
                })

        retained_edges: List[Dict[str, Any]] = []
        for edge in candidate_edges:
            edge_session = edge.get("session_id", "")
            if session_id and edge_session != session_id:
                continue

            u = edge.get("source_id") or edge.get("from")
            v = edge.get("target_id") or edge.get("to")
            if not u or not v:
                continue

            rel = edge.get("relation", "RELATES_TO")
            base_w = float(edge.get("weight", 1.0))
            eff_w = float(edge.get("effective_weight", base_w))
            edge_width = max(2, min(5, round(eff_w * 1.5)))

            # High-contrast institutional relation palette (inspired by Graphify & Bloomberg terminal)
            if rel in ("TRIGGERED_ANOMALY", "VOLATILITY_SURGE"):
                edge_color_rgba = "rgba(244, 63, 94, 0.85)"  # Rose / Coral
                edge_highlight = "#FDA4AF"
            elif rel in ("CATALYZED_BY", "ANNOUNCED_ACTION", "DISCLOSURE"):
                edge_color_rgba = "rgba(16, 185, 129, 0.85)"  # Emerald / Mint
                edge_highlight = "#6EE7B7"
            elif rel in ("ACCUMULATED_BY", "NET_FLOW", "BROKER_TRANSACTION"):
                edge_color_rgba = "rgba(168, 85, 247, 0.85)"  # Purple / Violet
                edge_highlight = "#D8B4FE"
            elif rel in ("MENTIONED_IN_BRIEFING", "RESEARCHED", "INVESTIGATED"):
                edge_color_rgba = "rgba(56, 189, 248, 0.80)"  # Sky Blue / Cyan
                edge_highlight = "#7DD3FC"
            elif rel in ("OPERATES_IN_SECTOR", "SECTOR_PEER", "INDUSTRY"):
                edge_color_rgba = "rgba(245, 158, 11, 0.80)"  # Amber / Gold
                edge_highlight = "#FDE68A"
            else:
                edge_color_rgba = "rgba(148, 163, 184, 0.70)"  # Crisp Slate Blue
                edge_highlight = "#FCD535"

            is_superseded = bool(edge.get("is_superseded", False))
            if is_superseded:
                edge_color_rgba = "rgba(100, 116, 139, 0.40)"
                edge_highlight = "#EF4444"

            retained_edges.append({
                "id": f"{u}_{rel}_{v}",
                "from": u,
                "to": v,
                "label": "",
                "title": f"[{rel}] {u} ➔ {v} (Effective Weight: {eff_w})",
                "relation": rel,
                "font": {"size": 9, "color": "#94A3B8", "strokeWidth": 2, "strokeColor": "#0B0E14", "align": "horizontal"},
                "arrows": {"to": {"enabled": True, "scaleFactor": 0.65}},
                "color": {
                    "color": edge_color_rgba,
                    "highlight": edge_highlight,
                    "hover": "#FCD535",
                },
                "dashes": [4, 4] if is_superseded else False,
                "width": edge_width,
                "weight": base_w,
                "effective_weight": eff_w,
                "context_snippet": edge.get("context_snippet", ""),
                "session_id": edge_session,
                "last_observed_at": edge.get("last_observed_at", ""),
                "is_superseded": is_superseded,
            })

        active_node_ids = set()
        for e in retained_edges:
            if e.get("from"):
                active_node_ids.add(e["from"])
            if e.get("to"):
                active_node_ids.add(e["to"])

        # Calculate node degrees for degree centrality scaling
        node_degrees: Dict[str, int] = {}
        for e in retained_edges:
            if e.get("from"):
                node_degrees[e["from"]] = node_degrees.get(e["from"], 0) + 1
            if e.get("to"):
                node_degrees[e["to"]] = node_degrees.get(e["to"], 0) + 1

        scoped = bool(session_id or ticker)
        allowed_types = {t.upper() for t in node_types} if node_types is not None else None

        nodes: List[Dict[str, Any]] = []
        for node_data in candidate_nodes:
            node_id = node_data.get("id")
            if not node_id:
                continue

            if scoped and node_id not in active_node_ids:
                continue

            ntype = str(node_data.get("node_type", "ENTITY")).upper()
            if allowed_types is not None and ntype not in allowed_types:
                continue

            raw_label = str(node_data.get("label") or node_id)
            formatted_label = format_node_label(raw_label)
            style = NODE_TYPE_STYLES.get(ntype, DEFAULT_NODE_STYLE)

            deg = node_degrees.get(node_id, 1)
            base_size = style.get("size", 16)
            dynamic_size = min(34, max(base_size, base_size + int(math.sqrt(deg) * 2)))

            nodes.append({
                "id": node_id,
                "label": formatted_label,
                "raw_label": raw_label,
                "title": f"[{ntype}] {raw_label} ({deg} connections)",
                "group": ntype,
                "color": style["color"],
                "font": style["font"],
                "shape": style["shape"],
                "size": dynamic_size,
                "borderWidth": 2,
                "borderWidthSelected": 3.5,
                "margin": style.get("margin", 10),
                "widthConstraint": style.get("widthConstraint", {"maximum": 150, "minimum": 80}),
                "last_observed_at": node_data.get("last_observed_at", ""),
                "metadata": node_data.get("metadata", {}),
            })

        retained_node_ids = {n["id"] for n in nodes}
        final_edges = [
            e for e in retained_edges
            if e["from"] in retained_node_ids and e["to"] in retained_node_ids
        ]

        return {
            "nodes": nodes,
            "edges": final_edges,
            "stats": stats,
            "generated_at": datetime.now(timezone.utc).isoformat() + "Z",
        }

    def generate_html(
        self,
        session_id: Optional[str] = None,
        ticker: Optional[str] = None,
        depth: int = 1,
        node_types: Optional[List[str]] = None,
        embed: bool = False,
        title: str = "Niskava Agent Market Intelligence",
    ) -> str:
        """Generate a complete, self-contained HTML page string."""
        data = self.export_graph_data(
            session_id=session_id,
            ticker=ticker,
            depth=depth,
            node_types=node_types,
        )
        raw_json = json.dumps(data, ensure_ascii=False)
        if session_id:
            sess_str = session_id
        elif ticker:
            sess_str = f"Ego Graph: {ticker}"
        else:
            sess_str = "All Active Sessions"

        html = HTML_TEMPLATE.replace("__TITLE__", title)
        html = html.replace("__SESSION_ID__", sess_str)
        html = html.replace("__RAW_JSON__", raw_json)

        if embed:
            embed_css = "<style>.sidebar { display: none !important; } .top-bar { display: none !important; } .canvas-area { width: 100vw; height: 100vh; }</style>"
            html = html.replace("</head>", f"    {embed_css}\n</head>")

        return html

    def export_to_file(
        self,
        output_path: str = "~/.niskava/graph.html",
        session_id: Optional[str] = None,
        ticker: Optional[str] = None,
        depth: int = 1,
        node_types: Optional[List[str]] = None,
        embed: bool = False,
        title: str = "Niskava Agent Market Intelligence",
    ) -> str:
        """Generate and save interactive HTML to disk."""
        path = os.path.expanduser(output_path)
        os.makedirs(os.path.dirname(os.path.abspath(path)), exist_ok=True)

        html_content = self.generate_html(
            session_id=session_id,
            ticker=ticker,
            depth=depth,
            node_types=node_types,
            embed=embed,
            title=title,
        )
        with open(path, "w", encoding="utf-8") as f:
            f.write(html_content)

        return path