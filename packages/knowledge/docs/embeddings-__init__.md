---
type: Documentation
domain: knowledge
origin: packages/knowledge/embeddings/__init__.py
last_modified: 2026-01-31
generated: true
source: packages/knowledge/embeddings/__init__.py
generated_at: 2026-01-31T09:47:55.953610
hash: 8155797cfbbac00bb501a46542405e653e9cc589aede51dfac27cdb69312dafd
---

## Knowledge Embeddings Package Documentation

This package provides tools for generating embeddings from text data, a key component in many knowledge-based applications. Embeddings represent text as numerical vectors, allowing for semantic similarity comparisons and enabling tasks like search, clustering, and question answering. We offer multiple embedding models and a routing function to select the appropriate model for your needs.

**Module Responsibilities:**

The primary responsibility of this module is to expose a clean and consistent API for accessing different embedding models. It acts as a central point for generating embeddings and managing model selection.

**Key Classes:**

*   **`GeminiEmbedding`**: This class encapsulates the Gemini embedding model. It provides a method for generating embeddings using the Gemini API. It accepts text as input and returns a corresponding embedding vector.

*   **`Gemma12bItEmbedding`**: This class encapsulates the Gemma 12B Italian embedding model. Similar to `GeminiEmbedding`, it takes text input and produces an embedding vector, specifically tailored for Italian language data.

**Important Functions:**

*   **`get_embeddings_for_docs(docs: list[str]) -> list[list[float]]`**: This function serves as a router, intelligently selecting an embedding model based on the input documents. It accepts a list of documents (strings) and returns a list of embeddings, where each embedding is a list of floats. The function handles the logic of choosing the best model for the provided text.

**Type Hints:**

Throughout the package, type hints are used extensively (e.g., `docs: list[str]`, `-> list[list[float]]`). These hints improve code readability and maintainability, and allow for static analysis to catch potential errors. They clearly define the expected input and output types for each function and method.

**Design Decisions:**

The package is structured to promote flexibility and extensibility. By separating the embedding models into individual classes, we allow for easy addition of new models without modifying existing code. The `get_embeddings_for_docs` function provides a single entry point for generating embeddings, abstracting away the complexity of model selection. The `__all__` variable explicitly defines the public API of the package, controlling which classes and functions are exposed to users.