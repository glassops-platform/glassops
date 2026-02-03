---
type: Documentation
domain: knowledge
last_modified: 2026-02-02
generated: true
source: packages/knowledge/drift/detect_drift.py
generated_at: 2026-02-02T22:27:47.770479
hash: dee44b64dd3a391ae28d74379a8cb4f62cf1733a64945e310fcbe0a6ce6a6ac2
---

## Knowledge Drift Detection Documentation

This module is responsible for comparing newly ingested document embeddings against previously established embeddings to identify potential drift in the knowledge base. Drift, in this context, refers to significant changes in the content or meaning of documents that could impact the performance of retrieval-augmented generation (RAG) systems.

**Core Functionality:**

The primary function, `detect_drift`, analyzes a list of document embeddings and identifies documents that exhibit characteristics indicative of drift. Currently, the implementation simulates drift detection by identifying near-duplicate documents, which suggests redundancy or conflicting information. A report detailing the findings is generated.

**Key Components:**

*   **`cosine_similarity(a, b)`:** This function calculates the cosine similarity between two vectors `a` and `b`. It returns a value between -1 and 1, where 1 indicates perfect similarity and 0 indicates orthogonality (no similarity). This function is a utility for measuring the similarity between embeddings.
    *   `a`: A numpy array representing the first embedding vector.
    *   `b`: A numpy array representing the second embedding vector.
    *   Returns: A float representing the cosine similarity between `a` and `b`.

*   **`detect_drift(embeddings, threshold=0.85)`:** This is the main function for drift detection. It takes a list of document embeddings as input and returns a list of document paths that are considered to have drifted.
    *   `embeddings`: A list of tuples, where each tuple contains a document dictionary (`doc_dict`) and its corresponding embedding vector. The `doc_dict` is expected to have a "hash" key for content identification and a "path" key for document location.
    *   `threshold`: A float representing the similarity threshold. Documents with a similarity score below this threshold are considered to have drifted. The default value is 0.85.
    *   Returns: A list of strings, where each string is the path to a document that has drifted.

**Drift Detection Process:**

1.  **Near-Duplicate Detection:** The function iterates through the provided embeddings, maintaining a dictionary (`seen_hashes`) to track document hashes and their corresponding paths. If a duplicate hash is encountered, the corresponding document paths are flagged as potential conflicts.
2.  **Report Generation:** A markdown report is generated at `../docs/generated/drift_report.md`. This report summarizes the findings of the drift detection process.
    *   If near-duplicate documents are found, the report lists the conflicting document paths.
    *   If no conflicts are detected, the report indicates that all indexed documents appear unique.
    *   A "Drift Status" section currently states that no significant semantic drift has been detected, as full drift detection is not yet implemented.
3.  **Return Value:** The function currently returns an empty list (`drifted`) as the full drift detection logic is not yet implemented. In a future version, this list will contain the paths of documents identified as having drifted based on semantic similarity comparisons.

**Design Considerations:**

*   **Type Hints:** The code uses type hints (e.g., `embeddings: list[tuple[dict, np.ndarray]]`, `threshold: float`) to improve code readability and maintainability. These hints help clarify the expected data types for function arguments and return values.
*   **Future Expansion:** The current implementation is a placeholder for a more robust drift detection mechanism. The `TODO` comment indicates that the function will eventually load previous embedding snapshots and compare them to the current embeddings to identify semantic drift.
*   **Report-Driven Approach:** The module adopts a report-driven approach to communicate drift detection results. This allows for easy monitoring and analysis of the knowledge base health.
*   **Hash-Based Duplicate Detection:** The current method for identifying potential conflicts relies on document hashes. This is a quick and efficient way to detect exact duplicates, but it does not account for near-duplicates or semantic changes.