---
type: Documentation
domain: knowledge
origin: packages/knowledge/embeddings/__init__.py
last_modified: 2026-01-26
generated: true
source: packages/knowledge/embeddings/__init__.py
generated_at: 2026-01-26T14:10:46.880Z
hash: fd97b6c7fd00ff00aac9846e3424bc38ae90d1092971d7b11cbafb6823e2fb82
---

## Knowledge Embeddings Package Documentation

This document details the `knowledge.embeddings` package, which provides functionality for generating vector embeddings from text data. Embeddings are numerical representations of text that capture semantic meaning, enabling efficient similarity comparisons and powering various knowledge-intensive applications.

**Core Components:**

*   **`GeminiEmbedding`:**  A class implementing embedding generation using the Gemini model. This provides access to Google’s Gemini embedding capabilities.
*   **`Gemma12bItEmbedding`:** A class implementing embedding generation using the Gemma 12B Italian model. This provides access to embeddings specifically tailored for Italian language text.
*   **`get_embeddings_for_docs(docs)`:** A function that accepts a list of documents (`docs`) and returns a list of corresponding embeddings. This acts as a router, intelligently selecting the appropriate embedding model based on document characteristics or configuration (implementation details are abstracted).

**Usage:**

This package is designed to be easily integrated into applications requiring text understanding and semantic search.  Users can instantiate the embedding classes directly or utilize the `get_embeddings_for_docs` function for streamlined embedding generation.

**Dependencies:**

The functionality of this package relies on the underlying Gemini and Gemma models.  Ensure appropriate access and configuration for these models are in place before utilizing the package.

**`__all__` List:**

The `__all__` list explicitly defines the public API of the `knowledge.embeddings` package, ensuring that only the specified classes and functions are exposed for external use. This promotes code clarity and maintainability.