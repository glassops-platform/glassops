---
type: Documentation
domain: knowledge
origin: packages/knowledge/embeddings/router_embedding.py
last_modified: 2026-01-31
generated: true
source: packages/knowledge/embeddings/router_embedding.py
generated_at: 2026-01-31T09:49:01.776704
hash: 49a09ab49ded32a47ef1514aba557a72be4f60b26c06cb948ff12be0d674bc98
---

## Router Embedding Documentation

This module provides a routing mechanism for generating embeddings from a collection of documents. It prioritizes a primary embedding model and automatically falls back to a secondary model if rate limits are encountered with the primary. This ensures continuous operation even when the primary model is unavailable due to usage restrictions.

**Module Responsibilities:**

The primary responsibility of this module is to abstract the complexity of managing multiple embedding models and handling potential failures. It offers a single function, `get_embeddings_for_docs`, that accepts a list of documents and returns a list of embeddings paired with their original documents.

**Key Classes:**

*   **`GeminiEmbedding`**: This class (imported from `gemini_embedding.py`) represents the primary embedding model. It is responsible for generating embeddings from text.
*   **`Gemma12bItEmbedding`**: This class (imported from `gemma_12b_it_embedding.py`) represents the fallback embedding model. It provides an alternative embedding generation capability when the primary model is unavailable.
*   **`RPDLimitError`**: A custom exception class. It signals that the primary embedding model has reached its rate limit or other usage restrictions.

**Important Functions:**

*   **`get_embeddings_for_docs(docs, batch_size=10)`**: This function is the main entry point for generating embeddings.
    *   **Parameters:**
        *   `docs`: A list of dictionaries, where each dictionary represents a document and is expected to have a key named `"content"` containing the text to be embedded.  The type is `list[dict]`.
        *   `batch_size`: An integer specifying the number of documents to process in each batch. Defaults to 10.  This parameter controls the size of requests sent to the embedding models.
    *   **Behavior:**
        1.  Initializes instances of the `GeminiEmbedding` (primary) and `Gemma12bItEmbedding` (fallback) classes.
        2.  Iterates through the input `docs` in batches of size `batch_size`.
        3.  For each batch, it attempts to generate embeddings using the `GeminiEmbedding` model.
        4.  If a `RPDLimitError` is raised (indicating a rate limit or other issue with the primary model), it falls back to using the `Gemma12bItEmbedding` model for that batch.
        5.  The function extends the `embeddings` list with tuples containing the original document and its corresponding embedding.
        6.  Prints progress updates to the console during processing.
        7.  Returns a list of tuples, where each tuple contains a document (dictionary) and its embedding (list of floats). The type is `list[tuple[dict, list[float]]]`.
    *   **Example:**

        ```python
        docs = [{"content": "This is the first document."}, {"content": "This is the second document."}]
        embeddings = get_embeddings_for_docs(docs, batch_size=1)
        print(embeddings)
        ```

**Type Hints:**

The code extensively uses type hints (e.g., `docs: list[dict]`, `batch_size: int`) to improve code readability and maintainability. These hints clarify the expected data types for function parameters and return values, aiding in static analysis and error detection.

**Design Decisions and Patterns:**

*   **Fallback Mechanism:** The core design pattern is a fallback mechanism. This ensures resilience by providing an alternative embedding model when the primary model is unavailable.
*   **Batch Processing:** Processing documents in batches improves efficiency by reducing the number of requests made to the embedding models.
*   **Exception Handling:** The `RPDLimitError` exception allows for graceful handling of rate limits and other errors from the primary embedding model.
*   **Clear Separation of Concerns:** The module focuses solely on routing embedding requests and does not include the implementation details of the individual embedding models. This promotes modularity and allows for easy swapping of models.