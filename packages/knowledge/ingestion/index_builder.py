# index_builder.py
# Builds or updates the vector store with embeddings

import chromadb
from chromadb.config import Settings
import os

def build_or_update_index(embeddings):
    """
    embeddings: list of tuples (doc_dict, embedding_vector)
    """
    persist_dir = os.path.join(os.getcwd(), "glassops_index")
    
    # Initialize Chroma Client with persistence
    client = chromadb.PersistentClient(path=persist_dir)
    
    # Create or get collection
    collection = client.get_or_create_collection(
        name="glassops_knowledge",
        metadata={"hnsw:space": "cosine"}
    )
    
    print(f"DEBUG: Using ChromaDB at {persist_dir}")
    
    ids = []
    documents = []
    metadatas = []
    embedding_vectors = []
    
    for doc, emb in embeddings:
        # doc is { "path": ..., "content": ..., "hash": ... }
        doc_id = doc["hash"] # Use hash as ID to avoid duplicates? Or path?
        # Better to combine path + hash if we want versions, 
        # but for now let's use path as ID to overwrite/update easiest.
        # Actually using hash as ID makes it immutable-ish.
        # Let's use path for update-in-place behavior.
        
        ids.append(doc["path"]) 
        documents.append(doc["content"])
        
        # Merge system metadata with extracted doc metadata
        meta = {
            "path": doc["path"], 
            "hash": doc["hash"]
        }
        # Copy relevant fields from doc if present (excluding content/path/hash which are handled)
        for k, v in doc.items():
            if k not in ["content", "path", "hash"] and isinstance(v, (str, int, float, bool)):
                meta[k] = v
                
        metadatas.append(meta)
        embedding_vectors.append(emb)

    if not ids:
        print("No documents to index.")
        return

    # ChromaDB upsert
    try:
        collection.upsert(
            ids=ids,
            embeddings=embedding_vectors,
            documents=documents,
            metadatas=metadatas
        )
        print(f"[SUCCESS] Successfully indexed {len(ids)} documents in ChromaDB.")
    except Exception as e:
        print(f"[ERROR] Error indexing documents: {e}")
