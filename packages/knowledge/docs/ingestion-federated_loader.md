---
type: Documentation
domain: knowledge
origin: packages/knowledge/ingestion/federated_loader.py
last_modified: 2026-01-26
generated: true
source: packages/knowledge/ingestion/federated_loader.py
generated_at: 2026-01-26T14:08:56.264Z
hash: d8f8723cd2c6f186c687cbe57c2ba7a7db043f32baedb5b95deb38a2643d0766
---

## Federated Document Loader Documentation

This document details the functionality of the Federated Document Loader, a Python module designed to ingest, chunk, and hash documentation from a repository.  It is a core component for building knowledge bases from source code documentation, architectural decision records (ADRs), and other markdown-based documentation.

**Purpose:**

The primary goal of this module is to automatically discover documentation files within a specified directory structure, break them down into manageable chunks, and generate a unique hash for each chunk. This process prepares the documentation for efficient storage, retrieval, and use in applications like question answering systems or knowledge search.

**Key Features:**

*   **Automated Discovery:**  The loader automatically identifies relevant documentation files based on predefined file patterns.
*   **Flexible Chunking:**  Documentation is intelligently split into chunks based on markdown headers (## and #) to preserve contextual information.  If no headers are present, the entire file is treated as a single chunk.
*   **Content Hashing:**  Each chunk is hashed using SHA256 to ensure data integrity and enable efficient change detection.
*   **Robust Error Handling:**  The loader gracefully handles file reading errors and ignores specified directories.
*   **Metadata Preservation:**  The original file path is stored alongside each chunk for provenance tracking.

**Functionality:**

The module provides a single primary function:

*   `discover_and_chunk_docs(root_dir=".")`:
    *   **Input:**  `root_dir` (string, optional): The root directory to scan for documentation. Defaults to the current directory.
    *   **Process:**
        1.  **File Discovery:**  Scans the `root_dir` for files matching the following patterns:
            *   `docs/**/*.md`
            *   `packages/**/docs/**/*.md`
            *   `packages/**/adr/**/*.md`
            *   `packages/**/README.md`
        2.  **Filtering:** Excludes files located within the following directories: `node_modules`, `venv`, `vnev`, `.git`, `__pycache__`, `dist`, and `site-packages`.
        3.  **Chunking:** Reads the content of each identified file and splits it into chunks based on markdown headers (##, then #).
        4.  **Hashing:** Calculates the SHA256 hash of each chunk.
        5.  **Metadata Creation:** Creates a dictionary for each chunk containing:
            *   `path`: A unique identifier for the chunk (e.g., `file/path.md#chunk-0`).
            *   `source_file`: The original file path.
            *   `content`: The text content of the chunk.
            *   `hash`: The SHA256 hash of the chunk.
    *   **Output:** A list of dictionaries, where each dictionary represents a single documentation chunk.

**Dependencies:**

*   `hashlib`:  For generating SHA256 hashes.
*   `glob`: For file path pattern matching.
*   `os`: For interacting with the operating system (file paths, directory traversal).
*   `re`: For regular expression operations used in chunking.

**Usage:**

```python
from packages.knowledge.ingestion.federated_loader import discover_and_chunk_docs

docs = discover_and_chunk_docs(root_dir="/path/to/your/repository")

for doc in docs:
    print(f"Chunk Path: {doc['path']}")
    print(f"Chunk Hash: {doc['hash']}")
    # Further processing of the 'content' can be done here
```

**Considerations:**

*   **Chunk Size:** The chunking strategy aims to create semantically meaningful chunks based on headers.  However, very large files with minimal structure may result in large chunks.
*   **Encoding:** The module assumes UTF-8 encoding for all documentation files.
*   **Performance:** For very large repositories, the file discovery and processing steps may take a significant amount of time.