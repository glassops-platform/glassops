# generation/validator.py
"""
Content validator for generated documentation.
Detects conversational filler and other quality issues.
"""

import re
from typing import List


class Validator:
    """Validates generated documentation for quality issues."""

    BANNED_PHRASES = [
        "Here is the document",
        "I hope this helps",
        "Let me know if",
        "Feel free to",
        "As requested",
        "Sure, here is",
        "Here's the",
    ]

    BANNED_WORDS = [
        "utilize",
        "crucial",
        "showcasing",
        "delve",
        "underscores",
        "watershed",
        "groundbreaking",
    ]

    @classmethod
    def validate(cls, content: str, file_path: str = "") -> List[str]:
        """
        Validate generated content for quality issues.

        Args:
            content: The generated documentation content.
            file_path: Optional file path for context.

        Returns:
            List of validation error/warning messages.
        """
        errors = []

        # 1. Check Frontmatter
        if not content.startswith("---"):
            errors.append("Missing frontmatter block")

        content_lower = content.lower()

        # 2. Check for Conversational Filler
        for phrase in cls.BANNED_PHRASES:
            if phrase.lower() in content_lower:
                errors.append(f'Conversational phrase detected: "{phrase}"')

        # 3. Check for Banned Words
        for word in cls.BANNED_WORDS:
            if word.lower() in content_lower:
                errors.append(f'Banned word detected: "{word}"')

        # 4. Check for "NobleForge" mentions
        if "nobleforge" in content_lower or "noble forge" in content_lower:
            errors.append('Banned term detected: "NobleForge"')

        return errors
