# generation/adapters/apex_adapter.py
"""
Salesforce Apex adapter for documentation generation.
"""

from pathlib import Path
from typing import List

from .base import BaseAdapter


class ApexAdapter(BaseAdapter):
    """Adapter for Salesforce Apex classes and triggers."""

    TARGET_CHUNK_SIZE = 24000

    def can_handle(self, file_path: Path) -> bool:
        return file_path.suffix in {".cls", ".trigger"}

    def parse(self, file_path: Path, content: str) -> List[str]:
        # Check for metadata file
        metadata_content = None
        meta_path = file_path.with_name(file_path.name + "-meta.xml")
        if meta_path.exists():
            try:
                metadata_content = meta_path.read_text(encoding="utf-8")
            except Exception as e:
                print(f"[WARNING] Failed to read metadata for {file_path}: {e}")

        if len(content) <= self.TARGET_CHUNK_SIZE:
            return [self._format_chunk(file_path, content, metadata=metadata_content)]

        chunks = []
        current_chunk = ""
        chunk_count = 1

        for line in content.split('\n'):
            if len(current_chunk) + len(line) > self.TARGET_CHUNK_SIZE:
                if current_chunk.strip():
                    chunks.append(self._format_chunk(file_path, current_chunk, chunk_count, metadata=metadata_content if chunk_count == 1 else None))
                    chunk_count += 1
                current_chunk = line + '\n'
            else:
                current_chunk += line + '\n'

        if current_chunk.strip():
            chunks.append(self._format_chunk(file_path, current_chunk, chunk_count if chunk_count > 1 else None, metadata=metadata_content if chunk_count == 1 else None))

        return chunks if chunks else [self._format_chunk(file_path, content, metadata=metadata_content)]

    def validate_content(self, content: str) -> List[str]:
        return []

    def _format_chunk(self, file_path: Path, content: str, part: int = None, metadata: str = None) -> str:
        part_suffix = f" (Part {part})" if part else ""
        file_type = "Apex Trigger" if file_path.suffix == ".trigger" else "Apex Class"

        formatted = f"File: {file_path} ({file_type}){part_suffix}\n\nContent:\n```apex\n{content}\n```"

        if metadata:
            formatted += f"\n\nMetadata:\n```xml\n{metadata}\n```"

        return formatted

    def get_prompt(self, file_path: Path, parsed_content: str) -> str:
        return f"""You are a Salesforce architect. Document the provided Apex code. Explain:
- The class/trigger purpose and responsibilities
- Key methods and their behavior
- Governor limit considerations
- Integration points with other Salesforce components
- Test coverage requirements

IMPORTANT: Output valid Markdown only. No conversational text. Do NOT wrap the output in ```markdown code blocks. Do not mention "NobleForge" or "Noble Forge" anywhere.

STRICT RULES:
- Do NOT use emojis.
- Do NOT use the words: utilize, crucial, showcasing, delve, underscores, watershed, groundbreaking.
- Use "We" or "I" when referring to the project maintainers.

Generate documentation for this Apex code:
{parsed_content}"""
