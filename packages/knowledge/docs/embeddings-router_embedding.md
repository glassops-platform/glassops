---
type: Documentation
domain: knowledge
origin: packages/knowledge/embeddings/router_embedding.py
last_modified: 2026-01-31
generated: true
source: packages/knowledge/embeddings/router_embedding.py
generated_at: 2026-01-31T08:53:33.496478
hash: 49a09ab49ded32a47ef1514aba557a72be4f60b26c06cb948ff12be0d674bc98
---

## Router Embedding Documentation

This module provides a routing mechanism for generating embeddings from a collection of documents. It prioritizes a primary embedding model and automatically falls back to a secondary model if rate limits are encountered with the primary. This ensures continuous operation even when the primary model experiences capacity constraints.

**Key Classes:**

* **`RPDLimitError`**: A custom exception class. This is raised by the `GeminiEmbedding` class when its rate limits are reached. We catch this exception to trigger the fallback mechanism.
* **`GeminiEmbedding`**: (From `gemini_embedding.py`) This class encapsulates the logic for interacting with the Gemini embedding model. It is the preferred embedding provider.
* **`Gemma12bItEmbedding`**: (From `gemma_12b_it_embedding.py`) This class encapsulates the logic for interacting with the Gemma 12B Italian embedding model. It serves as a fallback option when the Gemini model is unavailable due to rate limits.

**Important Functions:**

* **`get_embeddings_for_docs(docs, batch_size=10)`**: This is the primary function of the module. It takes a list of documents (`docs`) as input and returns a list of tuples, where each tuple contains a document and its corresponding embedding.

    *   **`docs`**: A list of dictionaries. Each dictionary represents a document and is expected to have a key named `"content"` which holds the text to be embedded.
    *   **`batch_size`**: An optional integer parameter that controls the number of documents processed in each batch. The default value is 10. Processing in batches helps manage resource usage and potentially improve performance.
    *   **Behavior**: The function first instantiates instances of both the `GeminiEmbedding` and `Gemma12bItEmbedding` classes. It then iterates through the input documents in batches. For each batch, it attempts to generate embeddings using the `GeminiEmbedding` model. If a `RPDLimitError` is raised (indicating the Gemini model has reached its rate limit), it falls back to using the `Gemma12bItEmbedding` model for that batch. The function prints progress updates to the console during processing. Finally, it returns a list of (document, embedding) tuples.

**Type Hints:**

The code makes extensive use of type hints. For example, `docs: list` and `batch_size: int`. These hints improve code readability and allow for static analysis, helping to catch potential errors during development. They clearly define the expected data types for function parameters and return values.

**Design Decisions and Patterns:**

*   **Fallback Mechanism**: The core design pattern is a fallback mechanism. This enhances the robustness of the embedding process by providing an alternative when the primary service is unavailable.
*   **Batch Processing**: The use of batch processing improves efficiency by reducing the number of API calls.
*   **Exception Handling**: The `try...except` block specifically handles the `RPDLimitError`, allowing for graceful degradation and continued operation.
*   **Clear Separation of Concerns**: The module delegates the actual embedding generation to separate classes (`GeminiEmbedding` and `Gemma12bItEmbedding`), promoting modularity and maintainability. You can easily swap out these embedding providers without modifying the core routing logic.