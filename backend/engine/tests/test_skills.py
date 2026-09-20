"""Comprehensive unit tests for Layer 3 Modular Domain Skills and SkillsRegistry."""

import pytest
from pathlib import Path

from engine.sectors.client import SectorsAPIClient
from engine.skills.base import BaseSkill, SkillResult
from engine.skills.registry import SkillsRegistry


@pytest.fixture
def mock_sectors_client(tmp_path):
    db_file = tmp_path / "test_skills.db"
    return SectorsAPIClient(db_path=str(db_file), mock_mode=True)


@pytest.fixture
def registry():
    return SkillsRegistry()


def test_registry_discovery(registry):
    """Verify registry discovers all 6 domain skills automatically."""
    skills = registry.list_skills()
    skill_ids = {s.skill_id for s in skills}
    expected_skills = {
        "market-anomaly-recon",
        "event-causality-audit",
        "insider-bandarmology-forensic",
        "financial-health-stress-test",
        "mining-commodity-divergence",
        "peer-valuation-benchmark",
    }
    assert expected_skills.issubset(skill_ids), f"Missing skills: {expected_skills - skill_ids}"
    assert len(registry.get_all_tool_definitions()) >= 6


def test_skill_market_anomaly_recon(registry, mock_sectors_client):
    context = {"sectors_client": mock_sectors_client}
    result = registry.execute_skill("market-anomaly-recon", {"ticker": "ANTM", "days": 30}, context)
    assert isinstance(result, SkillResult)
    assert result.skill_id == "market-anomaly-recon"
    assert result.verification_status in ("SUPPORTED", "UNCERTAIN")
    assert result.confidence_score == 1.00
    assert "ticker" in result.metrics
    assert result.metrics["ticker"] == "ANTM"


def test_skill_event_causality_audit(registry, mock_sectors_client):
    context = {"sectors_client": mock_sectors_client}
    result = registry.execute_skill(
        "event-causality-audit",
        {"ticker": "ANTM", "anomaly_date": "2026-09-12"},
        context,
    )
    assert isinstance(result, SkillResult)
    assert result.skill_id == "event-causality-audit"
    assert "causality_label" in result.metrics
    assert len(result.evidence) >= 1


def test_skill_insider_bandarmology_forensic(registry, mock_sectors_client):
    context = {"sectors_client": mock_sectors_client}
    result = registry.execute_skill("insider-bandarmology-forensic", {"ticker": "ANTM"}, context)
    assert isinstance(result, SkillResult)
    assert result.skill_id == "insider-bandarmology-forensic"
    assert "concentration_ratio_c3" in result.metrics
    assert result.metrics["concentration_ratio_c3"] > 0
    assert "accumulation_regime" in result.metrics


def test_skill_financial_health_stress_test(registry, mock_sectors_client):
    context = {"sectors_client": mock_sectors_client}
    # Test normal health check
    res1 = registry.execute_skill("financial-health-stress-test", {"ticker": "ANTM"}, context)
    assert res1.verification_status == "SUPPORTED"
    assert "liquidity" in res1.metrics
    assert "solvency" in res1.metrics

    # Test rumor refutation
    res2 = registry.execute_skill(
        "financial-health-stress-test",
        {"ticker": "ANTM", "rumor_claim": "Rumor gagal bayar obligasi"},
        context,
    )
    assert res2.verification_status == "CONTRADICTED"
    assert "CONTRADICTED" in res2.summary


def test_skill_mining_commodity_divergence(registry, mock_sectors_client):
    context = {"sectors_client": mock_sectors_client}
    result = registry.execute_skill("mining-commodity-divergence", {"ticker": "ANTM", "commodity": "nickel"}, context)
    assert isinstance(result, SkillResult)
    assert result.skill_id == "mining-commodity-divergence"
    assert "pearson_r" in result.metrics
    assert "divergence_class" in result.metrics


def test_skill_peer_valuation_benchmark(registry, mock_sectors_client):
    context = {"sectors_client": mock_sectors_client}
    result = registry.execute_skill("peer-valuation-benchmark", {"ticker": "ANTM"}, context)
    assert isinstance(result, SkillResult)
    assert result.skill_id == "peer-valuation-benchmark"
    assert "target_pe" in result.metrics
    assert "valuation_posture" in result.metrics
    assert "Non-Advisory" in result.summary
