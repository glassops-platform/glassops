---
type: Documentation
domain: knowledge
origin: packages/knowledge/embeddings/__init__.py
last_modified: 2026-02-01
generated: true
source: packages/knowledge/embeddings/__init__.py
generated_at: 2026-02-01T19:28:30.447139
hash: 8155797cfbbac00bb501a46542405e653e9cc589aede51dfac27cdb69312dafd
---

## Knowledge Embeddings Package Documentation

This package provides tools for generating and routing document embeddings, which are numerical representations of text used for semantic search and other natural language processing tasks. It offers access to different embedding models and a routing mechanism to select the most appropriate model for a given task.

**Key Components:**

* **`GeminiEmbedding` Class:** This class interfaces with the Gemini embedding model. It takes text as input and returns its corresponding embedding vector. We designed this for leveraging Google’s Gemini model capabilities.

* **`Gemma12bItEmbedding` Class:** This class provides access to the Gemma 12B Italian embedding model. Similar to `GeminiEmbedding`, it converts text into embedding vectors, specifically tuned for the Italian language. This allows for language-specific semantic understanding.

* **`get_embeddings_for_docs` Function:** This function serves as a router for embedding requests. It accepts a list of documents (text strings) and returns a list of corresponding embeddings. The function intelligently selects the appropriate embedding model based on the input documents, potentially considering factors like language or content type. The function signature is: `get_embeddings_for_docs(docs: list[str]) -> list[list[float]]`. The type hint `list[str]` indicates the function expects a list of strings as input, and `list[list[float]]` signifies that it returns a list of lists, where each inner list represents the embedding vector for a document (a list of floating-point numbers).

**Design Considerations:**

The package follows a modular design, separating the implementation details of each embedding model into its own class. This allows for easy addition of new models without modifying existing code. The `get_embeddings_for_docs` function provides a single entry point for generating embeddings, abstracting away the complexity of model selection.

**Usage:**

You can access the embedding models directly:

```python
from knowledge.embeddings import GeminiEmbedding

gemini_embedder = GeminiEmbedding()
embedding = gemini_embedder.embed("This is a sample document.")
print(embedding)
```

Or, you can use the router function:

```python
from knowledge.embeddings import get_embeddings_for_docs

documents = ["This is the first document.", "And this is the second one."]
embeddings = get_embeddings_for_docs(documents)
print(embeddings)
```

**`__all__` Variable:**

The `__all__` list explicitly defines the public API of the package. This ensures that only the specified classes and functions are imported when using `from knowledge.embeddings import *`. This practice improves code clarity and maintainability.