---
type: Documentation
domain: knowledge
origin: packages/knowledge/ingestion/federated_loader.py
last_modified: 2026-01-31
generated: true
source: packages/knowledge/ingestion/federated_loader.py
generated_at: 2026-01-31T08:59:11.675105
hash: 16479342d05b9369478aa6657719d0de08cf932c3abacb79e515d7220a5e4e89
---

## Federated Document Loader Documentation

This module provides functionality for discovering, chunking, and hashing documentation files within a repository. It is designed to prepare documentation for ingestion into a knowledge base or similar system. The primary goal is to break down large documents into smaller, manageable chunks while preserving contextual information.

**Key Responsibilities:**

*   Locating documentation files (Markdown) within a specified directory structure.
*   Reading the content of these files.
*   Splitting the content into chunks based on document headers.
*   Generating a SHA256 hash for each chunk to ensure content integrity and enable efficient comparison.
*   Returning a list of dictionaries, each representing a chunk of documentation.

**Key Classes & Functions:**

*   **`hash_content(text: str) -> str`**: This function takes a string as input and returns its SHA256 hash as a hexadecimal string. It is used to generate unique identifiers for each document chunk, aiding in deduplication and change detection. The type hint `str` specifies that both the input and output are strings.

*   **`discover_and_chunk_docs(root_dir: str = ".") -> List[Dict]`**: This is the core function of the module. It performs the following steps:
    1.  **Document Discovery:** It searches for Markdown files (`.md`) within the `root_dir` and its subdirectories, using a set of predefined glob patterns. These patterns include `docs`, `packages/docs`, `packages/adr`, and `packages/README.md`.
    2.  **Path Handling:** It collects all discovered file paths, removes duplicates, and sorts them.
    3.  **Ignored Directories:** It filters out files located within common directories that should be excluded from the knowledge base (e.g., `node_modules`, `.git`).
    4.  **File Reading & Chunking:** For each valid document, it reads the content and splits it into chunks. The chunking strategy prioritizes splitting by Level 2 headers (`##`) to maintain context. If no Level 2 headers are found, it falls back to splitting by Level 1 headers (`#`). If no headers are present, the entire file is treated as a single chunk.
    5.  **Chunk Metadata:** For each chunk, it creates a dictionary containing the following information:
        *   `path`: A unique identifier for the chunk, combining the original file path with a chunk index (e.g., `path/to/file.md#chunk-0`).
        *   `source_file`: The original file path from which the chunk was extracted.
        *   `content`: The text content of the chunk.
        *   `hash`: The SHA256 hash of the chunk's content.
    6.  **Error Handling:** It includes a `try-except` block to gracefully handle potential errors during file reading and processing, logging warnings for any files that cannot be processed.
    7.  **Return Value:** The function returns a list of these dictionaries, representing all the extracted and processed document chunks. The type hint `List[Dict]` indicates that the return value is a list of dictionaries.

**Design Decisions & Patterns:**

*   **Glob Patterns:** The use of glob patterns provides a flexible way to specify the locations of documentation files.
*   **Header-Based Chunking:** The chunking strategy based on document headers aims to preserve semantic context within each chunk. This is more effective than arbitrary splitting.
*   **SHA256 Hashing:** The use of SHA256 hashing ensures content integrity and allows for efficient deduplication of chunks.
*   **Type Hints:** The inclusion of type hints (`str`, `List`, `Dict`) improves code readability and maintainability, and enables static analysis tools to catch potential errors.
*   **Robust Error Handling:** The `try-except` block ensures that the process continues even if some files cannot be read or processed.
*   **Ignored Directories:** Explicitly ignoring common development directories prevents irrelevant content from being included in the knowledge base.
*   **Chunk ID Generation:** The chunk ID format (`path#chunk-i`) provides a clear and unique identifier for each chunk, linking it back to its original source file.