---
type: Documentation
domain: knowledge
origin: packages/knowledge/config/config.json
last_modified: 2026-01-31
generated: true
source: packages/knowledge/config/config.json
generated_at: 2026-01-31T11:03:39.513707
hash: 20fe2a53c1392c0cfc0142b90c8d3cc228fc21039203bc303e72df5b61d28dde
---

# Knowledge Configuration

This document details the configuration options for the knowledge retrieval system. This system powers intelligent responses within the GlassOps platform by indexing documentation and providing context to language models.

## Overview

The configuration file defines how documentation is processed, stored, and retrieved. It specifies the embedding models used to create vector representations of the documentation, the vector database for storage, the source documents to index, and parameters controlling the retrieval process.

## Configuration Parameters

### `embedding_models`

This section configures the embedding models used to convert text into vector representations.

*   `primary` (string, required): The primary embedding model to use. Currently set to `gemini-embedding-1.0`. This model is preferred for generating embeddings.
*   `fallback` (string, required): The fallback embedding model to use if the primary model is unavailable. Currently set to `gemma-3-12b-it`.

### `vector_store`

This section configures the vector database used to store and retrieve document embeddings.

*   `type` (string, required): The type of vector database. Currently set to `chroma`.
*   `persist_dir` (string, required): The directory where the vector database will store its data. Currently set to `glassops-index`.  You should ensure this directory is writable.

### `federated_doc_paths`

This is a list of file paths or glob patterns that specify the documentation sources to be indexed.

*   `federated_doc_paths` (array of strings, required):  Each string represents a path or pattern.
    *   `docs/`: Indexes the contents of the `docs` directory.
    *   `packages/**/adr`: Indexes all files with the `.adr` extension within any `packages` subdirectory.
    *   `packages/**/docs`: Indexes all files within `docs` subdirectories within any `packages` subdirectory.

### `retrieval_triggers`

This section maps specific query types (triggers) to a specific documentation file. This allows for targeted retrieval of information related to specific system events or reports.

*   `audit` (string, required): Path to the documentation file for audit-related queries. Currently points to `packages/knowledge/docs/generated/drift_report.md`.
*   `backup` (string, required): Path to the documentation file for backup-related queries. Currently points to `packages/knowledge/docs/generated/drift_report.md`.
*   `legacy` (string, required): Path to the documentation file for legacy-related queries. Currently points to `packages/knowledge/docs/generated/drift_report.md`.
*   `overlap` (string, required): Path to the documentation file for overlap-related queries. Currently points to `packages/knowledge/docs/generated/drift_report.md`.
*   `drift` (string, required): Path to the documentation file for drift-related queries. Currently points to `packages/knowledge/docs/generated/drift_report.md`.

### `batch_size`

This parameter controls the number of documents processed in each batch during indexing.

*   `batch_size` (integer, required):  The number of documents to process in a single batch. Currently set to `10`.  Adjusting this value can impact indexing performance.

### `drift_threshold`

This parameter defines the similarity threshold used to identify significant changes (drift) between documentation versions.

*   `drift_threshold` (float, required): A value between 0 and 1 representing the similarity threshold. Currently set to `0.85`. Lower values indicate a higher sensitivity to changes.

### `system_context`

This parameter provides the initial context given to the language model when answering questions.

*   `system_context` (string, required): A string containing the system prompt. This prompt instructs the language model on its role, the available documentation sources, and how to handle specific query types (overlap, backup, legacy, drift).  We maintain this context to ensure consistent and accurate responses.