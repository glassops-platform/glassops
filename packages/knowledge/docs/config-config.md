---
type: Documentation
domain: knowledge
origin: packages/knowledge/config/config.json
last_modified: 2026-02-01
generated: true
source: packages/knowledge/config/config.json
generated_at: 2026-02-01T19:27:27.114812
hash: 20fe2a53c1392c0cfc0142b90c8d3cc228fc21039203bc303e72df5b61d28dde
---

# Knowledge Configuration

This document details the configuration options for the knowledge retrieval system. This system powers intelligent responses within the GlassOps platform by indexing documentation and providing context to language models.

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

This array defines the file paths and patterns to be included in the knowledge base. The system recursively searches these locations for documentation files.

*   `federated_doc_paths` (array of strings, required): A list of paths to documentation sources.
    *   `docs/`: Includes all files within the `docs/` directory.
    *   `packages/**/adr`: Includes all files with the `.adr` extension within any `packages/` subdirectory.
    *   `packages/**/docs`: Includes all files within any `packages/` subdirectory's `docs/` directory.

### `retrieval_triggers`

This section maps specific query types (triggers) to a specific documentation file. This allows for targeted retrieval of information related to specific system events or reports.

*   `audit` (string, required): Path to the documentation file for audit-related queries. Currently points to `packages/knowledge/docs/generated/drift_report.md`.
*   `backup` (string, required): Path to the documentation file for backup-related queries. Currently points to `packages/knowledge/docs/generated/drift_report.md`.
*   `legacy` (string, required): Path to the documentation file for legacy-related queries. Currently points to `packages/knowledge/docs/generated/drift_report.md`.
*   `overlap` (string, required): Path to the documentation file for overlap-related queries. Currently points to `packages/knowledge/docs/generated/drift_report.md`.
*   `drift` (string, required): Path to the documentation file for drift-related queries. Currently points to `packages/knowledge/docs/generated/drift_report.md`.

### `batch_size`

This parameter controls the number of documents processed in each batch during indexing.

*   `batch_size` (integer, required): Specifies the batch size. Currently set to `10`.  Adjusting this value can impact performance.

### `drift_threshold`

This parameter defines the similarity score threshold used to identify significant changes (drift) between documentation versions.

*   `drift_threshold` (float, required): Specifies the drift threshold. Currently set to `0.85`.  Values closer to 1 indicate higher similarity.

### `system_context`

This string provides the language model with contextual information about the knowledge base and how to respond to specific queries.

*   `system_context` (string, required): A multi-line string containing instructions for the language model.  This context guides the model's responses, particularly when handling queries related to documentation drift, overlap, backup, and legacy content.  It instructs the model to prioritize information from the `drift_report.md` file when relevant.