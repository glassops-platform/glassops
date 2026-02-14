
from pathlib import Path
from typing import List
from .base import BaseAdapter

class AuraAdapter(BaseAdapter):
    """
    Adapter for processing Salesforce Aura Components.
    """

    def can_handle(self, file_path: Path) -> bool:
        # Check if file is in an 'aura' directory
        if "aura" not in file_path.parts:
            return False

        # Common Aura file extensions
        valid_extensions = {".cmp", ".app", ".evt", ".intf", ".tokens", ".js", ".css", ".auradoc"}
        return file_path.suffix.lower() in valid_extensions

    def parse(self, file_path: Path, content: str) -> List[str]:
        # Treat each file in the bundle as a single chunk

        # Check for metadata file (only for main definition files)
        metadata_content = ""
        if file_path.suffix in {".cmp", ".app", ".evt", ".intf", ".tokens"}:
            meta_path = file_path.with_name(file_path.name + "-meta.xml")
            if meta_path.exists():
                try:
                    meta_text = meta_path.read_text(encoding="utf-8")
                    metadata_content = f"\n\nMetadata:\n```xml\n{meta_text}\n```"
                except Exception as e:
                    print(f"[WARNING] Failed to read metadata for {file_path}: {e}")

        return [content + metadata_content]

    def get_prompt(self, file_path: Path, parsed_content: str) -> str:
        # Fallback prompt if not found in config
        return f"Document the following Aura component file:\n\n{parsed_content}"

    def validate_content(self, content: str) -> List[str]:
        # Basic validation could be added here
        return []
