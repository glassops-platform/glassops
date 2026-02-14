
from pathlib import Path
from typing import List
from .base import BaseAdapter

class VisualforceAdapter(BaseAdapter):
    """
    Adapter for processing Salesforce Visualforce Pages and Components.
    """

    def can_handle(self, file_path: Path) -> bool:
        # Visualforce extensions
        return file_path.suffix.lower() in {".page", ".component"}

    def parse(self, file_path: Path, content: str) -> List[str]:
        # Treat whole file as one chunk
        return [content]

    def get_prompt(self, file_path: Path, parsed_content: str) -> str:
        # Fallback prompt
        return f"Document the following Visualforce file:\n\n{parsed_content}"

    def validate_content(self, content: str) -> List[str]:
        # Basic XML/HTML validation could go here
        return []
