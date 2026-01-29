---
type: Documentation
domain: knowledge
origin: packages/knowledge/main.py
last_modified: 2026-01-28
generated: true
source: packages/knowledge/main.py
generated_at: 2026-01-28T22:47:20.774856
hash: 7e73371cb5cef0a28a05248d302b108fff4651aa1e1a0eacbb6a40b57634ac4d
---

## GlassOps Knowledge Pipeline Documentation

This document describes the GlassOps Knowledge Pipeline, a system designed to ingest, process, and query documentation from various sources. It supports automated documentation generation, semantic drift detection, and retrieval-augmented generation (RAG) for answering questions based on the knowledge base.

**Module Purpose and Responsibilities:**

The `main.py` module serves as the entry point and orchestrator for the entire knowledge pipeline. It handles argument parsing, calls individual components for document loading, embedding generation, index building, drift detection, and query execution.  It also includes functionality for generating documentation from source code.

**Key Classes and Their Roles:**

*   **`Generator` (from `knowledge.generation`):** This class is responsible for generating documentation from source code files based on provided patterns. It takes the root directory of the project as input and processes files matching the specified glob patterns.
*   Other classes are primarily functions within modules, but represent core pipeline components:
    *   `discover_and_chunk_docs` (from `knowledge.ingestion.federated_loader`): Locates and prepares documentation for embedding.
    *   `get_embeddings_for_docs` (from `knowledge.embeddings.router_embedding`): Creates vector representations of the documentation.
    *   `build_or_update_index` (from `knowledge.ingestion.index_builder`): Manages the vector store.
    *   `detect_drift` (from `knowledge.drift.detect_drift`): Identifies changes in the semantic meaning of the documentation.
    *   `query_index` (from `knowledge.rag.query_engine`): Executes queries against the vector store and returns relevant results.

**Important Functions and Their Behavior:**

*   **`run_generate(patterns: list[str]) -> None`:**  This function initiates the documentation generation process. It takes a list of file patterns (glob strings) as input and uses the `Generator` class to create documentation for files matching those patterns.
*   **`run_pipeline()`:** This is the main function that drives the entire pipeline. It parses command-line arguments, determines the appropriate course of action (querying, indexing, generating documentation), and calls the necessary functions to execute the selected tasks.
*   **`discover_and_chunk_docs() -> list`:** This function discovers documentation from various sources and splits it into smaller chunks suitable for embedding. The specific sources and chunking strategy are defined within the `knowledge.ingestion.federated_loader` module.
*   **`get_embeddings_for_docs(docs: list, batch_size: int) -> list`:** This function computes embeddings for a list of document chunks. It uses a router to select an embedding model (Gemini is primary, Gemma is a fallback). The `batch_size` parameter controls the number of documents processed in each batch.
*   **`build_or_update_index(embeddings: list) -> None`:** This function builds or updates a vector store (index) using the provided embeddings. The vector store is used for efficient similarity search during query time.
*   **`detect_drift(embeddings: list, threshold: float) -> list`:** This function detects semantic drift in the documentation by comparing the current embeddings to a baseline. Documents with a drift score above the specified `threshold` are flagged as drifted.
*   **`query_index(query: str) -> str`:** This function executes a query against the vector store and returns the most relevant results. It uses a retrieval-augmented generation (RAG) approach to combine the retrieved results with a language model to generate a comprehensive answer.

**Type Hints and Their Significance:**

The code makes extensive use of type hints (e.g., `patterns: list[str]`, `batch_size: int`, `-> None`, `-> list`). These hints improve code readability, maintainability, and allow for static analysis to catch potential errors early in the development process. They clearly define the expected data types for function arguments and return values.

**Notable Patterns and Design Decisions:**

*   **Configuration Management:** The pipeline loads configuration parameters from a `config.json` file, allowing for easy customization of settings such as batch size and drift threshold.  It also supports loading environment variables from a `.env` file.
*   **Command-Line Interface:** The pipeline is designed to be run from the command line, with options for querying the knowledge base, forcing re-indexing, and generating documentation.
*   **Modularity:** The pipeline is broken down into smaller, reusable modules, each responsible for a specific task. This promotes code organization and maintainability.
*   **Embedding Router:** The use of an embedding router allows for flexibility in choosing the appropriate embedding model based on factors such as cost, performance, and availability.
*   **Semantic Drift Detection:** The inclusion of semantic drift detection helps to ensure that the knowledge base remains up-to-date and accurate.
*   **Error Handling:** While not explicitly shown in this overview, robust error handling is expected within the individual modules to ensure the pipeline's stability.
*   **Path Management:** The code uses `pathlib.Path` for robust and platform-independent path manipulation.