---
type: Documentation
domain: knowledge
origin: packages/knowledge/rag/query_engine.py
last_modified: 2026-02-01
generated: true
source: packages/knowledge/rag/query_engine.py
generated_at: 2026-02-01T19:36:02.494184
hash: 09603013c88193d7faf419e2b18b13bf32b3d9de8e28c293f2d3e97f0923446f
---

## GlassOps Knowledge Query Engine Documentation

This module provides a Retrieval-Augmented Generation (RAG) system for querying a knowledge base. It combines information retrieval from a vector database (ChromaDB) with a large language model (Gemini) to provide informed answers to user questions.

**Module Responsibilities:**

The primary function of this module is to accept a user query, retrieve relevant documents from a knowledge base, and generate a concise answer using a language model. It also incorporates a mechanism for injecting content from specific files based on keywords found in the query.

**Key Classes & Data Structures:**

This module does not define any custom classes. It relies on external libraries like `chromadb` and `google.genai`. The core data structures are lists of strings representing document chunks and their corresponding IDs.

**Important Functions:**

*   **`query_index(query: str, n_results: int = 5) -> str`**: This is the main function of the module. It takes a user query (a string) and an optional number of results to retrieve (defaulting to 5). It returns a string containing the generated answer, or an error message if something goes wrong.

    1.  **Embedding Generation:** The function first generates an embedding vector for the input query using the `get_embeddings_for_docs` function. This embedding represents the semantic meaning of the query.
    2.  **ChromaDB Query:** It then queries a ChromaDB collection named "glassops\_knowledge" using the query embedding. ChromaDB is a vector database that stores document embeddings for efficient similarity search.
    3.  **Context Construction:** The function retrieves the most relevant document chunks and their IDs from ChromaDB.
    4.  **Trigger-Based File Injection:** It checks for predefined keywords in the query. If a keyword is found, it attempts to inject the content of a corresponding file (specified in a `config.json` file) into the context. This allows for dynamic inclusion of up-to-date information.
    5.  **Answer Generation:** Finally, it uses the Gemini language model to generate an answer based on the combined context (retrieved documents and potentially injected files). The prompt provided to the model instructs it to answer based *only* on the provided context.
    6.  **Error Handling:** The function includes robust error handling to catch potential issues during embedding generation, database querying, file injection, and answer generation. It returns informative error messages to the user.

**Type Hints:**

The function signatures use type hints (e.g., `query: str`, `n_results: int`) to improve code readability and maintainability. These hints specify the expected data types for function arguments and return values.

**Notable Patterns and Design Decisions:**

*   **RAG Architecture:** The module implements a standard RAG pipeline, combining information retrieval and generation.
*   **Configuration-Driven Behavior:** The system context for the language model and the file injection triggers are loaded from a `config.json` file. This allows for easy customization without modifying the code.
*   **Error Handling:** Comprehensive error handling is implemented throughout the function to provide informative messages to the user and prevent unexpected crashes.
*   **Contextual Prompting:** The prompt sent to the Gemini model is carefully crafted to instruct it to answer based solely on the provided context, minimizing the risk of hallucination.
*   **Modular Embedding Function:** The embedding generation is delegated to a separate function (`get_embeddings_for_docs`) to promote code reuse and maintainability.
*   **File Injection Priority:** Injected files are prepended to the context, giving them higher priority during answer generation.
*   **Environment Variable for API Key:** The Google API key is read from an environment variable (`GOOGLE_API_KEY`) for security and flexibility.
*   **Model Specification:** The model used for generation is configurable via the `model_name` variable, currently set to 'gemma-3-12b-it'.