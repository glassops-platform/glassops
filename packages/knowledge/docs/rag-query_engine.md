---
type: Documentation
domain: knowledge
origin: packages/knowledge/rag/query_engine.py
last_modified: 2026-01-31
generated: true
source: packages/knowledge/rag/query_engine.py
generated_at: 2026-01-31T09:01:03.765181
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
    *   **Context Construction:** The retrieved documents are assembled into a context string, which will be provided to the language model.
    *   **Trigger-Based File Injection:** Before passing the context to the LLM, the function checks for specific keywords in the query. If a keyword is found, it attempts to load and prepend the content of a corresponding file (defined in a `config.json` file) to the context. This allows for dynamic inclusion of up-to-date information.
    *   **Gemini Integration:** The function uses the Gemini language model to generate an answer based on the constructed context and the original query. It uses a system prompt to guide the model's behavior.
    *   **Error Handling:** The function includes robust error handling to catch potential issues during embedding generation, ChromaDB querying, file loading, and language model interaction.  Informative error messages are returned to the user.

**Type Hints:**

The code makes extensive use of type hints (e.g., `query: str`, `n_results: int`) to improve code readability and maintainability. These hints specify the expected data types for function arguments and return values, enabling static analysis and helping to prevent errors.

**Notable Patterns & Design Decisions:**

*   **RAG Architecture:** The module implements a standard RAG pipeline, combining information retrieval with language model generation.
*   **Configuration-Driven Behavior:** The system prompt and file injection triggers are loaded from a `config.json` file, allowing for easy customization without modifying the code.
*   **Modular Design:** The embedding generation is handled by a separate function (`get_embeddings_for_docs`), promoting code reuse and separation of concerns.
*   **Error Handling:** Comprehensive error handling is implemented throughout the function to provide informative messages to the user and prevent unexpected crashes.
*   **Contextual Injection:** The ability to inject content from external files based on query keywords provides a mechanism for incorporating dynamic or frequently updated information into the knowledge base.
*   **Environment Variable for API Key:** The Gemini API key is retrieved from an environment variable (`GOOGLE_API_KEY`), enhancing security and flexibility.
*   **Model Specification:** The code explicitly defines the Gemini model to be used (`gemma-3-12b-it`), allowing for easy switching between different models.