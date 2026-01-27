---
type: Documentation
domain: knowledge
origin: packages/knowledge/ingestion/index_builder.py
last_modified: 2026-01-26
generated: true
source: packages/knowledge/ingestion/index_builder.py
generated_at: 2026-01-26T14:09:14.662Z
hash: 39cd37b37940da1b0b27c716bcc04f941ca3c4ed3407426dd21610052be0113d
---

## GlassOps Knowledge Index Builder Documentation

**1. Introduction**

The Knowledge Index Builder is a Python component responsible for creating and maintaining a vector store index of knowledge base documents. This index enables efficient semantic search and retrieval of information. It leverages ChromaDB, an open-source vector database, for storing and querying document embeddings.

**2. Functionality**

The core function, `build_or_update_index`, processes a list of document-embedding pairs and populates or updates the ChromaDB index.  The process involves:

*   **Initialization:** Establishes a connection to a ChromaDB instance configured for persistent storage. The index data is stored locally in a directory named `glassops_index` within the current working directory.
*   **Collection Management:**  Retrieves an existing ChromaDB collection named `glassops_knowledge` or creates a new one if it doesn't exist. The collection is configured to use cosine similarity for vector comparisons (`hnsw:space = "cosine"`).
*   **Data Preparation:** Extracts relevant information (document content, path, hash, and embedding vector) from the input data. Document paths are used as unique identifiers within the index. Metadata, including document path and hash, is stored alongside each embedding for enhanced context and traceability.
*   **Index Update:**  Utilizes ChromaDB’s `upsert` operation to add new documents and update existing ones.  `upsert` efficiently handles both insertion and update operations based on the document ID (path).
*   **Error Handling:** Includes basic error handling to catch and report exceptions during the indexing process.

**3. Input**

The `build_or_update_index` function accepts a single argument:

*   `embeddings`: A list of tuples. Each tuple contains:
    *   `doc_dict`: A dictionary representing a document with the following keys:
        *   `path`: (string) The file path of the document.  Used as the unique identifier in the index.
        *   `content`: (string) The textual content of the document.
        *   `hash`: (string) A hash value representing the document's content.
    *   `embedding_vector`: (list of floats) The numerical vector representation (embedding) of the document's content.

**4. Output**

The function does not explicitly return a value.  Its primary effect is to modify the ChromaDB index.  It provides console output indicating success or failure, along with the number of documents indexed.

**5. Dependencies**

*   **ChromaDB:**  A Python client for ChromaDB is required (`chromadb`).
*   **Python Standard Library:**  The `os` module is used for file path manipulation.

**6. Configuration**

*   **Persistence Directory:** The ChromaDB index is stored in a directory named `glassops_index` in the current working directory.
*   **Collection Name:** The ChromaDB collection is named `glassops_knowledge`.
*   **Similarity Metric:** Cosine similarity is used for vector comparisons within the ChromaDB collection.

**7. Considerations**

*   **Document Identification:** The document `path` is used as the unique identifier in the index. This allows for easy updating of documents if their content changes.
*   **Immutability:** Using the document `hash` as the ID would provide a degree of immutability, preventing accidental overwrites. The current implementation prioritizes update-in-place behavior.
*   **Error Handling:** The current error handling is basic. More robust error handling and logging may be necessary for production environments.