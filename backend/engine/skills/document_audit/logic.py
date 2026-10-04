"""Deterministic execution logic for document-audit skill.

Complies strictly with:
- Law 1: Deterministic Before Generative
- Law 2: Strict Financial Non-Advisory Boundary (3-Tier Taxonomy)
- Law 4: Local-First Data Sovereignty
- Law 5: Credit Budget & Anti-Token Explosion
"""

import os
from pathlib import Path
from typing import Any, Dict, List, Optional

from engine.skills.base import BaseSkill, SkillResult
from engine.skills.document_audit.parser import (
    get_document_page,
    parse_document,
    search_document,
)


class DocumentAuditSkill(BaseSkill):
    """Executes deterministic document inspection, page extraction, and keyword auditing."""

    def __init__(self, skill_dir: Optional[Path] = None):
        super().__init__(skill_dir or Path(__file__).parent)

    def get_tool_definition(self) -> Dict[str, Any]:
        return {
            "name": "skill_document_audit",
            "description": "Audit, inspect, and analyze uploaded local financial disclosures, reports, and files.",
            "parameters": {
                "type": "object",
                "properties": {
                    "doc_path": {
                        "type": "string",
                        "description": "Local path of the uploaded document (PDF, TXT, CSV, etc.) to inspect.",
                    },
                    "query": {
                        "type": "string",
                        "description": "Optional keyword or search query within the document.",
                    },
                    "page": {
                        "type": "integer",
                        "description": "Optional 1-indexed page number to extract specifically.",
                    },
                    "ticker": {
                        "type": "string",
                        "description": "Optional IDX stock ticker associated with the document.",
                    },
                },
                "required": ["doc_path"],
            },
        }

    def execute(self, arguments: Dict[str, Any], context: Dict[str, Any]) -> SkillResult:
        doc_path = arguments.get("doc_path") or arguments.get("file_path", "")
        query = arguments.get("query")
        page = arguments.get("page")
        ticker = arguments.get("ticker", "").upper().strip()

        if not doc_path:
            return SkillResult(
                skill_id=self.skill_id,
                verification_status="UNCERTAIN",
                confidence_score=0.65,
                metrics={"error": "Missing doc_path parameter"},
                evidence=[],
                summary="Audit dokumen gagal: path file tidak disertakan.",
            )

        clean_path = os.path.expanduser(str(doc_path).strip())
        if not os.path.exists(clean_path):
            return SkillResult(
                skill_id=self.skill_id,
                verification_status="UNCERTAIN",
                confidence_score=0.65,
                metrics={"error": f"File not found: {clean_path}"},
                evidence=[],
                summary=f"Dokumen tidak ditemukan di penyimpanan lokal: {clean_path}.",
            )

        try:
            parsed = parse_document(clean_path)
        except Exception as exc:
            return SkillResult(
                skill_id=self.skill_id,
                verification_status="UNCERTAIN",
                confidence_score=0.65,
                metrics={"error": str(exc)},
                evidence=[],
                summary=f"Gagal memproses dokumen {os.path.basename(clean_path)}: {str(exc)}.",
            )

        emitter = context.get("emitter")
        if emitter and callable(emitter):
            emitter({
                "event": "document_parsed",
                "filename": parsed["filename"],
                "format": parsed["format"],
                "page_count": parsed["page_count"],
                "file_size": parsed["file_size"],
            })

        evidence: List[Dict[str, Any]] = []
        metrics: Dict[str, Any] = {
            "filename": parsed["filename"],
            "format": parsed["format"],
            "file_size": parsed["file_size"],
            "page_count": parsed["page_count"],
        }
        if ticker:
            metrics["ticker"] = ticker

        if page is not None:
            try:
                page_int = int(page)
                page_text = get_document_page(parsed, page_int)
                if page_text is not None:
                    metrics["inspected_page"] = page_int
                    evidence.append({
                        "type": "DOCUMENT_PAGE_EXTRACT",
                        "page": page_int,
                        "content_length": len(page_text),
                        "snippet": page_text[:500],
                        "source": parsed["filename"],
                    })
                    summary = f"Audit {parsed['filename']}: Berhasil mengekstrak teks halaman {page_int} ({len(page_text)} karakter)."
                else:
                    summary = f"Audit {parsed['filename']}: Halaman {page_int} berada di luar jangkauan (total {parsed['page_count']} halaman)."
            except (ValueError, TypeError):
                summary = f"Audit {parsed['filename']}: Nomor halaman '{page}' tidak valid."
        elif query:
            matches = search_document(parsed, str(query).strip(), max_matches=5)
            metrics["search_query"] = query
            metrics["match_count"] = len(matches)
            for m in matches:
                evidence.append({
                    "type": "DOCUMENT_SEARCH_MATCH",
                    "page": m["page_number"],
                    "snippet": m["snippet"],
                    "source": parsed["filename"],
                })
            summary = f"Audit {parsed['filename']}: Ditemukan {len(matches)} kemunculan kata kunci '{query}' pada {parsed['page_count']} halaman."
        else:
            first_preview = parsed["pages"][0]["text"][:400] if parsed["pages"] else ""
            evidence.append({
                "type": "DOCUMENT_METADATA",
                "page_count": parsed["page_count"],
                "snippet": first_preview,
                "source": parsed["filename"],
            })
            snippet_preview = f" Cuplikan: {first_preview[:120]}..." if first_preview else ""
            summary = f"Audit {parsed['filename']}: Terbaca {parsed['page_count']} halaman (format: {parsed['format']}, ukuran: {parsed['file_size']} bytes).{snippet_preview}"

        return SkillResult(
            skill_id=self.skill_id,
            verification_status="SUPPORTED",
            confidence_score=1.00,
            metrics=metrics,
            evidence=evidence,
            summary=summary,
        )
