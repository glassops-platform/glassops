# router-embedding.py
# Routes embedding requests based on quota / fallback

from .gemini_embedding import GeminiEmbedding
from .gemma_12b_it_embedding import Gemma12bItEmbedding

class RPDLimitError(Exception):
    pass

def get_embeddings_for_docs(docs, batch_size=10):
    primary = GeminiEmbedding()
    fallback = Gemma12bItEmbedding()
    embeddings = []

    for i in range(0, len(docs), batch_size):
        print(f"  Processed {i}/{len(docs)}...", end='\r')
        batch = docs[i:i+batch_size]
        try:
            emb = primary.get_embeddings([d["content"] for d in batch])
            embeddings.extend(zip(batch, emb))
        except RPDLimitError:
            emb = fallback.get_embeddings([d["content"] for d in batch])
            embeddings.extend(zip(batch, emb))
    print(f"  Processed {len(docs)}/{len(docs)}... Done.")
    return embeddings
