---
type: Documentation
domain: agent
origin: packages/tools/agent/src/llm-client.ts
last_modified: 2026-01-26
generated: true
source: packages/tools/agent/src/llm-client.ts
generated_at: 2026-01-26T14:13:07.277Z
hash: 61f93a71599bcb0cd1ae08d2fd797fdfe210724ede3ffe02519c39e7b4e52ecd
---

## Gemini Large Language Model (LLM) Client

This document details the GeminiClient, a component designed to interact with the Google Gemini LLM API. It provides a streamlined interface for sending prompts and receiving text-based responses.

**Overview**

The GeminiClient facilitates communication with the Gemini API, handling authentication, request formatting, rate limiting, and error management. It is intended for use in applications requiring natural language processing capabilities.

**Functionality**

*   **Initialization:** The client requires a Google API key, which it attempts to retrieve from a `.env` file located in the project’s directory or any parent directory.  You must set the `GOOGLE_API_KEY` environment variable.  It also accepts a model name as a parameter, defaulting to ‘gemma-3-27b-it’.
*   **Prompting:** The `generateContent` method sends a text prompt to the Gemini API and returns the generated text response.
*   **Rate Limiting:** To prevent exceeding API usage limits, the client incorporates rate limiting. It enforces a minimum interval between requests (currently 4 seconds, approximately 15 requests per minute) and logs when waiting.
*   **Error Handling:** The client includes robust error handling. It catches potential issues during API calls, including network errors, invalid responses, and rate limit errors (429 and 503 status codes).  Retry logic with exponential backoff is implemented for rate limit errors, attempting up to three retries. Detailed error messages are logged to the console.
*   **Response Parsing:** The client parses the JSON response from the Gemini API, extracting the generated text. It logs detailed information if a valid text response is not received.

**Key Components**

*   **`ChatMessage` Interface:** Defines the structure of messages exchanged with the LLM, containing a `role` (user or model) and an array of `parts`, each with a `text` field.
*   **`GeminiClient` Class:** The core class responsible for interacting with the Gemini API.
    *   `apiKey`: Stores the Google API key.
    *   `model`: Stores the selected Gemini model.
    *   `baseUrl`: The API endpoint URL.
    *   `generateContent(prompt: string, retryCount: number = 0)`:  Sends a prompt to the Gemini API and returns the generated content.

**Configuration**

*   **API Key:**  Set the `GOOGLE_API_KEY` environment variable in a `.env` file. The client searches for this file in the current directory and its parent directories.
*   **Model:** You can specify the desired Gemini model during client instantiation: `new GeminiClient('your-model-name')`.

**Usage**

1.  Instantiate the `GeminiClient`:

    ```typescript
    const client = new GeminiClient();
    ```

2.  Generate content:

    ```typescript
    const prompt = "Write a short story about a cat.";
    const response = await client.generateContent(prompt);
    console.log(response);
    ```

**Dependencies**

*   `axios`: For making HTTP requests to the Gemini API.
*   `dotenv`: For loading environment variables from a `.env` file.
*   `fs`: For file system operations (finding the `.env` file).
*   `path`: For manipulating file paths.