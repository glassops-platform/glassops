---
type: Documentation
domain: knowledge
origin: packages/knowledge/llm/client.py
last_modified: 2026-01-31
generated: true
source: packages/knowledge/llm/client.py
generated_at: 2026-01-31T09:55:45.827027
hash: 5c3004e9f42636ddf1fc589893d5cbcf16779f191ba71625688658e4b79db9f5
---

## GlassOps Knowledge Pipeline - LLM Client Documentation

This document details the functionality of the LLMClient, a component designed to provide a consistent interface for interacting with the Google Generative AI models within the GlassOps Knowledge Pipeline. It manages API calls, implements retry mechanisms for common errors, and enforces rate limiting to ensure responsible and reliable usage.

### Module Responsibilities

The `llm/client.py` module encapsulates all interactions with the Google Generative AI API. Its primary responsibilities include:

*   Establishing a connection to the LLM service.
*   Sending prompts to the LLM and receiving responses.
*   Handling potential errors during API calls, including transient issues like rate limits and server errors.
*   Implementing rate limiting to prevent exceeding API usage quotas.
*   Providing a simple and reusable interface for other components of the Knowledge Pipeline.

### Key Classes

#### `LLMClient`

This class is the core of the module. It manages the connection to the LLM and provides the `generate` method for requesting content.

*   **Attributes:**
    *   `client`: An instance of the `genai.Client` object, representing the connection to the Google Generative AI API.  It is initialized to `None` if the API key is not found.
    *   `model`: A string specifying the name of the LLM model to use (default: "gemma-3-27b-it").
    *   `_request_history`: A list of dictionaries used to track recent requests for rate limiting purposes. Each dictionary contains the request timestamp and token count.
    *   `_rpm_limit`: An integer defining the requests per minute limit (default: 28).
    *   `_tpm_limit`: An integer defining the tokens per minute limit (default: 14000).

*   **Initialization (`__init__`)**:
    *   Loads the Google API key from the environment variables (specifically, `GOOGLE_API_KEY`).
    *   Initializes the `genai.Client` if the API key is found. If the key is missing, it prints a warning and disables the client.
    *   Sets the default model name.
    *   Initializes the request history list.
    *   Sets the rate limits for requests per minute and tokens per minute.

### Important Functions

#### `_estimate_tokens(text: str) -> int`

This private function provides a rough estimate of the number of tokens in a given text string. It assumes approximately 4 characters per token. This estimation is used for rate limiting.

*   **Parameters:**
    *   `text`: The input string.
*   **Return Value:** An integer representing the estimated token count.

#### `_throttle(estimated_tokens: int) -> None`

This private function implements the rate limiting logic. It checks if the current request would exceed the defined RPM or TPM limits. If a limit is approaching, it pauses execution until sufficient headroom is available.

*   **Parameters:**
    *   `estimated_tokens`: The estimated number of tokens for the current request.
*   **Return Value:** None.

#### `generate(prompt: str, max_retries: int = 3, temperature: float = 0.2, max_output_tokens: int = 8192) -> Optional[str]`

This is the primary function for interacting with the LLM. It sends a prompt to the model and returns the generated text.

*   **Parameters:**
    *   `prompt`: The text prompt to send to the LLM.
    *   `max_retries`: The maximum number of times to retry the request if a transient error occurs (default: 3).
    *   `temperature`: A value controlling the randomness of the generated text (default: 0.2). Lower values produce more deterministic output.
    *   `max_output_tokens`: The maximum number of tokens to generate in the response (default: 8192).
*   **Return Value:**
    *   The generated text string if the request is successful.
    *   `None` if the request fails after multiple retries or if the LLM client is not initialized.

### Design Decisions and Patterns

*   **Retry Logic:** The `generate` function incorporates a retry mechanism with exponential backoff to handle transient errors (429 and 503 errors, or errors containing "overloaded"). This improves the robustness of the client.
*   **Rate Limiting:** The `_throttle` function enforces rate limits to prevent exceeding API quotas. It tracks requests and tokens within a sliding window (60 seconds) and pauses execution if necessary.
*   **Type Hints:** The code uses type hints (e.g., `str`, `int`, `Optional[str]`) to improve code readability and maintainability. These hints help clarify the expected data types for function parameters and return values.
*   **Environment Variables:** The API key is loaded from an environment variable (`GOOGLE_API_KEY`), which promotes security and allows for easy configuration without modifying the code.
*   **Error Handling:** The code includes comprehensive error handling, logging informative messages when errors occur, and distinguishing between retryable and non-retryable errors.