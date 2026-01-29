---
type: Documentation
domain: knowledge
origin: packages/knowledge/ingestion/index_builder.py
last_modified: 2026-01-28
generated: true
source: packages/knowledge/ingestion/index_builder.py
generated_at: 2026-01-28T22:46:07.999394
hash: 9c54c23a508d4b3fccdf0915b486a5525e52745057e432f05a5f2dd21d3dbe91
---

## Glassops Knowledge Index Builder Documentation

This document details the functionality of the knowledge index builder, a component responsible for creating and maintaining a vector store of knowledge documents using ChromaDB. This vector store enables efficient similarity searches for question answering and knowledge retrieval.

**Module Purpose:**

The primary responsibility of this module is to ingest document embeddings and store them in a ChromaDB collection. It handles the initialization of the ChromaDB client, collection creation (if it doesn’t exist), and the population of the collection with document data and their corresponding vector representations. The module supports updating the index with new or modified documents.

**Key Classes & Roles:**

*   **`chromadb.PersistentClient`:** This class from the ChromaDB library is used to establish a connection to a ChromaDB instance with persistent storage. The `PersistentClient` ensures that the vector store data is saved to disk and available across sessions.
*   **`chromadb.Collection`:** Represents a collection within ChromaDB where document embeddings are stored and managed. We create or retrieve a collection named "glassops_knowledge" to hold our knowledge base.

**Important Functions & Behavior:**

*   **`build_or_update_index(embeddings)`:** This is the core function of the module. It accepts a list of tuples, where each tuple contains a document dictionary and its associated embedding vector.
    *   **Input:** `embeddings` – A list of tuples. Each tuple contains a dictionary representing a document (with keys like "path", "content", and "hash") and a list of floats representing the document’s embedding vector. Type hint: `list[tuple[dict, list[float]]]`.
    *   **Process:**
        1.  Determines the persistence directory for the ChromaDB database ("glassops\_index" within the current working directory).
        2.  Initializes a `chromadb.PersistentClient` to connect to the ChromaDB instance.
        3.  Retrieves an existing collection named "glassops\_knowledge" or creates a new one if it doesn’t exist. The collection is configured to use cosine similarity for distance calculations (`metadata={"hnsw:space": "cosine"}`).
        4.  Extracts document IDs, content, metadata, and embedding vectors from the input `embeddings` list. The document's "path" is used as the unique identifier for each document. Metadata includes both the document "path" and its "hash".
        5.  Handles the case where the input `embeddings` list is empty, printing a message and returning early.
        6.  Uses the `upsert` method of the ChromaDB collection to add or update documents in the vector store. The `upsert` operation efficiently handles both new documents and updates to existing documents based on their IDs.
    *   **Output:** None. The function modifies the ChromaDB collection directly. Prints success or error messages to the console.

**Type Hints:**

Type hints are used throughout the code to improve readability and maintainability. They specify the expected data types for function arguments and return values. For example, `embeddings: list[tuple[dict, list[float]]]` clearly indicates that the `embeddings` argument should be a list of tuples, where each tuple contains a dictionary and a list of floats.

**Notable Patterns & Design Decisions:**

*   **Persistence:** The use of `chromadb.PersistentClient` ensures that the vector store is saved to disk, allowing for data persistence across sessions.
*   **Upsert for Updates:** The `upsert` operation is used to efficiently handle both the addition of new documents and the updating of existing documents. Using the document "path" as the ID allows for easy updates when the content at that path changes.
*   **Metadata Storage:** Storing both the document "path" and "hash" as metadata provides flexibility for future use cases, such as versioning or content verification.
*   **Error Handling:** A `try-except` block is used to catch potential exceptions during the `upsert` operation, providing informative error messages to the user.
*   **ID Strategy:** The document path is used as the ID for indexing. This simplifies updates, as changes to a document at the same path will overwrite the existing entry. Using the hash would create immutable entries.
*   **Cosine Similarity:** The collection is configured to use cosine similarity, a common metric for measuring the similarity between vector embeddings.