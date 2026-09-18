"""Subprocess IPC entrypoint called by Go Core Daemon.

Emits JSON Lines to sys.stdout per docs/20-architecture/04-ipc-and-api-contract.md.
"""

import argparse
import json
import os
import sys
from typing import Any, Dict

from engine.agent.react_agent import NiskavaReActAgent
from engine.agent.tools import NiskavaToolRegistry


def emit_jsonl(event_dict: Dict[str, Any]) -> None:
    """Print a single JSON object followed by newline and flush immediately."""
    line = json.dumps(event_dict)
    sys.stdout.write(line + "\n")
    sys.stdout.flush()


def main() -> None:
    parser = argparse.ArgumentParser(description="Niskava Python Agent Engine IPC Runner")
    parser.add_argument("--ticker", required=True, help="Target IDX ticker (e.g. ANTM)")
    parser.add_argument("--days", type=int, default=30, help="Observation window days")
    parser.add_argument("--session", default=None, help="Session ID (e.g. INV-2026-0042)")
    parser.add_argument("--db-path", default="~/.niskava/niskava.db", help="Path to local SQLite DB")
    parser.add_argument("--offline", action="store_true", help="Force offline mock mode")

    args = parser.parse_args()

    mock_mode = (
        args.offline
        or os.environ.get("MOCK_SECTORS", "0") in ("1", "true", "True")
        or os.environ.get("NISKAVA_OFFLINE", "0") in ("1", "true", "True")
    )

    try:
        registry = NiskavaToolRegistry(
            db_path=args.db_path,
            mock_mode=mock_mode,
        )
        agent = NiskavaReActAgent(
            tool_registry=registry,
            emitter=emit_jsonl,
            mock_mode=mock_mode,
        )
        agent.investigate(
            ticker=args.ticker,
            days=args.days,
            session_id=args.session,
        )
    except Exception as exc:
        emit_jsonl({
            "event": "session_error",
            "session_id": args.session or "UNKNOWN",
            "error": str(exc),
        })
        sys.exit(1)


if __name__ == "__main__":
    main()
