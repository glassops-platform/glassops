---
type: Documentation
domain: agent
origin: packages/tools/agent/src/llm-client.ts
last_modified: 2026-02-01
generated: true
source: packages/tools/agent/src/llm-client.ts
generated_at: 2026-02-01T19:50:49.445194
hash: 6936f5e724e055e48121feb3598ccbec69a7f115857cf8236e6897d4bc162ef2
---

## glassops-agent LLM Client Documentation

This document details the functionality of the Large Language Model (LLM) client, specifically designed for interacting with the Gemini API. It provides an overview of the client’s features, setup, and usage.

### Overview

The LLM client facilitates communication with the Gemini API to generate text-based responses from provided prompts. It incorporates rate limiting and error handling to ensure reliable operation and adherence to API usage guidelines. The client is designed to be easily integrated into applications requiring natural language processing capabilities.

### Setup and Configuration

1.  **Environment Variables:** The client requires a `GOOGLE_API_KEY` to be set in a `.env` file. The client automatically searches for a `.env` file in parent directories, starting from the current working directory.  If no `.env` file is found, a warning is logged.

2.  **Installation:** Ensure you have the necessary dependencies installed: `axios`, `dotenv`, and `fs`.

3.  **Initialization:**  You initialize the client by creating an instance of the `GeminiClient` class. You can optionally specify the desired model during initialization. The default model is `gemma-3-27b-it`.

    ```typescript
    const client = new GeminiClient('gemma-3-27b-it');
    ```

### Core Functionality

The primary function of this client is to generate content based on a given prompt.

#### `GeminiClient.generateContent(prompt: string, retryCount: number = 0): Promise<string>`

This asynchronous function sends a prompt to the Gemini API and returns the generated text.

*   **Parameters:**
    *   `prompt`:  The input text prompt for the LLM. (string)
    *   `retryCount`:  An internal parameter used for handling rate limiting and API errors. You do not need to specify this value during normal use. (number, default: 0)

*   **Return Value:** A `Promise` that resolves with the generated text string.

*   **Error Handling:** The function includes robust error handling:
    *   **Rate Limiting:**  The client implements rate limiting to avoid exceeding API quotas. If a request is throttled, it will automatically wait and retry.
    *   **API Errors:**  Handles common API errors (429, 503) with exponential backoff and retry logic.
    *   **Invalid Responses:**  Detects and reports invalid response formats from the Gemini API.
    *   **General Errors:** Catches and logs other potential errors during the request process.

### Rate Limiting

To prevent exceeding API limits, the client incorporates rate limiting. It maintains a minimum interval between requests (currently 4 seconds, approximately 15 requests per minute).  When a request would violate this interval, the client pauses execution until the interval has elapsed.

### Data Structures

#### `ChatMessage` Interface

This interface defines the structure of a chat message, although it is not directly used in the `GeminiClient` class itself. It is included for potential future expansion.

*   `role`:  Indicates the role of the message sender ('user' or 'model').
*   `parts`: An array of text parts composing the message. Each part contains a `text` field.

    ```typescript
    interface ChatMessage {
      role: 'user' | 'model';
      parts: { text: string }[];
    }
    ```

### Important Considerations

*   Ensure your `GOOGLE_API_KEY` has the necessary permissions to access the Gemini API.
*   Monitor API usage to avoid unexpected costs or rate limiting issues.
*   The `maxOutputTokens` and `temperature` parameters within the `generateContent` function control the length and randomness of the generated text. Adjust these values as needed for your specific application.