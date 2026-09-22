"""Unit tests for Niskava Tool Registry & MCP Tool Dispatching."""

from engine.agent.tools import NiskavaToolRegistry


def test_tool_registry_definitions(tmp_path):
    registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
    defs = registry.get_tool_definitions()

    tool_names = [d["name"] for d in defs]
    assert len(defs) == 4
    assert "execute_skill" in tool_names
    assert "query_sectors" in tool_names
    assert "search_osint" in tool_names
    assert "query_memory" in tool_names


def test_tool_execution_direct_and_mcp_aliases(tmp_path):
    registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)

    # Test direct names
    candles = registry.execute_tool("get_daily_candles", {"ticker": "ANTM", "days": 30})
    assert len(candles) == 30

    anomalies = registry.execute_tool("compute_quant_anomalies", {"ticker": "ANTM"})
    assert len(anomalies) >= 1
    assert anomalies[0]["z_score"] > 2.5

    # Test MCP aliases
    candles_mcp = registry.execute_tool("sectors_get_daily_candles", {"ticker": "ANTM", "days": 30})
    assert len(candles_mcp) == 30

    report = registry.execute_tool("sectors_get_company_report", {"ticker": "ANTM"})
    assert report["symbol"] == "ANTM"

    flow = registry.execute_tool("sectors_get_foreign_flow", {"ticker": "ANTM"})
    assert len(flow) >= 1

    suspensions = registry.execute_tool("sectors_get_suspensions", {"ticker": "ANTM"})
    assert len(suspensions) >= 1

    corp_actions = registry.execute_tool("sectors_get_corporate_actions", {"ticker": "ANTM"})
    assert len(corp_actions) >= 1

    filings = registry.execute_tool("sectors_get_filings", {"ticker": "ANTM"})
    assert len(filings) >= 1

    brokers = registry.execute_tool("sectors_get_broker_summary", {"ticker": "ANTM"})
    assert "top_buyers" in brokers

    peers = registry.execute_tool("sectors_get_subsector_peers", {"subsector": "metals-and-minerals-mining"})
    assert "peer_count" in peers

    mining = registry.execute_tool("sectors_get_mining_detail", {"slug": "aneka-tambang"})
    assert mining["commodity"] == "NICKEL"


def test_skill_tool_execution(tmp_path):
    registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
    res = registry.execute_tool("skill_market_anomaly_recon", {"ticker": "ANTM", "days": 30})
    assert isinstance(res, dict)
    assert res["skill_id"] == "market-anomaly-recon"
    assert "metrics" in res


def test_all_tool_and_skill_definitions_exposed(tmp_path):
    registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
    defs = registry.get_tool_definitions()
    names = [d["name"] for d in defs]
    assert len(defs) == 4, f"Expected 4 gateway definitions, got {len(defs)}: {names}"

    # Verify domain skills are still registered in the skills registry
    skill_defs = registry.skills_registry.get_all_tool_definitions()
    skill_names = [s["name"] for s in skill_defs]
    assert len(skill_defs) == 6, f"Expected 6 domain skills, got {len(skill_defs)}: {skill_names}"
    expected_skills = [
        "skill_mining_commodity_divergence",
        "skill_event_causality_audit",
        "skill_insider_bandarmology_forensic",
        "skill_financial_health_stress_test",
        "skill_market_anomaly_recon",
        "skill_peer_valuation_benchmark",
    ]
    for skill_name in expected_skills:
        assert skill_name in skill_names, f"Missing skill in registry: {skill_name}"



class TestLeanToolDefinitions:
    """Verifies that get_tool_definitions returns exactly the 4 gateway primitives."""

    def test_returns_exactly_four_definitions(self, tmp_path):
        registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
        defs = registry.get_tool_definitions()
        assert len(defs) == 4, (
            f"Expected 4 gateway tool definitions, got {len(defs)}: "
            f"{[d['name'] for d in defs]}"
        )

    def test_definition_names_are_the_four_gateways(self, tmp_path):
        registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
        defs = registry.get_tool_definitions()
        names = {d["name"] for d in defs}
        assert names == {"execute_skill", "query_sectors", "search_osint", "query_memory"}

    def test_each_definition_has_required_schema_keys(self, tmp_path):
        registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
        defs = registry.get_tool_definitions()
        for d in defs:
            assert "name" in d
            assert "description" in d
            assert "parameters" in d
            assert "properties" in d["parameters"]

    def test_query_sectors_description_lists_all_domains(self, tmp_path):
        registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
        defs = registry.get_tool_definitions()
        qs = next(d for d in defs if d["name"] == "query_sectors")
        desc = qs["description"]
        for domain in ["candles", "fundamentals", "foreign_flow", "suspensions",
                       "filings", "broker_summary", "corporate_actions"]:
            assert domain in desc, f"Domain '{domain}' missing from query_sectors description"

    def test_execute_skill_description_lists_all_skill_ids(self, tmp_path):
        registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
        defs = registry.get_tool_definitions()
        es = next(d for d in defs if d["name"] == "execute_skill")
        desc = es["description"]
        for skill_id in [
            "market_anomaly_recon",
            "event_causality_audit",
            "insider_bandarmology_forensic",
            "financial_health_stress_test",
            "mining_commodity_divergence",
            "peer_valuation_benchmark",
        ]:
            assert skill_id in desc, f"skill_id '{skill_id}' missing from execute_skill description"


def test_query_sectors_default_domain_fallback(tmp_path):
    """Verify query_sectors defaults domain gracefully when omitted or empty."""
    import pytest
    from engine.agent.tools import NiskavaToolRegistry
    registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)

    # Empty domain with stock ticker defaults to 'candles'
    res = registry.query_sectors(domain="", ticker="ANTM")
    assert isinstance(res, list)
    assert len(res) > 0

    # Empty domain with IHSG defaults to 'news' or 'candles'
    res_ihsg = registry.query_sectors(domain="", ticker="IHSG")
    assert res_ihsg is not None


def test_query_sectors_unknown_domain_raises_english_error(tmp_path):
    """Verify invalid domain error message is in English."""
    import pytest
    from engine.agent.tools import NiskavaToolRegistry
    registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
    with pytest.raises(ValueError) as exc:
        registry.query_sectors(domain="invalid_domain_xyz", ticker="ANTM")
    assert "Unknown domain" in str(exc.value)
    assert "Domain tidak dikenal" not in str(exc.value)



