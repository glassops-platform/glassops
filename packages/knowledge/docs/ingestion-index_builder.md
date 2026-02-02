---
type: Documentation
domain: knowledge
origin: packages/knowledge/ingestion/index_builder.py
last_modified: 2026-02-01
generated: true
source: packages/knowledge/ingestion/index_builder.py
generated_at: 2026-02-01T19:34:25.053102
hash: 9c54c23a508d4b3fccdf0915b486a5525e52745057e432f05a5f2dd21d3dbe91
---

# Glassops Knowledge Index Builder Documentation

This document describes the functionality of the `index_builder` module, which is responsible for creating and updating a vector store used for knowledge retrieval. The module leverages ChromaDB for storing and querying embeddings.

## Module Purpose

The primary purpose of this module is to ingest document embeddings and store them in a ChromaDB collection. This allows for efficient similarity searches based on the semantic meaning of the document content. The module handles both initial index creation and subsequent updates to the index.

## Key Classes and Roles

The module directly interacts with the `chromadb` library. While it doesn't define custom classes, the core component is the ChromaDB `PersistentClient` and `Collection`.

*   **`chromadb.PersistentClient`**: This class provides a client interface to ChromaDB, enabling persistent storage of the vector index to disk.
*   **`chromadb.Collection`**: Represents a collection within ChromaDB where embeddings and associated metadata are stored.

## Important Functions and Their Behavior

### `build_or_update_index(embeddings)`

This function is the main entry point for building or updating the knowledge index.

*   **Purpose**:  Takes a list of document-embedding pairs and stores them in the ChromaDB collection.
*   **Parameters**:
    *   `embeddings`: A list of tuples, where each tuple contains a document dictionary (`doc_dict`) and its corresponding embedding vector. The `doc_dict` is expected to have keys "path", "content", and "hash".
*   **Behavior**:
    1.  Determines the persistence directory for the ChromaDB database, defaulting to a directory named "glassops\_index" in the current working directory.
    2.  Initializes a `chromadb.PersistentClient` to connect to the ChromaDB database.
    3.  Retrieves or creates a collection named "glassops\_knowledge" within ChromaDB. The collection is configured to use cosine similarity for distance calculations (`metadata={"hnsw:space": "cosine"}`).
    4.  Iterates through the provided `embeddings` list. For each document and embedding:
        *   Extracts the document path, content, and hash.
        *   Uses the document path as the unique identifier (`id`) for the document in ChromaDB. This design choice allows for easy updates to documents based on their path.
        *   Appends the document ID, content, metadata (including path and hash), and embedding vector to separate lists.
    5.  If no documents are provided (empty `embeddings` list), prints a message and returns.
    6.  Uses the `collection.upsert()` method to add or update the documents in the ChromaDB collection. The `upsert` operation handles both inserting new documents and updating existing ones based on their IDs.
    7.  Prints a success or error message indicating the outcome of the indexing process.
*   **Type Hints**:
    *   `embeddings`: `list[tuple[dict, list[float]]]` – A list of tuples, where each tuple contains a dictionary representing document metadata and a list of floats representing the embedding vector.
*   **Design Decisions**:
    *   The document path is used as the ID for upserting documents. This simplifies updates, as changes to a document at the same path will overwrite the existing entry. Using the hash would create immutable entries.
    *   The ChromaDB collection is configured with cosine similarity, which is appropriate for semantic similarity searches.
    *   Error handling is included to catch potential exceptions during the `upsert` operation.

## Notable Patterns and Design Decisions

The module follows a straightforward pattern of extracting data from the input `embeddings`, preparing it for ChromaDB, and then using the `upsert` operation to populate the vector store. The use of a persistent ChromaDB client ensures that the index is saved to disk and can be reused across sessions. The choice of document path as the ID prioritizes update functionality over immutability.