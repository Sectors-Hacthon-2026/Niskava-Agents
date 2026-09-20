"""Unit tests for Dual-Engine OSINT harvester and XML context wrapping."""

from engine.osint.harvester import DualEngineOSINTHarvester, OSINTItem


def test_osint_mock_harvest():
    harvester = DualEngineOSINTHarvester(mock_mode=True)
    items = harvester.harvest("ANTM")
    assert len(items) >= 1
    assert "Smelter" in items[0].title
    assert items[0].is_disclosure is True


def test_wrap_in_evidence_context():
    harvester = DualEngineOSINTHarvester(mock_mode=True)
    items = [
        OSINTItem(
            title="Uji Coba Smelter Baru",
            source_name="IDX Channel",
            source_url="https://idxchannel.com",
            publication_date="2026-09-12",
            snippet="Penyelesaian hilirisasi nikel",
            source_type="DISCLOSURE",
        )
    ]
    xml_output = harvester.wrap_in_evidence_context(items)
    assert "<evidence_context>" in xml_output
    assert "</evidence_context>" in xml_output
    assert "Uji Coba Smelter Baru" in xml_output
    assert 'type="DISCLOSURE"' in xml_output
