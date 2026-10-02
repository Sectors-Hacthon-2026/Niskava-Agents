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
                js_parts.append(f"    // --- {f} ---\n" + fp.read().strip())
        else:
            print(f"[WARN] JS file {f} not found!")
    combined_js = "(function() {\n    'use strict';\n\n" + "\n\n".join(js_parts) + "\n})();"

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
    <link rel="icon" type="image/png" href="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAYAAABzenr0AAAFAklEQVR4nMWXT2hUVxTGf/fe9+b/jKlOlMTWLupCrBBMowbBIIOREhTcVVsX0p2tiLiQuOlaKUIXbWkDttYKlqEbUetCFP8hUls0tcaIKTY0iX9ikklmnMx7ee/eLpKMMSZvbCr4wVvd98733e+8e+458Johgha926v6XgWJ9e712lnXgogXrn9Q8yoETMabScgLDni3V/W9KuLpeHyh5sF0EXKu5EpJBBohDJZSLyVg4foHNdPTWnagErlSCqM9PM9lzHVwnVF83wfAsm1sO0w0lkQpC8/XGGNmFTLViRn/gZnIS8Vhik/zZdKpcB0H13FwSkUWpKsZc0qEIgm0nl3EJARU3r0/NspwbuDZBwKebVAAz4jC4TCu6xIKhUnOSzObhkkX5MzLU6A9CvkhhBBIKTCMkwsBUoJSIIRAKYUQAsdxiMViOE6JMXe0YvhAAVLAcK4ffyKnWhsSUcnGNUmMgc8+TrO+PoExBt/3McaQyWTIZrM0NDSQH8lhqeA9Bv4DrlvC9302rEoQi0g831Bdpbh0s8jGNXE6u13OXc9j2zZbtmxh165dNDU1AbBu3TrWrl3LP72PUVZobg6MuS7bmqvIFXyu/fmUO/dLnLqSZ+mbIcK24qezI+z85FNu3bpFNpulqakJrTWlUolkMkkmk0HI4CMaKGDrhio6/i4xXNA0r06y/J0IG9YkAMnJKyNs3/4hX3/1JUuWLOHq1av09vYipcS2bQCamzcGkkNACt6qTXDj3iNGCh7r6uL8fGEEd0ygjUFrSCQSHDz4Of39/bS3t9PR0cHu3bvxPI8TJ04Qj8c5duxHBMFHcVYH+vpy3LwzxOrlEc5cK1AsabxyDdC0trZSW1tLNBrlxo0bdHV10dPTg23bLFq0iNbWVrLZLLF4Ym4O+MYgBDiewXU1H2z9kKHBJ1y+fJmVK1eyd+9etNZEIhE2b95MsVikubmZTCaDbdtYlkUoFAmsiIEClLIwRhKLWAwXDe81rKJYyFFdXc3Ro0fH7ZMSKSXLli0DoL6+fuK4anbs2MF33/9AqeTM0QEtiCaq+KvH4f3GGIcOfcH5s6cZyeeRUqK1Lr9rjCk/k2uNjY2sXt1AR+d9hJi97Qg8Bclkkl87HKRUpEPdbPtoB7lcjr6+PqSUZXuFEBOVUlIoFDhz5gz79u2jp6cXVeGmDBTguaNgfH7vLLL07RR/tP9GS0sLhw8fRmuN53nldydFxONxFi9ezMWLF3n06MlzTv1nAb72MRiiYYXW4/YmUlWcPPXL+MdS4nkeWo+X6sHBQY4cOcLx48dJpVJEY/FA8ooClLIBwVDep7PbRWtDKBTlXtd99uzZw8DAAJZllfOeSCRoaWkhlUoxOjqKkJVv++DLSFpYlmK44HOvxy3n3bLDfPNtG3V1dRw4cIDu7m6UUliWxfz5C9i5cyfJ5LyyS3MWoA0k56XHr9qJuqCUhdaadPVi+p8MsH//flasWMGmTZtoa2ujvf0md+/e5eGjfoC514FJCGkx742FaO3hjD5FTwR0XJcF6RpGco8pFAqcPn2ac+cvYYcjCASWHXopAS/dE0opML4LMvRcUCkMQwMPUUoxP12LO+bNFqKMqT1h5Y5oApalECr0wo60EcSTVQgh8PzgIzcTnitRQS4opWZsSGG8BlgKPL+y5dNng9c+mMxYpF/1aPb4Qs0DeMnRbCYh/xdBw+lrx78nNl1lgzFI8QAAAABJRU5ErkJggg==">
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
    <link rel="icon" type="image/png" href="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAYAAABzenr0AAAFAklEQVR4nMWXT2hUVxTGf/fe9+b/jKlOlMTWLupCrBBMowbBIIOREhTcVVsX0p2tiLiQuOlaKUIXbWkDttYKlqEbUetCFP8hUls0tcaIKTY0iX9ikklmnMx7ee/eLpKMMSZvbCr4wVvd98733e+8e+458Johgha926v6XgWJ9e712lnXgogXrn9Q8yoETMabScgLDni3V/W9KuLpeHyh5sF0EXKu5EpJBBohDJZSLyVg4foHNdPTWnagErlSCqM9PM9lzHVwnVF83wfAsm1sO0w0lkQpC8/XGGNmFTLViRn/gZnIS8Vhik/zZdKpcB0H13FwSkUWpKsZc0qEIgm0nl3EJARU3r0/NspwbuDZBwKebVAAz4jC4TCu6xIKhUnOSzObhkkX5MzLU6A9CvkhhBBIKTCMkwsBUoJSIIRAKYUQAsdxiMViOE6JMXe0YvhAAVLAcK4ffyKnWhsSUcnGNUmMgc8+TrO+PoExBt/3McaQyWTIZrM0NDSQH8lhqeA9Bv4DrlvC9302rEoQi0g831Bdpbh0s8jGNXE6u13OXc9j2zZbtmxh165dNDU1AbBu3TrWrl3LP72PUVZobg6MuS7bmqvIFXyu/fmUO/dLnLqSZ+mbIcK24qezI+z85FNu3bpFNpulqakJrTWlUolkMkkmk0HI4CMaKGDrhio6/i4xXNA0r06y/J0IG9YkAMnJKyNs3/4hX3/1JUuWLOHq1av09vYipcS2bQCamzcGkkNACt6qTXDj3iNGCh7r6uL8fGEEd0ygjUFrSCQSHDz4Of39/bS3t9PR0cHu3bvxPI8TJ04Qj8c5duxHBMFHcVYH+vpy3LwzxOrlEc5cK1AsabxyDdC0trZSW1tLNBrlxo0bdHV10dPTg23bLFq0iNbWVrLZLLF4Ym4O+MYgBDiewXU1H2z9kKHBJ1y+fJmVK1eyd+9etNZEIhE2b95MsVikubmZTCaDbdtYlkUoFAmsiIEClLIwRhKLWAwXDe81rKJYyFFdXc3Ro0fH7ZMSKSXLli0DoL6+fuK4anbs2MF33/9AqeTM0QEtiCaq+KvH4f3GGIcOfcH5s6cZyeeRUqK1Lr9rjCk/k2uNjY2sXt1AR+d9hJi97Qg8Bclkkl87HKRUpEPdbPtoB7lcjr6+PqSUZXuFEBOVUlIoFDhz5gz79u2jp6cXVeGmDBTguaNgfH7vLLL07RR/tP9GS0sLhw8fRmuN53nldydFxONxFi9ezMWLF3n06MlzTv1nAb72MRiiYYXW4/YmUlWcPPXL+MdS4nkeWo+X6sHBQY4cOcLx48dJpVJEY/FA8ooClLIBwVDep7PbRWtDKBTlXtd99uzZw8DAAJZllfOeSCRoaWkhlUoxOjqKkJVv++DLSFpYlmK44HOvxy3n3bLDfPNtG3V1dRw4cIDu7m6UUliWxfz5C9i5cyfJ5LyyS3MWoA0k56XHr9qJuqCUhdaadPVi+p8MsH//flasWMGmTZtoa2ujvf0md+/e5eGjfoC514FJCGkx742FaO3hjD5FTwR0XJcF6RpGco8pFAqcPn2ac+cvYYcjCASWHXopAS/dE0opML4LMvRcUCkMQwMPUUoxP12LO+bNFqKMqT1h5Y5oApalECr0wo60EcSTVQgh8PzgIzcTnitRQS4opWZsSGG8BlgKPL+y5dNng9c+mMxYpF/1aPb4Qs0DeMnRbCYh/xdBw+lrx78nNl1lgzFI8QAAAABJRU5ErkJggg==">
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
    <script src="js/app.js"></script>
</body>
</html>
"""
    dest_index = os.path.join(BASE_DIR, "index.html")
    with open(dest_index, "w", encoding="utf-8") as fp:
        fp.write(modular_index)
    print(f"[OK] Generated {dest_index} ({len(modular_index.splitlines())} lines)")

if __name__ == "__main__":
    build()
