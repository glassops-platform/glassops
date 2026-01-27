---
type: Documentation
domain: knowledge
origin: packages/knowledge/main.py
last_modified: 2026-01-26
generated: true
source: packages/knowledge/main.py
generated_at: 2026-01-26T14:05:22.543Z
hash: 5d41e9d3d8607148f1424259cfb9527ae1a07fa000c2fd7e26bcf294e21ed089
---

# GlassOps Knowledge Pipeline Documentation

## Overview

The GlassOps Knowledge Pipeline is a system designed to ingest, process, and query documentation from various sources (federated documentation). It leverages embeddings and a vector store to enable semantic search and Retrieval-Augmented Generation (RAG) for answering questions based on the knowledge base.  The pipeline includes drift detection to identify outdated or semantically changed documents.

## Core Functionality

The pipeline performs the following key steps:

1.  **Documentation Discovery & Chunking:**  Identifies and retrieves documentation from multiple sources, then divides it into manageable chunks.
2.  **Embedding Generation:**  Creates vector embeddings for each document chunk using a router that prioritizes Gemini and falls back to Gemma. Embeddings represent the semantic meaning of the text.
3.  **Vector Store Management:** Builds or updates a vector store (database) containing the embeddings, enabling efficient similarity searches.
4.  **Semantic Drift Detection:**  Monitors the knowledge base for semantic drift, identifying documents whose meaning has changed significantly over time.
5.  **Retrieval-Augmented Generation (RAG):**  Answers user queries by retrieving relevant document chunks from the vector store and using them to augment the query to a language model.

## Usage

The pipeline can be executed from the command line.  Here's a breakdown of the available options:

*   `python packages/knowledge/main.py`: Runs the entire pipeline, including documentation discovery, embedding generation, index building, drift detection, and an example RAG query.
*   `python packages/knowledge/main.py --query "your question"` or `python packages/knowledge/main.py "your question"`:  Performs a RAG query against the existing knowledge base.  The query is provided either via the `--query` flag or as positional arguments.
*   `python packages/knowledge/main.py --index`: Forces a complete re-indexing of the documentation, including discovery, embedding generation, and vector store update.  This is useful when the source documentation has changed significantly.

## Configuration

The pipeline's behavior is configurable through a `config.json` file located at `packages/knowledge/config/config.json`.  Key configurable parameters include:

*   `batch_size`:  The number of documents processed in each batch during embedding generation.  Defaults to 10.
*   `drift_threshold`: The threshold used for semantic drift detection.  Documents with a drift score above this threshold are flagged. Defaults to 0.85.

The pipeline also supports loading environment variables from a `.env` file located in the project root directory.

## Technical Details

*   **Dependencies:** The pipeline relies on several Python packages, including `dotenv`, and internal modules from the `knowledge` package.
*   **Embedding Router:** The pipeline uses a router to select the embedding model. It prioritizes Gemini and falls back to Gemma if Gemini is unavailable.
*   **Vector Store:** The specific vector store implementation is not detailed in the provided code, but it is assumed to be managed by the `build_or_update_index` function.
*   **Drift Detection:** The `detect_drift` function identifies documents that have undergone semantic changes based on a configurable threshold.
*   **Error Handling:** The code does not explicitly demonstrate error handling, but a production implementation should include robust error handling and logging.
*   **Path Management:** Uses `pathlib` for robust and platform-independent path manipulation.
*   **sys.path Modification:** Dynamically adds the `packages` directory to `sys.path` to allow execution from the project root.