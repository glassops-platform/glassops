---
type: Documentation
domain: knowledge
origin: packages/knowledge/ingestion/federated_loader.py
last_modified: 2026-01-28
generated: true
source: packages/knowledge/ingestion/federated_loader.py
generated_at: 2026-01-28T22:45:46.084407
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

*   **`hash_content(text: str) -> str`**: This function takes a string as input and returns its SHA256 hash as a hexadecimal string. It is used to generate unique identifiers for each document chunk, allowing for efficient content comparison and deduplication. The type hint `str` specifies that both the input and output are strings.

*   **`discover_and_chunk_docs(root_dir: str = ".") -> List[Dict]`**: This is the core function of the module. It performs the following steps:
    1.  **Document Discovery:** It searches for Markdown files (`.md`) within the `root_dir` (defaulting to the current directory) using a set of predefined glob patterns. These patterns include `docs`, `packages/docs`, `packages/adr`, and `packages/README.md` directories.
    2.  **Path Handling:** It handles potential duplicate file paths and sorts the discovered paths for consistent processing.
    3.  **Ignored Directories:** It skips files located within common directories that should not be included in the knowledge base (e.g., `node_modules`, `.git`).
    4.  **File Reading:** It reads the content of each Markdown file, handling potential file reading errors gracefully. Empty files are skipped.
    5.  **Chunking:** It splits the document content into chunks based on header levels. It prioritizes splitting by Level 2 headers (`##`) to maintain context within sections. If no Level 2 headers are found, it falls back to splitting by Level 1 headers (`#`). If no headers are present, the entire file is treated as a single chunk.
    6.  **Chunk Metadata:** For each chunk, it creates a dictionary containing the following information:
        *   `path`: A unique identifier for the chunk, combining the original file path with a chunk index (e.g., `path/to/file.md#chunk-0`).
        *   `source_file`: The original file path from which the chunk was extracted.
        *   `content`: The text content of the chunk.
        *   `hash`: The SHA256 hash of the chunk's content.
    7.  **Return Value:** It returns a list of these dictionaries, representing all the extracted and processed documentation chunks. The type hint `List[Dict]` indicates that the function returns a list of dictionaries.

**Design Decisions & Patterns:**

*   **Glob Patterns:** The use of glob patterns provides a flexible way to specify the locations of documentation files.
*   **Header-Based Chunking:** Splitting documents based on headers is a semantic approach that aims to preserve context and create more meaningful chunks.
*   **Error Handling:** The `try...except` block ensures that the process continues even if some files cannot be read.
*   **Hashing:** The use of SHA256 hashing provides a reliable way to identify and compare document chunks.
*   **Type Hints:** Type hints are used throughout the code to improve readability and maintainability, and to enable static analysis.
*   **Clear Metadata:** The inclusion of both a unique `path` and the original `source_file` allows for easy tracking and referencing of chunks.
*   **Robust Ignoring:** The explicit list of ignored directories prevents unwanted files from being processed.