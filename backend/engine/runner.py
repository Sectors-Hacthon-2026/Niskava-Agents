"""Subprocess IPC entrypoint called by Go Core Daemon.

Emits JSON Lines to sys.stdout per docs/20-architecture/04-ipc-and-api-contract.md.
"""

import argparse
import json
import os
import signal
import sys
from typing import Any, Dict

from engine.agent.react_agent import NiskavaReActAgent
from engine.agent.tools import NiskavaToolRegistry


# Ensure UTF-8 output across all operating systems (especially Windows consoles)
if hasattr(sys.stdout, "reconfigure"):
    sys.stdout.reconfigure(encoding="utf-8")
if hasattr(sys.stderr, "reconfigure"):
    sys.stderr.reconfigure(encoding="utf-8")


def _handle_signal(sig: int, frame: Any) -> None:
    sys.exit(0)


if hasattr(signal, "SIGINT"):
    signal.signal(signal.SIGINT, _handle_signal)
if hasattr(signal, "SIGTERM"):
    signal.signal(signal.SIGTERM, _handle_signal)


def emit_jsonl(event_dict: Dict[str, Any]) -> None:
    """Print a single JSON object followed by newline and flush immediately."""
    line = json.dumps(event_dict)
    sys.stdout.write(line + "\n")
    sys.stdout.flush()


def load_dotenv_fallback() -> None:
    """Lightweight .env loader using standard library only."""
    for candidate in (".env", os.path.expanduser("~/.niskava/.env")):
        if os.path.exists(candidate):
            try:
                with open(candidate, "r", encoding="utf-8") as f:
                    for line in f:
                        line = line.strip()
                        if not line or line.startswith("#") or "=" not in line:
                            continue
                        k, v = line.split("=", 1)
                        k = k.strip()
                        v = v.strip().strip("'\"")
                        # Only set non-empty values into os.environ so empty keys do not shadow valid ones
                        if k and v and (k not in os.environ or not os.environ[k]):
                            os.environ[k] = v
            except OSError:
                pass


def load_config_yaml_fallback() -> None:
    """Optional fallback to load credentials from ~/.niskava/config.yaml if not set in os.environ."""
    config_path = os.path.expanduser("~/.niskava/config.yaml")
    if not os.path.exists(config_path):
        return
    try:
        import yaml
        with open(config_path, "r", encoding="utf-8") as f:
            cfg = yaml.safe_load(f)
            if not isinstance(cfg, dict):
                return
            auth = cfg.get("auth", {})
            if isinstance(auth, dict):
                mapping = {
                    "SECTORS_API_KEY": auth.get("sectors_api_key"),
                    "SECTORS_BASE_URL": auth.get("sectors_base_url"),
                    "AI_PROVIDER": auth.get("ai_provider"),
                    "GEMINI_API_KEY": auth.get("gemini_api_key"),
                    "GEMINI_MODEL": auth.get("gemini_model"),
                    "OPENAI_API_KEY": auth.get("openai_api_key"),
                    "OPENAI_BASE_URL": auth.get("openai_base_url"),
                    "OPENAI_MODEL": auth.get("openai_model"),
                    "ANTHROPIC_API_KEY": auth.get("anthropic_api_key"),
                    "OLLAMA_BASE_URL": auth.get("ollama_base_url"),
                    "OLLAMA_MODEL": auth.get("ollama_model"),
                }
                for k, v in mapping.items():
                    if v and isinstance(v, str) and (k not in os.environ or not os.environ[k]):
                        os.environ[k] = v
            prefs = cfg.get("preferences", {})
            if isinstance(prefs, dict):
                if prefs.get("language") and "NISKAVA_LANG" not in os.environ:
                    os.environ["NISKAVA_LANG"] = str(prefs["language"])
                if prefs.get("offline_mode") and "NISKAVA_OFFLINE" not in os.environ:
                    os.environ["NISKAVA_OFFLINE"] = "1" if prefs["offline_mode"] else "0"
    except Exception:
        pass


