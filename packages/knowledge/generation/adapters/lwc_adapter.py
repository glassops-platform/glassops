# generation/adapters/lwc_adapter.py
"""
Salesforce Lightning Web Component adapter for documentation generation.
"""

from pathlib import Path
from typing import List

from .base import BaseAdapter


class LWCAdapter(BaseAdapter):
    """Adapter for Salesforce Lightning Web Components."""

    TARGET_CHUNK_SIZE = 24000

    def can_handle(self, file_path: Path) -> bool:
        # LWC files are in lwc/ directories and are .js, .html, or .css
        if "lwc" not in file_path.parts:
            return False
        return file_path.suffix in {".js", ".html", ".css"}

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
        lang = "javascript" if file_path.suffix == ".js" else file_path.suffix[1:]
        return f"File: {file_path} (Lightning Web Component){part_suffix}\n\nContent:\n```{lang}\n{content}\n```"

    def get_prompt(self, file_path: Path, parsed_content: str) -> str:
        return f"""You are a Salesforce Lightning expert. Document the provided Lightning Web Component file. Explain:
- The component's purpose and functionality
- Public properties (@api decorated)
- Wire adapters and their data sources
- Event handling (dispatching and listening)
- Lifecycle hooks used
- CSS styling approach (if CSS file)

IMPORTANT: Output valid Markdown only. No conversational text. Do NOT wrap the output in ```markdown code blocks. Do not mention "NobleForge" or "Noble Forge" anywhere.

STRICT RULES:
- Do NOT use emojis.
- Do NOT use the words: utilize, crucial, showcasing, delve, underscores, watershed, groundbreaking.
- Use "We" or "I" when referring to the project maintainers.

Generate documentation for this LWC file:
{parsed_content}"""
