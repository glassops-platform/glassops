# main.py
"""
GlassOps Knowledge Pipeline
- Discover federated documentation
- Compute embeddings (with router/fallback)
- Build/update vector store
- Detect semantic drift
- RAG query
- Documentation generation from source code
"""

import json
import sys
import os
from pathlib import Path

# Ensure 'packages' is in sys.path so we can import 'knowledge'
# This allows running from root like: python packages/knowledge/main.py
PACKAGE_ROOT = Path(__file__).parent.parent
if str(PACKAGE_ROOT) not in sys.path:
    sys.path.append(str(PACKAGE_ROOT))

from knowledge.ingestion.federated_loader import discover_and_chunk_docs
from knowledge.embeddings.router_embedding import get_embeddings_for_docs
from knowledge.ingestion.index_builder import build_or_update_index
from knowledge.drift.detect_drift import detect_drift
from knowledge.rag.query_engine import query_index
from knowledge.generation import Generator

# Optional: load config
from dotenv import load_dotenv
# Load .env from project root (../../.env relative to this file)
ROOT_DIR = Path(__file__).parent.parent.parent
load_dotenv(ROOT_DIR / ".env")

CONFIG_PATH = Path(__file__).parent / "config" / "config.json"
with open(CONFIG_PATH, "r", encoding="utf-8") as f:
    config = json.load(f)


import argparse

def run_generate(patterns: list[str]) -> None:
    """Run documentation generation for the given patterns."""
    print("[INFO] Starting documentation generation...")
    generator = Generator(str(ROOT_DIR))
    generator.run(patterns)


def run_pipeline():
    parser = argparse.ArgumentParser(description="GlassOps Knowledge Pipeline")
    parser.add_argument("--query", "-q", type=str, help="Run a RAG query against the knowledge base")
    parser.add_argument("query_pos", nargs="*", help="Positional query string (joined by space)")
    parser.add_argument("--index", "-i", action="store_true", help="Force re-indexing of documents")
    parser.add_argument("--generate", "-g", action="store_true", help="Generate documentation from source code")
    parser.add_argument("--pattern", "-p", type=str, action="append", dest="patterns",
                        help="Glob pattern(s) for --generate (can be specified multiple times)")
    args = parser.parse_args()

    # Documentation generation mode
    if args.generate:
        patterns = args.patterns if args.patterns else [
            "packages/**/*.go",
            "packages/**/*.py",
            "packages/**/*.ts",
            "packages/**/*.js",
            "packages/**/*.mjs",
            "packages/**/*.tsx",
            "packages/**/*.jsx",
            "packages/**/*.yml",
            "packages/**/*.yaml",
            "packages/**/*.json",
            "packages/**/*.tf",
            "packages/**/*.cls",
            "packages/**/*.trigger",
            "packages/**/Dockerfile",
            "!**/node_modules/**",
            "!**/dist/**",
            "!**/venv/**",
            "!**/__pycache__/**",
        ]
        run_generate(patterns)
        return

    # Consolidate query from flag OR positional args
    final_query = args.query
    if not final_query and args.query_pos:
        final_query = " ".join(args.query_pos)

    print("Starting GlassOps Knowledge Pipeline...")
    
    # If a query is provided and we are NOT forcing an index, jump straight to query
    if final_query and not args.index:
         # TODO: verify index exists? For now assume yes if the user is asking.
         pass
    else:
        # Step 1: Discover federated docs
        print("Discovering docs...")
        docs = discover_and_chunk_docs()
        print(f"Found {len(docs)} docs.")

        # Step 2: Compute embeddings using router (Gemini primary, fallback Gemma)
        print("Generating embeddings...")
        embeddings = get_embeddings_for_docs(docs, batch_size=config.get("batch_size", 10))
        print(f"Generated embeddings for {len(embeddings)} docs.")

        # Step 3: Build or update vector store
        print("Building/updating vector store...")
        build_or_update_index(embeddings)
        print("Vector store updated.")

        # Step 4: Detect semantic drift
        print("Checking for semantic drift...")
        drifted_docs = detect_drift(embeddings, threshold=config.get("drift_threshold", 0.85))
        if drifted_docs:
            print("Semantic drift detected in these docs:")
            for d in drifted_docs:
                print(f"  - {d}")
        else:
            print("No semantic drift detected.")

    # Step 5: RAG query (Example OR User provided)
    if final_query:
        print(f"Query: {final_query}")
        response = query_index(final_query)
        print(f"\nRAG Response:\n{response}\n")
    elif args.index:
        print("Re-indexing complete. Use --query '...' to ask questions.")
    else:
        # Default behavior: Run pipeline + Example Query
        print("Running example RAG query...")
        example_query = "List all ADRs related to workflow layer"
        response = query_index(example_query)
        print(f"RAG response: {response}")

    print("GlassOps Knowledge Pipeline finished!")


if __name__ == "__main__":
    run_pipeline()
