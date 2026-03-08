---
type: Documentation
domain: knowledge
last_modified: 2026-03-08
generated: true
source: packages/knowledge/rag/query_engine.py
generated_at: 2026-03-08T22:47:03.423576
hash: ef028d8bb839e7f41f8cf7bbea87e5828cd2f691b3362054c7a99ea20c94e0b0
---

## GlassOps Knowledge Query Engine Documentation

This module provides a Retrieval-Augmented Generation (RAG) system for querying a knowledge base. It combines information retrieval from a vector database with a large language model to provide informed answers to user questions.

**Module Responsibilities:**

The primary responsibility of this module is to accept a user query, retrieve relevant information from a knowledge base, and generate a concise and informative answer using a language model. It handles embedding the query, querying a vector database (ChromaDB), constructing a context from the retrieved results, and prompting a language model (Gemini) to generate the final answer.

**Key Classes & Roles:**

*   **`chromadb.PersistentClient`**:  This class from the ChromaDB library manages the connection to the vector database. It allows for storing and retrieving vector embeddings and associated documents.
*   **`chromadb.Collection`**: Represents a collection within ChromaDB, used to organize and query the knowledge base.
*   **`google.genai.Client`**: This class from the Google Gemini API provides access to the language model for generating responses.

**Important Functions & Behavior:**

*   **`query_index(query: str, n_results: int = 5) -> str`**: This is the main function of the module.
    *   **Input:**
        *   `query`: The user's question as a string.
        *   `n_results`: The number of relevant documents to retrieve from the vector database (defaults to 5).
    *   **Process:**
        1.  **Embedding:** Converts the input `query` into a vector embedding using the `get_embeddings_for_docs` function.
        2.  **ChromaDB Query:** Queries the ChromaDB vector database using the query embedding to find the most similar documents. The database location is determined by a configuration file, defaulting to "glassops\_index".
        3.  **Context Construction:** Extracts the relevant text (`documents`) and identifiers (`ids`) from the ChromaDB query results.
        4.  **Trigger-Based File Injection:** Checks for keywords in the query against a configuration file. If a keyword is found, the corresponding file content is prepended to the context, providing additional information.
        5.  **Gemini Prompting:** Constructs a prompt for the Gemini language model, including system context (defined in a configuration file or a default value), the retrieved context, and the user's query.
        6.  **Response Generation:** Sends the prompt to the Gemini model and returns the generated response along with the sources of the information.
    *   **Output:** A string containing the generated answer, or an error message if any step fails.

*   **`get_embeddings_for_docs(docs: list[dict]) -> list[list[float]]`**: (This function is imported from another module.) This function is responsible for generating vector embeddings for a list of documents. It is used to convert the user query into a vector representation that can be compared to the embeddings stored in ChromaDB.

**Type Hints:**

The code uses type hints (e.g., `query: str`, `n_results: int`) to improve code readability and maintainability. These hints specify the expected data types for function arguments and return values, allowing for static analysis and error detection.

**Notable Patterns & Design Decisions:**

*   **Configuration Management:** The module relies on a `config.json` file to manage settings such as the vector database location, system context for the language model, and trigger keywords for file injection. This allows for easy customization without modifying the code.
*   **Error Handling:** The code includes `try...except` blocks to handle potential errors during embedding generation, database access, file loading, and language model interaction. This ensures that the system can gracefully handle unexpected situations.
*   **Trigger Mechanism:** The trigger mechanism allows for injecting specific files into the context based on keywords in the user's query. This is useful for providing additional information relevant to specific topics.
*   **Modular Design:** The use of imported functions (e.g., `get_embeddings_for_docs`) promotes modularity and code reuse.
*   **Contextual Prompting:** The prompt sent to the Gemini model includes a system context to guide the model's behavior and ensure that the generated answers are relevant and accurate.
*   **Source Attribution:** The generated response includes a list of sources, allowing users to verify the information and understand where it came from.