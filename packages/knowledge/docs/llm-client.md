---
type: Documentation
domain: knowledge
last_modified: 2026-02-02
generated: true
source: packages/knowledge/llm/client.py
generated_at: 2026-02-02T22:31:19.187907
hash: 473c09a4e3a5d67b47cc170370f8d926e5f2acdaf9cda945edbe96e7665bb578
---

## GlassOps Knowledge Pipeline: LLM Client Documentation

This document details the functionality of the LLM Client module, designed to provide a consistent interface for interacting with large language models (LLMs), specifically Google’s Generative AI models. It handles API communication, implements retry mechanisms for resilience, and incorporates rate limiting to ensure responsible usage.

### Module Responsibilities

The primary responsibility of this module is to abstract the complexities of interacting with the Google Generative AI API. It provides a simple `generate` function for obtaining text completions from a given prompt, while managing potential issues like temporary API errors and rate limits.  The module is designed to be reusable across different components of the GlassOps Knowledge Pipeline.

### Key Classes

#### `LLMClient`

This class encapsulates the logic for interacting with the LLM.

*   **Purpose:**  Provides a centralized point for making requests to the LLM, handling retries, and enforcing rate limits.
*   **Initialization (`__init__`)**:
    *   `model: str = "gemma-3-27b-it"`: Specifies the LLM model to use. Defaults to "gemma-3-27b-it".
    *   Loads the Google API key from the environment (using a `.env` file in the project root). If the key is not found, the client is disabled.
    *   Initializes the `genai.Client` object if the API key is valid.
    *   Initializes internal data structures for request history (`_request_history`), RPM limit (`_rpm_limit`), and TPM limit (`_tpm_limit`).
*   **Attributes:**
    *   `client`: An instance of `genai.Client` if the API key is valid, otherwise `None`.
    *   `model`: The name of the LLM model being used.
    *   `_request_history`: A list of dictionaries storing the timestamp and token count of recent requests, used for rate limiting.
    *   `_rpm_limit`: The maximum number of requests per minute allowed (set to 28 as a safety buffer).
    *   `_tpm_limit`: The maximum number of tokens processed per minute allowed (set to 14000 as a safety buffer).

### Important Functions

#### `_estimate_tokens(text: str) -> int`

*   **Purpose:** Provides a rough estimate of the number of tokens in a given text string.
*   **Arguments:**
    *   `text: str`: The input text string.
*   **Return Value:** An integer representing the estimated token count.  The estimation is based on a simple rule of 4 characters per token.

#### `_throttle(estimated_tokens: int) -> None`

*   **Purpose:** Implements rate limiting to prevent exceeding the API's RPM and TPM limits.
*   **Arguments:**
    *   `estimated_tokens: int`: The estimated number of tokens for the upcoming request.
*   **Behavior:**
    *   Maintains a history of recent requests and their token counts.
    *   Checks if the current request would exceed the RPM or TPM limits.
    *   If a limit is approaching, it pauses execution using `time.sleep()` until sufficient headroom is available.
    *   Updates the request history with the current request's information.

#### `generate(prompt: str, max_retries: int = 3, temperature: float = 0.2, max_output_tokens: int = 8192) -> Optional[str]`

*   **Purpose:** Generates text content from a given prompt using the configured LLM.
*   **Arguments:**
    *   `prompt: str`: The input prompt for the LLM.
    *   `max_retries: int = 3`: The maximum number of times to retry the request if a transient error occurs.
    *   `temperature: float = 0.2`: Controls the randomness of the generated text (lower values are more deterministic).
    *   `max_output_tokens: int = 8192`: The maximum number of tokens to generate in the response.
*   **Return Value:**
    *   `str`: The generated text content if the request is successful.
    *   `None`: If the request fails after multiple retries or if the API key is not configured.
*   **Behavior:**
    *   Estimates the token count of the prompt and output.
    *   Calls the `_throttle` function to ensure rate limits are respected.
    *   Implements a retry loop with exponential backoff for transient errors (429, 503, or "overloaded").
    *   Sends the prompt to the LLM using `self.client.models.generate_content()`.
    *   Handles potential exceptions during the API call and logs errors.
    *   Returns the generated text if successful, or `None` if all retries fail.

### Type Hints

The code extensively uses type hints (e.g., `str`, `int`, `Optional[str]`) to improve code readability and maintainability. These hints clarify the expected data types for function arguments and return values, aiding in static analysis and error detection.

### Design Decisions

*   **Rate Limiting:** The implementation of rate limiting is a key design decision to ensure responsible API usage and prevent errors caused by exceeding API limits.
*   **Retry Logic:** The inclusion of retry logic with exponential backoff enhances the robustness of the client by automatically handling transient errors.
*   **Configuration via Environment Variables:**  Loading the API key from an environment variable promotes security and allows for easy configuration without modifying the code.
*   **Token Estimation:** The simple token estimation method provides a reasonable approximation for rate limiting purposes.