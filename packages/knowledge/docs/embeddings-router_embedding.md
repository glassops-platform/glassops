---
type: Documentation
domain: knowledge
origin: packages/knowledge/embeddings/router_embedding.py
last_modified: 2026-01-28
generated: true
source: packages/knowledge/embeddings/router_embedding.py
generated_at: 2026-01-28T22:40:14.293849
hash: 49a09ab49ded32a47ef1514aba557a72be4f60b26c06cb948ff12be0d674bc98
---

## Router Embedding Documentation

This module provides a routing mechanism for generating embeddings from a collection of documents. It prioritizes a primary embedding model and seamlessly falls back to a secondary model if rate limits are encountered with the primary. This ensures continuous operation even when facing API restrictions.

**Module Responsibilities:**

The primary responsibility of this module is to abstract the complexity of managing multiple embedding models and their potential limitations. It handles batch processing of documents and intelligently switches between models to maximize throughput and reliability.

**Key Classes:**

*   **`RPDLimitError` (Exception):** A custom exception class raised when the primary embedding service encounters a rate limit or quota issue. This signals the need to switch to the fallback model.
*   **`GeminiEmbedding` (from `gemini_embedding`):** Represents the primary embedding service. It is responsible for generating embeddings using the Gemini model.
*   **`Gemma12bItEmbedding` (from `gemma_12b_it_embedding`):** Represents the fallback embedding service. It generates embeddings using the Gemma 12B Italian model. This model is used when the primary service is unavailable due to rate limits.

**Important Functions:**

*   **`get_embeddings_for_docs(docs: list[dict], batch_size: int = 10) -> list[tuple[dict, list[float]]]`:**
    *   **Purpose:** This function takes a list of documents and generates embeddings for each document.
    *   **Parameters:**
        *   `docs` (list of dictionaries): A list of documents, where each document is a dictionary containing a "content" key with the text to be embedded.
        *   `batch_size` (int, optional): The number of documents to process in each batch. Defaults to 10.  Adjusting this value can impact performance and rate limit behavior.
    *   **Behavior:**
        1.  It initializes instances of the `GeminiEmbedding` (primary) and `Gemma12bItEmbedding` (fallback) classes.
        2.  It iterates through the documents in batches.
        3.  For each batch, it attempts to generate embeddings using the primary `GeminiEmbedding` service.
        4.  If a `RPDLimitError` is raised (indicating a rate limit), it falls back to using the `Gemma12bItEmbedding` service for that batch.
        5.  The function pairs each original document with its corresponding embedding and appends it to the `embeddings` list.
        6.  Finally, it returns a list of tuples, where each tuple contains a document and its embedding.
    *   **Return Value:** A list of tuples. Each tuple contains a document (dictionary) and its corresponding embedding (list of floats).
    *   **Type Hints:** The function uses type hints to clearly define the expected input and output types, improving code readability and maintainability.

**Design Decisions and Patterns:**

*   **Fallback Mechanism:** The core design pattern is a fallback mechanism. This ensures resilience against API rate limits or service outages.
*   **Batch Processing:** Processing documents in batches improves efficiency and reduces the number of API calls.
*   **Clear Error Handling:** The custom `RPDLimitError` exception provides a specific signal for rate limit events, allowing for targeted fallback behavior.
*   **Loose Coupling:** The module relies on separate classes for each embedding service, promoting modularity and making it easier to add or replace embedding models in the future.
*   **Progress Indicator:** The print statement within the loop provides a simple progress indicator to the user, showing the processing status. You may want to replace this with a more sophisticated logging mechanism in a production environment.