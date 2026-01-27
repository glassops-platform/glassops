# GlassOps Knowledge Package

This package manages the Retrieval-Augmented Generation (RAG) pipeline for GlassOps. It handles document discovery, embedding generation, vector indexing, and drift detection.

## Structure

- **ingestion/**: Fetches and chunks markdown files from the monorepo.
- **embeddings/**: Generates vector embeddings using Google Gemini (via `google-generativeai`).
- **drift/**: Detects semantic drift in documentation over time.
- **rag/**: Provides a query engine for retrieving context-aware answers.

## Usage

### 1. Setup

Ensure you have a `.env` file in the project root with your API key:

```env
GOOGLE_API_KEY=your_key_here
```

### 2. Run the Pipeline

To scan docs, generate embeddings, and update the index:

```bash
# Using npm script (easiest)
npm run knowledge:pipeline
```

### 3. Query the Knowledge Base

To avoid argument parsing issues with npm (especially on Windows), use the Python executable directly:

```bash
# Direct Python command (Recommended)
packages\knowledge\venv\Scripts\python.exe packages/knowledge/main.py --query "What is the update policy for ADRs?"

# NPM alternative (may require extra escaping on Windows)
npm run knowledge:pipeline -- --query "What is the update policy for ADRs?"
```

### 4. Force Re-indexing

To force a full re-index:

```bash
npm run knowledge:pipeline -- --index
```
