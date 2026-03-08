---
type: Documentation
domain: knowledge
last_modified: 2026-03-08
generated: true
source: packages/knowledge/llm/client.py
generated_at: 2026-03-08T22:46:45.809337
hash: d799a10e04493879c0725b4dc786ed6e3f1c675bc48b65a246d1e6c6c59cba7a
---

## GlassOps Knowledge Pipeline - LLM Client Documentation

This document details the purpose, functionality, and design of the LLM Client module, a component responsible for interacting with the Google Generative AI models. It provides a reusable interface for generating text from prompts, incorporating error handling and rate limit management.

**Module Purpose and Responsibilities:**

The LLM Client module centralizes the interaction with the Google Generative AI API. It handles:

*   API Authentication: Securely manages the API key for accessing the LLM.
*   Prompt Submission: Sends text prompts to the specified LLM model.
*   Response Handling: Processes the LLM’s text responses.
*   Error Management: Implements retry logic for transient errors like rate limits and server unavailability.
*   Rate Limiting: Enforces rate limits to prevent exceeding the API’s usage quotas, ensuring stable operation.
*   Request History: Maintains a record of recent requests for accurate rate limit calculations.

**Key Classes and Roles:**

*   **`LLMClient`**: This is the core class of the module. It encapsulates the LLM interaction logic.
    *   `__init__(self, model: str = "gemma-3-27b-it")`: Initializes the client. It loads the Google API key from environment variables. If the key is missing, the client disables itself. It also sets the default model to "gemma-3-27b-it".  The `model` parameter allows you to specify a different model.
    *   `_estimate_tokens(self, text: str) -> int`: Provides a rough estimate of the number of tokens in a given text string. This is used for rate limiting.
    *   `_throttle(self, estimated_tokens: int) -> None`: Implements the rate limiting logic. It checks if the current request would exceed the defined RPM (requests per minute) and TPM (tokens per minute) limits. If a limit is approaching, it pauses execution until the rate limit window allows another request.
    *   `generate(self, prompt: str, max_retries: int = 5, temperature: float = 0.2, max_output_tokens: int = 8192) -> Optional[str]`:  The primary method for generating text. It sends a prompt to the LLM, handles potential errors through retries, and returns the generated text. It also incorporates rate limiting before sending the request.

**Important Functions and Their Behavior:**

*   **`generate(prompt, max_retries, temperature, max_output_tokens)`**: This function is the main entry point for interacting with the LLM.
    *   `prompt` (str): The input text that will be sent to the LLM.
    *   `max_retries` (int, default=5): Specifies the maximum number of times to retry the request if a transient error occurs.
    *   `temperature` (float, default=0.2): Controls the randomness of the generated text. Lower values produce more predictable output.
    *   `max_output_tokens` (int, default=8192): Sets the maximum number of tokens the LLM can generate in its response.
    *   Returns: The generated text as a string, or `None` if the request fails after multiple retries.

*   **`_throttle(estimated_tokens)`**: This private function manages the rate limits. It maintains a history of recent requests and their token usage. Before making a new request, it checks if the request would exceed the configured RPM and TPM limits. If so, it pauses execution until the rate limit window allows another request.

**Type Hints and Their Significance:**

The code extensively uses type hints (e.g., `str`, `int`, `Optional[str]`). These hints improve code readability and maintainability. They allow static analysis tools to detect potential type errors, reducing the risk of runtime issues.  The `Optional[str]` return type in the `generate` function indicates that the function may return a string or `None` if the generation fails.

**Notable Patterns and Design Decisions:**

*   **Retry Logic:** The `generate` function implements a retry mechanism with exponential backoff to handle transient errors. This improves the robustness of the client.
*   **Rate Limiting:** The `_throttle` function proactively manages rate limits to prevent API errors and ensure fair usage.
*   **Configuration via Environment Variables:** The API key is loaded from an environment variable (`GOOGLE_API_KEY`), promoting secure configuration and separation of concerns.
*   **Modular Design:** The `LLMClient` class encapsulates all LLM interaction logic, making it reusable and testable.
*   **Token Estimation:** The `_estimate_tokens` function provides a basic token estimation to assist with rate limiting.
*   **Error Handling:** The code includes specific checks for retryable errors (429, 503) and provides informative error messages.
*   **Request History:** The `_request_history` list is used to track recent requests for accurate rate limit calculations.
*   **Defensive Programming:** The code checks for a missing API key and disables the client if it's not found, preventing unexpected errors.