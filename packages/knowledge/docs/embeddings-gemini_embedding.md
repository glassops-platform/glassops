---
type: Documentation
domain: knowledge
last_modified: 2026-02-02
generated: true
source: packages/knowledge/embeddings/gemini_embedding.py
generated_at: 2026-02-02T22:28:15.403853
hash: f4e72660eea1301307f4dc67667b1a3510684cdde0de250c21941755d2af5edd
---

## GeminiEmbedding Documentation

This module provides a class for generating text embeddings using the Gemini models offered through the Google Generative AI API. Embeddings are vector representations of text, useful for semantic search, clustering, and other machine learning tasks.

**Module Responsibilities:**

The primary responsibility of this module is to encapsulate the logic for interacting with the Gemini embedding models. It handles API key management, error handling, and provides a consistent interface for obtaining embeddings from text data. It includes a fallback mechanism to sequential processing and mock data generation if the API is unavailable or encounters issues.

**GeminiEmbedding Class:**

The `GeminiEmbedding` class is the core component of this module.

*   **`__init__(self)`:**
    *   Initializes the `GeminiEmbedding` object.
    *   Retrieves the Google API key from the `GOOGLE_API_KEY` environment variable.
    *   If the API key is not found, a warning message is printed, and the class will return mock embeddings.
    *   If the API key is present and the `google.generativeai` library is imported successfully, it configures the Gemini API with the provided key.
    *   Type hints are not used for self, as is standard practice.

*   **`get_embeddings(self, texts: list[str]) -> list[list[float]]`:**
    *   This function takes a list of strings (`texts`) as input and returns a list of embeddings, where each embedding is a list of floats.
    *   It first attempts a batched call to the Gemini API to generate embeddings for all input texts simultaneously. This is the preferred method for performance.
    *   It handles potential `Exception`s during the batched call. If a batched call fails, it falls back to sequential processing.
    *   If the API key is not set or the `google.generativeai` library is not available, it generates mock embeddings (random 768-dimensional vectors) for each input text.
    *   The function uses the `models/text-embedding-004` model for generating embeddings.
    *   The `task_type` is set to "retrieval\_document".
    *   The return type is explicitly annotated as `list[list[float]]`, indicating a list of lists of floating-point numbers, representing the embeddings.
    *   The function includes logic to handle different response structures from the API, ensuring compatibility with both older and newer versions of the `google.generativeai` SDK. It checks if the response contains an 'embedding' key and verifies the structure of the returned data.

**Design Decisions and Patterns:**

*   **Environment Variable for API Key:** The API key is loaded from an environment variable (`GOOGLE_API_KEY`) to avoid hardcoding sensitive information in the code.
*   **Fallback Mechanism:** The code includes a fallback mechanism to sequential processing and mock data generation to ensure robustness in case of API errors or unavailability.
*   **Error Handling:**  The code includes `try...except` blocks to catch potential exceptions during API calls and handle them gracefully.  Errors during sequential embedding are logged, and a random vector is returned to maintain alignment with the input list length.
*   **Type Hints:** Type hints (`list[str]`, `list[list[float]]`) are used to improve code readability and maintainability, and to enable static analysis.
*   **Warning Suppression:**  The code suppresses `FutureWarning`s from the `google.generativeai` and `google.auth` modules to avoid noisy output.
*   **Batching:** The code prioritizes batched embedding calls for improved performance.
*   **Mock Data:** When the API is unavailable, the code generates mock embeddings to allow for testing and development without requiring an active API connection.