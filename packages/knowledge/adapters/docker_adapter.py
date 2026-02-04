# generation/adapters/docker_adapter.py
"""
Dockerfile adapter for documentation generation.
"""

from pathlib import Path
from typing import List

from .base import BaseAdapter


class DockerAdapter(BaseAdapter):
    """Adapter for Dockerfiles."""

    TARGET_CHUNK_SIZE = 24000

    def can_handle(self, file_path: Path) -> bool:
        return file_path.name == "Dockerfile" or file_path.name.startswith("Dockerfile.")

    def parse(self, file_path: Path, content: str) -> List[str]:
        # Dockerfiles are usually small, no chunking needed
        return [self._format_chunk(file_path, content)]

    def validate_content(self, content: str) -> List[str]:
        return []

    def _format_chunk(self, file_path: Path, content: str, part: int = None) -> str:
        part_suffix = f" (Part {part})" if part else ""
        return f"File: {file_path}{part_suffix}\n\nContent:\n```dockerfile\n{content}\n```"

    def get_prompt(self, file_path: Path, parsed_content: str) -> str:
        return f"""You are a DevOps expert. Document the provided Dockerfile. Explain:
- The base image and why it was chosen
- Each stage (if multi-stage build)
- Key instructions and their purpose
- Security considerations
- How to build and run the container

IMPORTANT: Output valid Markdown only. No conversational text. Do NOT wrap the output in ```markdown code blocks. Do not mention "NobleForge" or "Noble Forge" anywhere.

STRICT RULES:
- Do NOT use emojis.
- Do NOT use the words: utilize, crucial, showcasing, delve, underscores, watershed, groundbreaking.
- Use "We" or "I" when referring to the project maintainers.

Generate documentation for this Dockerfile:
{parsed_content}"""
