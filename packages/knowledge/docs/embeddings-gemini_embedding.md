---
type: Documentation
domain: knowledge
origin: packages/knowledge/embeddings/gemini_embedding.py
last_modified: 2026-01-31
generated: true
source: packages/knowledge/embeddings/gemini_embedding.py
generated_at: 2026-01-31T08:53:06.480312
hash: a0b56efa7a899a2d9e217b1e561575b4fc5fa0176d7c2575ebbb8eacc83645a3
---

## Gemini Embedding Module Documentation

This module provides a way to generate embeddings for text data using the Google Gemini models. Embeddings are numerical representations of text that capture its semantic meaning, enabling tasks like semantic search and similarity comparison.

**Module Responsibilities:**

The primary responsibility of this module is to interface with the Google Gemini API to produce embeddings from input text. It handles API key management, error handling, and provides a fallback mechanism for when the API is unavailable or encounters issues. It also includes a mock embedding generation capability for testing and development when the API key is not configured.

**GeminiEmbedding Class:**

The `GeminiEmbedding` class encapsulates the logic for interacting with the Gemini API.

*   **`__init__(self)`:**
    *   Initializes the `GeminiEmbedding` object.
    *   Retrieves the Google API key from the `GOOGLE_API_KEY` environment variable.
    *   If the API key is not found, it prints a warning message indicating that mock data will be returned.
    *   If the API key is present and the `google.generativeai` library is installed, it configures the Gemini API with the provided key.
*   **`get_embeddings(self, texts: list[str]) -> list[list[float]]`:**
    *   This is the core function of the class. It takes a list of strings (`texts`) as input and returns a list of embeddings, where each embedding is a list of floats.
    *   **API Interaction:** If a valid API key is available and the `google.generativeai` library is accessible, it attempts to generate embeddings using the Gemini API. It first tries a batched call to the API for efficiency. If the batched call fails, it falls back to processing each text sequentially.
    *   **Error Handling:**  Includes `try...except` blocks to catch potential exceptions during API calls. If an error occurs during sequential processing, it prints a warning message and generates a random vector as a placeholder to maintain the expected output structure.
    *   **Mock Data:** If the API key is not set, or if the `google.generativeai` library is not installed, it generates mock embeddings – lists of 768 random floating-point numbers – for each input text. This allows the application to function even without access to the Gemini API.
    *   **Type Hints:** The function signature `(self, texts: list[str]) -> list[list[float]]` uses type hints to clearly indicate that the input `texts` is expected to be a list of strings, and the output will be a list of lists of floats. This improves code readability and helps with static analysis.
    *   **Batching Logic:** The code attempts to use a batched embedding call (`genai.embed_content` with a list of texts) for improved performance. It handles potential variations in the API response format (single embedding vs. list of embeddings) to ensure compatibility with different Gemini SDK versions.

**Important Considerations:**

*   **API Key:** The `GOOGLE_API_KEY` environment variable must be set with a valid Google Gemini API key for the module to function correctly with the actual API.
*   **Dependencies:** The `google.generativeai` library must be installed (`pip install google-generativeai`) to use the Gemini API.
*   **Fallback Mechanism:** The sequential processing and mock data generation provide robustness in case of API unavailability or errors.
*   **Embedding Dimension:** The generated embeddings are 768-dimensional vectors. This is consistent with the `text-embedding-004` model.
*   **Warnings:** The code suppresses deprecation warnings from the `google.generativeai` and `google.auth` libraries to avoid cluttering the logs.
*   **Model Selection:** The code explicitly uses the "models/text-embedding-004" model. You can modify this to use other available Gemini embedding models if needed.