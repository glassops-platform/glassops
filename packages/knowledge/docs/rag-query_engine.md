---
type: Documentation
domain: knowledge
origin: packages/knowledge/rag/query_engine.py
last_modified: 2026-01-31
generated: true
source: packages/knowledge/rag/query_engine.py
generated_at: 2026-01-31T09:56:35.293144
hash: 09603013c88193d7faf419e2b18b13bf32b3d9de8e28c293f2d3e97f0923446f
---

## GlassOps Knowledge Query Engine Documentation

This module provides a Retrieval-Augmented Generation (RAG) system for querying a knowledge base. It combines information retrieval from a vector database (ChromaDB) with a large language model (Gemini) to provide informed answers to user questions.

**Module Responsibilities:**

The primary function of this module is to accept a user query, retrieve relevant documents from a knowledge base, and generate a concise answer using a language model. It also incorporates a mechanism for injecting content from specific files based on keywords found in the query.

**Key Classes & Data Structures:**

This module does not define any custom classes. It relies on external libraries like `chromadb` and `google.genai`. The core data structures are lists of strings representing document chunks and their corresponding IDs, managed by the ChromaDB client.

**Important Functions:**

*   **`query_index(query: str, n_results: int = 5) -> str`**: This is the main function of the module. It takes a user `query` (string) and an optional `n_results` parameter (integer, default is 5) specifying the number of relevant documents to retrieve. It returns a string containing the generated answer, or an error message if something goes wrong.

    1.  **Embedding Generation:** The function first generates an embedding vector for the input `query` using the `get_embeddings_for_docs` function. This embedding represents the semantic meaning of the query.
    2.  **ChromaDB Query:** It then queries a ChromaDB collection named "glassops\_knowledge" using the generated query embedding. The `n_results` parameter controls the number of documents retrieved.
    3.  **Context Construction:** The retrieved documents and their IDs are extracted from the ChromaDB results.
    4.  **Trigger-Based File Injection:** The function checks for predefined keywords in the query. If a keyword is found, it attempts to inject the content of a corresponding file (specified in a `config.json` file) into the context. This allows for dynamic inclusion of up-to-date information.
    5.  **Answer Generation:** Finally, the function uses the Gemini language model to generate an answer based on the combined context (retrieved documents and potentially injected files) and the original query. The `GOOGLE_API_KEY` environment variable must be set for this step to work.
    6.  **Error Handling:** The function includes robust error handling, returning informative messages if embedding generation fails, the index is not found, the API key is missing, or the language model encounters an error.

**Type Hints:**

The code uses type hints extensively (e.g., `query: str`, `n_results: int`) to improve code readability and maintainability. These hints specify the expected data types for function parameters and return values, aiding in static analysis and error detection.

**Notable Patterns and Design Decisions:**

*   **RAG Architecture:** The module implements a standard RAG pipeline, separating the retrieval and generation stages.
*   **Configuration-Driven Behavior:** The system context for the language model and the file injection triggers are loaded from a `config.json` file, allowing for easy customization without modifying the code.
*   **Environment Variable for API Key:** The Gemini API key is read from the `GOOGLE_API_KEY` environment variable, promoting security and flexibility.
*   **Error Handling:** Comprehensive error handling is implemented throughout the function to provide informative messages to the user in case of failures.
*   **Debug Logging:** Print statements are included for debugging purposes, providing insights into the query process and file injection events.
*   **File Injection Priority:** Injected files are prepended to the context, giving them higher priority during answer generation.
*   **Model Specification:** The code explicitly specifies the 'gemma-3-12b-it' model for generation.