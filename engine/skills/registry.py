"""Dynamic Modular Domain Skills Registry for Niskava Agent.

Discovers, registers, and dispatches Layer 3 Domain Skills.
Exposes them as deterministic tools for ReAct Agent and MCP server.
"""

import importlib
import logging
from pathlib import Path
from typing import Any, Dict, List, Optional

from engine.skills.base import BaseSkill, SkillResult

logger = logging.getLogger("niskava.skills.registry")


class SkillsRegistry:
    """Registry managing dynamic discovery and execution of analytical Domain Skills."""

    def __init__(self, skills_dir: Optional[Path] = None):
        self.skills_dir = skills_dir or Path(__file__).parent
        self._skills: Dict[str, BaseSkill] = {}
        self.discover_skills()

    def register(self, skill: BaseSkill) -> None:
        """Register a single Domain Skill instance."""
        self._skills[skill.skill_id] = skill
        # Also alias with underscore
        self._skills[skill.skill_id.replace("-", "_")] = skill

    def discover_skills(self) -> None:
        """Scan skills directory and instantiate all valid skill packages."""
        if not self.skills_dir.exists():
            return

        for item in self.skills_dir.iterdir():
            if item.is_dir() and not item.name.startswith(("_", ".")):
                logic_file = item / "logic.py"
                if logic_file.exists():
                    try:
                        module_name = f"engine.skills.{item.name}.logic"
                        mod = importlib.import_module(module_name)
                        for attr_name in dir(mod):
                            attr = getattr(mod, attr_name)
                            if (
                                isinstance(attr, type)
                                and issubclass(attr, BaseSkill)
                                and attr is not BaseSkill
                            ):
                                skill_instance = attr(skill_dir=item)
                                self.register(skill_instance)
                                break
                    except Exception as e:
                        logger.warning(f"Failed to load skill from {item.name}: {e}")

    def get_skill(self, skill_id: str) -> Optional[BaseSkill]:
        """Retrieve skill by kebab-case or snake_case ID."""
        key = skill_id.lower().replace("_", "-")
        return self._skills.get(key) or self._skills.get(skill_id)

    def list_skills(self) -> List[BaseSkill]:
        """Return unique registered skills."""
        seen = set()
        unique = []
        for s in self._skills.values():
            if s.skill_id not in seen:
                seen.add(s.skill_id)
                unique.append(s)
        return unique

    def get_all_tool_definitions(self) -> List[Dict[str, Any]]:
        """Return function calling tool definitions for all registered skills."""
        return [skill.get_tool_definition() for skill in self.list_skills()]

    def get_skills_prompt_guidance(self) -> str:
        """Generate formatted prompt string describing all available skills for ReAct Agent."""
        lines = ["=== AVAILABLE DOMAIN SKILLS (Layer 3 SOPs) ==="]
        for skill in self.list_skills():
            triggers_str = ", ".join(skill.metadata.triggers[:2]) if skill.metadata.triggers else "General analysis"
            lines.append(f"- {skill.skill_id}: {skill.description}")
            lines.append(f"  Triggers: {triggers_str}")
        return "\n".join(lines)

    def execute_skill(
        self,
        skill_id: str,
        arguments: Dict[str, Any],
        context: Optional[Dict[str, Any]] = None,
    ) -> SkillResult:
        """Dispatch execution to the requested skill."""
        skill = self.get_skill(skill_id)
        if not skill:
            raise ValueError(f"Skill '{skill_id}' is not registered in SkillsRegistry.")
        return skill.execute(arguments=arguments, context=context or {})
