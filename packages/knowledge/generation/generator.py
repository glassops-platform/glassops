# generation/generator.py
"""
Documentation generator orchestrator.
Scans files, selects adapters, invokes LLM, and writes output.
"""

import glob
import hashlib
import json
import os
import re
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, List, Optional

import pathspec
import yaml

from knowledge.llm.client import LLMClient
from knowledge.generation.adapters import (
    BaseAdapter,
    GoAdapter,
    PythonAdapter,
    TypeScriptAdapter,
    YAMLAdapter,
    JSONAdapter,
    DockerAdapter,
    TerraformAdapter,
    ApexAdapter,
    LWCAdapter,
)
from knowledge.generation.validator import Validator


class Generator:
    """
    Orchestrates documentation generation for a codebase.
    """

    # Directories to ignore during scanning
    IGNORED_DIRS = {
        "node_modules", "dist", "build", ".git", "__pycache__",
        "venv", ".venv", "site-packages", ".turbo", "coverage",
        "glassops_site", "docs_staging"
    }

    # Mapping from file extension to prompt key
    EXTENSION_TO_PROMPT_KEY = {
        ".go": "go",
        ".py": "py",
        ".ts": "ts",
        ".js": "ts",
        ".mjs": "ts",
        ".tsx": "ts",
        ".jsx": "ts",
        ".yml": "yml",
        ".yaml": "yml",
        ".json": "json",
        ".tf": "tf",
        ".cls": "apex",
        ".trigger": "apex",
    }

    def __init__(self, root_dir: str, output_dir: Optional[str] = None):
        """
        Initialize the generator.

        Args:
            root_dir: Root directory of the repository.
            output_dir: Optional output directory for generated docs.
                        If None, docs are placed alongside source files.
        """
        self.root_dir = Path(root_dir).resolve()
        self.output_dir = Path(output_dir).resolve() if output_dir else None
        self.llm = LLMClient()
        self.cache_path = self.root_dir / "config" / "doc-cache.json"
        self.prompts_path = Path(__file__).parent.parent / "config" / "prompts.yml"
        self.cache: Dict[str, dict] = {}
        self.prompts: Dict[str, Any] = {}
        self.gitignore_spec = self._load_gitignore()

        # Initialize all adapters (order matters - first match wins)
        self.adapters: List[BaseAdapter] = [
            LWCAdapter(),       # Must come before TypeScriptAdapter for lwc/ dirs
            GoAdapter(),
            PythonAdapter(),
            TypeScriptAdapter(),
            YAMLAdapter(),
            JSONAdapter(),
            DockerAdapter(),
            TerraformAdapter(),
            ApexAdapter(),
        ]

    def _load_cache(self) -> None:
        """Load the documentation cache from disk."""
        try:
            if self.cache_path.exists():
                self.cache = json.loads(self.cache_path.read_text(encoding="utf-8"))
                print(f"📦 Loaded cache ({len(self.cache)} entries)")
        except Exception as e:
            print(f"⚠️ Failed to load cache: {e}")
            self.cache = {}

    def _save_cache(self) -> None:
        """Save the documentation cache to disk."""
        try:
            self.cache_path.parent.mkdir(parents=True, exist_ok=True)
            self.cache_path.write_text(
                json.dumps(self.cache, indent=2),
                encoding="utf-8"
            )
        except Exception as e:
            print(f"⚠️ Failed to save cache: {e}")

    def _load_prompts(self) -> None:
        """Load prompts configuration from YAML file."""
        try:
            if self.prompts_path.exists():
                raw = yaml.safe_load(self.prompts_path.read_text(encoding="utf-8"))
                self.prompts = raw.get("prompts", {})
                print(f"📝 Loaded prompts from {self.prompts_path.name}")
            else:
                print(f"⚠️ Prompts file not found: {self.prompts_path}")
                self.prompts = {}
        except Exception as e:
            print(f"⚠️ Failed to load prompts: {e}")
            self.prompts = {}

    def _get_prompt_for_file(self, file_path: Path, parsed_content: str) -> Optional[str]:
        """
        Get the prompt for a file from the prompts config.
        Returns None if no prompt config is found (falls back to adapter).
        """
        if not self.prompts:
            return None

        # Determine prompt key
        suffix = file_path.suffix.lower()
        
        # Special cases
        if file_path.name.startswith("Dockerfile"):
            prompt_key = "dockerfile"
        elif "lwc" in file_path.parts:
            prompt_key = "lwc"
        else:
            prompt_key = self.EXTENSION_TO_PROMPT_KEY.get(suffix)

        if not prompt_key or prompt_key not in self.prompts:
            prompt_key = "default"

        prompt_config = self.prompts.get(prompt_key)
        if not prompt_config:
            return None

        # Get shared rules
        shared_rules = self.prompts.get("_shared_rules", "")

        # Build the prompt
        system = prompt_config.get("system", "")
        user = prompt_config.get("user", "")

        # Replace placeholders
        system = system.replace("{{shared_rules}}", shared_rules)
        user = user.replace("{{content}}", parsed_content)

        return f"{system}\n\n{user}"

    def _load_gitignore(self) -> Optional[pathspec.PathSpec]:
        """Load .gitignore patterns."""
        gitignore_path = self.root_dir / ".gitignore"
        if gitignore_path.exists():
            try:
                with open(gitignore_path, "r", encoding="utf-8") as f:
                    return pathspec.PathSpec.from_lines("gitwildmatch", f)
            except Exception as e:
                print(f"⚠️ Failed to load .gitignore: {e}")
        return None

    def _should_ignore(self, path: Path) -> bool:
        """Check if a path should be ignored."""
        # Check hardcoded ignores first
        parts = path.parts
        if any(ignored in parts for ignored in self.IGNORED_DIRS):
            return True
        
        # Check .gitignore
        if self.gitignore_spec:
            try:
                # relative_to can fail if path is not relative to root_dir
                rel_path = path.relative_to(self.root_dir).as_posix()
                if self.gitignore_spec.match_file(rel_path):
                    return True
            except ValueError:
                pass
                
        return False

    def _generate_frontmatter(self, source_path: Path, content: str) -> str:
        """
        Generate YAML front matter for the documentation file.
        Matches the format from the TypeScript agent.
        """
        relative_path = source_path.relative_to(self.root_dir).as_posix()
        content_hash = hashlib.sha256(content.encode("utf-8")).hexdigest()
        today = datetime.now().strftime("%Y-%m-%d")
        now_iso = datetime.now().isoformat()

        # Infer type
        doc_type = "Documentation"
        if "adr" in relative_path.lower() or source_path.stem.lower().startswith("adr"):
            doc_type = "ADR"

        # Infer domain from package structure
        domain = "global"
        parts = relative_path.split("/")
        if parts[0] == "packages" and len(parts) > 1:
            # Handle grouped packages (adapters, tools)
            groups = ["adapters", "tools"]
            if len(parts) > 2 and parts[1] in groups:
                domain = parts[2]
            else:
                domain = parts[1]

        return f"""---
type: {doc_type}
domain: {domain}
origin: {relative_path}
last_modified: {today}
generated: true
source: {relative_path}
generated_at: {now_iso}
hash: {content_hash}
---

"""

    def _find_adapter(self, file_path: Path) -> Optional[BaseAdapter]:
        """Find the appropriate adapter for a file."""
        for adapter in self.adapters:
            if adapter.can_handle(file_path):
                return adapter
        return None

    def _get_output_path(self, source_path: Path) -> Path:
        """
        Determine the output path for generated documentation.

        Strategy:
        1. If output_dir is set, mirror the source structure there.
        2. Otherwise, place docs in the package's docs/ directory.
        """
        relative = source_path.relative_to(self.root_dir)

        if self.output_dir:
            # Mirror structure in output_dir
            output_path = self.output_dir / relative.with_suffix(".md")
        else:
            # Find the package root (first dir under packages/)
            parts = relative.parts
            if "packages" in parts:
                pkg_idx = parts.index("packages")
                if len(parts) > pkg_idx + 1:
                    # Handle grouped packages (adapters, tools)
                    groups = ["adapters", "tools"]
                    if len(parts) > pkg_idx + 2 and parts[pkg_idx + 1] in groups:
                        pkg_parts = parts[:pkg_idx + 3]  # packages/tools/agent
                    else:
                        pkg_parts = parts[:pkg_idx + 2]  # packages/<pkg_name>

                    file_parts = parts[len(pkg_parts):]  # rest of path

                    # Flatten the filename: parent-dir-filename.md
                    if len(file_parts) > 1:
                        # Prefix with parent directory name
                        flat_name = f"{file_parts[-2]}-{file_parts[-1]}"
                    else:
                        flat_name = file_parts[-1] if file_parts else source_path.name

                    output_path = self.root_dir / Path(*pkg_parts) / "docs" / Path(flat_name).with_suffix(".md")
                else:
                    output_path = source_path.with_suffix(".md")
            else:
                # Files outside packages/ -> root docs/
                output_path = self.root_dir / "docs" / relative.with_suffix(".md")

        return output_path

    def scan_files(self, patterns: List[str]) -> List[Path]:
        """
        Scan for files matching the given glob patterns.

        Args:
            patterns: List of glob patterns (relative to root_dir).

        Returns:
            List of matching file paths.
        """
        matched_files = set()

        for pattern in patterns:
            # Handle negation patterns (starting with !)
            if pattern.startswith("!"):
                continue  # We'll handle exclusions separately

            full_pattern = str(self.root_dir / pattern)
            for match in glob.glob(full_pattern, recursive=True):
                path = Path(match)
                if path.is_file() and not self._should_ignore(path):
                    matched_files.add(path)

        # Apply exclusion patterns
        for pattern in patterns:
            if pattern.startswith("!"):
                exclude_pattern = pattern[1:]  # Remove the !
                full_pattern = str(self.root_dir / exclude_pattern)
                for match in glob.glob(full_pattern, recursive=True):
                    path = Path(match)
                    matched_files.discard(path)

        return sorted(matched_files)

    def _clean_llm_output(self, output: str) -> str:
        """Clean up LLM output by removing markdown code block wrappers."""
        # Remove ```markdown ... ``` or ``` ... ``` wrappers
        output = re.sub(r'^\s*```[a-z]*\s*', '', output, flags=re.IGNORECASE)
        output = re.sub(r'\s*```\s*$', '', output)
        return output.strip()

    def generate_for_file(self, file_path: Path) -> Optional[str]:
        """
        Generate documentation for a single file.

        Args:
            file_path: Path to the source file.

        Returns:
            Generated documentation string, or None on failure.
        """
        adapter = self._find_adapter(file_path)
        if not adapter:
            print(f"⏭️ No adapter for: {file_path.name}")
            return None

        try:
            content = file_path.read_text(encoding="utf-8-sig")
            # Normalize line endings to LF for consistent hashing across OS
            content = content.replace("\r\n", "\n")
        except Exception as e:
            print(f"❌ Failed to read {file_path}: {e}")
            return None

        if not content.strip():
            print(f"⏭️ Empty file: {file_path.name}")
            return None

        # Check cache
        relative_path = file_path.relative_to(self.root_dir).as_posix()
        content_hash = hashlib.sha256(content.encode("utf-8")).hexdigest()

        if relative_path in self.cache:
            if self.cache[relative_path].get("hash") == content_hash:
                print(f"⏭️ Unchanged (cached): {file_path.name}")
                return None  # Skip, already generated

        # Parse into chunks
        chunks = adapter.parse(file_path, content)
        print(f"📄 Processing {file_path.name} ({len(chunks)} chunk(s))")

        # Generate documentation for each chunk
        outputs = []
        for i, chunk in enumerate(chunks):
            # Try to get prompt from config first, fall back to adapter
            prompt = self._get_prompt_for_file(file_path, chunk)
            if not prompt:
                prompt = adapter.get_prompt(file_path, chunk)
            
            result = self.llm.generate(prompt)

            if result:
                # Clean up LLM output
                result = self._clean_llm_output(result)
                outputs.append(result)
                if len(chunks) > 1:
                    print(f"   ✓ Chunk {i + 1}/{len(chunks)}")
            else:
                print(f"   ⚠️ Chunk {i + 1}/{len(chunks)} failed")

        if not outputs:
            return None

        # Post-process and combine
        return adapter.post_process(file_path, outputs)

    def run(self, patterns: List[str]) -> None:
        """
        Run the documentation generator.

        Args:
            patterns: List of glob patterns to process.
        """
        print(f"🔍 Scanning from: {self.root_dir}")
        self._load_cache()
        self._load_prompts()

        files = self.scan_files(patterns)

        if not files:
            print("❌ No files matched the patterns.")
            return

        print(f"📁 Found {len(files)} file(s) to process\n")

        success_count = 0
        skip_count = 0
        total_files = len(files)

        try:
            for idx, file_path in enumerate(files, 1):
                relative_path = file_path.relative_to(self.root_dir).as_posix()
                print(f"👉 Processing {idx}/{total_files}: {relative_path}")
                
                doc = self.generate_for_file(file_path)

                if doc:
                    output_path = self._get_output_path(file_path)
                    output_path.parent.mkdir(parents=True, exist_ok=True)

                    # Read original content for hashing
                    original_content = file_path.read_text(encoding="utf-8-sig")
                    original_content = original_content.replace("\r\n", "\n")
                    content_hash = hashlib.sha256(original_content.encode("utf-8")).hexdigest()
                    frontmatter = self._generate_frontmatter(file_path, original_content)
                    final_content = frontmatter + doc

                    # Validate content
                    errors = Validator.validate(final_content, str(output_path))
                    if errors:
                        print(f"   ⚠️ Validation warnings:")
                        for err in errors:
                            print(f"      - {err}")

                    output_path.write_text(final_content, encoding="utf-8")
                    print(f"   💾 Saved: {output_path.relative_to(self.root_dir)}\n")

                    # Update cache
                    relative_path = file_path.relative_to(self.root_dir).as_posix()
                    self.cache[relative_path] = {
                        "hash": content_hash,
                        "generatedFiles": [output_path.relative_to(self.root_dir).as_posix()],
                        "timestamp": datetime.now().isoformat(),
                    }

                    success_count += 1
                else:
                    print(f"   ⏩ Skipped (no content generated)\n")
                    skip_count += 1
        finally:
            self._save_cache()

        print(f"\n✅ Generation complete: {success_count} generated, {skip_count} skipped")
