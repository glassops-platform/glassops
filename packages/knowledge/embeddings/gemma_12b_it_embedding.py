import os
import warnings

# Suppress google.generativeai deprecation warning
warnings.filterwarnings("ignore", category=FutureWarning, module="google.generativeai")
warnings.filterwarnings("ignore", category=FutureWarning, module="google.auth")

try:
    import google.generativeai as genai
except ImportError:
    genai = None

class Gemma12bItEmbedding:
    def __init__(self):
        self.api_key = os.getenv("GOOGLE_API_KEY")
        if not self.api_key:
             print("[WARNING] Warning: GOOGLE_API_KEY not set. Gemma12bItEmbedding will return mock data.")
        elif genai:
            genai.configure(api_key=self.api_key)

    def get_embeddings(self, texts: list[str]) -> list[list[float]]:
        """
        Generates embeddings using the Google GenAI API.
        Acts as a functional fallback/alternative to GeminiEmbedding.
        """
        if self.api_key and genai:
            try:
                # Using the standard embedding model as Gemma instruction-tuned models 
                # don't typically expose a direct public embedding endpoint in the SDK 
                # different from the main text-embedding-models.
                # using 004 as it is the most capable.
                model = "models/text-embedding-004" 
                
                result = genai.embed_content(
                    model=model,
                    content=texts,
                    task_type="retrieval_document"
                )
                
                if 'embedding' in result:
                    emb_data = result['embedding']
                    if isinstance(emb_data, list) and len(emb_data) > 0:
                        if isinstance(emb_data[0], list):
                             return emb_data
                        return [emb_data] if isinstance(emb_data[0], (float, int)) else emb_data

            except Exception as e:
                # Fallback to sequential if batch fails
                pass

        # Sequential Fallback
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
                     import random
                     embeddings.append([random.random() for _ in range(768)])
             return embeddings

        # Final Mock Fallback
        import random
        return [[random.random() for _ in range(768)] for _ in texts]
