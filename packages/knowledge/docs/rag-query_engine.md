---
type: Documentation
domain: knowledge
origin: packages/knowledge/rag/query_engine.py
last_modified: 2026-01-28
generated: true
source: packages/knowledge/rag/query_engine.py
generated_at: 2026-01-28T22:48:07.081711
hash: 09603013c88193d7faf419e2b18b13bf32b3d9de8e28c293f2d3e97f0923446f
---

## GlassOps Knowledge Query Engine Documentation

This module provides a Retrieval-Augmented Generation (RAG) system for querying a knowledge base. It combines information retrieval from a vector database (ChromaDB) with a large language model (Gemini) to provide informed answers to user questions.

**Module Responsibilities:**

The primary function of this module is to accept a user query, retrieve relevant documents from a knowledge base, and generate a concise answer using a language model. It also incorporates a mechanism for injecting content from specific files based on keywords found in the query.

**Key Classes & Data Structures:**

This module primarily relies on external libraries (chromadb, google.genai) and does not define custom classes.  ChromaDB collections are used to store and query embeddings of the knowledge base documents.

**Important Functions:**

*   **`query_index(query: str, n_results: int = 5) -> str`**: This is the main function of the module. It takes a user query (string) and an optional number of results to retrieve (integer, default is 5). It returns a string containing the generated answer, or an error message if something goes wrong.
    *   **Embedding Generation:** The function first generates an embedding vector for the input query using the `get_embeddings_for_docs` function (defined elsewhere). This embedding represents the semantic meaning of the query.
    *   **ChromaDB Querying:** It then queries a ChromaDB collection named "glassops\_knowledge" using the query embedding. The `n_results` parameter controls the number of relevant documents retrieved.
    *   **Context Construction:** The retrieved documents are assembled into a context string, separated by delimiters.  Document IDs (sources) are also collected.
    *   **Trigger-Based File Injection:**  The function checks for keywords in the query against a configuration file (`config.json`). If a keyword is found, the corresponding file (specified by a relative path in the configuration) is read and prepended to the context. This allows for dynamic inclusion of up-to-date information.
    *   **Gemini Answer Generation:** Finally, the function uses the Gemini language model to generate an answer based on the constructed context and the original query.  A system prompt provides instructions to the model, defining its role and behavior.
    *   **Error Handling:** The function includes robust error handling, returning informative messages if embedding generation fails, the index is not found, the API key is missing, or the language model encounters an error.

**Type Hints:**

The code makes extensive use of type hints (e.g., `query: str`, `n_results: int`) to improve code readability and maintainability. These hints specify the expected data types for function arguments and return values, enabling static analysis and helping to prevent errors.

**Notable Patterns and Design Decisions:**

*   **RAG Architecture:** The module implements a standard RAG pipeline, combining retrieval and generation for improved accuracy and relevance.
*   **Configuration-Driven Behavior:** The use of a `config.json` file allows for customization of the system prompt and trigger-based file injection without modifying the code.
*   **Error Handling:** Comprehensive error handling is implemented to provide informative messages to the user and facilitate debugging.
*   **Modular Design:** The embedding generation is delegated to a separate function (`get_embeddings_for_docs`), promoting code reuse and separation of concerns.
*   **Context Injection:** The ability to inject content from external files based on query keywords provides a flexible mechanism for incorporating dynamic information into the knowledge base.
*   **Environment Variables:** The API key is loaded from an environment variable (`GOOGLE_API_KEY`), enhancing security and portability.
*   **Path Handling:** Uses `pathlib` for robust and platform-independent path manipulation.