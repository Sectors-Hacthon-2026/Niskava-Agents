"""Base specification and protocol for Layer 3 Modular Domain Skills.

Complies strictly with:
- Law 1: Deterministic Before Generative (NumPy firewall before LLM)
- Law 2: Strict Financial Non-Advisory Boundary (3-Tier Taxonomy: SUPPORTED, UNCERTAIN, CONTRADICTED)
- Discrete Confidence Rubric (1.00, 0.95, 0.85, 0.75, 0.65, 0.55)
"""

from abc import ABC, abstractmethod
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, List, Literal, Optional
from pydantic import BaseModel, Field


def _parse_yaml_frontmatter(raw_text: str) -> Dict[str, Any]:
    """Lightweight zero-dependency YAML frontmatter parser for SKILL.md."""
    data: Dict[str, Any] = {}
    current_list_key = None
    for line in raw_text.splitlines():
        trimmed = line.strip()
        if not trimmed or trimmed.startswith("#"):
            continue
        if trimmed.startswith("- ") and current_list_key:
            data.setdefault(current_list_key, []).append(trimmed[2:].strip().strip("'\""))
        elif ":" in trimmed:
            k, v = trimmed.split(":", 1)
            k = k.strip()
            v = v.strip().strip("'\"")
            if not v:
                current_list_key = k
                data[k] = []
            else:
                current_list_key = None
                data[k] = v
    return data


class SkillMetadata(BaseModel):
    """Metadata extracted from SKILL.md YAML frontmatter."""
    name: str
    description: str
    triggers: List[str] = Field(default_factory=list)
    tools: List[str] = Field(default_factory=list)


class SkillResult(BaseModel):
    """Standardized output schema for every Domain Skill investigation."""
    skill_id: str
    verification_status: Literal["SUPPORTED", "UNCERTAIN", "CONTRADICTED"]
    confidence_score: float
    metrics: Dict[str, Any] = Field(default_factory=dict)
    evidence: List[Dict[str, Any]] = Field(default_factory=list)
    summary: str
    timestamp: str = Field(default_factory=lambda: datetime.now().isoformat() + "Z")

    def to_dict(self) -> Dict[str, Any]:
        return self.model_dump()


class BaseSkill(ABC):
    """Abstract base class for all Niskava Layer 3 Domain Skills."""

    def __init__(self, skill_dir: Path):
        self.skill_dir = skill_dir
        self.skill_id = skill_dir.name.replace("_", "-")
        self.metadata, self.markdown_instructions = self._load_skill_md()

    def _load_skill_md(self) -> tuple[SkillMetadata, str]:
        """Parse SKILL.md frontmatter and markdown body."""
        md_file = self.skill_dir / "SKILL.md"
        if not md_file.exists():
            return (
                SkillMetadata(
                    name=self.skill_id,
                    description=f"Analytical skill {self.skill_id}",
                    triggers=[],
                    tools=[],
                ),
                "",
            )

        content = md_file.read_text(encoding="utf-8")
        if content.startswith("---"):
            parts = content.split("---", 2)
            if len(parts) >= 3:
                frontmatter_raw = parts[1]
                body = parts[2].strip()
                data = _parse_yaml_frontmatter(frontmatter_raw)
                meta = SkillMetadata(
                    name=data.get("name", self.skill_id),
                    description=data.get("description", ""),
                    triggers=data.get("triggers", []),
                    tools=data.get("tools", []),
                )
                return meta, body
        return (
            SkillMetadata(
                name=self.skill_id,
                description=f"Analytical skill {self.skill_id}",
                triggers=[],
                tools=[],
            ),
            content,
        )

    @property
    def name(self) -> str:
        return self.metadata.name

    @property
    def description(self) -> str:
        return self.metadata.description

    @abstractmethod
    def execute(self, arguments: Dict[str, Any], context: Dict[str, Any]) -> SkillResult:
        """Execute the deterministic analytical procedure."""
        pass

    @abstractmethod
    def get_tool_definition(self) -> Dict[str, Any]:
        """Return JSON-schema compatible tool definition for ReAct LLM function calling."""
        pass
