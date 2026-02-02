---
type: Documentation
domain: knowledge
origin: packages/knowledge/ingestion/federated_loader.py
last_modified: 2026-02-01
generated: true
source: packages/knowledge/ingestion/federated_loader.py
generated_at: 2026-02-01T19:34:04.697992
hash: 16479342d05b9369478aa6657719d0de08cf932c3abacb79e515d7220a5e4e89
---

## Federated Document Loader Documentation

This module provides functionality for discovering, chunking, and hashing documentation files within a repository. It is designed to prepare documentation for ingestion into a knowledge base or similar system. The primary goal is to break down large documents into smaller, manageable chunks while preserving contextual information.

**Key Responsibilities:**

*   Locating documentation files based on defined patterns.
*   Reading and processing the content of these files.
*   Splitting documents into chunks based on header levels (##, then #).
*   Generating a SHA256 hash for each chunk to ensure content integrity and enable efficient comparison.
*   Returning a list of dictionaries, each representing a chunk of documentation.

**Key Functions:**

*   `hash_content(text: str) -> str`: This function takes a string as input and returns its SHA256 hash as a hexadecimal string. This is used to uniquely identify each chunk of documentation. The `text` parameter is type-hinted as a string, ensuring that the input is always a string.

*   `discover_and_chunk_docs(root_dir: str = ".") -> List[Dict]`: This is the core function of the module. It performs the following steps:
    1.  **Discovery:** It searches for Markdown (`.md`) and README files within the specified `root_dir` (defaults to the current directory) using a set of predefined glob patterns. These patterns include `docs/**/*.md`, `packages/**/docs/**/*.md`, `packages/**/adr/**/*.md`, and `packages/**/README.md`.
    2.  **Filtering:** It filters out files located within ignored directories such as `node_modules`, `venv`, `.git`, and others. This prevents irrelevant files from being processed.
    3.  **Reading:** It reads the content of each identified file, handling potential encoding errors with `utf-8`.
    4.  **Chunking:** It splits the content into chunks based on header levels. It first attempts to split by Level 2 headers (`##`), then by Level 1 headers (`#`), and finally falls back to using the entire file if no headers are found. A helper function `split_by_header` is used to perform the splitting based on regular expressions.
    5.  **Hashing:** It calculates the SHA256 hash of each chunk using the `hash_content` function.
    6.  **Output:** It returns a list of dictionaries. Each dictionary contains the following keys:
        *   `path`: A unique identifier for the chunk, combining the original file path with a chunk index (e.g., `path/to/file.md#chunk-0`).
        *   `source_file`: The original file path from which the chunk was extracted.
        *   `content`: The text content of the chunk.
        *   `hash`: The SHA256 hash of the chunk's content.

    The function is type-hinted to return a `List[Dict]`, indicating that it returns a list of dictionaries. The `root_dir` parameter is type-hinted as a string.

**Design Decisions and Patterns:**

*   **Glob Patterns:** The use of glob patterns allows for flexible and scalable document discovery.
*   **Header-Based Chunking:** Splitting documents based on headers helps to preserve contextual information and create more meaningful chunks.
*   **Hashing:**  Hashing ensures data integrity and allows for efficient content comparison and deduplication.
*   **Error Handling:** The code includes basic error handling to gracefully handle files that cannot be read.
*   **Type Hints:** The use of type hints improves code readability and maintainability, and enables static analysis.
*   **Directory Ignoring:** The explicit list of ignored directories prevents the ingestion of unwanted files.
*   **Chunk ID Generation:** The `path` field in the output dictionary provides a unique identifier for each chunk, linking it back to its original source file.
*   **Helper Function:** The `split_by_header` function encapsulates the logic for splitting text by regular expression patterns, promoting code reuse and readability.

**Usage:**

You can use the `discover_and_chunk_docs` function to process documentation files in a given directory. For example:

```python
from federated_loader import discover_and_chunk_docs

docs = discover_and_chunk_docs(root_dir="/path/to/your/repository")

for doc in docs:
    print(f"Chunk Path: {doc['path']}")
    print(f"Chunk Hash: {doc['hash']}")
    # You can then use the 'content' field to further process the chunk