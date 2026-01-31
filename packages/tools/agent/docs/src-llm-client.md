---
type: Documentation
domain: agent
origin: packages/tools/agent/src/llm-client.ts
last_modified: 2026-01-31
generated: true
source: packages/tools/agent/src/llm-client.ts
generated_at: 2026-01-31T09:23:30.464008
hash: 6936f5e724e055e48121feb3598ccbec69a7f115857cf8236e6897d4bc162ef2
---

## glassops-agent LLM Client Documentation

This document details the functionality of the Large Language Model (LLM) Client, designed to interact with the Gemini API. It provides a clear overview for both technical and non-technical users.

**Overview**

The LLM Client facilitates communication with a language model to generate text based on provided prompts. It handles API key management, request formatting, rate limiting, and error handling to provide a reliable interface for accessing LLM capabilities.  I am designed to be easily integrated into applications requiring text generation.

**Key Features**

*   **Gemini API Integration:** Connects to the Gemini API for text generation.
*   **API Key Management:** Securely retrieves the API key from environment variables.  You must set the `GOOGLE_API_KEY` environment variable.
*   **Rate Limiting:** Implements rate limiting to prevent exceeding API usage limits and ensure fair usage. A minimum interval of 4 seconds between requests is enforced.
*   **Error Handling:** Robust error handling with retry logic for common API errors (429 and 503 status codes). Detailed error messages are logged for debugging.
*   **Configurable Model:** Allows selection of different Gemini models via the constructor. The default model is 'gemma-3-27b-it'.
*   **Configurable Parameters:** Sets `maxOutputTokens` to 8192 and `temperature` to 0.2 for controlling the generated text.

**Installation & Setup**

1.  Ensure you have an active Gemini API key.
2.  Set the `GOOGLE_API_KEY` environment variable to your API key. This can be done by creating a `.env` file in the project root or setting the environment variable directly in your system. The client automatically searches for a `.env` file in parent directories.
3.  Install the necessary dependencies: `axios`, `dotenv`, and `fs`.

**Usage**

1.  **Import the Client:**
    ```typescript
    import { GeminiClient } from './llm-client';
    ```

2.  **Instantiate the Client:**
    ```typescript
    const client = new GeminiClient(); // Uses default model 'gemma-3-27b-it'
    // or
    const client = new GeminiClient('another-model'); // Specify a different model
    ```

3.  **Generate Content:**
    ```typescript
    const prompt = 'Write a short story about a robot.';
    try {
      const response = await client.generateContent(prompt);
      console.log(response);
    } catch (error) {
      console.error('Error generating content:', error);
    }
    ```

**Interfaces**

*   `ChatMessage`: Represents a message in a conversation.
    *   `role`:  Indicates the role of the message sender ('user' or 'model').
    *   `parts`: An array of text parts composing the message. Each part contains a `text` field.

**Important Considerations**

*   **API Key Security:** Protect your `GOOGLE_API_KEY`. Do not commit it to version control.
*   **Rate Limits:** Be mindful of the Gemini API rate limits. The client includes rate limiting, but exceeding limits may still result in errors.
*   **Error Handling:** Implement appropriate error handling in your application to gracefully handle potential API errors.
*   **Cost:** Using the Gemini API may incur costs. Review the Gemini API pricing documentation for details.
*   **Model Selection:** Choose the appropriate Gemini model based on your specific needs and budget.