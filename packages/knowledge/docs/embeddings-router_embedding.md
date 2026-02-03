---
type: Documentation
domain: knowledge
last_modified: 2026-02-02
generated: true
source: packages/knowledge/embeddings/router_embedding.py
generated_at: 2026-02-02T22:28:51.865340
hash: 49a09ab49ded32a47ef1514aba557a72be4f60b26c06cb948ff12be0d674bc98
---

## Router Embedding Documentation

This module provides a routing mechanism for generating embeddings from a collection of documents. It prioritizes a primary embedding model and seamlessly falls back to a secondary model if rate limits are encountered with the primary. This ensures continuous operation even when facing API restrictions.

**Module Responsibilities:**

The primary responsibility of this module is to abstract the complexity of managing multiple embedding models and their associated limitations. It handles batch processing of documents and intelligently switches between models to maximize throughput and reliability.

**Key Classes:**

*   **`RPDLimitError`**: A custom exception class. This is raised by the `GeminiEmbedding` class when its rate limits are reached. We catch this exception to trigger the fallback mechanism.
*   **`GeminiEmbedding`**: (From `gemini_embedding.py`) This class encapsulates the logic for interacting with the Gemini embedding model. It is the primary embedding provider.
*   **`Gemma12bItEmbedding`**: (From `gemma_12b_it_embedding.py`) This class encapsulates the logic for interacting with the Gemma 12b IT embedding model. It serves as the fallback embedding provider.

**Important Functions:**

*   **`get_embeddings_for_docs(docs, batch_size=10)`**: This is the core function of the module. It takes a list of documents (`docs`) and an optional `batch_size` as input.

    *   **Parameters:**
        *   `docs`: A list of dictionaries, where each dictionary represents a document and is expected to have a `"content"` key containing the text to be embedded.  Type: `list[dict]`.
        *   `batch_size`: The number of documents to process in each batch.  Defaults to 10. Type: `int`.
    *   **Behavior:**
        1.  Initializes instances of `GeminiEmbedding` (primary) and `Gemma12bItEmbedding` (fallback).
        2.  Iterates through the `docs` list in batches of the specified `batch_size`.
        3.  For each batch, it first attempts to generate embeddings using the `GeminiEmbedding` model.
        4.  If a `RPDLimitError` is raised (indicating the Gemini model has reached its rate limit), it falls back to using the `Gemma12bItEmbedding` model for that batch.
        5.  The function extends the `embeddings` list with tuples containing the original document and its corresponding embedding.
        6.  Prints progress updates to the console during processing.
        7.  Returns a list of tuples, where each tuple contains a document (dictionary) and its embedding (list of floats). Type: `list[tuple[dict, list[float]]]`.
    *   **Example:**

        ```python
        docs = [{"content": "This is the first document."}, {"content": "This is the second document."}]
        embeddings = get_embeddings_for_docs(docs, batch_size=1)
        print(embeddings)
        ```

**Type Hints:**

The code extensively uses type hints (e.g., `docs: list[dict]`, `batch_size: int`). These hints improve code readability and allow for static analysis, helping to catch potential errors during development.

**Design Decisions and Patterns:**

*   **Fallback Mechanism:** The core design pattern is a fallback mechanism. This enhances the robustness of the embedding process by providing an alternative when the primary model is unavailable due to rate limits.
*   **Batch Processing:** Processing documents in batches improves efficiency by reducing the number of API calls. The `batch_size` parameter allows you to tune performance based on your specific needs and the API limits of the embedding models.
*   **Exception Handling:** The use of a custom exception (`RPDLimitError`) allows for specific and targeted error handling, making the code more maintainable and easier to understand.
*   **Clear Separation of Concerns:** The module focuses solely on routing embedding requests. The actual embedding logic is encapsulated within the `GeminiEmbedding` and `Gemma12bItEmbedding` classes, promoting modularity and reusability.