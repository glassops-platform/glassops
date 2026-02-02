---
type: Documentation
domain: knowledge
origin: packages/knowledge/embeddings/router_embedding.py
last_modified: 2026-02-01
generated: true
source: packages/knowledge/embeddings/router_embedding.py
generated_at: 2026-02-01T19:29:19.040315
hash: 49a09ab49ded32a47ef1514aba557a72be4f60b26c06cb948ff12be0d674bc98
---

## Router Embedding Documentation

This module provides a routing mechanism for generating embeddings from a collection of documents. It prioritizes a primary embedding model and automatically falls back to a secondary model if rate limits are encountered with the primary. This ensures continuous operation even when the primary model is unavailable due to usage restrictions.

**Module Responsibilities:**

The primary responsibility of this module is to abstract the complexity of managing multiple embedding models and handling potential rate limiting issues. It offers a single function, `get_embeddings_for_docs`, that accepts a list of documents and returns a list of embeddings paired with their original documents.

**Key Classes:**

*   **`GeminiEmbedding`**: This class (imported from `gemini_embedding.py`) represents the primary embedding model. It is responsible for generating embeddings using the Gemini API.
*   **`Gemma12bItEmbedding`**: This class (imported from `gemma_12b_it_embedding.py`) represents the fallback embedding model. It is responsible for generating embeddings using the Gemma 12B Italian model.
*   **`RPDLimitError`**: A custom exception class used to signal that the primary embedding model has reached its rate limit.

**Important Functions:**

*   **`get_embeddings_for_docs(docs, batch_size=10)`**:
    *   **Purpose:** This function takes a list of documents and generates embeddings for each document.
    *   **Parameters:**
        *   `docs`: A list of dictionaries, where each dictionary represents a document and contains a `"content"` key with the text content of the document.  The type hint is `list[dict]`.
        *   `batch_size`: An integer specifying the number of documents to process in each batch. Defaults to 10.  The type hint is `int`.
    *   **Behavior:**
        1.  It initializes instances of the `GeminiEmbedding` (primary) and `Gemma12bItEmbedding` (fallback) classes.
        2.  It iterates through the documents in batches of the specified `batch_size`.
        3.  For each batch, it attempts to generate embeddings using the `GeminiEmbedding` model.
        4.  If a `RPDLimitError` is raised (indicating a rate limit), it falls back to using the `Gemma12bItEmbedding` model for that batch.
        5.  The function extends the `embeddings` list with tuples containing the original document and its corresponding embedding.
        6.  Finally, it returns the `embeddings` list.
    *   **Return Value:** A list of tuples, where each tuple contains a document (dictionary) and its corresponding embedding (list of floats). The type hint is `list[tuple[dict, list[float]]]`.
    *   **Error Handling:** The function handles `RPDLimitError` exceptions, gracefully falling back to the secondary embedding model.

**Notable Patterns and Design Decisions:**

*   **Fallback Mechanism:** The design incorporates a clear fallback mechanism to ensure resilience against rate limits or unavailability of the primary embedding model.
*   **Batch Processing:** Processing documents in batches improves efficiency and can help mitigate rate limiting issues.
*   **Exception Handling:** The use of a custom exception (`RPDLimitError`) provides a specific and informative way to handle rate limit errors.
*   **Type Hints:** The code uses type hints to improve readability and maintainability, and to enable static analysis. This helps to catch potential errors early in the development process.
*   **Progress Indicator:** The `print` statement within the loop provides a simple progress indicator to the user, showing the number of documents processed.