---
type: Documentation
domain: knowledge
origin: packages/knowledge/llm/client.py
last_modified: 2026-02-01
generated: true
source: packages/knowledge/llm/client.py
generated_at: 2026-02-01T19:35:10.862927
hash: 5c3004e9f42636ddf1fc589893d5cbcf16779f191ba71625688658e4b79db9f5
---

## GlassOps Knowledge Pipeline - LLM Client Documentation

This document details the functionality of the LLMClient, a component designed to provide a consistent interface for interacting with the Google Generative AI models within the GlassOps Knowledge Pipeline. It manages API calls, implements retry mechanisms for common errors, and enforces rate limiting to ensure responsible and reliable usage.

### Module Responsibilities

The `llm/client.py` module encapsulates all logic related to communicating with the Google Generative AI API. Its primary responsibilities include:

*   Establishing a connection to the LLM service.
*   Sending prompts to the LLM and receiving responses.
*   Handling potential errors during API calls, including retrying transient failures.
*   Managing request rates to avoid exceeding API limits (Rate Limiting).
*   Providing a simple and reusable interface for other components of the Knowledge Pipeline.

### Key Classes

#### `LLMClient`

This class is the central component of the module. It handles the interaction with the Google Generative AI API.

*   **Attributes:**
    *   `client`: An instance of the `genai.Client` object, representing the connection to the LLM service.  It is initialized to `None` if the API key is not found.
    *   `model`: A string specifying the name of the LLM model to use (default: "gemma-3-27b-it").
    *   `_request_history`: A list of dictionaries used to track recent requests for rate limiting purposes. Each dictionary contains the request timestamp and token count.
    *   `_rpm_limit`: An integer defining the requests per minute limit (default: 28).
    *   `_tpm_limit`: An integer defining the tokens per minute limit (default: 14000).

*   **Initialization (`__init__`)**:
    *   Loads the Google API key from the environment variables (using a `.env` file in the project root).
    *   Initializes the `genai.Client` if the API key is found. If the API key is missing, the client is set to `None`, effectively disabling the LLM functionality.
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

This is the primary function for generating text from a given prompt. It handles the API call, retry logic, and rate limiting.

*   **Parameters:**
    *   `prompt`: The text prompt to send to the LLM.
    *   `max_retries`: The maximum number of times to retry the request if a transient error occurs (default: 3).
    *   `temperature`: A value controlling the randomness of the generated text (default: 0.2). Lower values produce more deterministic output.
    *   `max_output_tokens`: The maximum number of tokens to generate in the response (default: 8192).
*   **Return Value:**
    *   The generated text string if the request is successful.
    *   `None` if the request fails after multiple retries or if the LLM client is not initialized (due to a missing API key).

### Type Hints

The code extensively uses type hints (e.g., `str`, `int`, `Optional[str]`) to improve code readability and maintainability. These hints specify the expected data types for function parameters and return values, allowing for static analysis and early detection of potential errors.

### Design Decisions and Patterns

*   **Retry Logic:** The `generate` function incorporates a retry mechanism with exponential backoff to handle transient errors such as network issues or temporary API overload.
*   **Rate Limiting:** The `_throttle` function implements a basic rate limiting strategy to prevent exceeding the API's usage limits. It tracks recent requests and pauses execution if necessary.
*   **Environment Variables:** The API key is loaded from an environment variable, promoting secure configuration management and preventing hardcoding of sensitive information.
*   **Error Handling:** The code includes robust error handling, logging informative messages when errors occur, and distinguishing between retryable and non-retryable errors.
*   **Defensive Programming:** The client checks for the presence of the API key and disables functionality gracefully if it's missing. It also includes checks for empty responses from the LLM.