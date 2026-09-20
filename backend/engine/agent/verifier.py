"""Fact Verification Gate & Guardrail Enforcement for Niskava Agent Engine.

Complies with:
- Law 2 (Strict Financial Non-Advisory Boundary): Redacts BUY/SELL recommendations, price targets.
- Section 5 of AGENTS.md (Confidence Scoring Rubric Invariant):
  1.00 = Extracted directly from Sectors API or IDXnet regulatory disclosures.
  0.95 = Official corporate press releases with explicit timestamp.
  0.85 = Accredited mainstream financial press (Kontan, Bisnis.com, CNBC Indonesia).
  0.65 = Unverified market commentary or rumors.
"""

import re
from typing import Any, Dict, List, Optional


PROHIBITED_ADVISORY_PATTERNS = [
    r"\bREKOMENDASI\s+(BELI|JUAL|HOLD|ACCUMULATE|REDUCE)\b",
    r"\bTARGET\s+HARGA\b",
    r"\bPRICE\s+TARGET\b",
    r"\bBUY\s+RECOMMENDATION\b",
    r"\bSELL\s+RECOMMENDATION\b",
    r"\bDISARANKAN\s+(MEMBELI|MENJUAL)\b",
]

DISCLAIMER_TEXT = (
    "DISCLAIMER: Niskava Agent adalah platform OSINT dan intelijen pasar faktual, "
    "bukan penasihat investasi. Seluruh temuan bersifat investigatif dan tidak boleh "
    "dianggap sebagai rekomendasi beli/jual atau nasihat keuangan personal."
)


class FactVerificationGate:
    """Second-Pass Fact Checker and Financial Guardrail Verifier."""

    @staticmethod
    def sanitize_non_advisory_text(text: str) -> str:
        """Scan text and redact prohibited investment advisory phrasing."""
        if not text:
            return text

        sanitized = text
        for pattern in PROHIBITED_ADVISORY_PATTERNS:
            sanitized = re.sub(pattern, "[REDACTED_NON_ADVISORY_BOUNDARY]", sanitized, flags=re.IGNORECASE)

        return sanitized

    @staticmethod
    def calibrate_confidence_score(
        status: str,
        sources: List[Dict[str, Any]],
        initial_score: float,
    ) -> float:
        """Calibrate confidence score based on standardized rubric invariant.
        
        Rubric:
        - 1.00: Official Sectors API or IDXnet disclosures.
        - 0.95: Official corporate press release.
        - 0.85: Accredited mainstream press (Kontan, Bisnis, CNBC Indonesia, etc.).
        - 0.65: Unverified market commentary / rumors.
        """
        if status != "SUPPORTED":
            return min(initial_score, 0.65)

        has_official = False
        has_press_release = False
        has_mainstream = False

        for src in sources:
            stype = str(src.get("source_type", "")).upper()
            sname = str(src.get("source_name", "")).lower()
            surl = str(src.get("source_url", "")).lower()

            if stype in ("OFFICIAL_DISCLOSURE", "SECTORS_API") or "idxnet" in sname or "idx.co.id" in surl:
                has_official = True
            elif "press release" in sname or "siaran pers" in sname or stype == "PRESS_RELEASE":
                has_press_release = True
            elif any(m in sname or m in surl for m in ("kontan", "bisnis", "cnbc", "investor", "bloomberg", "reuters", "marketbeat")):
                has_mainstream = True

        if has_official:
            return 1.00
        elif has_press_release:
            return 0.95
        elif has_mainstream:
            return 0.85
        else:
            return 0.65

    @classmethod
    def verify_finding(
        cls,
        finding: Dict[str, Any],
        raw_evidence_snippets: Optional[List[str]] = None,
        quant_context: Optional[Dict[str, Any]] = None,
    ) -> Dict[str, Any]:
        """Perform second-pass verification audit on an emitted finding."""
        verified = dict(finding)

        # 1. Sanitize text for advisory terms
        verified["title"] = cls.sanitize_non_advisory_text(verified.get("title", ""))
        verified["claim_text"] = cls.sanitize_non_advisory_text(verified.get("claim_text", ""))

        status = verified.get("verification_status", "UNCERTAIN")
        evidence = verified.get("evidence", [])
        raw_snippets = raw_evidence_snippets or []

        # 2. Check if claim marked SUPPORTED has actual supporting evidence in raw snippets or quant
        if status == "SUPPORTED":
            claim_text = (verified.get("claim_text", "") + " " + verified.get("title", "")).lower()
            
            # Simple keyword matching across raw evidence snippets
            has_matching_evidence = False
            if evidence:
                has_matching_evidence = True
            elif quant_context and any(k in claim_text for k in ("volume", "z-score", "return", "foreign", "anomali")):
                has_matching_evidence = True
            elif raw_snippets:
                # Check if any significant words in claim exist in raw snippets
                words = [w for w in re.findall(r"\w+", claim_text) if len(w) > 4]
                if words and any(any(w in snip.lower() for w in words) for snip in raw_snippets):
                    has_matching_evidence = True

            if not has_matching_evidence:
                status = "UNCERTAIN"
                verified["verification_status"] = "UNCERTAIN"
                verified["verification_note"] = "Downgraded to UNCERTAIN: No matching empirical text snippet found in OSINT or Quant context."

        # 3. Calibrate confidence score according to rubric
        initial_score = float(verified.get("confidence_score", 0.75))
        calibrated_score = cls.calibrate_confidence_score(status, evidence, initial_score)
        verified["confidence_score"] = calibrated_score

        # 4. Attach disclaimer
        verified["disclaimer"] = DISCLAIMER_TEXT

        return verified
