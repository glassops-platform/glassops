import re
import re
from typing import List, Optional
from pathlib import Path

from ..adapters import (
    BaseAdapter,
    GoAdapter,
    PythonAdapter,
    LWCAdapter,
    ApexAdapter,
    YAMLAdapter,
    JSONAdapter,
    DockerAdapter,
    TerraformAdapter
)


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
    def get_adapter_for_lang(cls, lang: str) -> Optional[BaseAdapter]:
        """Factory to get adapter based on language string."""
        lang = lang.lower()
        if lang in ["python", "py"]:
            return PythonAdapter()
        elif lang in ["go", "golang"]:
            return GoAdapter()
        elif lang in ["html", "xml", "svg"]:
             # LWC Adapter handles HTML validation for now, could act as generic XML validator too
             return LWCAdapter()
        return None

    @classmethod
    def extract_code_blocks(cls, content: str) -> List[tuple[str, str]]:
        """
        Extract code blocks and their language from markdown content.
        
        Returns:
            List of (language, code) tuples.
        """
        # Matches ```lang ... ```
        pattern = r"```(\w+)\n(.*?)```"
        return re.findall(pattern, content, re.DOTALL)

    @classmethod
    def validate(cls, content: str, file_path: str = "") -> dict:
        """
        Validate generated content for quality issues and syntax.

        Args:
            content: The generated documentation content.
            file_path: Optional file path for context.

        Returns:
            Dictionary with keys 'passes', 'warnings', 'errors'.
        """
        results = {
            "passes": [],
            "warnings": [],
            "errors": []
        }

        # 1. Check Frontmatter
        if not content.startswith("---"):
            results["errors"].append("Missing frontmatter block")
        else:
            results["passes"].append("Frontmatter block present")

        content_lower = content.lower()

        # 2. Check for Conversational Filler
        found_phrases = []
        for phrase in cls.BANNED_PHRASES:
            if phrase.lower() in content_lower:
                found_phrases.append(phrase)
                results["warnings"].append(f'Conversational phrase detected: "{phrase}"')
        
        if not found_phrases:
            results["passes"].append("No conversational filler detected")

        # 3. Check for Banned Words
        found_words = []
        for word in cls.BANNED_WORDS:
            if word.lower() in content_lower:
                found_words.append(word)
                results["warnings"].append(f'Banned word detected: "{word}"')

        if not found_words:
            results["passes"].append("No banned words detected")

        # 4. Check for "NobleForge" mentions
        if "nobleforge" in content_lower or "noble forge" in content_lower:
            results["warnings"].append('Banned term detected: "NobleForge"')
        else:
            results["passes"].append('No "NobleForge" mentions')

        # 5. Delegate Code Block Validation to Adapters
        blocks = cls.extract_code_blocks(content)
        if not blocks:
            results["passes"].append("No code blocks to validate")
        
        for i, (lang, code) in enumerate(blocks):
            adapter = cls.get_adapter_for_lang(lang)
            if adapter:
                block_errors = adapter.validate_content(code)
                if block_errors:
                    for err in block_errors:
                        results["errors"].append(f"Code block {i+1} ({lang}) error: {err}")
                else:
                    results["passes"].append(f"Code block {i+1} ({lang}) passed validation")
            else:
                # results["warnings"].append(f"Skipped validation for code block {i+1} ({lang}): No adapter found")
                pass

        return results

    @staticmethod
    def print_report(results: dict):
        """Pretty-print the validation report."""
        if results["errors"] or results["warnings"]:
            print(f"   [WARNING] Validation Report:")
            
            if results["errors"]:
                print(f"      [ERROR] Errors:")
                for err in results["errors"]:
                    print(f"         - {err}")
                    
            if results["warnings"]:
                print(f"      [WARNING] Warnings:")
                for warn in results["warnings"]:
                    print(f"         - {warn}")

            if results["passes"]:
                    print(f"      [PASS] Passes: {len(results['passes'])} checks passed")
        else:
            print(f"   [SUCCESS] Validation Passed ({len(results['passes'])} checks)")
