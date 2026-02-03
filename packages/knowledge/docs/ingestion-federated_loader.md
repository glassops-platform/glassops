---
type: Documentation
domain: knowledge
last_modified: 2026-02-02
generated: true
source: packages/knowledge/ingestion/federated_loader.py
generated_at: 2026-02-02T22:30:16.913920
hash: f661f42a2a0ba297001606a606746e5498fb799fe7172ea92b496ebfd91fa0db
---

## Federated Document Loader Documentation

This module provides functionality for discovering, chunking, and hashing documentation files within a repository. It is designed to prepare documentation for use with large language models (LLMs) and knowledge retrieval systems. We aim to ingest documentation from various sources within a project to build a comprehensive knowledge base.

**Key Responsibilities:**

*   **Document Discovery:** Locates relevant documentation files based on predefined patterns.
*   **Content Chunking:** Divides large documents into smaller, more manageable chunks based on semantic headers.
*   **Content Hashing:** Generates SHA256 hashes for each chunk to ensure data integrity and enable efficient duplicate detection.
*   **Metadata Handling:** Extracts and associates frontmatter metadata with document chunks.

### Core Functions

**1. `hash_content(text: str) -> str`**

This function calculates the SHA256 hash of a given text string.

*   **Parameters:**
    *   `text` (str): The input text string to be hashed.
*   **Return Value:**
    *   `str`: The SHA256 hash of the input text, represented as a hexadecimal string.
*   **Purpose:** Provides a consistent and unique identifier for each document chunk.

**2. `discover_and_chunk_docs(root_dir: str = ".") -> List[Dict]`**

This is the primary function of the module. It scans a specified directory (or the current directory if none is provided) for documentation files, chunks their content, and returns a list of dictionaries containing metadata about each chunk.

*   **Parameters:**
    *   `root_dir` (str, optional): The root directory to start the document search. Defaults to the current directory (".")
*   **Return Value:**
    *   `List[Dict]`: A list of dictionaries, where each dictionary represents a document chunk and contains the following keys:
        *   `path` (str): A unique identifier for the chunk, combining the original file path and a chunk index (e.g., "path/to/file.md#chunk-0").
        *   `source_file` (str): The original file path from which the chunk was extracted.
        *   `content` (str): The text content of the chunk.
        *   `hash` (str): The SHA256 hash of the chunk's content.
*   **Behavior:**
    1.  **Document Discovery:** Uses a set of predefined file patterns (e.g., "docs/**/*.md", "packages/**/README.md") to locate potential documentation files within the specified `root_dir`.  The `glob` module is used for recursive file searching.
    2.  **Path Deduplication:** Removes duplicate file paths to avoid processing the same document multiple times.
    3.  **Ignored Directories:** Skips files located within specified ignored directories (e.g., "node_modules", ".git") to avoid including irrelevant content.
    4.  **File Reading:** Reads the content of each identified documentation file, handling UTF-8 encoding.
    5.  **Frontmatter Parsing:** Attempts to parse YAML frontmatter from the beginning of the file. If successful, the frontmatter metadata is extracted and stored. The frontmatter is then removed from the content before chunking.
    6.  **Content Chunking:** Splits the document content into chunks based on header levels. It prioritizes splitting by Level 2 headers (`##`) to preserve context within sections. If no Level 2 headers are found, it falls back to splitting by Level 1 headers (`#`).
    7.  **Chunk Metadata Creation:** Creates a dictionary for each chunk, including its path, source file, content, and hash.
    8.  **Error Handling:** Includes `try...except` blocks to gracefully handle potential errors during file reading, YAML parsing, and other operations.  Warnings are printed to the console for failed operations.

### Design Decisions and Patterns

*   **Type Hints:** The code extensively uses type hints (e.g., `str`, `List[Dict]`) to improve code readability and maintainability. Type hints also enable static analysis tools to catch potential errors.
*   **SHA256 Hashing:** The use of SHA256 hashing ensures the integrity of the document chunks and allows for efficient duplicate detection.
*   **Semantic Chunking:** The chunking strategy based on headers aims to create meaningful chunks that preserve context and are suitable for LLM processing.
*   **Frontmatter Support:** The module supports parsing YAML frontmatter to extract metadata associated with documentation files.
*   **Robust Error Handling:** The inclusion of `try...except` blocks and warning messages ensures that the module can handle unexpected errors gracefully.
*   **Glob Pattern Flexibility:** The use of glob patterns allows for easy customization of the document discovery process.
*   **Helper Function for Splitting:** The `split_by_header` function encapsulates the logic for splitting text by regular expression patterns, promoting code reuse and readability.