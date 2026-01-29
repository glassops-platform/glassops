# knowledge/__init__.py
# Expose the main pipeline entrypoint
from .main import run_pipeline

# Optionally, expose submodules too
from .embeddings import get_embeddings_for_docs
from .ingestion import discover_and_chunk_docs, build_or_update_index
from .drift import detect_drift
from .rag import query_index
from .utils import hash_file, batch_items

__all__ = [
    "run_pipeline",
    "get_embeddings_for_docs",
    "discover_and_chunk_docs",
    "build_or_update_index",
    "detect_drift",
    "query_index",
    "hash_file",
    "batch_items"
]
