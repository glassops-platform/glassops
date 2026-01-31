---
type: Documentation
domain: knowledge
origin: packages/knowledge/main.py
last_modified: 2026-01-31
generated: true
source: packages/knowledge/main.py
generated_at: 2026-01-31T09:56:08.552277
hash: 7e73371cb5cef0a28a05248d302b108fff4651aa1e1a0eacbb6a40b57634ac4d
---

## GlassOps Knowledge Pipeline Documentation

This document describes the GlassOps Knowledge Pipeline, a system designed to manage and query documentation from various sources. It encompasses discovery, embedding, indexing, drift detection, and retrieval-augmented generation (RAG) capabilities.

**Module Purpose:**

The primary purpose of this module is to provide a centralized knowledge base for GlassOps, enabling efficient documentation management and intelligent querying. It automates the process of collecting documentation, converting it into a searchable format, and responding to natural language queries.

**Key Classes:**

*   **Generator:** This class handles documentation generation from source code. It takes a root directory as input and extracts documentation based on specified patterns. The `run()` method executes the documentation generation process.

**Important Functions:**

*   **`discover_and_chunk_docs()`:** This function discovers documentation from federated sources and divides it into smaller, manageable chunks. It returns a list of document objects.
*   **`get_embeddings_for_docs(docs, batch_size)`:** This function computes embeddings for a list of document chunks. It uses a router to select an embedding model (Gemini is primary, Gemma is fallback). The `batch_size` parameter controls the number of documents processed in each batch. It returns a list of embedding vectors.
*   **`build_or_update_index(embeddings)`:** This function builds or updates a vector store index using the provided embeddings. This index enables efficient similarity searches.
*   **`detect_drift(embeddings, threshold)`:** This function detects semantic drift in the knowledge base by comparing the current embeddings to a baseline. The `threshold` parameter determines the sensitivity of the drift detection. It returns a list of documents identified as having drifted.
*   **`query_index(query)`:** This function performs a RAG query against the vector store index. It takes a natural language query as input and returns a relevant response.
*   **`run_generate(patterns)`:** This function initiates the documentation generation process using the `Generator` class. It accepts a list of glob patterns to specify which files to process.
*   **`run_pipeline()`:** This is the main function that orchestrates the entire pipeline. It handles command-line arguments, calls the appropriate functions, and manages the overall workflow.

**Type Hints:**

The code extensively uses type hints (e.g., `list[str]`, `str`, `None`) to improve code readability and maintainability. These hints specify the expected data types for function arguments and return values, aiding in static analysis and error detection.

**Configuration:**

The pipeline's behavior is configurable through a `config.json` file. This file allows you to adjust parameters such as the batch size for embedding generation and the drift detection threshold. The configuration file is loaded from `packages/knowledge/config/config.json`. Environment variables can override these settings via a `.env` file located in the project root.

**Command-Line Interface:**

The `run_pipeline()` function provides a command-line interface with the following options:

*   `--query` or `-q`:  Run a RAG query against the knowledge base.
*   `query_pos`: Positional arguments representing the query string.
*   `--index` or `-i`: Force re-indexing of documents.
*   `--generate` or `-g`: Generate documentation from source code.
*   `--pattern` or `-p`: Glob pattern(s) for documentation generation.

**Notable Patterns and Design Decisions:**

*   **Modular Design:** The pipeline is structured into distinct modules (ingestion, embeddings, indexing, drift, RAG, generation) to promote code organization and reusability.
*   **Router-Based Embedding:** The use of a router allows for flexibility in selecting the appropriate embedding model based on availability and performance.
*   **Semantic Drift Detection:** The inclusion of semantic drift detection helps maintain the quality and relevance of the knowledge base over time.
*   **Configuration Management:** The use of a configuration file and environment variables enables easy customization of the pipeline's behavior.
*   **Argument Parsing:** The `argparse` module is used to provide a flexible and user-friendly command-line interface.
*   **Path Management:** The `pathlib` module is used for robust and platform-independent path manipulation.
*   **Error Handling:** While not explicitly shown in this excerpt, a production system would include comprehensive error handling and logging.