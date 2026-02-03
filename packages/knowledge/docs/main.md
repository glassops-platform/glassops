---
type: Documentation
domain: knowledge
last_modified: 2026-02-02
generated: true
source: packages/knowledge/main.py
generated_at: 2026-02-02T22:31:43.282548
hash: 0fd894863622081aa14243c7e23901027617e12b3e94fd4789761d43f303802a
---

## GlassOps Knowledge Pipeline Documentation

This document describes the GlassOps Knowledge Pipeline, a system designed to manage and query documentation from various sources within a software project. The pipeline automates documentation discovery, embedding generation, index creation, drift detection, and retrieval-augmented generation (RAG) querying.

**Module Purpose:**

The `knowledge` package provides tools for building and interacting with a knowledge base derived from source code and other documentation. It supports automated documentation generation, semantic drift monitoring, and efficient information retrieval through vector search.

**Key Classes and Their Roles:**

*   **`Generator`**: This class handles the generation of documentation from source code files. It takes a root directory as input and processes files matching specified patterns.
*   Other classes are primarily functions within modules, but represent core pipeline stages:
    *   `discover_and_chunk_docs`: Responsible for locating and dividing documentation into manageable chunks.
    *   `get_embeddings_for_docs`: Computes vector embeddings for each document chunk.
    *   `build_or_update_index`: Creates or updates a vector store (index) using the generated embeddings.
    *   `detect_drift`: Identifies documents that have undergone significant semantic changes.
    *   `query_index`: Executes a RAG query against the vector store to retrieve relevant information.

**Important Functions and Their Behavior:**

*   **`run_generate(patterns: list[str]) -> None`**:  This function initiates the documentation generation process. It takes a list of file patterns (globs) as input, instructing the `Generator` class to process files matching those patterns. The function prints informational messages to the console during execution.
*   **`run_pipeline()`**: This is the main function that orchestrates the entire knowledge pipeline. It parses command-line arguments, controls the execution flow, and calls other functions to perform specific tasks.
    *   It supports the following command-line arguments:
        *   `--query` or `-q`:  Specifies a query string to execute against the knowledge base.
        *   `query_pos`: Allows providing the query as positional arguments (joined by spaces).
        *   `--index` or `-i`: Forces a re-indexing of all documents.
        *   `--generate` or `-g`: Triggers documentation generation.
        *   `--pattern` or `-p`: Specifies file patterns for documentation generation (can be used multiple times).
*   **`discover_and_chunk_docs() -> list`**: This function discovers documentation files based on predefined or user-provided patterns and splits them into smaller chunks for embedding. The return value is a list of document chunks.
*   **`get_embeddings_for_docs(docs: list, batch_size: int) -> list`**: This function takes a list of document chunks and generates vector embeddings for each chunk. It uses a router to select an embedding model (Gemini is primary, Gemma is fallback). The `batch_size` parameter controls the number of documents processed in each batch.
*   **`build_or_update_index(embeddings: list) -> None`**: This function builds or updates a vector store (index) using the provided embeddings. The index allows for efficient similarity search.
*   **`detect_drift(embeddings: list, threshold: float) -> list`**: This function detects semantic drift by comparing the current embeddings to a baseline. It returns a list of documents that have drifted beyond the specified `threshold`.
*   **`query_index(query: str) -> str`**: This function executes a RAG query against the vector store. It retrieves the most relevant documents based on the query and generates a response.

**Type Hints and Their Significance:**

The code extensively uses type hints (e.g., `patterns: list[str]`, `batch_size: int`) to improve code readability and maintainability. Type hints allow static analysis tools to catch potential errors and provide better code completion suggestions. They also serve as documentation, clarifying the expected data types for function arguments and return values.

**Notable Patterns and Design Decisions:**

*   **Configuration Management:** The pipeline loads configuration parameters from a `config.json` file, allowing for easy customization of settings such as batch size and drift threshold.  It also supports loading environment variables from a `.env` file.
*   **Command-Line Interface:** The `argparse` module is used to create a command-line interface, providing a flexible way to interact with the pipeline.
*   **Modularity:** The pipeline is structured into separate modules (e.g., `ingestion`, `embeddings`, `drift`, `rag`, `generation`), each responsible for a specific task. This promotes code reuse and maintainability.
*   **Embedding Router/Fallback:** The `get_embeddings_for_docs` function uses a router to select an embedding model. This allows for easy switching between different models and provides a fallback mechanism in case the primary model is unavailable.
*   **Semantic Drift Detection:** The inclusion of semantic drift detection helps ensure the knowledge base remains up-to-date and accurate.
*   **Path Management:** The code uses `pathlib.Path` for robust and platform-independent path manipulation.
*   **Error Handling:** While not explicitly shown in the provided snippet, a production system would include comprehensive error handling and logging.