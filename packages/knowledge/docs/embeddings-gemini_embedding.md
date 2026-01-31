---
type: Documentation
domain: knowledge
origin: packages/knowledge/embeddings/gemini_embedding.py
last_modified: 2026-01-31
generated: true
source: packages/knowledge/embeddings/gemini_embedding.py
generated_at: 2026-01-31T09:48:18.397170
hash: a0b56efa7a899a2d9e217b1e561575b4fc5fa0176d7c2575ebbb8eacc83645a3
---

## Gemini Embedding Module Documentation

This module provides a way to generate embeddings for text using Google’s Gemini models. Embeddings are numerical representations of text that capture its semantic meaning, enabling tasks like semantic search and similarity comparison. The module handles both batched and sequential embedding requests, and includes a fallback mechanism to provide mock data when the API key is missing or when errors occur during API calls.

**Key Classes:**

*   **`GeminiEmbedding`**: This class encapsulates the logic for interacting with the Gemini embedding API.
    *   **`__init__(self)`**: The constructor initializes the `GeminiEmbedding` object. It attempts to retrieve the Google API key from the environment variable `GOOGLE_API_KEY`. If the API key is not found, a warning message is printed, and the class will return mock embeddings. If the API key is present and the `google.generativeai` library is available, it configures the library with the provided API key.

**Important Functions:**

*   **`get_embeddings(self, texts: list[str]) -> list[list[float]]`**: This function takes a list of strings (`texts`) as input and returns a list of embeddings, where each embedding is a list of floats.
    *   **Input:** `texts`: A list of strings to be embedded. The type hint `list[str]` specifies that the input must be a list of strings.
    *   **Output:** A list of lists of floats, representing the embeddings for each input text. The type hint `list[list[float]]` indicates this structure.
    *   **Behavior:**
        1.  **API Key Check:** First, it checks if the `GOOGLE_API_KEY` is set and the `genai` library is available. If not, it proceeds to the fallback mechanism.
        2.  **Batched Embedding Attempt:** It attempts to generate embeddings for all input texts in a single batch using `genai.embed_content`. This is the preferred method for performance. The `task_type` is set to "retrieval\_document".
        3.  **Result Handling:** The function handles different possible return structures from the API. It checks if the result contains an 'embedding' key and verifies the structure of the data within. It accommodates cases where the API might return a single embedding (if a single string was input) or a list of embeddings (if a list of strings was input).
        4.  **Sequential Embedding Fallback:** If the batched embedding attempt fails (due to an exception), the function falls back to processing each text individually in a loop. This provides resilience against API errors.  Error handling within the loop prints a warning message and appends a random vector to maintain alignment if a specific chunk fails to embed.
        5.  **Mock Embedding Fallback:** If the `GOOGLE_API_KEY` is not set or the `genai` library is unavailable, the function generates mock embeddings – lists of 768 random floats – for each input text. This allows the application to function even without access to the Gemini API.

**Type Hints:**

The code uses type hints (e.g., `texts: list[str]`, `-> list[list[float]]`) to improve code readability and maintainability. Type hints specify the expected data types for function arguments and return values, enabling static analysis tools to detect potential type errors.

**Design Decisions and Patterns:**

*   **Fallback Mechanism:** The module implements a robust fallback mechanism to handle API key absence, API errors, and library import failures. This ensures that the application can continue to function, albeit with mock data, even in adverse conditions.
*   **Batched vs. Sequential Processing:** The module prioritizes batched embedding requests for performance but falls back to sequential processing when necessary.
*   **Error Handling:** The code includes `try...except` blocks to catch potential exceptions during API calls and provides informative warning messages.
*   **Environment Variable Configuration:** The API key is loaded from an environment variable (`GOOGLE_API_KEY`), which is a secure and flexible way to manage sensitive credentials.
*   **Warning Suppression:** The code suppresses specific warnings from the `google.generativeai` and `google.auth` libraries to reduce noise in the logs.