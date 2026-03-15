# knowledge/embeddings/__init__.py
# Expose embedding APIs and router

from .gemini_embedding import GeminiEmbedding
from .gemma_12b_it_embedding import Gemma12bItEmbedding
from .router_embedding import get_embeddings_for_docs

__all__ = [
    "GeminiEmbedding",
    "Gemma12bItEmbedding",
    "get_embeddings_for_docs"
]
