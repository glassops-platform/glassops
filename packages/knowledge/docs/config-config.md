---
type: Documentation
domain: knowledge
last_modified: 2026-03-08
generated: true
source: packages/knowledge/config/config.json
generated_at: 2026-03-08T22:44:43.964088
hash: aaff9e175e11e0dded44338711bfcb4ee33643c9da856b153889339a4ccf21d3
---

# Knowledge Configuration

This document details the configuration options for the knowledge retrieval system. This system powers intelligent responses by indexing documentation and related files, and providing relevant context to language models. We maintain this configuration to control the behavior of the knowledge base.

## Embedding Models

This section defines the models used to create vector embeddings from text. These embeddings are used for semantic search.

*   **`primary`** (string, *required*): Specifies the primary embedding model. Currently set to `"gemini-embedding-1.0"`. This model is preferred for generating embeddings.
*   **`fallback`** (string, *required*): Specifies a fallback embedding model. Currently set to `"gemma-3-12b-it"`. This model is used if the primary model is unavailable or encounters an error.

## Vector Store

This section configures the vector database used to store and retrieve embeddings.

*   **`type`** (string, *required*): Specifies the type of vector store. Currently set to `"chroma"`.
*   **`persist_dir`** (string, *required*): Specifies the directory where the vector store will persist data. Currently set to `"glassops_index"`.

## Federated Document Paths

This section defines the file paths to be included in the knowledge base. The system recursively searches these paths for documentation.

*   **`federated_doc_paths`** (array of strings, *required*): A list of glob patterns representing the paths to documentation.
    *   `"docs/"`: Includes all files in the `docs/` directory.
    *   `"packages/**/adr"`: Includes all files with the `.adr` extension within any `packages/` subdirectory.
    *   `"packages/**/docs"`: Includes all files within `docs` directories inside any `packages/` subdirectory.

## Retrieval Triggers

This section maps specific query types to a designated document for providing context.

*   **`retrieval_triggers`** (object, *required*): A mapping of trigger keywords to a specific documentation file.
    *   `"audit"` (string, *required*): Points to `"packages/knowledge/docs/generated/drift_report.md"`.
    *   `"backup"` (string, *required*): Points to `"packages/knowledge/docs/generated/drift_report.md"`.
    *   `"legacy"` (string, *required*): Points to `"packages/knowledge/docs/generated/drift_report.md"`.
    *   `"overlap"` (string, *required*): Points to `"packages/knowledge/docs/generated/drift_report.md"`.
    *   `"drift"` (string, *required*): Points to `"packages/knowledge/docs/generated/drift_report.md"`.

## Batch Processing and Drift Detection

These settings control the batch size for processing documents and the threshold for drift detection.

*   **`batch_size`** (integer, *required*): Specifies the number of documents to process in each batch. Currently set to `10`.
*   **`drift_threshold`** (number, *required*): Specifies the threshold for identifying significant changes (drift) between documentation versions. Currently set to `0.85`.

## System Context

This section defines the context provided to the language model to guide its responses.

*   **`system_context`** (string, *required*): A multi-line string containing instructions and information about the documentation structure and how to handle specific query types. This context informs the language model about the purpose of the documentation and how to respond to user inquiries related to "overlap", "backup", "legacy", and "drift". It also instructs the model to refer to the `drift_report.md` file when appropriate.