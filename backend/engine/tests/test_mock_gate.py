"""Tests that mock mode is blocked in non-test environments."""
import os
import importlib
import pytest


def test_sectors_client_blocks_mock_without_testing_gate(tmp_path, monkeypatch):
    """SectorsAPIClient must raise RuntimeError if mock_mode=True and NISKAVA_TESTING != '1'."""
    monkeypatch.delenv("NISKAVA_TESTING", raising=False)
    monkeypatch.setenv("MOCK_SECTORS", "1")
    monkeypatch.setenv("SECTORS_API_KEY", "")

    import backend.engine.sectors.client as client_mod
    importlib.reload(client_mod)

    with pytest.raises(RuntimeError, match="SECTORS_API_KEY is required"):
        client_mod.SectorsAPIClient(db_path=str(tmp_path / "test.db"), mock_mode=True)


def test_sectors_client_allows_mock_with_testing_gate(tmp_path, monkeypatch):
    """SectorsAPIClient must allow mock_mode=True when NISKAVA_TESTING=1."""
    monkeypatch.setenv("NISKAVA_TESTING", "1")
    monkeypatch.setenv("MOCK_SECTORS", "1")
    monkeypatch.setenv("SECTORS_API_KEY", "test-fixture-key-not-real")

    import backend.engine.sectors.client as client_mod
    importlib.reload(client_mod)

    client = client_mod.SectorsAPIClient(db_path=str(tmp_path / "test.db"), mock_mode=True)
    assert client.mock_mode is True


def test_sectors_client_live_mode_with_real_key(tmp_path, monkeypatch):
    """SectorsAPIClient with a real key must be in live mode regardless of gate."""
    monkeypatch.delenv("NISKAVA_TESTING", raising=False)
    monkeypatch.setenv("SECTORS_API_KEY", "abc123realkey")
    monkeypatch.setenv("MOCK_SECTORS", "0")

    import backend.engine.sectors.client as client_mod
    importlib.reload(client_mod)

    client = client_mod.SectorsAPIClient(db_path=str(tmp_path / "test.db"))
    assert client.mock_mode is False
