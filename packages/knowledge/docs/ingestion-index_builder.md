---
type: Documentation
domain: knowledge
last_modified: 2026-03-08
generated: true
source: packages/knowledge/ingestion/index_builder.py
generated_at: 2026-03-08T22:46:23.356556
hash: 6b7ca399ed8c26d9ffe78d8581ced02e9dd2abff833b53086e4c79306f0385ac
---

## Knowledge Ingestion Index Builder Documentation

This module is responsible for building and updating a vector store index using ChromaDB. The index stores document embeddings, enabling efficient similarity searches for knowledge retrieval.

**Key Components:**

* **`build_or_update_index(embeddings)` Function:** This is the primary function within the module. It takes a list of embeddings and their associated documents as input and updates the ChromaDB vector store.

    * **`embeddings` Parameter:** A list of tuples. Each tuple contains a document dictionary (`doc_dict`) and its corresponding embedding vector. The `doc_dict` is expected to have keys like "path", "content", and "hash", representing the document's location, text content, and a unique hash value, respectively. Additional metadata may also be present.

    * **Configuration Loading:** The function attempts to load configuration settings from a `config.json` file located in the `config` directory (relative to the module’s location). Specifically, it looks for the `persist_dir` setting within the `vector_store` section of the configuration. If the configuration file is not found or an error occurs during loading, it defaults to using "glassops_index" as the persistence directory.

    * **ChromaDB Initialization:** It initializes a ChromaDB `PersistentClient`, connecting to a specified persistence directory. This ensures that the vector store is saved to disk and can be reused across sessions.

    * **Collection Management:** It retrieves an existing ChromaDB collection named "glassops_knowledge" or creates a new one if it doesn't exist. The collection is configured to use cosine similarity for distance calculations (`"hnsw:space": "cosine"`).

    * **Data Preparation:** The function iterates through the input `embeddings` list, extracting relevant information from each document dictionary:
        * **IDs:** Document paths are used as unique identifiers for each document within the vector store.
        * **Documents:** The document content is extracted.
        * **Metadata:** A metadata dictionary is created for each document, including the document path, hash, and any other relevant string, integer, float, or boolean fields present in the original document dictionary.
        * **Embedding Vectors:** The embedding vector associated with the document is extracted.

    * **ChromaDB Upsert:** The function uses the `upsert` method of the ChromaDB collection to add or update documents in the vector store. The `upsert` operation efficiently handles both new documents and updates to existing documents based on their IDs.

    * **Error Handling:** The function includes `try...except` blocks to handle potential errors during configuration loading and ChromaDB indexing, providing informative error messages.

**Type Hints:**

The `build_or_update_index` function uses type hints to improve code readability and maintainability. For example, `embeddings: list of tuples (doc_dict, embedding_vector)` clearly indicates the expected input type.

**Design Decisions:**

* **Persistence:** The use of `chromadb.PersistentClient` ensures that the vector store is saved to disk, allowing for persistent storage and reuse of the index.
* **Document Identification:** Document paths are used as IDs for upsert operations. This allows for easy updating of documents if their content changes. Using the hash value as an ID would make the index more immutable.
* **Metadata Handling:** The function dynamically extracts metadata from the document dictionary, allowing for flexible storage of document-specific information.
* **Configuration:** The module reads configuration settings from a JSON file, making it easy to customize the behavior of the index builder without modifying the code.
* **Error Handling:** Comprehensive error handling is implemented to provide informative messages in case of failures.

**Usage:**

You should call `build_or_update_index` with a list of tuples, where each tuple contains a document dictionary and its corresponding embedding vector. The document dictionary should include at least "path", "content", and "hash" keys. The function will then build or update the ChromaDB vector store with the provided data.