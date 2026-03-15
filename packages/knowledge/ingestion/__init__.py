# knowledge/ingestion/__init__.py
# Expose ingestion APIs

from .federated_loader import discover_and_chunk_docs
from .index_builder import build_or_update_index

__all__ = [
    "discover_and_chunk_docs",
    "build_or_update_index"
]
