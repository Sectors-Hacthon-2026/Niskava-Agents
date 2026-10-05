---
name: document-audit
description: Audit and inspect uploaded local financial documents, corporate disclosures, and prospectuses without context bloat.
triggers:
  - User uploaded financial reports, disclosures, or prospectus files
  - Document auditing, page extraction, and financial claim verification
tools:
  - inspect_document
---

# Document Audit Skill

## 1. Overview
The `document-audit` skill deterministically parses and audits local documents (PDF, TXT, CSV, JSON, MD) uploaded by analysts and researchers. In strict compliance with Law 4 (Local-First Data Sovereignty) and Law 5 (Credit Budget & Anti-Token Explosion), all document data is extracted and searched locally using deterministic Python parsers without cloud leakage or massive token consumption.

## 2. Analytical Procedure (SOP)
1. **Local Ingestion & Format Detection**:
   - Safely read uploaded document from local storage (`~/.niskava/uploads/`).
   - Extract metadata (filename, file size, page count, document format).
2. **Deterministic Extraction & Page Chunking**:
   - For PDF documents, extract text per page via `pypdf` with graceful per-page exception recovery.
   - For structured text/CSV/markdown/JSON, segment into ~2,000-character virtual pages.
3. **Targeted Inspection & Search**:
   - If a specific `page` is requested, return page-isolated text.
   - If a `query` keyword is provided, search case-insensitively and extract contextual snippets with page references.
4. **Three-Tier Verification Taxonomy (Law 2)**:
   - `SUPPORTED`: Claims matching exact figures and text in official uploaded documents.
   - `UNCERTAIN`: Claims where context is ambiguous or requires external market validation.
   - `CONTRADICTED`: Claims refuted by contradictory numbers or statements in the document.
5. **Discrete Confidence Rubric**:
   - Score `1.00`: Direct extraction from formal financial reports or official IDXnet disclosures.
