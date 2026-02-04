import os
import time
import warnings

# Suppress google.generativeai deprecation warning
warnings.filterwarnings("ignore", category=FutureWarning, module="google.generativeai")
warnings.filterwarnings("ignore", category=FutureWarning, module="google.auth")

try:
    import google.generativeai as genai
except ImportError:
    genai = None

class GeminiEmbedding:
    def __init__(self):
        self.api_key = os.getenv("GOOGLE_API_KEY")
        if not self.api_key:
             print("[WARNING] Warning: GOOGLE_API_KEY not set. GeminiEmbedding will return mock data.")
        elif genai:
            genai.configure(api_key=self.api_key)

    def get_embeddings(self, texts: list[str]) -> list[list[float]]:
        if self.api_key and genai:
            try:
                # Try batched call first
                model = "models/text-embedding-004" 
                
                # Check for list support in embed_content (older SDK behavior)
                # or use batch_embed_contents if available (newer SDK)
                # Ideally, we just try to pass the list.
                
                result = genai.embed_content(
                    model=model,
                    content=texts,
                    task_type="retrieval_document"
                )
                
                # If result contains 'embedding', it might be a single embedding (if texts was string)
                # OR a list of embeddings? 
                # According to docs, passing a list of strings returns a dict with 'embedding' key which is a list of lists.
                if 'embedding' in result:
                    emb_data = result['embedding']
                    # Verify structure
                    if isinstance(emb_data, list) and len(emb_data) > 0:
                        # If list of lists (batch)
                        if isinstance(emb_data[0], list):
                             return emb_data
                        # If list of floats (single), but we passed a list?
                        # This happens if 'texts' was a single string, but type hint says list[str].
                        # If we passed list of 1 string, it usually returns list of 1 vector.
                        # Let's handle generic case:
                        return [emb_data] if isinstance(emb_data[0], (float, int)) else emb_data

            except Exception as e:
                # Pass through to fallback
                # print(f"DEBUG: Batch embedding failed ({e}), switching to sequential.")
                pass

        # Fallback: Sequential processing
        if self.api_key and genai:
             embeddings = []
             for text in texts:
                 try:
                     result = genai.embed_content(
                        model="models/text-embedding-004",
                        content=text,
                        task_type="retrieval_document"
                     )
                     embeddings.append(result['embedding'])
                 except Exception as e:
                     print(f"[ERROR] Error embedding chunk: {e}")
                     # Random fallback for failed chunk to keep alignment
                     import random
                     embeddings.append([random.random() for _ in range(768)])
             return embeddings

        # Mock 768-dim vectors
        import random
        return [[random.random() for _ in range(768)] for _ in texts]
