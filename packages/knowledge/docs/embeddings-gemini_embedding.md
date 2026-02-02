---
type: Documentation
domain: knowledge
origin: packages/knowledge/embeddings/gemini_embedding.py
last_modified: 2026-02-01
generated: true
source: packages/knowledge/embeddings/gemini_embedding.py
generated_at: 2026-02-01T19:28:50.826485
hash: a0b56efa7a899a2d9e217b1e561575b4fc5fa0176d7c2575ebbb8eacc83645a3
---

## GeminiEmbedding Documentation

This module provides a class for generating text embeddings using the Google Gemini API. Embeddings are numerical representations of text that capture semantic meaning, enabling tasks like semantic search and similarity comparison.

**Module Responsibilities:**

The primary responsibility of this module is to encapsulate the interaction with the Gemini API for generating embeddings. It handles API key management, error handling, and provides a fallback mechanism for when the API is unavailable or encounters issues. It also includes a mock embedding generation capability for testing and development when the API key is not configured.

**GeminiEmbedding Class:**

The `GeminiEmbedding` class is the core component of this module.

*   **`__init__(self)`:**
    *   Initializes the `GeminiEmbedding` object.
    *   Retrieves the Google API key from the `GOOGLE_API_KEY` environment variable.
    *   If the API key is not found, it prints a warning message indicating that mock data will be returned.
    *   If the API key is found and the `google.generativeai` library is available, it configures the Gemini API with the provided key.
    *   Type hints: None
*   **`get_embeddings(self, texts: list[str]) -> list[list[float]]`:**
    *   Generates embeddings for a list of input texts.
    *   **Parameters:**
        *   `texts` (list[str]): A list of strings for which to generate embeddings.
    *   **Return Value:**
        *   list[list[float]]: A list of embeddings, where each embedding is a list of floats. Each inner list represents the embedding vector for the corresponding text in the input list.
    *   **Behavior:**
        1.  First, it attempts a batched call to the Gemini API using `genai.embed_content` with the `texts` list directly. This is the preferred method for efficiency.
        2.  If the batched call fails (due to exceptions or API behavior), it falls back to sequential processing, generating embeddings for each text individually within a loop. This ensures that even if some texts fail to embed, the process continues for the remaining texts. A random vector is returned for any text that fails to embed during sequential processing to maintain alignment with the input list length.
        3.  If the API key is not set or the `google.generativeai` library is not available, it generates mock embeddings – random 768-dimensional vectors – for each input text. This allows the application to function even without access to the Gemini API.
        4.  The function handles potential variations in the API response format, ensuring that the returned data is always a list of lists of floats, representing the embeddings.
    *   Type hints: `texts` is explicitly typed as `list[str]`, and the return value is `list[list[float]]`, ensuring type safety and clarity.

**Design Decisions and Patterns:**

*   **Fallback Mechanism:** The implementation includes a robust fallback mechanism to handle API failures or unavailability. This ensures that the application remains functional even in adverse conditions.
*   **Mock Data:** The provision of mock data allows for testing and development without requiring access to the Gemini API.
*   **Error Handling:** The code includes `try...except` blocks to catch potential exceptions during API calls and handle them gracefully.
*   **API Key Management:** The API key is retrieved from an environment variable, promoting security and configuration flexibility.
*   **Batched vs. Sequential Processing:** The code prioritizes batched API calls for efficiency but falls back to sequential processing when necessary.
*   **Type Hints:** The use of type hints enhances code readability and maintainability, and helps to prevent errors.
*   **Warning Suppression:** The code suppresses specific warnings from the `google.generativeai` and `google.auth` libraries to reduce noise in the logs.