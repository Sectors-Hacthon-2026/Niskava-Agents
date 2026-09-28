#!/usr/bin/env python3
"""
Niskava Web Workspace Builder & Bundler
Combines modular CSS and JS components into production distribution files:
- backend/core/server/workspace.html (embedded in single-binary Go Core daemon)
- clients/web/dist/workspace.html (optional standalone distribution)
"""

import os
import sys

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
CSS_DIR = os.path.join(BASE_DIR, "css")
JS_DIR = os.path.join(BASE_DIR, "js")
ROOT_DIR = os.path.abspath(os.path.join(BASE_DIR, "../.."))
DEST_WORKSPACE = os.path.join(ROOT_DIR, "backend/core/server/workspace.html")

CSS_ORDER = [
    "variables.css",
    "layout.css",
    "chat.css",
    "graph.css",
    "modals.css",
]

JS_ORDER = [
    "00_state.js",
    "01_i18n.js",
    "02_ui.js",
    "03_charts.js",
    "04_sessions.js",
    "05_graph.js",
    "06_chat.js",
    "07_settings.js",
    "08_bootstrap.js",
]

def build():
    # 1. Combine CSS
    css_parts = []
    for f in CSS_ORDER:
        path = os.path.join(CSS_DIR, f)
        if os.path.exists(path):
            with open(path, "r", encoding="utf-8") as fp:
                css_parts.append(f"        /* --- {f} --- */\n" + fp.read().strip())
        else:
            print(f"[WARN] CSS file {f} not found!")
    combined_css = "\n\n".join(css_parts)

    # Also update clients/web/css/styles.css for standalone development
    with open(os.path.join(CSS_DIR, "styles.css"), "w", encoding="utf-8") as fp:
        fp.write(combined_css + "\n")

    # 2. Combine JS
    js_parts = []
    for f in JS_ORDER:
        path = os.path.join(JS_DIR, f)
        if os.path.exists(path):
            with open(path, "r", encoding="utf-8") as fp:
                js_parts.append(f"        // --- {f} ---\n" + fp.read().strip())
        else:
            print(f"[WARN] JS file {f} not found!")
    combined_js = "\n\n".join(js_parts)

    # Also update clients/web/js/app.js for standalone development
    with open(os.path.join(JS_DIR, "app.js"), "w", encoding="utf-8") as fp:
        fp.write(combined_js + "\n")

    # 3. Read Layout HTML
    layout_path = os.path.join(BASE_DIR, "layout.html")
    with open(layout_path, "r", encoding="utf-8") as fp:
        layout_html = fp.read().strip()

    # 4. Construct unified workspace.html for Go embed
    workspace_html = f"""<!DOCTYPE html>
<html lang="id" data-theme="light">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>NISKAVA — AI for Brighter Investments</title>
    <link rel="icon" type="image/png" href="data:image/png;base64,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">
    <!-- Single consistent eye-pleasing font: Plus Jakarta Sans -->
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;600&family=Plus+Jakarta+Sans:ital,wght@0,300;0,400;0,500;0,600;0,700;0,800;1,400;1,600&display=swap" rel="stylesheet">
    <!-- TradingView Lightweight Charts (Standalone production bundle) -->
    <script src="https://unpkg.com/lightweight-charts@4.1.1/dist/lightweight-charts.standalone.production.js"></script>
    <style>
{combined_css}
    </style>
</head>
<body>
{layout_html}
    <script>
{combined_js}
    </script>
</body>
</html>
"""

    with open(DEST_WORKSPACE, "w", encoding="utf-8") as fp:
        fp.write(workspace_html)
    print(f"[OK] Generated {DEST_WORKSPACE} ({len(workspace_html.splitlines())} lines)")

    # 5. Also update clients/web/index.html with clean external links for local browser/IDE editing
    modular_index = f"""<!DOCTYPE html>
<html lang="id" data-theme="light">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>NISKAVA — AI for Brighter Investments</title>
    <link rel="icon" type="image/png" href="data:image/png;base64,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">
    <!-- Single consistent eye-pleasing font: Plus Jakarta Sans -->
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;600&family=Plus+Jakarta+Sans:ital,wght@0,300;0,400;0,500;0,600;0,700;0,800;1,400;1,600&display=swap" rel="stylesheet">
    <!-- TradingView Lightweight Charts (Standalone production bundle) -->
    <script src="https://unpkg.com/lightweight-charts@4.1.1/dist/lightweight-charts.standalone.production.js"></script>
    <link rel="stylesheet" href="css/variables.css">
    <link rel="stylesheet" href="css/layout.css">
    <link rel="stylesheet" href="css/chat.css">
    <link rel="stylesheet" href="css/graph.css">
    <link rel="stylesheet" href="css/modals.css">
</head>
<body>
{layout_html}
    <script src="js/00_state.js"></script>
    <script src="js/01_i18n.js"></script>
    <script src="js/02_ui.js"></script>
    <script src="js/03_charts.js"></script>
    <script src="js/04_sessions.js"></script>
    <script src="js/05_graph.js"></script>
    <script src="js/06_chat.js"></script>
    <script src="js/07_settings.js"></script>
    <script src="js/08_bootstrap.js"></script>
</body>
</html>
"""
    dest_index = os.path.join(BASE_DIR, "index.html")
    with open(dest_index, "w", encoding="utf-8") as fp:
        fp.write(modular_index)
    print(f"[OK] Generated {dest_index} ({len(modular_index.splitlines())} lines)")

if __name__ == "__main__":
    build()
