# generation/adapters/python.py
"""
Python language adapter for documentation generation.
"""

from pathlib import Path
from typing import List

from .base import BaseAdapter


class PythonAdapter(BaseAdapter):
    """Adapter for Python source files."""

    TARGET_CHUNK_SIZE = 24000  # ~6k tokens

    def can_handle(self, file_path: Path) -> bool:
        return file_path.suffix == ".py"

    def parse(self, file_path: Path, content: str) -> List[str]:
        """
        Parse Python file content into chunks.
        Uses line-based chunking with awareness of class/function boundaries.
        """
        if len(content) <= self.TARGET_CHUNK_SIZE:
            return [self._format_chunk(file_path, content)]

        chunks = []
        current_chunk = ""
        chunk_count = 1

        for line in content.split('\n'):
            # Check if adding this line would exceed chunk size
            if len(current_chunk) + len(line) > self.TARGET_CHUNK_SIZE:
                # Try to break at a class or function definition
                if current_chunk.strip():
                    chunks.append(self._format_chunk(file_path, current_chunk, chunk_count))
                    chunk_count += 1
                current_chunk = line + '\n'
            else:
                current_chunk += line + '\n'

        # Add remaining content
        if current_chunk.strip():
            chunks.append(self._format_chunk(file_path, current_chunk, chunk_count if chunk_count > 1 else None))

        return chunks if chunks else [self._format_chunk(file_path, content)]

    def _format_chunk(self, file_path: Path, content: str, part: int = None) -> str:
        """Format a chunk with file context."""
        part_suffix = f" (Part {part})" if part else ""
        return f"File: {file_path}{part_suffix}\n\nCode Content:\n```python\n{content}\n```"

    def get_prompt(self, file_path: Path, parsed_content: str) -> str:
        return f"""You are a principal architect. Your job is to translate the provided Python source code into a high-level, concise, but comprehensive document that is easily understood by both highly technical and non-technical audiences.

IMPORTANT: Generate ONLY the document content itself. Do NOT include any conversational filler, preambles (e.g., "Here is the document..."), post-generation suggestions, or follow-up questions. Do NOT wrap the output in ```markdown code blocks. Do not mention "NobleForge" or "Noble Forge" anywhere.

STRICT RULES:
- Do NOT use emojis.
- Do NOT use the words: utilize, crucial, showcasing, delve, underscores, watershed, groundbreaking.
- Use "We" or "I" when referring to the project maintainers.
- You MAY use "You" when giving specific instructions to the user.

Focus on:
- Module purpose and responsibilities
- Key classes and their roles
- Important functions and their behavior
- Type hints and their significance
- Any notable patterns or design decisions

Generate documentation for the following Python file:
{parsed_content}"""
