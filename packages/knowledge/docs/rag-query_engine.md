---
type: Documentation
domain: knowledge
origin: packages/knowledge/rag/query_engine.py
last_modified: 2026-01-26
generated: true
source: packages/knowledge/rag/query_engine.py
generated_at: 2026-01-26T14:11:57.820Z
hash: 784f3cbb59399621add5034da54bb06994c7a91e466764d2c999dbf7de700422
---

# GlassOps Knowledge Retrieval and Query Engine

This document details the functionality of the GlassOps Knowledge Retrieval and Query Engine, a system designed to answer user questions based on a knowledge base of documents. It leverages vector embeddings, ChromaDB, and the Gemma large language model (LLM) to provide accurate and contextually relevant responses.

## Overview

The query engine operates in four primary stages:

1.  **Query Embedding:** Converts the user's question into a vector embedding, a numerical representation that captures the semantic meaning of the query.
2.  **Knowledge Retrieval:** Searches a vector database (ChromaDB) for documents with embeddings similar to the query embedding.  This identifies relevant knowledge chunks.
3.  **Context Construction:**  Assembles the retrieved documents into a coherent context.  Critically, this stage includes a mechanism to inject specific files (e.g., audit reports, drift analyses) based on keywords in the query, ensuring important information is always considered.
4.  **Answer Generation:**  Uses the Gemma LLM to generate a concise answer to the user's question, based solely on the constructed context.

## Components

*   **ChromaDB:** A vector database used to store and efficiently search document embeddings.  The index is stored locally in the `glassops_index` directory.
*   **`get_embeddings_for_docs`:** A function (defined elsewhere in the `knowledge.embeddings` package) responsible for generating vector embeddings from text.
*   **Gemma (Google Generative AI):** A large language model used to generate human-readable answers from the retrieved context.  Requires a `GOOGLE_API_KEY` environment variable to be set.
*   **Configuration File (`config.json`):**  A JSON file used to customize the system's behavior, including:
    *   `system_context`:  Provides the LLM with a persona and instructions.
    *   `retrieval_triggers`: Defines keywords and corresponding file paths. If a keyword is present in the query, the associated file's content is prepended to the context.

## Functionality

### `query_index(query, n_results=5)`

This function is the primary entry point for querying the knowledge base.

**Parameters:**

*   `query` (string): The user's question.
*   `n_results` (int, optional): The number of relevant documents to retrieve from ChromaDB. Defaults to 5.

**Return Value:**

*   (string): The generated answer to the query, including the sources used to formulate the response.  Returns error messages if issues occur during embedding, database access, or answer generation.

**Workflow:**

1.  **Embeds the query** using `get_embeddings_for_docs`.
2.  **Connects to the ChromaDB** instance located in the `glassops_index` directory.
3.  **Queries ChromaDB** for the `n_results` most similar documents based on the query embedding.
4.  **Constructs a context** from the retrieved documents.
5.  **Checks for and injects relevant files** based on keywords in the query and the `retrieval_triggers` configuration.
6.  **Generates an answer** using the Gemma LLM, providing the context and the query as input.
7.  **Returns the answer** along with a list of the source documents used.

## Error Handling

The system includes robust error handling:

*   **Embedding Errors:** Catches exceptions during embedding generation and returns an informative error message.
*   **Index Errors:** Checks for the existence of the ChromaDB index and returns an error if it's missing.
*   **API Key Errors:** Verifies that the `GOOGLE_API_KEY` environment variable is set before attempting to generate an answer.
*   **Configuration Errors:** Handles potential errors when loading the `config.json` file.
*   **Model Errors:**  Provides debugging information, including a list of available models, if the specified Gemma model fails to load or generate content.
*   **File Injection Errors:** Logs warnings if trigger files cannot be loaded or processed.

## Important Considerations

*   **`GOOGLE_API_KEY`:**  This environment variable *must* be set for the answer generation stage to function.
*   **`glassops_index`:** The ChromaDB index must be created before querying. This is typically done via a separate indexing process.
*   **`config.json`:**  Customizing this file allows for fine-tuning the system's behavior, including the LLM's persona and the injection of critical reports.
*   **Trigger Mechanism:** The file injection mechanism provides a powerful way to ensure that specific documents are always considered, even if their vector similarity to the query is low.