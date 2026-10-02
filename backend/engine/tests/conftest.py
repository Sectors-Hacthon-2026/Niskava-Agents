"""Pytest session-level fixture: activates NISKAVA_TESTING=1 and mock mode
for the entire test suite. This is the single authoritative gate that
allows mock/offline behavior — end-user code checks for this flag.
"""
import os
import pytest


def pytest_configure(config: pytest.Config) -> None:  # noqa: ARG001
    """Set testing environment variables before any test is collected."""
    os.environ["NISKAVA_TESTING"] = "1"
    os.environ["MOCK_SECTORS"] = "1"
    os.environ["NISKAVA_OFFLINE"] = "1"
    # Provide a dummy key so tests that inspect key presence still work
    if not os.environ.get("SECTORS_API_KEY"):
        os.environ["SECTORS_API_KEY"] = "test-fixture-key-not-real"
