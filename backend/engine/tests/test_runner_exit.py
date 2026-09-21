"""Test bahwa runner.py tidak menggunakan sys.exit(1) setelah emit session_error.

Memverifikasi ketidaksesuaian arsitektur A: sys.exit(1) menyebabkan
double error di Go side (cmd.Wait() + event session_error).
"""
import ast
import os
import pytest


def _get_runner_source() -> str:
    """Baca source code runner.py."""
    runner_path = os.path.join(
        os.path.dirname(__file__), "..", "runner.py"
    )
    with open(runner_path, "r", encoding="utf-8") as f:
        return f.read()


def test_runner_does_not_exit_with_code_1_after_session_error():
    """runner.py tidak boleh memanggil sys.exit(1) setelah emit session_error."""
    source = _get_runner_source()
    tree = ast.parse(source)

    exit_one_found = False
    for node in ast.walk(tree):
        if isinstance(node, ast.ExceptHandler):
            body_src = ast.unparse(node)
            if "session_error" in body_src and "sys.exit(1)" in body_src:
                exit_one_found = True
                break

    assert not exit_one_found, (
        "runner.py tidak boleh memanggil sys.exit(1) setelah emit session_error. "
        "Gunakan sys.exit(0) agar Go tidak double-report error."
    )


def test_runner_exits_cleanly_after_session_error():
    """runner.py harus menggunakan sys.exit(0) di blok except setelah session_error."""
    source = _get_runner_source()
    tree = ast.parse(source)

    has_exit_zero_in_except = False
    for node in ast.walk(tree):
        if isinstance(node, ast.ExceptHandler):
            body_src = ast.unparse(node)
            if "session_error" in body_src and "sys.exit(0)" in body_src:
                has_exit_zero_in_except = True
                break

    assert has_exit_zero_in_except, (
        "runner.py harus menggunakan sys.exit(0) di blok except setelah "
        "emit session_error, bukan sys.exit(1)."
    )
