# generation/adapters/yaml_adapter.py
"""
YAML configuration adapter for documentation generation.
"""

from pathlib import Path
from typing import List

from .base import BaseAdapter


class YAMLAdapter(BaseAdapter):
    """Adapter for YAML configuration files."""

    TARGET_CHUNK_SIZE = 24000

    def can_handle(self, file_path: Path) -> bool:
        return file_path.suffix in {".yml", ".yaml"}

    def parse(self, file_path: Path, content: str) -> List[str]:
        if len(content) <= self.TARGET_CHUNK_SIZE:
            return [self._format_chunk(file_path, content)]

        chunks = []
        current_chunk = ""
        chunk_count = 1

        for line in content.split('\n'):
            if len(current_chunk) + len(line) > self.TARGET_CHUNK_SIZE:
                if current_chunk.strip():
                    chunks.append(self._format_chunk(file_path, current_chunk, chunk_count))
                    chunk_count += 1
                current_chunk = line + '\n'
            else:
                current_chunk += line + '\n'

        if current_chunk.strip():
            chunks.append(self._format_chunk(file_path, current_chunk, chunk_count if chunk_count > 1 else None))

        return chunks if chunks else [self._format_chunk(file_path, content)]

    def _format_chunk(self, file_path: Path, content: str, part: int = None) -> str:
        part_suffix = f" (Part {part})" if part else ""
        return f"File: {file_path}{part_suffix}\n\nContent:\n```yaml\n{content}\n```"

    def get_prompt(self, file_path: Path, parsed_content: str) -> str:
        return f"""You are a DevOps engineer and technical writer. Your task is to document the provided YAML configuration. Explain the purpose of the configuration, the structure, and what each key controls.

IMPORTANT: Output valid Markdown only. No conversational text. Do NOT wrap the output in ```markdown code blocks. Do not mention "NobleForge" or "Noble Forge" anywhere.

STRICT RULES:
- Do NOT use emojis.
- Do NOT use the words: utilize, crucial, showcasing, delve, underscores, watershed, groundbreaking.
- Use "We" or "I" when referring to the project maintainers.

Generate documentation for this YAML content:
{parsed_content}"""
