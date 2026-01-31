---
type: Documentation
domain: knowledge
origin: packages/knowledge/main.py
last_modified: 2026-01-31
generated: true
source: packages/knowledge/main.py
generated_at: 2026-01-31T11:04:33.396733
hash: 11c84f152af5919603a1faa11b95ec8eb3bfd840332ca9614c74bddd3bec3cf6
---

## GlassOps Knowledge Pipeline Documentation

This document describes the GlassOps Knowledge Pipeline, a system designed to manage and query documentation from various sources. It encompasses discovery, embedding, indexing, drift detection, and retrieval-augmented generation (RAG) capabilities.

**Module Purpose:**

The primary goal of this module is to provide a centralized knowledge base for a software project, enabling efficient information retrieval and documentation generation. It automates the process of collecting documentation from diverse sources, converting it into a searchable format, and answering questions based on that knowledge.

**Key Classes and Their Roles:**

*   **`Generator`**: This class handles the generation of documentation from source code files. It takes a root directory as input and processes files matching specified patterns.
*   Other classes are primarily functions within modules, but represent core components:
    *   `discover_and_chunk_docs`: Responsible for locating and dividing documentation into manageable chunks.
    *   `get_embeddings_for_docs`: Converts documentation chunks into numerical representations (embeddings) for semantic search.
    *   `build_or_update_index`: Creates or updates a vector store (index) using the generated embeddings.
    *   `detect_drift`: Identifies documentation that has undergone significant semantic changes.
    *   `query_index`: Executes a search against the vector store and returns relevant information.

**Important Functions and Their Behavior:**

*   **`run_generate(patterns: list[str]) -> None`**:  This function initiates the documentation generation process. It accepts a list of file patterns (e.g., `"packages/**/*.py"`) as input, instructing the `Generator` class to process files matching those patterns. The function prints status messages to the console.
*   **`run_pipeline()`**: This is the main function that orchestrates the entire knowledge pipeline. It parses command-line arguments, performs the following steps (based on arguments):
    1.  **Documentation Discovery & Chunking**: Locates and splits documentation into smaller, manageable pieces.
    2.  **Embedding Generation**: Creates vector embeddings for each documentation chunk, representing their semantic meaning.  It uses a router to select between embedding models (Gemini is primary, Gemma is fallback). The `batch_size` is configurable via a `config.json` file.
    3.  **Index Building/Updating**: Constructs or updates a vector store (index) using the generated embeddings, enabling efficient similarity searches.
    4.  **Semantic Drift Detection**:  Identifies documentation that has significantly changed in meaning, potentially indicating outdated or inaccurate information. A `drift_threshold` is configurable via `config.json`.
    5.  **RAG Query**:  Processes a user-provided query against the vector store, retrieves relevant documentation, and generates a response. If no query is provided, it runs an example query.
*   **`discover_and_chunk_docs() -> list`**: This function discovers documentation files and splits them into smaller chunks. The exact implementation details of discovery are not specified here, but it is responsible for identifying relevant files.
*   **`get_embeddings_for_docs(docs: list, batch_size: int) -> list`**: This function takes a list of documentation chunks and generates embeddings for each chunk. The `batch_size` parameter controls the number of documents processed in each batch.
*   **`build_or_update_index(embeddings: list) -> None`**: This function builds or updates the vector store with the provided embeddings.
*   **`detect_drift(embeddings: list, threshold: float) -> list`**: This function detects semantic drift in the documentation by comparing the embeddings to a baseline. The `threshold` parameter determines the sensitivity of the drift detection.
*   **`query_index(query: str) -> str`**: This function performs a semantic search against the vector store using the provided query and returns the most relevant results.

**Type Hints and Their Significance:**

The code extensively uses type hints (e.g., `patterns: list[str]`, `embeddings: list`, `threshold: float`). These hints improve code readability, maintainability, and allow for static analysis to catch potential errors. They clearly define the expected data types for function arguments and return values.

**Notable Patterns and Design Decisions:**

*   **Configuration Management**: The pipeline loads configuration parameters (e.g., `batch_size`, `drift_threshold`) from a `config.json` file, allowing for easy customization without modifying the code.  It also supports loading environment variables from a `.env` file.
*   **Command-Line Interface**: The `argparse` module is used to create a command-line interface, providing flexibility in how the pipeline is executed. Users can specify options such as running a query, forcing re-indexing, or generating documentation.
*   **Modular Design**: The pipeline is structured into separate modules (e.g., `ingestion`, `embeddings`, `drift`, `rag`, `generation`), each responsible for a specific task. This promotes code reusability and maintainability.
*   **Router-Based Embedding**: The `get_embeddings_for_docs` function uses a router to select an appropriate embedding model. This allows for flexibility and fallback mechanisms in case the primary model is unavailable.
*   **Semantic Drift Detection**: The inclusion of semantic drift detection is a proactive measure to ensure the knowledge base remains accurate and up-to-date.
*   **File Pattern Flexibility**: The documentation generation process supports multiple file patterns, allowing it to process a wide range of source code and documentation formats.