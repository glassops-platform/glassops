# llm/client.py
"""
Shared LLM client for GlassOps Knowledge Pipeline.
Handles API interaction, retry logic, and rate limiting.
"""

import os
import time
from typing import Optional
from pathlib import Path

from google import genai
from google.genai import types
from dotenv import load_dotenv

# Load .env from project root
ROOT_DIR = Path(__file__).parent.parent.parent.parent
load_dotenv(ROOT_DIR / ".env")


class LLMClient:
    """
    A reusable client for interacting with Google Generative AI.
    Includes retry logic for transient errors (429, 503).
    """

    def __init__(self, model: str = "gemma-3-27b-it"):
        api_key = os.getenv("GOOGLE_API_KEY", "").strip().strip("'\"")
        if not api_key:
            raise ValueError("GOOGLE_API_KEY not found in environment.")

        self.client = genai.Client(api_key=api_key)
        self.model = model
        self._request_history: list[dict] = []
        self._rpm_limit = 28  # Safety buffer below 30
        self._tpm_limit = 14000  # Safety buffer below 15000

    def _estimate_tokens(self, text: str) -> int:
        """Rough token estimation (4 chars per token)."""
        return len(text) // 4

    def _throttle(self, estimated_tokens: int) -> None:
        """
        Simple throttle to stay within RPM/TPM limits.
        Blocks if we're approaching the limit.
        """
        now = time.time()
        window_size = 60  # 1 minute

        # Clean old entries
        self._request_history = [
            entry for entry in self._request_history
            if now - entry["time"] < window_size
        ]

        # Check RPM
        if len(self._request_history) >= self._rpm_limit:
            oldest = self._request_history[0]
            wait_time = (oldest["time"] + window_size) - now
            if wait_time > 0:
                print(f"⏳ RPM Limit: Waiting {wait_time:.1f}s...")
                time.sleep(wait_time)
                return self._throttle(estimated_tokens)  # Re-check

        # Check TPM
        current_tokens = sum(e["tokens"] for e in self._request_history)
        if current_tokens + estimated_tokens > self._tpm_limit:
            # Find when we'll have enough headroom
            freed_tokens = 0
            wait_time = 0
            for entry in self._request_history:
                freed_tokens += entry["tokens"]
                if current_tokens - freed_tokens + estimated_tokens <= self._tpm_limit:
                    wait_time = (entry["time"] + window_size) - now
                    break

            if wait_time > 0:
                print(f"⏳ TPM Limit ({current_tokens}/{self._tpm_limit}): Waiting {wait_time:.1f}s...")
                time.sleep(wait_time)
                return self._throttle(estimated_tokens)

        self._request_history.append({"time": time.time(), "tokens": estimated_tokens})

    def generate(
        self,
        prompt: str,
        max_retries: int = 3,
        temperature: float = 0.2,
        max_output_tokens: int = 8192,
    ) -> Optional[str]:
        """
        Generate content from a prompt with retry logic.

        Args:
            prompt: The prompt to send to the model.
            max_retries: Maximum number of retries on transient errors.
            temperature: Sampling temperature.
            max_output_tokens: Max tokens for the response.

        Returns:
            The generated text, or None on failure.
        """
        estimated_tokens = self._estimate_tokens(prompt) + 100  # Buffer for output
        self._throttle(estimated_tokens)

        backoffs = [10, 30, 60]  # Retry delays in seconds

        for attempt in range(max_retries + 1):
            try:
                response = self.client.models.generate_content(
                    model=self.model,
                    contents=prompt,
                    config=types.GenerateContentConfig(
                        temperature=temperature,
                        max_output_tokens=max_output_tokens,
                    ),
                )

                if response.text:
                    return response.text

                # No text but no exception - check finish reason
                print(f"⚠️ Gemini returned no text. Finish reason: {response.candidates[0].finish_reason if response.candidates else 'Unknown'}")
                return None

            except Exception as e:
                error_str = str(e)
                is_retryable = "429" in error_str or "503" in error_str or "overloaded" in error_str.lower()

                if is_retryable and attempt < max_retries:
                    wait = backoffs[min(attempt, len(backoffs) - 1)]
                    print(f"⚠️ Retryable error ({error_str[:50]}...). Retrying in {wait}s (attempt {attempt + 1}/{max_retries})...")
                    time.sleep(wait)
                    continue

                print(f"❌ LLM Error: {e}")
                return None

        return None
