---
type: Documentation
domain: agent
origin: packages/tools/agent/src/llm-client.ts
last_modified: 2026-01-31
generated: true
source: packages/tools/agent/src/llm-client.ts
generated_at: 2026-01-31T10:20:44.500225
hash: 6936f5e724e055e48121feb3598ccbec69a7f115857cf8236e6897d4bc162ef2
---

## glassops-agent LLM Client Documentation

This document details the functionality of the Large Language Model (LLM) client, specifically designed for interacting with the Gemini API. It provides an overview of the client’s features, setup, and usage.

### Overview

The LLM client facilitates communication with the Gemini API to generate text-based responses from provided prompts. It incorporates rate limiting and error handling to ensure reliable operation and adherence to API usage guidelines. The client is designed to be easily integrated into applications requiring natural language processing capabilities.

### Setup and Configuration

1.  **Environment Variables:** The client requires a `GOOGLE_API_KEY` to be set in a `.env` file. The client automatically searches for a `.env` file in parent directories, starting from the current working directory.  If no `.env` file is found, a warning is logged.

2.  **Installation:** Ensure you have the necessary dependencies installed: `axios`, `dotenv`, and `fs`.

3.  **Initialization:** You initialize the client by creating an instance of the `GeminiClient` class. You can optionally specify the desired model during initialization. The default model is `gemma-3-27b-it`.

    ```typescript
    const client = new GeminiClient('gemma-3-27b-it');
    ```

### Core Functionality

The primary function of this client is to generate content based on a given prompt.

#### `GeminiClient` Class

*   **`constructor(model: string = 'gemma-3-27b-it')`**:  Initializes the Gemini client.
    *   `model`: (Optional) Specifies the Gemini model to use. Defaults to `gemma-3-27b-it`.
    *   Throws an error if the `GOOGLE_API_KEY` is not found in the environment.

*   **`generateContent(prompt: string, retryCount: number = 0): Promise<string>`**:  Sends a prompt to the Gemini API and returns the generated text.
    *   `prompt`: The text prompt to send to the Gemini API.
    *   `retryCount`: (Optional) Internal parameter for retry logic. You should not need to specify this directly.
    *   Implements rate limiting to avoid exceeding API limits.
    *   Includes retry logic with exponential backoff for 429 (Overloaded) and 503 (Service Unavailable) errors.
    *   Returns a `Promise` that resolves with the generated text.
    *   Throws an error if the API request fails or returns an invalid response.

### Data Structures

*   **`ChatMessage` Interface:** Defines the structure of a chat message. While not directly used in the `GeminiClient` class, it is exported for potential use in related components.
    *   `role`:  Indicates the role of the message sender ('user' or 'model').
    *   `parts`: An array of text parts composing the message. Each part contains a `text` field.

### Error Handling

The client includes robust error handling:

*   **Missing API Key:** Throws an error if the `GOOGLE_API_KEY` environment variable is not set.
*   **API Errors:** Logs detailed error messages for API failures, including status codes and response data.
*   **Rate Limiting:**  Pauses execution when rate limits are approached, logging the wait time.
*   **Invalid Responses:**  Handles cases where the Gemini API returns an unexpected response format, logging the `finishReason` and full response for debugging.
*   **Retry Mechanism:** Automatically retries requests on transient errors (429 and 503) with increasing delays.

### Rate Limiting

To prevent exceeding API limits, the client incorporates rate limiting. It enforces a minimum interval of 4 seconds between requests, allowing approximately 15 requests per minute. This helps avoid hitting the 15,000 tokens per minute limit.