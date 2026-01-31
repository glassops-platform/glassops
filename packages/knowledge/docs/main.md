---
type: Documentation
domain: knowledge
origin: packages/knowledge/main.py
last_modified: 2026-01-31
generated: true
source: packages/knowledge/main.py
generated_at: 2026-01-31T09:00:33.587627
hash: 7e73371cb5cef0a28a05248d302b108fff4651aa1e1a0eacbb6a40b57634ac4d
---

## GlassOps Knowledge Pipeline Documentation

This document describes the GlassOps Knowledge Pipeline, a system designed to manage and query documentation from various sources. It encompasses document discovery, embedding generation, index management, drift detection, and retrieval-augmented generation (RAG) querying.

**Module Purpose:**

The primary purpose of this module is to provide a centralized knowledge base for GlassOps, enabling efficient information retrieval and supporting informed decision-making. It automates the process of collecting, processing, and querying documentation, reducing manual effort and improving accuracy.

**Key Classes and Roles:**

*   **`Generator`:** This class is responsible for automatically generating documentation from source code files. It takes a root directory as input and processes files matching specified patterns.
*   Other classes (within imported modules) handle specific tasks:
    *   `discover_and_chunk_docs` (from `knowledge.ingestion.federated_loader`): Locates and divides documentation into manageable chunks.
    *   `get_embeddings_for_docs` (from `knowledge.embeddings.router_embedding`): Creates numerical representations (embeddings) of the document chunks.
    *   `build_or_update_index` (from `knowledge.ingestion.index_builder`): Constructs or updates a vector store using the generated embeddings.
    *   `detect_drift` (from `knowledge.drift.detect_drift`): Identifies documents where the meaning has changed significantly over time.
    *   `query_index` (from `knowledge.rag.query_engine`): Executes queries against the vector store and retrieves relevant information.

**Important Functions and Their Behavior:**

*   **`run_generate(patterns: list[str])`:**  Initiates the documentation generation process. It accepts a list of file patterns (globs) to identify source code files for documentation extraction. The function instantiates the `Generator` class and calls its `run` method with the provided patterns.
*   **`run_pipeline()`:** This is the main function that orchestrates the entire knowledge pipeline. It parses command-line arguments, performs document ingestion, embedding generation, index building, drift detection, and RAG querying.
    *   It uses `argparse` to handle command-line arguments for querying, re-indexing, generating documentation, and specifying file patterns.
    *   It loads configuration parameters from a `config.json` file.
    *   It conditionally executes pipeline steps based on the provided arguments.
    *   If no query is provided, it runs an example query.
*   **`discover_and_chunk_docs()`:**  (Imported function) Discovers documentation from various sources and splits it into smaller, more manageable chunks.
*   **`get_embeddings_for_docs(docs, batch_size)`:** (Imported function) Computes embeddings for the provided document chunks using a router that prioritizes Gemini and falls back to Gemma. The `batch_size` parameter controls the number of documents processed in each batch.
*   **`build_or_update_index(embeddings)`:** (Imported function) Creates or updates a vector store (index) using the generated embeddings. This index enables efficient similarity searches.
*   **`detect_drift(embeddings, threshold)`:** (Imported function) Detects semantic drift by comparing the embeddings of documents over time. The `threshold` parameter determines the sensitivity of the drift detection.
*   **`query_index(query)`:** (Imported function) Executes a query against the vector store and returns the most relevant documents or information.

**Type Hints and Their Significance:**

The code extensively uses type hints (e.g., `patterns: list[str]`, `query: str`) to improve code readability and maintainability. Type hints allow static analysis tools to verify the correctness of the code and help prevent runtime errors. They also serve as documentation, clarifying the expected data types for function arguments and return values.

**Notable Patterns and Design Decisions:**

*   **Configuration Management:** The pipeline loads configuration parameters from a `config.json` file, allowing for easy customization without modifying the code. Environment variables are also loaded from a `.env` file.
*   **Command-Line Interface:** The use of `argparse` provides a flexible and user-friendly command-line interface for interacting with the pipeline.
*   **Modular Design:** The pipeline is structured into separate modules (e.g., `ingestion`, `embeddings`, `drift`, `rag`, `generation`) with well-defined responsibilities, promoting code reuse and maintainability.
*   **Embedding Router:** The `get_embeddings_for_docs` function employs a router to select the appropriate embedding model (Gemini or Gemma), providing flexibility and resilience.
*   **Error Handling:** While not explicitly shown in this snippet, a production system would include robust error handling and logging mechanisms.
*   **Path Management:** The code uses `pathlib.Path` for robust and platform-independent path manipulation.
*   **Conditional Execution:** The pipeline's behavior is determined by command-line arguments, allowing users to selectively execute specific steps (e.g., re-indexing, querying, documentation generation).