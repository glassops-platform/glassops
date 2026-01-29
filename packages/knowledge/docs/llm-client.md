---
type: Documentation
domain: knowledge
origin: packages/knowledge/llm/client.py
last_modified: 2026-01-28
generated: true
source: packages/knowledge/llm/client.py
generated_at: 2026-01-28T22:46:51.975239
hash: 3667a6526a67a319a53c0b2c2c225b58b5079367f817698164710ab84f22aa81
---

## GlassOps Knowledge Pipeline - LLM Client Documentation

This document details the functionality of the LLM Client module, designed to provide a consistent and reliable interface for interacting with the Google Generative AI models within the GlassOps Knowledge Pipeline. It handles API communication, implements retry mechanisms for common errors, and manages request rates to avoid exceeding API limits.

**Module Responsibilities:**

The primary responsibility of this module is to abstract the complexities of interacting with the Google Generative AI API. It provides a simple `generate` function for obtaining text completions from a given prompt, while internally managing error handling, retries, and rate limiting.  This ensures robust and predictable behavior when querying the LLM.

**Key Classes:**

*   **`LLMClient`**: This class encapsulates all interactions with the Google Generative AI service.
    *   **`__init__(self, model: str = "gemma-3-27b-it")`**:  The constructor initializes the client with the specified model name (defaulting to "gemma-3-27b-it"). It retrieves the Google API key from the environment variables and validates its presence. It also initializes internal data structures for tracking request history and managing rate limits.
    *   **`_request_history: list[dict]`**: A private list used to store information about recent requests, including their timestamps and token counts. This is used for rate limiting.
    *   **`_rpm_limit: int = 28`**:  A private attribute defining the requests per minute limit. A safety buffer is applied to stay below the actual API limit of 30.
    *   **`_tpm_limit: int = 14000`**: A private attribute defining the tokens per minute limit. A safety buffer is applied to stay below the actual API limit of 15000.

**Important Functions:**

*   **`_estimate_tokens(self, text: str) -> int`**: This private function provides a rough estimate of the number of tokens in a given text string. It uses a simple heuristic of 4 characters per token. This estimation is used for rate limiting purposes.
*   **`_throttle(self, estimated_tokens: int) -> None`**: This private function implements the rate limiting logic. It checks if the current request would exceed the defined RPM or TPM limits. If a limit is approaching, it pauses execution until sufficient headroom is available. It maintains a history of recent requests to track token usage.
*   **`generate(self, prompt: str, max_retries: int = 3, temperature: float = 0.2, max_output_tokens: int = 8192) -> Optional[str]`**: This is the primary function for generating text from a prompt.
    *   **`prompt: str`**: The input text prompt for the LLM.
    *   **`max_retries: int = 3`**:  The maximum number of times to retry the request if a transient error occurs.
    *   **`temperature: float = 0.2`**:  A parameter controlling the randomness of the generated text. Lower values result in more deterministic outputs.
    *   **`max_output_tokens: int = 8192`**: The maximum number of tokens to generate in the response.
    *   **Return Value**: Returns the generated text as a string if successful, or `None` if the request fails after multiple retries.

**Type Hints:**

The code extensively uses type hints (e.g., `str`, `int`, `Optional[str]`) to improve code readability and maintainability. These hints clarify the expected data types for function arguments and return values, aiding in static analysis and error detection.

**Design Decisions and Patterns:**

*   **Retry Logic:** The `generate` function incorporates a retry mechanism with exponential backoff to handle transient errors such as rate limits (429 errors) and service unavailability (503 errors).
*   **Rate Limiting:** The `_throttle` function implements a basic rate limiting strategy to prevent exceeding the Google Generative AI API limits. It tracks token usage over a sliding window and pauses execution when necessary.
*   **Environment Variables:** The API key is loaded from an environment variable (`GOOGLE_API_KEY`), promoting secure configuration management and avoiding hardcoding sensitive information.
*   **Error Handling:** The code includes robust error handling, logging informative messages when errors occur, and distinguishing between retryable and non-retryable errors.
*   **Token Estimation:** A simple token estimation method is used to approximate the number of tokens in the prompt and generated response, which is essential for accurate rate limiting.