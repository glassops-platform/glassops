---
type: Documentation
domain: knowledge
origin: packages/knowledge/ingestion/index_builder.py
last_modified: 2026-01-31
generated: true
source: packages/knowledge/ingestion/index_builder.py
generated_at: 2026-01-31T09:55:06.078841
hash: 9c54c23a508d4b3fccdf0915b486a5525e52745057e432f05a5f2dd21d3dbe91
---

# Glassops Knowledge Index Builder Documentation

This document describes the functionality of the `index_builder` module, which is responsible for creating and maintaining a vector store of knowledge using ChromaDB. This vector store enables efficient similarity searches for question answering and knowledge retrieval.

## Module Purpose

The primary purpose of this module is to ingest document embeddings and store them in a ChromaDB collection. The module handles initialization of the ChromaDB client, creation or retrieval of the collection, and the population of the collection with document data and their corresponding embeddings. It supports updating the index with new or modified documents.

## Key Classes and Roles

The module directly interacts with the `chromadb` library. While it doesn’t define custom classes, it leverages the following core components from `chromadb`:

*   **`chromadb.PersistentClient`**: This class provides a client interface to ChromaDB, enabling persistent storage of the vector index on disk.
*   **`chromadb.Collection`**: Represents a collection within ChromaDB, where document embeddings and metadata are stored and managed.

## Important Functions and Their Behavior

### `build_or_update_index(embeddings)`

This function is the core of the module. It takes a list of embeddings as input and builds or updates the ChromaDB index.

**Parameters:**

*   `embeddings` (list of tuples): A list where each tuple contains a document dictionary (`doc_dict`) and its corresponding embedding vector. The `doc_dict` is expected to have keys "path", "content", and "hash".

**Behavior:**

1.  **Persistence Directory:** Determines the directory for persistent storage of the ChromaDB index. The default location is a subdirectory named `glassops_index` within the current working directory.
2.  **ChromaDB Client Initialization:** Initializes a `chromadb.PersistentClient` to connect to ChromaDB, utilizing the specified persistence directory.
3.  **Collection Management:** Retrieves an existing collection named "glassops\_knowledge" or creates a new one if it doesn't exist. The collection is configured to use cosine similarity for distance calculations (`metadata={"hnsw:space": "cosine"}`).
4.  **Data Preparation:** Iterates through the input `embeddings` list, extracting document IDs, content, metadata, and embedding vectors. The document "path" is used as the ID for upserting. Metadata includes both the document "path" and "hash".
5.  **Upsert Operation:** Uses the `collection.upsert()` method to add or update documents in the ChromaDB collection. This operation efficiently handles both new documents and updates to existing ones.
6.  **Error Handling:** Includes a `try...except` block to catch potential exceptions during the `upsert` operation and prints an error message if an issue occurs.
7.  **Empty Input Handling:** Checks if the input `embeddings` list is empty. If so, it prints a message and returns without performing any indexing operations.

**Type Hints:**

The function uses type hints to improve code readability and maintainability. For example, `embeddings: list of tuples (doc_dict, embedding_vector)` clearly indicates the expected input type.

## Notable Patterns and Design Decisions

*   **Persistence:** The use of `chromadb.PersistentClient` ensures that the vector index is stored on disk, allowing it to be reused across multiple sessions.
*   **ID Strategy:** The document "path" is used as the ID for upserting documents into ChromaDB. This design choice allows for easy updating of documents based on their path. Using the hash would create immutable entries.
*   **Metadata Inclusion:** The inclusion of both "path" and "hash" in the metadata provides valuable information for tracking document provenance and identifying potential changes.
*   **Cosine Similarity:** Configuring the collection with `hnsw:space="cosine"` specifies that cosine similarity should be used to measure the distance between embedding vectors, which is appropriate for many natural language processing tasks.
*   **Error Handling:** The inclusion of a `try...except` block around the `upsert` operation provides basic error handling and prevents the program from crashing if an issue occurs during indexing.