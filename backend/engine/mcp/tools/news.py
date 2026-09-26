"""Sectors News MCP Tool Definitions and Dispatcher.

Complies strictly with:
- Law 2: Strict Financial Non-Advisory Boundary (Factual observation)
- Sectors News & Disclosure Engine Architecture (Sectors API v2 /v2/news/)
- Anti-Prompt Injection: XML-wrapped <evidence_context>
"""

from typing import Any, Dict, List, Optional
from engine.mcp.tools.sectors import sanitize_ticker
from engine.sectors.client import SectorsAPIClient
from engine.sectors.news_engine import NewsItem, SectorsNewsEngine

# Backward compatibility alias
NewsHarvester = SectorsNewsEngine


def get_news_tool_definitions() -> List[Dict[str, Any]]:
    """Return standardized MCP schemas for Sectors news tools."""
    return [
        {
            "name": "news_harvest_market_news",
            "description": "Harvest curated exchange news and corporate disclosures directly from Sectors Financial API v2 (/v2/news/).",
            "inputSchema": {
                "type": "object",
                "properties": {
                    "ticker": {
                        "type": "string",
                        "description": "4-letter IDX stock ticker symbol (e.g. ANTM, BBRI)",
                    },
                    "company_name": {
                        "type": "string",
                        "description": "Official company name (optional, e.g. 'Aneka Tambang')",
                    },
                },
                "required": ["ticker"],
            },
        },
        {
            "name": "news_extract_article_content",
            "description": "Extract clean article text from URL or raw HTML using Trafilatura, isolated inside <evidence_context> tags for anti-prompt injection defense.",
            "inputSchema": {
                "type": "object",
                "properties": {
                    "url_or_html": {
                        "type": "string",
                        "description": "Web article URL or raw HTML string to clean",
                    },
                },
                "required": ["url_or_html"],
            },
        },
    ]


# Backward-compat alias for schemas
get_osint_tool_definitions = get_news_tool_definitions


def execute_news_tool(
    harvester: SectorsNewsEngine,
    sectors_client: Optional[SectorsAPIClient],
    name: str,
    arguments: Dict[str, Any],
) -> Any:
    """Execute a news tool call against SectorsNewsEngine."""
    if name in ("news_harvest_market_news", "osint_harvest_market_news"):
        ticker = sanitize_ticker(arguments.get("ticker", ""))
        company_name = arguments.get("company_name")

        sectors_news = []
        if sectors_client:
            try:
                if not company_name:
                    report = sectors_client.get_company_report(ticker)
                    company_name = report.get("company_name", ticker)
                sectors_news = sectors_client.get_news(ticker)
            except Exception:
                pass

        items: List[NewsItem] = harvester.harvest(
            ticker=ticker,
            company_name=company_name,
            sectors_news_items=sectors_news,
        )
        return [item.model_dump() for item in items]

    if name in ("news_extract_article_content", "osint_extract_article_content"):
        url_or_html = arguments.get("url_or_html", "")
        clean_text = harvester.sanitize_article_text(url_or_html)
        if not clean_text:
            return {
                "clean_text": "",
                "evidence_context": "<evidence_context>\n  <item error='Extraction failed or empty content'/>\n</evidence_context>",
            }
        wrapped = (
            f"<evidence_context>\n"
            f"  <item type='ARTICLE' source='web'>\n"
            f"    <content><![CDATA[{clean_text}]]></content>\n"
            f"  </item>\n"
            f"</evidence_context>"
        )
        return {
            "clean_text": clean_text,
            "evidence_context": wrapped,
        }

    raise ValueError(f"Unknown news tool: {name}")


# Backward-compat alias for executor
execute_osint_tool = execute_news_tool
