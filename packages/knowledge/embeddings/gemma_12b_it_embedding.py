# gemma_12b_it_embedding.py
# Fallback embedding model

class Gemma12bItEmbedding:
    def get_embeddings(self, texts: list[str]) -> list[list[float]]:
        # Mock implementation
        import random
        # 768 dim to match Gemini usually, or whatever Gemma uses
        return [[random.random() for _ in range(768)] for _ in texts]
