"""Deterministic document parser for financial disclosures and reports.

Complies with:
- Law 1: Deterministic Before Generative (pure extraction, zero estimation)
- Law 4: Local-First Data Sovereignty (strictly local file parsing)
- Law 5: Credit Budget & Anti-Token Explosion (page-by-page chunking)
"""

import os
from typing import Any, Dict, List, Optional


def parse_document(file_path: str, max_pages: int = 150) -> Dict[str, Any]:
    """Deterministically parse a local document into structured pages and text.

    Supports:
    - PDF (.pdf) via pypdf
    - Text/Markdown/CSV/JSON (.txt, .md, .csv, .json) with ~2000-char virtual pages
    - Fallback for unsupported or binary formats

    Args:
        file_path: Local path to the file.
        max_pages: Maximum number of pages to extract (default: 150).

    Returns:
        Dict with filename, file_path, format, file_size, page_count, pages, and text.

    Raises:
        FileNotFoundError: If file does not exist on disk.
    """
    clean_path = os.path.expanduser(file_path)
    if not os.path.exists(clean_path):
        raise FileNotFoundError(f"File not found: {clean_path}")

    ext = os.path.splitext(clean_path)[1].lower()
    filename = os.path.basename(clean_path)
    file_size = os.path.getsize(clean_path)

    pages: List[Dict[str, Any]] = []

    if ext == ".pdf":
        try:
            from pypdf import PdfReader
            reader = PdfReader(clean_path)
            total = min(len(reader.pages), max_pages)
            if total == 0:
                pages.append({
                    "page_number": 1,
                    "text": "[Empty PDF document]",
                    "char_count": 0,
                })
            else:
                for i in range(total):
                    try:
                        page_text = reader.pages[i].extract_text() or ""
                    except Exception as page_exc:
                        page_text = f"[Error extracting text from page {i + 1}: {page_exc}]"
                    pages.append({
                        "page_number": i + 1,
                        "text": page_text.strip(),
                        "char_count": len(page_text),
                    })
        except Exception as exc:
            pages.append({
                "page_number": 1,
                "text": f"[Error extracting PDF: {str(exc)}]",
                "char_count": 0,
            })
    elif ext in (".txt", ".csv", ".json", ".md"):
        try:
            with open(clean_path, "r", encoding="utf-8", errors="replace") as f:
                content = f.read()

            chunk_size = 2000
            if not content.strip():
                pages.append({
                    "page_number": 1,
                    "text": "",
                    "char_count": 0,
                })
            else:
                chunks = [content[i:i + chunk_size] for i in range(0, len(content), chunk_size)]
                total = min(len(chunks), max_pages)
                for i in range(total):
                    chunk = chunks[i]
                    pages.append({
                        "page_number": i + 1,
                        "text": chunk.strip(),
                        "char_count": len(chunk),
                    })
        except Exception as exc:
            pages.append({
                "page_number": 1,
                "text": f"[Error reading text file: {str(exc)}]",
                "char_count": 0,
            })
    else:
        pages.append({
            "page_number": 1,
            "text": f"[Binary or unsupported format '{ext}']",
            "char_count": 0,
        })

    all_text = "\n\n".join([f"--- Page {p['page_number']} ---\n{p['text']}" for p in pages])
    return {
        "filename": filename,
        "file_path": clean_path,
        "format": ext.lstrip("."),
        "file_size": file_size,
        "page_count": len(pages),
        "pages": pages,
        "text": all_text,
    }


def search_document(
    parsed: Dict[str, Any],
    query: str,
    max_matches: int = 5,
) -> List[Dict[str, Any]]:
    """Search pages case-insensitively and return snippets with page numbers.

    Args:
        parsed: Dict returned by parse_document.
        query: Search string keyword or phrase.
        max_matches: Maximum number of snippet matches to return (default: 5).

    Returns:
        List of dicts with page_number and snippet.
    """
    if not query or not query.strip():
        return []

    matches: List[Dict[str, Any]] = []
    q_lower = query.lower().strip()

    for p in parsed.get("pages", []):
        text = p.get("text", "")
        text_lower = text.lower()
        if q_lower in text_lower:
            start_pos = 0
            while len(matches) < max_matches:
                idx = text_lower.find(q_lower, start_pos)
                if idx == -1:
                    break
                snippet_start = max(0, idx - 80)
                snippet_end = min(len(text), idx + len(q_lower) + 120)
                snippet = text[snippet_start:snippet_end].replace("\n", " ").strip()
                matches.append({
                    "page_number": p["page_number"],
                    "snippet": f"...{snippet}...",
                })
                start_pos = idx + len(q_lower)

            if len(matches) >= max_matches:
                break

    return matches


def get_document_page(parsed: Dict[str, Any], page_num: int) -> Optional[str]:
    """Retrieve text of a specific 1-indexed page.

    Args:
        parsed: Dict returned by parse_document.
        page_num: 1-indexed page number.

    Returns:
        Page text string, or None if page_num is invalid or out of range.
    """
    if page_num < 1:
        return None

    for p in parsed.get("pages", []):
        if p.get("page_number") == page_num:
            return p.get("text", "")

    return None
