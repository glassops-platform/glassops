---
type: Documentation
domain: knowledge
origin: packages/knowledge/llm/client.py
last_modified: 2026-01-31
generated: true
source: packages/knowledge/llm/client.py
generated_at: 2026-01-31T09:00:09.906275
hash: 5c3004e9f42636ddf1fc589893d5cbcf16779f191ba71625688658e4b79db9f5
---

## GlassOps Knowledge Pipeline - LLM Client Documentation

This document details the functionality of the LLM Client, a component designed for interacting with the Google Generative AI models within the GlassOps Knowledge Pipeline. It provides a reusable interface for sending prompts to the model, handling potential errors through retry logic, and managing API request rates to avoid exceeding limits.

**Module Responsibilities:**

The primary responsibility of this module is to abstract the complexities of interacting with the Google Generative AI API. It handles authentication, request formatting, error handling, and rate limiting, providing a simplified interface for other components of the Knowledge Pipeline.

**Key Classes:**

*   **`LLMClient`**: This class serves as the central point of interaction with the Google Generative AI service. It encapsulates the API client, manages request history for rate limiting, and provides a `generate` method for submitting prompts and receiving responses.

    *   **`__init__(self, model: str = "gemma-3-27b-it")`**:  The constructor initializes the LLMClient. It retrieves the Google API key from the environment, creates a `genai.Client` instance if the key is found, and sets the default model to "gemma-3-27b-it". If the API key is missing, the client is disabled, and a warning message is printed. It also initializes internal data structures for request history and rate limiting. The `model` parameter allows you to specify which Google Generative AI model to use.
    *   `_request_history: list[dict]`: A private list used to store information about recent requests, including their timestamps and token counts, for rate limiting purposes.
    *   `_rpm_limit = 28`: A private integer defining the requests per minute limit.
    *   `_tpm_limit = 14000`: A private integer defining the tokens per minute limit.

**Important Functions:**

*   **`_estimate_tokens(self, text: str) -> int`**: This private function provides a rough estimate of the number of tokens in a given text string. It uses a simple heuristic of 4 characters per token. This estimation is used for rate limiting.
*   **`_throttle(self, estimated_tokens: int) -> None`**: This private function implements rate limiting logic. It checks the number of requests made within the last minute (RPM) and the total number of tokens used within the last minute (TPM) against predefined limits. If the limits are approaching, it pauses execution until sufficient headroom is available.
*   **`generate(self, prompt: str, max_retries: int = 3, temperature: float = 0.2, max_output_tokens: int = 8192) -> Optional[str]`**: This is the core function for generating text from a given prompt. It takes the prompt as input, along with optional parameters for controlling the generation process (maximum retries, temperature, and maximum output tokens). It handles rate limiting before making the API call. It includes retry logic for transient errors (HTTP 429 and 503 errors, or "overloaded" messages). The function returns the generated text if successful, or `None` if it fails after multiple retries.

    *   `prompt: str`: The input text prompt for the LLM.
    *   `max_retries: int`: The maximum number of times to retry the request if it fails due to a transient error.
    *   `temperature: float`: Controls the randomness of the generated text. Lower values produce more predictable output.
    *   `max_output_tokens: int`: Limits the length of the generated text.

**Type Hints:**

The code extensively uses type hints (e.g., `str`, `int`, `Optional[str]`) to improve code readability and maintainability. These hints specify the expected data types for function arguments and return values, allowing for static analysis and early detection of potential errors.

**Design Decisions and Patterns:**

*   **Retry Logic:** The `generate` function incorporates a retry mechanism to handle transient errors, improving the robustness of the client.
*   **Rate Limiting:** The `_throttle` function implements rate limiting to prevent exceeding the API usage limits, ensuring stable operation.
*   **Configuration via Environment Variables:** The API key is loaded from an environment variable (`GOOGLE_API_KEY`), promoting secure configuration management.
*   **Error Handling:** The code includes comprehensive error handling, logging informative messages when errors occur, and distinguishing between retryable and non-retryable errors.
*   **Token Estimation:** A simple token estimation method is used to proactively manage token usage for rate limiting.
*   **Safety Buffers:** The RPM and TPM limits are set with safety buffers to avoid accidental exceeding of the actual API limits.