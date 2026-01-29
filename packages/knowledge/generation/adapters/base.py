# generation/adapters/base.py
"""
Base adapter interface for documentation generation.
Each language adapter must implement this interface.
"""

from abc import ABC, abstractmethod
from pathlib import Path
from typing import List


class BaseAdapter(ABC):
    """Abstract base class for language-specific documentation adapters."""

    @abstractmethod
    def can_handle(self, file_path: Path) -> bool:
        """
        Check if this adapter can handle the given file.

        Args:
            file_path: Path to the source file.

        Returns:
            True if this adapter can process the file.
        """
        pass

    @abstractmethod
    def parse(self, file_path: Path, content: str) -> List[str]:
        """
        Parse and chunk the file content for processing.

        Args:
            file_path: Path to the source file.
            content: Raw file content.

        Returns:
            List of content chunks suitable for LLM processing.
        """
        pass

    @abstractmethod
    def get_prompt(self, file_path: Path, parsed_content: str) -> str:
        """
        Generate the LLM prompt for documentation generation.

        Args:
            file_path: Path to the source file.
            parsed_content: A single chunk of parsed content.

        Returns:
            The prompt string for the LLM.
        """
        pass

    def post_process(self, file_path: Path, outputs: List[str]) -> str:
        """
        Combine multiple LLM outputs into final documentation.
        Default implementation joins with newlines.

        Args:
            file_path: Path to the source file.
            outputs: List of LLM-generated outputs for each chunk.

        Returns:
            Final combined documentation string.
        """
        return "\n\n".join(outputs)
