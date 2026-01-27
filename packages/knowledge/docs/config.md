---
type: Documentation
domain: knowledge
origin: packages/knowledge/config/config.json
last_modified: 2026-01-26
generated: true
source: packages/knowledge/config/config.json
generated_at: 2026-01-26T05:09:08.693Z
hash: 9b46966ca6f27e8329973e83e1d00448e16c343041a88f1e393ea5bb5ca8b9c4
---

# Knowledge Configuration

This document details the configuration settings for the knowledge management system. This configuration governs how documents are processed, embedded, stored, and retrieved. It defines the core components and parameters used for building and maintaining a knowledge base.

## Overview

The configuration is structured into several key sections: embedding models, vector store settings, document source paths, batch processing size, and drift detection threshold. These settings work together to enable semantic search and retrieval of information from a collection of documents.

## Schema Details

### `embedding_models`

This section defines the models used for generating vector embeddings from text. Embeddings are numerical representations of text that capture semantic meaning, enabling similarity comparisons.

*   **`primary`** (String, *Required*): Specifies the primary embedding model to use. This model is preferred for generating embeddings.  Example: `"gemini-embedding-1.0"`
*   **`fallback`** (String, *Required*): Specifies a fallback embedding model. This model is used if the primary model is unavailable or encounters an error. Example: `"gemma-3-12b-it"`

**Use Case:**  Allows for model redundancy and graceful degradation. If the preferred embedding model is unavailable, the system can automatically switch to the fallback model, ensuring continued functionality.

### `vector_store`

This section configures the vector database used for storing and querying embeddings.

*   **`type`** (String, *Required*): Specifies the type of vector database to use. Currently, only `"chroma"` is supported. Example: `"chroma"`
*   **`persist_dir`** (String, *Required*): Specifies the directory where the vector database will store its data. This allows for persistent storage of embeddings across sessions. Example: `"glassops-index"`

**Use Case:**  Provides a scalable and efficient way to store and retrieve vector embeddings, enabling fast semantic search.

### `federated_doc_paths`

This section defines the file system paths to be scanned for documents. The system recursively searches these paths for documents to include in the knowledge base.

*   **`federated_doc_paths`** (Array of Strings, *Required*): An array of glob patterns representing the paths to documents.  Supports wildcard characters for flexible path specification.
    *   `"docs/"`:  Scans the `docs` directory.
    *   `"packages/**/adr"`: Scans all `adr` directories within any subdirectory of `packages`.
    *   `"packages/**/docs"`: Scans all `docs` directories within any subdirectory of `packages`.

**Use Case:**  Allows the knowledge base to be built from a distributed set of documents located in various directories, promoting modularity and organization.

### `batch_size`

This section defines the number of documents processed in a single batch during embedding and indexing.

*   **`batch_size`** (Integer, *Required*): Specifies the number of documents to process in each batch.  Larger batch sizes can improve performance but may require more memory. Example: `10`

**Use Case:**  Optimizes the processing of large document collections by dividing the work into manageable batches.

### `drift_threshold`

This section defines the threshold for detecting drift in document embeddings.

*   **`drift_threshold`** (Float, *Required*): Specifies the threshold value. If the similarity between a new document's embedding and the existing embeddings in the vector store falls below this threshold, it indicates potential drift. Example: `0.85`

**Use Case:**  Helps identify when the content of the knowledge base has significantly changed, potentially requiring re-indexing or retraining of embedding models.  This ensures the knowledge base remains relevant and accurate.