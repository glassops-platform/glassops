# generation/adapters/go.py
"""
Go language adapter for documentation generation.
"""

import re
from pathlib import Path
from typing import List

from .base import BaseAdapter


class GoAdapter(BaseAdapter):
    """Adapter for Go source files."""

    TARGET_CHUNK_SIZE = 24000  # ~6k tokens (conservative)

    def can_handle(self, file_path: Path) -> bool:
        return file_path.suffix == ".go"

    def parse(self, file_path: Path, content: str) -> List[str]:
        """
        Parse Go file content into chunks.
        Attempts to split by function/type declarations for semantic boundaries.
        Falls back to line-based chunking for large files.
        """
        if len(content) <= self.TARGET_CHUNK_SIZE:
            return [self._format_chunk(file_path, content)]

        # Try to split by top-level declarations (func, type, const, var)
        # This regex finds the start of top-level declarations
        declaration_pattern = re.compile(
            r'^(?:func |type |const |var |\n// )',
            re.MULTILINE
        )

        chunks = []
        current_chunk = ""
        chunk_count = 1

        lines = content.split('\n')
        current_start = 0

        for i, line in enumerate(lines):
            # Check if this line starts a new declaration
            is_declaration_start = declaration_pattern.match(line) is not None

            if is_declaration_start and len(current_chunk) > self.TARGET_CHUNK_SIZE // 2:
                # Save current chunk and start new one
                if current_chunk.strip():
                    chunks.append(self._format_chunk(file_path, current_chunk, chunk_count))
                    chunk_count += 1
                current_chunk = line + '\n'
            else:
                current_chunk += line + '\n'

                # Force split if chunk gets too large
                if len(current_chunk) > self.TARGET_CHUNK_SIZE:
                    chunks.append(self._format_chunk(file_path, current_chunk, chunk_count))
                    chunk_count += 1
                    current_chunk = ""

        # Don't forget the last chunk
        if current_chunk.strip():
            chunks.append(self._format_chunk(file_path, current_chunk, chunk_count if chunk_count > 1 else None))

        return chunks if chunks else [self._format_chunk(file_path, content)]

    def _format_chunk(self, file_path: Path, content: str, part: int = None) -> str:
        """Format a chunk with file context."""
        part_suffix = f" (Part {part})" if part else ""
        return f"File: {file_path}{part_suffix}\n\nCode Content:\n```go\n{content}\n```"

    def get_prompt(self, file_path: Path, parsed_content: str) -> str:
        return f"""You are a principal architect. Your job is to translate the provided Go source code into a high-level, concise, but comprehensive document that is easily understood by both highly technical and non-technical audiences.

IMPORTANT: Generate ONLY the document content itself. Do NOT include any conversational filler, preambles (e.g., "Here is the document..."), post-generation suggestions, or follow-up questions. Do NOT wrap the output in ```markdown code blocks. Do not mention "NobleForge" or "Noble Forge" anywhere.

STRICT RULES:
- Do NOT use emojis.
- Do NOT use the words: utilize, crucial, showcasing, delve, underscores, watershed, groundbreaking.
- Use "We" or "I" when referring to the project maintainers.
- You MAY use "You" when giving specific instructions to the user.

Focus on:
- Package purpose and responsibilities
- Key types, interfaces, and their roles
- Important functions and their behavior
- Error handling patterns
- Concurrency patterns (goroutines, channels) if present
- Any notable design decisions

Generate documentation for the following Go file:
{parsed_content}"""
