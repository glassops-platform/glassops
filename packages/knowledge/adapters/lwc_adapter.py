# generation/adapters/lwc_adapter.py
"""
Salesforce Lightning Web Component adapter for documentation generation.
"""

from pathlib import Path
from typing import List

import xml.etree.ElementTree as ET
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
        # Check for metadata file (only for main JS file)
        metadata_content = None
        if file_path.suffix == ".js":
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
        """
        Validate LWC content.
        - HTML: XML parsing
        - JS: Basic check (maybe node -c later, for now loose)
        - CSS: loose
        """
        # Heuristic: Detect type based on content? Or is this called with known extension?
        # The Validator extracts blocks and calls this. We need to know the 'language' context?
        # Actually, the Validator will call `_find_adapter` based on file extension if we pass a dummy path,
        # OR we need the adapter to sniff.
        
        # Taking a safe bet: If it looks like HTML (starts with <template), validate as XML.
        if content.strip().startswith("<template"):
            try:
                ET.fromstring(content)
            except ET.ParseError as e:
                return [f"LWC HTML Syntax Error: {e}"]
        
        return []

    def _format_chunk(self, file_path: Path, content: str, part: int = None, metadata: str = None) -> str:
        part_suffix = f" (Part {part})" if part else ""
        lang = "javascript" if file_path.suffix == ".js" else file_path.suffix[1:]

        formatted = f"File: {file_path} (Lightning Web Component){part_suffix}\n\nContent:\n```{lang}\n{content}\n```"

        if metadata:
            formatted += f"\n\nMetadata:\n```xml\n{metadata}\n```"

        return formatted

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