def main() -> None:
    load_dotenv_fallback()
    load_config_yaml_fallback()
    parser = argparse.ArgumentParser(description="Niskava Python Agent Engine IPC Runner")
    parser.add_argument("--prompt", default=None, help="Free-form conversational user prompt")
    parser.add_argument("--ticker", default=None, help="Target IDX ticker (e.g. ANTM)")
    parser.add_argument("--days", type=int, default=30, help="Observation window days")
    parser.add_argument("--session", default=None, help="Session ID (e.g. INV-2026-0042)")
    parser.add_argument("--db-path", default="~/.niskava/niskava.db", help="Path to local SQLite DB")
    parser.add_argument("--offline", action="store_true", help="Force offline mock mode")
    parser.add_argument("--export-graph-html", default=None, help="Export graph HTML to specified path")
    parser.add_argument("--depth", type=int, default=1, help="Hop depth radius for ego graph")
    parser.add_argument("--node-types", default=None, help="Comma-separated node types to filter")
    parser.add_argument("--embed", action="store_true", help="Render lightweight embedded view for iframes")
    parser.add_argument("--summary-graph", action="store_true", help="Output JSON text summary of graph to stdout")
    parser.add_argument("--language", "--lang", default=os.environ.get("NISKAVA_LANG", "id"), help="Interface and persona language ('id' or 'en')")

    args = parser.parse_args()

    mock_mode = (
        args.offline
        or os.environ.get("MOCK_SECTORS", "0") in ("1", "true", "True")
        or os.environ.get("NISKAVA_OFFLINE", "0") in ("1", "true", "True")
    )
    language = (args.language or "id").lower()
    os.environ["NISKAVA_LANG"] = language

    try:
        registry = NiskavaToolRegistry(
            db_path=args.db_path,
            mock_mode=mock_mode,
        )
        if args.summary_graph:
            from engine.memory.visualizer import GraphVisualizer
            viz = GraphVisualizer(memory=registry.memory)
            types = [t.strip().upper() for t in args.node_types.split(",")] if args.node_types else None
            data = viz.export_graph_data(
                session_id=args.session,
                ticker=args.ticker,
                depth=args.depth,
                node_types=types,
            )
            print(json.dumps({
                "total_nodes": len(data["nodes"]),
                "total_edges": len(data["edges"]),
                "nodes": [{"id": n["id"], "label": n.get("raw_label", n.get("label")), "group": n["group"]} for n in data["nodes"]],
                "edges": [{"from": e["from"], "to": e["to"], "rel": e["relation"], "weight": e["effective_weight"]} for e in data["edges"]],
                "stats": data.get("stats", {}),
            }, indent=2, ensure_ascii=False))
            return

        if args.export_graph_html:
            from engine.memory.visualizer import GraphVisualizer
            viz = GraphVisualizer(memory=registry.memory)
            types = [t.strip().upper() for t in args.node_types.split(",")] if args.node_types else None
            saved = viz.export_to_file(
                output_path=args.export_graph_html,
                session_id=args.session,
                ticker=args.ticker,
                depth=args.depth,
                node_types=types,
                embed=args.embed,
            )
            emit_jsonl({
                "event": "graph_exported",
                "session_id": args.session or "ALL",
                "file_path": saved,
            })
            return

        agent = NiskavaReActAgent(
            tool_registry=registry,
            emitter=emit_jsonl,
            mock_mode=mock_mode,
            language=language,
        )
        if args.prompt:
            agent.chat(
                user_prompt=args.prompt,
                session_id=args.session,
            )
        elif args.ticker:
            from engine.agent.pipeline import InvestigationPipeline
            pipeline = InvestigationPipeline(
                db_path=args.db_path,
                emitter=emit_jsonl,
                mock_mode=mock_mode,
            )
            pipeline.run(
                ticker=args.ticker,
                timeframe_days=args.days,
                session_id=args.session,
            )
        else:
            default_prompt = (
                "Conduct a comprehensive analysis of IDX stock movements today."
                if language == "en"
                else "Lakukan analisis menyeluruh terhadap pergerakan saham IDX hari ini."
            )
            agent.chat(
                user_prompt=default_prompt,
                session_id=args.session,
            )
    except Exception as exc:
        emit_jsonl({
            "event": "session_error",
            "session_id": args.session or "UNKNOWN",
            "error": str(exc),
        })
        # Exit 0 agar Go (cmd.Wait) tidak mengirim error kedua ke errChan.
        # Error sudah dilaporkan melalui event session_error di atas.
        sys.exit(0)


if __name__ == "__main__":
    main()
