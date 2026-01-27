# query_engine.py
import chromadb
import os
import google.generativeai as genai
import json
from pathlib import Path
from knowledge.embeddings.router_embedding import get_embeddings_for_docs

def query_index(query, n_results=5):
    """
    query: string
    returns: summarized answer from RAG
    """
    print(f"DEBUG: Querying for '{query}'...")
    
    # 1. Embed the query
    # We strip it into a list wrapper because our embedding function expects list[str]
    # unpacking the list of list result [ [0.1, ...] ] -> [0.1, ...]
    try:
        query_embeddings = get_embeddings_for_docs([{"content": query}])[0][1]
    except Exception as e:
        return f"Error generating embedding: {e}"

    # 2. Query ChromaDB
    persist_dir = os.path.join(os.getcwd(), "glassops_index")
    if not os.path.exists(persist_dir):
        return "Error: Index not found. Please run with --index first."
        
    client = chromadb.PersistentClient(path=persist_dir)
    collection = client.get_or_create_collection(name="glassops_knowledge")
    
    results = collection.query(
        query_embeddings=[query_embeddings],
        n_results=n_results
    )
    
    # 3. Construct Context
    context_chunks = results['documents'][0]
    sources = results['ids'][0]

    # Post-retrieval: Check for config-based file injection
    # This ensures critical reports (like drift/audit) are present if keywords match, 
    # overcoming vector similarity limitations.
    try:
        config_path = Path(__file__).parent.parent / "config" / "config.json"
        with open(config_path, "r", encoding="utf-8") as f:
            cfg = json.load(f)
            triggers = cfg.get("retrieval_triggers", {})
            
            project_root = Path(__file__).parent.parent.parent.parent
            injected_files = set()
            
            for keyword, rel_path in triggers.items():
                if keyword.lower() in query.lower():
                    # Resolve path relative to project root
                    abs_path = project_root / rel_path
                    if abs_path.exists() and str(abs_path) not in injected_files:
                        try:
                            content = abs_path.read_text(encoding="utf-8")
                            # Prepend to context (high priority)
                            context_chunks.insert(0, f"--- START SYSTEM REPORT ({rel_path}) ---\n{content}\n--- END SYSTEM REPORT ---\n")
                            sources.insert(0, str(abs_path))
                            injected_files.add(str(abs_path))
                            print(f"DEBUG: Trigger '{keyword}' detected. Injected {rel_path}.")
                        except Exception as e:
                            print(f"Warning: Failed to inject trigger file {rel_path}: {e}")

    except Exception as e:
        print(f"Warning: Trigger mechanism failed: {e}")

    if not context_chunks:
        return "I couldn't find any relevant information in the knowledge base."
    
    context_text = "\n\n---\n\n".join(context_chunks)
    
    # 4. Generate Answer with Gemini
    api_key = os.getenv("GOOGLE_API_KEY")
    if not api_key:
        return f"Context found ({len(context_chunks)} chunks), but GOOGLE_API_KEY not set for generation.\n\nTop Source: {sources[0]}"

    # Provide global context about the repository structure to the LLM
    # Load from config or use default
    config_path = Path(__file__).parent.parent / "config" / "config.json"
    
    # Default prompt if config fails
    system_context = "You are an expert for the GlassOps platform."
    
    try:
         with open(config_path, "r", encoding="utf-8") as f:
             cfg = json.load(f)
             if "system_context" in cfg:
                 system_context = cfg["system_context"]
    except Exception as e:
         # Fallback
         print(f"Warning: Could not load system_context from config: {e}")

    try:
        genai.configure(api_key=api_key)
        # Using user-specified model
        model_name = 'gemma-3-12b-it'
        model = genai.GenerativeModel(model_name)
        
        prompt = f"""{system_context}

Answer the user's question based strictly on the provided context.
If the answer is not in the context, say you don't know.

Context:
{context_text}

Question:
{query}

Answer:"""
        
        response = model.generate_content(prompt)
        return f"{response.text}\n\nSources:\n- " + "\n- ".join(sources)
        
    except Exception as e:
        # Debug helper: List available models
        try:
             models = list(genai.list_models())
             # Filter for generateContent support
             available = [m.name for m in models if 'generateContent' in m.supported_generation_methods]
             return f"Error with model '{model_name}': {e}\n\nAvailable Models (for generating content): {available}"
        except:
             return f"Error generating response: {e}\n\nContext:\n{context_text[:500]}..."
