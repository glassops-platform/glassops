---
type: Documentation
domain: knowledge
last_modified: 2026-02-02
generated: true
source: packages/knowledge/ingestion/index_builder.py
generated_at: 2026-02-02T22:30:34.509650
hash: cb8e1618acbacd4601a48d063e8a72c6243fb85b0172581dda2ab8353f46f698
---

# Glassops Knowledge Index Builder Documentation

This document describes the functionality of the `index_builder` module, which is responsible for creating and updating a vector store used for knowledge retrieval. The module leverages ChromaDB to store and query document embeddings.

## Module Purpose

The primary purpose of this module is to ingest document embeddings and store them in a ChromaDB collection. This allows for efficient similarity searches based on the semantic meaning of the documents. The module handles both initial index creation and updates to existing documents.

## Key Classes and Roles

The module directly interacts with the `chromadb` library. While it doesn't define custom classes, the core component is the ChromaDB `PersistentClient` and `Collection`.

*   **`chromadb.PersistentClient`**: This class provides a client interface to ChromaDB, enabling persistent storage of the vector index to disk.
*   **`chromadb.Collection`**: Represents a collection within ChromaDB where document embeddings and associated metadata are stored.

## Important Functions and Their Behavior

### `build_or_update_index(embeddings)`

This function is the main entry point for building or updating the knowledge index.

*   **Purpose**:  Takes a list of document-embedding pairs and stores them in the ChromaDB collection.
*   **Parameters**:
    *   `embeddings`: A list of tuples, where each tuple contains a document dictionary (`doc_dict`) and its corresponding embedding vector. The `doc_dict` is expected to have keys like "path", "content", and "hash".
*   **Behavior**:
    1.  **Persistence Directory**: Defines the directory where the ChromaDB data will be stored (`glassops_index` in the current working directory).
    2.  **ChromaDB Initialization**: Initializes a `PersistentClient` to connect to ChromaDB, ensuring data is saved to disk.
    3.  **Collection Management**: Retrieves an existing collection named "glassops\_knowledge" or creates a new one if it doesn't exist. The collection is configured to use cosine similarity for distance calculations (`metadata={"hnsw:space": "cosine"}`).
    4.  **Data Preparation**: Iterates through the input `embeddings` list, extracting document IDs, content, metadata, and embedding vectors. The document "path" is used as the ID for upserting. Metadata is constructed by combining the document's path and hash, along with any other relevant string, integer, float, or boolean fields present in the document dictionary.
    5.  **Upsert Operation**: Uses the `collection.upsert()` method to add or update documents in the ChromaDB collection.  This operation efficiently handles both new documents and updates to existing ones based on the document ID.
    6.  **Error Handling**: Includes a `try...except` block to catch potential exceptions during the `upsert` operation and print an error message.
*   **Type Hints**:
    *   `embeddings: list[tuple[dict, list[float]]]` – Specifies that the `embeddings` parameter is a list of tuples. Each tuple contains a dictionary (representing the document) and a list of floats (representing the embedding vector).
*   **Design Decisions**:
    *   The document "path" is used as the ID for upserting. This allows for easy updating of documents if their content changes. Using the hash would create immutable entries.
    *   Metadata is carefully constructed to include relevant document information while excluding the content itself, as the content is stored separately.
    *   The module includes debug print statements to aid in troubleshooting and monitoring.

## Notable Patterns and Design Decisions

The module follows a straightforward pattern of data preparation and storage. The use of ChromaDB's `upsert` operation simplifies the process of both creating and updating the index. The careful handling of metadata ensures that relevant document information is available for filtering and retrieval. The inclusion of error handling improves the robustness of the module.