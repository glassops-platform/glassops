---
type: Documentation
domain: knowledge
origin: packages/knowledge/embeddings/router_embedding.py
last_modified: 2026-01-26
generated: true
source: packages/knowledge/embeddings/router_embedding.py
generated_at: 2026-01-26T14:10:35.791Z
hash: 6ca89adeea1efe2c51cd06b358350c1e92bb78b0436fe0e0e6afb922f9c594de
---

## Router Embedding Documentation

**Document Version:** 1.0
**Date:** October 26, 2023
**Author:** AI Documentation Generator

### 1. Introduction

This document details the `router_embedding.py` module, responsible for generating embeddings for text documents using a prioritized approach. Embeddings are numerical representations of text, used for tasks like semantic search and similarity analysis. This module intelligently routes embedding requests between a primary, high-quality embedding model and a fallback model, managing potential rate limits and ensuring continuous operation.

### 2. Functionality

The core functionality of this module is to provide a robust and reliable embedding service. It prioritizes a preferred embedding model (`GeminiEmbedding`) but seamlessly switches to a fallback model (`Gemma12bItEmbedding`) if rate limits are encountered with the primary service. This ensures that embedding requests are always processed, even under high load or temporary service disruptions.

### 3. Components

*   **`GeminiEmbedding`:**  A class representing the primary embedding model.  This model is assumed to provide higher quality embeddings but may be subject to usage quotas.
*   **`Gemma12bItEmbedding`:** A class representing the fallback embedding model. This model provides a secondary option for generating embeddings when the primary model is unavailable due to rate limits.
*   **`RPDLimitError`:** A custom exception raised when the primary embedding model (`GeminiEmbedding`) exceeds its rate limit or quota.
*   **`get_embeddings_for_docs(docs, batch_size=10)`:** The primary function of the module. It takes a list of documents and generates embeddings for each document.

### 4. `get_embeddings_for_docs` Function Details

**Purpose:** Generates embeddings for a list of documents, utilizing a primary embedding model with automatic fallback to a secondary model if rate limits are hit.

**Parameters:**

*   `docs` (list): A list of dictionaries, where each dictionary represents a document and *must* contain a key named `"content"` holding the text to be embedded.
*   `batch_size` (int, optional): The number of documents to process in each batch.  Defaults to 10.  Larger batch sizes can improve throughput but may increase the likelihood of hitting rate limits.

**Return Value:**

*   `list`: A list of tuples. Each tuple contains the original document (dictionary) and its corresponding embedding (list of floats).

**Error Handling:**

*   The function handles `RPDLimitError` exceptions, which are raised by the `GeminiEmbedding` class when rate limits are exceeded.  Upon catching this exception, the function automatically switches to the `Gemma12bItEmbedding` model for the current batch.

**Workflow:**

1.  The function iterates through the input `docs` list in batches of size `batch_size`.
2.  For each batch, it attempts to generate embeddings using the `GeminiEmbedding` model.
3.  If the `GeminiEmbedding` model raises an `RPDLimitError`, the function falls back to using the `Gemma12bItEmbedding` model for that batch.
4.  The function extends the `embeddings` list with tuples of (document, embedding) for each processed document.
5.  A progress indicator is printed to the console during processing.
6.  Finally, the function returns the complete list of (document, embedding) tuples.

### 5. Dependencies

*   `gemini_embedding.py`
*   `gemma_12b_it_embedding.py`

### 6. Usage Example

```python
from packages.knowledge.embeddings.router_embedding import get_embeddings_for_docs

documents = [
    {"content": "This is the first document."},
    {"content": "This is the second document."},
    {"content": "And this is the third one."}
]

embeddings = get_embeddings_for_docs(documents, batch_size=2)

for doc, embedding in embeddings:
    print(f"Document: {doc['content']}")
    print(f"Embedding: {embedding[:5]}...") # Print only the first 5 elements for brevity