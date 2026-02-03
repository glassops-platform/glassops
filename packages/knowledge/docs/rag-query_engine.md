---
type: Documentation
domain: knowledge
last_modified: 2026-02-02
generated: true
source: packages/knowledge/rag/query_engine.py
generated_at: 2026-02-02T22:32:08.596349
hash: 09603013c88193d7faf419e2b18b13bf32b3d9de8e28c293f2d3e97f0923446f
---

## GlassOps Knowledge Query Engine Documentation

This module provides a Retrieval-Augmented Generation (RAG) system for querying a knowledge base. It combines information retrieval from a vector database (ChromaDB) with a large language model (Gemini) to provide informed answers to user questions.

**Module Responsibilities:**

The primary function of this module is to accept a user query, retrieve relevant documents from a knowledge base, and generate a concise answer using a language model. It also incorporates a mechanism for injecting content from specific files based on keywords found in the query.

**Key Classes and Their Roles:**

This module primarily utilizes external libraries (chromadb, google.genai) and does not define custom classes. ChromaDB’s `PersistentClient` and `Collection` are used for vector storage and retrieval.

**Important Functions and Their Behavior:**

*   **`query_index(query: str, n_results: int = 5) -> str`**: This is the main function of the module. It takes a user `query` (string) and an optional `n_results` parameter (integer, default is 5) specifying the number of relevant documents to retrieve. It returns a string containing the generated answer, or an error message if something goes wrong.

    1.  **Embedding Generation:** The function first generates an embedding vector for the input `query` using the `get_embeddings_for_docs` function. This embedding represents the semantic meaning of the query.
    2.  **ChromaDB Query:** It then queries a ChromaDB collection named "glassops\_knowledge" using the generated query embedding. The `n_results` parameter controls the number of documents retrieved.
    3.  **Context Construction:** The retrieved documents and their corresponding IDs are used to construct a context string.
    4.  **Trigger-Based File Injection:** The function checks for predefined keywords in the query. If a keyword is found, it attempts to inject the content of a corresponding file (specified in a `config.json` file) into the context. This allows for dynamic inclusion of up-to-date information.
    5.  **Answer Generation:** Finally, it uses the Gemini language model to generate an answer based on the constructed context and the original query. The function includes a system prompt to guide the model's behavior.
    6.  **Error Handling:** The function includes robust error handling to catch potential issues during embedding generation, ChromaDB querying, file injection, and answer generation. It returns informative error messages to the user.

**Type Hints and Their Significance:**

The code uses type hints (e.g., `query: str`, `n_results: int`) to improve code readability and maintainability. These hints specify the expected data types for function parameters and return values, allowing for static analysis and early detection of potential errors.

**Notable Patterns and Design Decisions:**

*   **RAG Architecture:** The module implements a standard RAG architecture, combining information retrieval with language model generation.
*   **Configuration-Based Behavior:** The system's behavior is partially configurable through a `config.json` file, allowing for customization of the system prompt and trigger-based file injection.
*   **Error Handling:** Comprehensive error handling is implemented throughout the function to provide informative error messages and prevent unexpected crashes.
*   **Modular Design:** The use of external libraries (chromadb, google.genai) promotes modularity and allows for easy replacement of components.
*   **Context Injection:** The trigger-based file injection mechanism provides a way to dynamically update the knowledge base with information from external sources.
*   **Environment Variables:** The API key for the Gemini model is loaded from an environment variable (`GOOGLE_API_KEY`), enhancing security and flexibility.
*   **Debugging:** Print statements are included for debugging purposes, providing insights into the query process and file injection events.