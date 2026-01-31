---
type: Documentation
domain: knowledge
origin: packages/knowledge/embeddings/__init__.py
last_modified: 2026-01-31
generated: true
source: packages/knowledge/embeddings/__init__.py
generated_at: 2026-01-31T08:52:45.268357
hash: 8155797cfbbac00bb501a46542405e653e9cc589aede51dfac27cdb69312dafd
---

## Knowledge Embeddings Package Documentation

This package provides tools for generating and routing document embeddings, which are numerical representations of text used for semantic search and other knowledge-based applications. We offer several embedding models and a routing function to select the most appropriate model for a given task.

**Module Responsibilities:**

The primary responsibility of this module is to expose a clean API for accessing different embedding models and a function for obtaining embeddings for a collection of documents. It acts as a central point for managing embedding-related functionality within the larger knowledge system.

**Key Classes:**

*   **`GeminiEmbedding`**: This class encapsulates the Gemini embedding model. It provides a way to generate embeddings using Google’s Gemini API. Instances of this class handle the specifics of interacting with the Gemini service.
*   **`Gemma12bItEmbedding`**: This class encapsulates the Gemma 12B Italian embedding model. It provides a way to generate embeddings using the Gemma 12B model, specifically tuned for Italian text. Instances of this class manage the model loading and embedding generation process.

**Important Functions:**

*   **`get_embeddings_for_docs(docs: list[str]) -> list[list[float]]`**: This function takes a list of documents (strings) as input and returns a list of embeddings, where each embedding is a list of floating-point numbers. This function acts as a router, selecting the best embedding model based on the input documents and returning the corresponding embeddings. The type hint `list[str]` clearly indicates the expected input type, and `list[list[float]]` specifies the structure of the output.

**Design Decisions and Patterns:**

*   **Class-Based Model Wrappers:** We employ classes (`GeminiEmbedding`, `Gemma12bItEmbedding`) to encapsulate each embedding model. This promotes modularity and allows for easy addition of new models in the future without modifying existing code. Each class handles the specific details of its corresponding model, such as API keys, model loading, and input/output formatting.
*   **Router Function:** The `get_embeddings_for_docs` function serves as a router, abstracting away the complexity of choosing the right embedding model. This simplifies the process for users, who only need to provide the documents and receive the embeddings.
*   **Type Hints:** We extensively use type hints (e.g., `list[str]`, `list[list[float]]`) to improve code readability and maintainability. Type hints also enable static analysis tools to catch potential errors early in the development process.
*   **`__all__` List:** The `__all__` list explicitly defines the public API of the module. This ensures that only the intended classes and functions are exposed to users when they import the package.