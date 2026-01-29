---
type: Documentation
domain: knowledge
origin: packages/knowledge/embeddings/__init__.py
last_modified: 2026-01-28
generated: true
source: packages/knowledge/embeddings/__init__.py
generated_at: 2026-01-28T22:39:13.074064
hash: 8155797cfbbac00bb501a46542405e653e9cc589aede51dfac27cdb69312dafd
---

## Knowledge Embeddings Package Documentation

This package provides tools for generating and routing document embeddings, which are numerical representations of text used for semantic search and other knowledge-based applications. We offer several embedding models and a routing function to select the most appropriate model for a given task.

**Module Responsibilities:**

The primary responsibility of this module is to expose a clean and consistent API for accessing different embedding models. It acts as a central point for managing and selecting embedding strategies.

**Key Classes:**

*   **`GeminiEmbedding`**: This class encapsulates the Gemini embedding model. It provides a method for converting text documents into embeddings using the Gemini API. Type hints within the class ensure correct data handling for input text and output embedding vectors.
*   **`Gemma12bItEmbedding`**: This class encapsulates the Gemma 12B Italian embedding model. It offers functionality to generate embeddings specifically tailored for Italian language text, leveraging the Gemma 12B model. Type hints are used to define expected input and output types.

**Important Functions:**

*   **`get_embeddings_for_docs(docs: list[str]) -> list[list[float]]`**: This function serves as a router, intelligently selecting the best embedding model based on the input documents. It accepts a list of text documents (`docs`) and returns a list of corresponding embeddings, where each embedding is a list of floating-point numbers. The function handles the complexities of model selection and ensures consistent output formatting.

**Design Decisions and Patterns:**

*   **Explicit API Exposure:** The `__all__` list explicitly defines the public API of the package, controlling which classes and functions are accessible to external users. This promotes a clear and maintainable interface.
*   **Class-Based Model Encapsulation:** Each embedding model is encapsulated within its own class. This allows for easy extension with new models and provides a structured way to manage model-specific configurations and logic.
*   **Type Hinting:** We extensively use type hints throughout the code. This improves code readability, helps prevent errors, and enables static analysis tools to verify the correctness of the code. You can benefit from these hints in your IDE for improved code completion and error detection.
*   **Routing Function:** The `get_embeddings_for_docs` function provides a single entry point for generating embeddings, abstracting away the details of model selection. This simplifies the user experience and allows us to easily change the routing logic without affecting client code.