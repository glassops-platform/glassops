---
type: Documentation
domain: knowledge
last_modified: 2026-02-02
generated: true
source: packages/knowledge/config/config.json
generated_at: 2026-02-02T22:26:57.983881
hash: 20fe2a53c1392c0cfc0142b90c8d3cc228fc21039203bc303e72df5b61d28dde
---

# Knowledge Configuration

This document details the configuration options for the knowledge retrieval system. This system powers intelligent responses within the GlassOps platform by indexing documentation and providing relevant context to language models.

## Overview

The configuration file defines how documentation is processed, stored, and retrieved. It specifies the embedding models used to create vector representations of the documentation, the vector database for storage, the source locations for documentation, and parameters controlling the retrieval process.

## Configuration Parameters

### `embedding_models`

This section configures the embedding models used to convert text into vector representations.

*   `primary` (string, required): Specifies the primary embedding model.  Currently set to `gemini-embedding-1.0`. This model is preferred for generating embeddings.
*   `fallback` (string, required): Specifies a fallback embedding model. Currently set to `gemma-3-12b-it`. This model is used if the primary model is unavailable or encounters an error.

### `vector_store`

This section configures the vector database used to store and retrieve document embeddings.

*   `type` (string, required): Specifies the type of vector database. Currently set to `chroma`.
*   `persist_dir` (string, required): Specifies the directory where the vector database will store its data. Currently set to `glassops-index`.  You should ensure this directory is writable.

### `federated_doc_paths`

This is a list of file paths or glob patterns that define the locations of documentation to be indexed.

*   `federated_doc_paths` (array of strings, required):  Each string represents a path or pattern.
    *   `docs/`: Indexes the main documentation directory.
    *   `packages/**/adr`: Indexes Architecture Decision Records (ADR) within any package.
    *   `packages/**/docs`: Indexes documentation within any package.

### `retrieval_triggers`

This section maps specific query types (triggers) to a specific documentation file.

*   `audit` (string, required): Path to the drift report for "audit" related queries.
*   `backup` (string, required): Path to the drift report for "backup" related queries.
*   `legacy` (string, required): Path to the drift report for "legacy" related queries.
*   `overlap` (string, required): Path to the drift report for "overlap" related queries.
*   `drift` (string, required): Path to the drift report for "drift" related queries.

All triggers currently point to `packages/knowledge/docs/generated/drift_report.md`.

### `batch_size`

This parameter controls the number of documents processed in each batch during indexing.

*   `batch_size` (integer, required):  Currently set to `10`.  Adjusting this value can impact indexing performance.

### `drift_threshold`

This parameter defines the threshold for determining significant drift between documentation versions.

*   `drift_threshold` (float, required): Currently set to `0.85`. This value is used in the drift report generation process.

### `system_context`

This parameter provides the initial context given to the language model when answering questions.

*   `system_context` (string, required): A multi-line string that sets the role of the language model and provides information about the documentation structure and how to handle specific query types (overlap, backup, legacy, drift).  It instructs the model to prioritize the `drift_report.md` file when relevant queries are detected.