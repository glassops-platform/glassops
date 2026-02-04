# federated_loader.py
# Scans the repo for Markdown/docs, chunks them, hashes content

import hashlib
import glob
import glob
import os
import re
from typing import List, Dict

def hash_content(text: str) -> str:
    return hashlib.sha256(text.encode("utf-8")).hexdigest()

def discover_and_chunk_docs(root_dir: str = ".") -> List[Dict]:
    """
    Returns a list of dicts:
    { "path": <file_path>, "content": <chunked_content>, "hash": <sha256> }
    """
    # Expanded patterns to catch more docs strings
    patterns = [
        "docs/**/*.md",
        "packages/**/docs/**/*.md", 
        "packages/**/adr/**/*.md",
        "packages/**/README.md"
    ]
    
    doc_paths = []
    for p in patterns:
        # Use recursive globbing
        full_pattern = os.path.join(root_dir, p)
        # glob.glob in python < 3.10 might need recursive=True explicit arg for **
        doc_paths.extend(glob.glob(full_pattern, recursive=True))

    # Deduplicate paths
    doc_paths = sorted(list(set(doc_paths)))
    
    docs = []
    print(f"DEBUG: Found {len(doc_paths)} potential documents.")

    for path in doc_paths:
        # Robust ignore logic
        ignored_dirs = ["node_modules", "venv", "vnev", ".git", "__pycache__", "dist", "site-packages"]
        if any(ignored in path.split(os.sep) for ignored in ignored_dirs):
            continue
            
        try:
            with open(path, "r", encoding="utf-8") as f:
                content = f.read()
            
            if not content.strip():
                continue

            if not content.strip():
                continue

            # Parse Frontmatter
            metadata = {}
            match = re.search(r"^---\n(.*?)\n---", content, re.DOTALL)
            if match:
                import yaml
                try:
                    metadata = yaml.safe_load(match.group(1))
                    # Remove frontmatter from content for chunking
                    content = content[match.end():].strip()
                except yaml.YAMLError as e:
                    print(f"[WARNING] Failed to parse frontmatter for {path}: {e}")

            # Smart Chunking Strategy:
            # 1. Split by Level 2 headers (##) to keep context grouped
            # 2. If no Level 2 headers, try Level 1 (#)
            # 3. Fallback to whole file
            
            # Simple splitter logic
            
            # Helper to split text by regex pattern but keep the delimiter
            def split_by_header(text, pattern):
                chunks = []
                last_pos = 0
                for match in re.finditer(pattern, text):
                    pos = match.start()
                    if pos > last_pos:
                        chunks.append(text[last_pos:pos].strip())
                    last_pos = pos
                chunks.append(text[last_pos:].strip())
                return [c for c in chunks if c]

            # Try splitting by ## first (Module/Section level)
            chunks = split_by_header(content, r'(?m)^##\s+')
            
            # If we only have 1 chunk (no ##), try splitting by # (Page level, though rarely multiple # per file)
            if len(chunks) <= 1:
                chunks = split_by_header(content, r'(?m)^#\s+')
            
            # Final fallback: if file is huge but structureless, we might need arbitrary splitting?
            # For now, let's stick to semantic headers.
            
            for i, chunk_text in enumerate(chunks):
                if not chunk_text: continue
                
                # Create a unique ID/path for the chunk
                # e.g. path/to/file.md#chunk-0
                chunk_id = f"{path}#chunk-{i}"
                
                doc_record = {
                    "path": chunk_id, # Store this so we know where it came from
                    "source_file": path, # Keep original path metadata
                    "content": chunk_text,
                    "hash": hash_content(chunk_text)
                }
                
                # Merge frontmatter metadata
                doc_record.update(metadata)
                
                docs.append(doc_record)

        except Exception as e:
            print(f"Warning: Could not read {path}: {e}")
            
    return docs
