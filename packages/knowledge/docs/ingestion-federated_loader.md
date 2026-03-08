---
type: Documentation
domain: knowledge
last_modified: 2026-03-08
generated: true
source: packages/knowledge/ingestion/federated_loader.py
generated_at: 2026-03-08T22:46:07.650169
hash: 8b0b50384046ebd768e0990496f28defad52e6b03065a5b40665b233716b9a97
---

## Federated Document Loader Documentation

This module is responsible for discovering, chunking, and hashing documentation files from various sources within a repository. It prepares the documentation for ingestion into a knowledge base. The primary goal is to break down large documents into smaller, manageable chunks while preserving contextual information.

**Key Classes & Data Structures:**

This module does not define any classes. It primarily uses dictionaries and lists to represent document data. The core data structure is a dictionary representing a document chunk, with the following keys:

*   `path`: A unique identifier for the chunk, combining the original file path and a chunk number (e.g., `docs/example.md#chunk-0`).
*   `source_file`: The original file path from which the chunk was extracted.
*   `content`: The text content of the chunk.
*   `hash`: The SHA256 hash of the chunk's content, used for deduplication and change detection.
*   (Optional) Frontmatter metadata: Any metadata parsed from the document's frontmatter (YAML).

**Important Functions:**

*   `hash_content(text: str) -> str`: This function takes a string as input and returns its SHA256 hash as a hexadecimal string. This is used to uniquely identify document chunks. The type hint `str` specifies that both the input and output are strings.
*   `discover_and_chunk_docs(root_dir: str = ".") -> List[Dict]`: This is the main function of the module. It performs the following steps:
    1.  **Configuration Loading:** Attempts to load document paths from a `config.json` file located in `root_dir/packages/knowledge/config`. If loading fails, it falls back to a set of default paths. The `root_dir` parameter specifies the base directory to search from, defaulting to the current directory.
    2.  **Pattern Generation:**  Constructs file patterns (using `glob`) based on the loaded or default paths. It includes patterns for `.md` files and `README.md` files.
    3.  **File Discovery:** Uses `glob` to find all files matching the generated patterns within the `root_dir`.
    4.  **Filtering:** Ignores files located within common directories like `node_modules`, `.git`, and `venv`.
    5.  **Content Extraction & Parsing:** Reads the content of each discovered file. It attempts to parse YAML frontmatter from the beginning of the file. If successful, the frontmatter is stored as metadata and removed from the main content.
    6.  **Chunking:** Splits the document content into smaller chunks based on header levels (##, then #). This aims to preserve context by grouping related information. If no headers are found, the entire file is treated as a single chunk.
    7.  **Hashing & Record Creation:** For each chunk, it calculates the SHA256 hash of the content and creates a dictionary containing the chunk's path, source file, content, and hash. Frontmatter metadata is also added to the dictionary.
    8.  **Return Value:** Returns a list of these dictionaries, each representing a document chunk. The type hint `List[Dict]` indicates that the function returns a list of dictionaries.

**Design Decisions & Patterns:**

*   **Configuration-Driven:** The module uses a configuration file (`config.json`) to define the locations of documentation. This allows for flexibility and easy customization without modifying the code.
*   **Fallback Mechanism:**  If the configuration file is missing or invalid, the module falls back to a set of hardcoded default paths.
*   **Globbing for File Discovery:** The `glob` module is used for efficient file discovery based on patterns.
*   **Header-Based Chunking:** The module prioritizes splitting documents based on header levels to maintain contextual integrity.
*   **SHA256 Hashing:** SHA256 is used for content hashing, providing a strong guarantee of uniqueness and enabling efficient change detection.
*   **Error Handling:** The code includes `try...except` blocks to handle potential errors during file reading, configuration loading, and YAML parsing, preventing the process from crashing. Warnings are printed to the console in case of errors.
*   **Type Hints:** Type hints are used throughout the code to improve readability and maintainability, and to enable static analysis.
*   **Frontmatter Support:** The module supports parsing YAML frontmatter to extract metadata from documents.
*   **Path Handling:** Uses `pathlib` for robust and platform-independent path manipulation.