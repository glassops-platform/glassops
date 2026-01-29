---
type: Documentation
domain: knowledge
origin: packages/knowledge/embeddings/gemini_embedding.py
last_modified: 2026-01-28
generated: true
source: packages/knowledge/embeddings/gemini_embedding.py
generated_at: 2026-01-28T22:39:34.472712
hash: 6759afc434c64427b7dfaac26c535d70358d0d1d4939e3e013d723cf076f01f8
---

## Gemini Embedding Module Documentation

This module provides functionality to generate embeddings for text data using Google’s Gemini models. Embeddings are numerical representations of text, useful for tasks like semantic search, clustering, and similarity comparisons.

**Module Responsibilities:**

The primary responsibility of this module is to interface with the Gemini API to produce embeddings from input text. It handles API key management, error handling, and provides fallback mechanisms to ensure functionality even when the API is unavailable or encounters issues.

**Key Classes:**

*   **GeminiEmbedding:** This class encapsulates the logic for interacting with the Gemini embedding API.
    *   **`__init__(self)`:** The constructor initializes the Gemini client. It retrieves the Google API key from the environment variable `GOOGLE_API_KEY`. If the API key is not found, it prints a warning and sets the client to `None`, which triggers mock data generation. If the API key is present, it creates a `genai.Client` instance.
    *   **`get_embeddings(self, texts: list[str]) -> list[list[float]]`:** This method takes a list of strings as input and returns a list of embeddings, where each embedding is a list of floats. The length of each embedding vector is 768.

**Important Functions:**

*   **`get_embeddings(self, texts: list[str]) -> list[list[float]]`:**
    1.  **API Client Check:** First, it verifies if the Gemini client is initialized. If not (due to a missing API key), it proceeds to generate mock embeddings.
    2.  **Batch Embedding Attempt:** If the client exists, it attempts to generate embeddings for all input texts in a single batch request to the Gemini API using the `text-embedding-004` model and setting the `task_type` to `RETRIEVAL_DOCUMENT`.
    3.  **Batch Embedding Error Handling:** If the batch request fails (e.g., due to network issues or API errors), it catches the exception, prints a warning, and falls back to sequential embedding.
    4.  **Sequential Embedding:** If batch embedding fails or is not possible, the method iterates through the input texts, generating an embedding for each text individually. It includes error handling for each individual embedding request, falling back to random vector generation if an error occurs.
    5.  **Mock Embedding Generation:** If the Gemini client is not initialized (no API key), the method generates random 768-dimensional vectors for each input text as a placeholder.
    6.  **Return Value:** The method returns a list of embeddings. Each embedding is a list of 768 floating-point numbers.

**Type Hints:**

The code uses type hints to improve readability and maintainability. For example:

*   `texts: list[str]` indicates that the `texts` parameter of the `get_embeddings` method is expected to be a list of strings.
*   `-> list[list[float]]` indicates that the `get_embeddings` method is expected to return a list of lists of floats.

These hints help with static analysis and can prevent type-related errors.

**Design Decisions and Patterns:**

*   **Fallback Mechanisms:** The code implements multiple fallback mechanisms to ensure robustness. It first attempts batch embedding, then falls back to sequential embedding, and finally to mock data generation. This ensures that the module can still function even if the API is unavailable or encounters issues.
*   **Error Handling:** The code includes error handling to catch exceptions during API calls. This prevents the program from crashing and provides informative error messages.
*   **API Key Management:** The API key is retrieved from an environment variable, which is a secure way to manage sensitive credentials.
*   **Configuration:** The `EmbedContentConfig` object is used to configure the embedding request, allowing you to specify the task type.
*   **768-Dimensional Embeddings:** The code consistently generates 768-dimensional embedding vectors, which is the standard output dimension for the `text-embedding-004` model. You should be aware of this dimension when using the embeddings in downstream tasks.