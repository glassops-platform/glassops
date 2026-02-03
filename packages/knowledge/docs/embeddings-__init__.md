---
type: Documentation
domain: knowledge
last_modified: 2026-02-02
generated: true
source: packages/knowledge/embeddings/__init__.py
generated_at: 2026-02-02T22:28:00.237064
hash: 8155797cfbbac00bb501a46542405e653e9cc589aede51dfac27cdb69312dafd
---

## Knowledge Embeddings Package Documentation

This package provides tools for generating embeddings from text data, a process that represents text as numerical vectors. These vectors capture the semantic meaning of the text, enabling applications like semantic search, document similarity analysis, and question answering. We offer several embedding models and a routing function to select the appropriate model for your needs.

**Key Components:**

* **`GeminiEmbedding` Class:** This class interfaces with the Gemini model to produce text embeddings. It handles the communication with the Gemini API and returns embedding vectors for input text.

* **`Gemma12bItEmbedding` Class:** This class utilizes the Gemma 12b IT (Instruction Tuned) model to generate text embeddings. Similar to `GeminiEmbedding`, it manages the API interaction and provides embedding vectors.  The IT designation indicates the model has been specifically tuned for instruction-following tasks, potentially improving embedding quality for certain applications.

* **`get_embeddings_for_docs` Function:** This function acts as a router, selecting the best embedding model based on the input documents and returning their corresponding embeddings.  

**Function Signatures and Behavior:**

* **`get_embeddings_for_docs(docs: list[str]) -> list[list[float]]`:**
    *   **Purpose:**  Generates embeddings for a list of documents.
    *   **Parameters:**
        *   `docs`: A list of strings, where each string represents a document.
    *   **Return Value:** A list of lists of floats. Each inner list represents the embedding vector for the corresponding document in the input list.

**Type Hints:**

Throughout the package, type hints (e.g., `list[str]`, `list[list[float]]`) are used to improve code readability and maintainability. They specify the expected data types for function parameters and return values, aiding in error detection and code understanding.

**Design Considerations:**

The package is designed with modularity in mind. Each embedding model is encapsulated in its own class, allowing for easy addition of new models without modifying existing code. The `get_embeddings_for_docs` function provides a single entry point for generating embeddings, abstracting away the complexity of model selection.

**Usage:**

You can access the embedding models and the routing function directly from the `knowledge.embeddings` module. For example:

```python
from knowledge.embeddings import get_embeddings_for_docs

documents = ["This is the first document.", "This is the second document."]
embeddings = get_embeddings_for_docs(documents)
print(embeddings)