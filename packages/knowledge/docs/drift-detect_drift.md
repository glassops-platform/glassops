---
type: Documentation
domain: knowledge
origin: packages/knowledge/drift/detect_drift.py
last_modified: 2026-01-31
generated: true
source: packages/knowledge/drift/detect_drift.py
generated_at: 2026-01-31T08:52:32.530018
hash: dee44b64dd3a391ae28d74379a8cb4f62cf1733a64945e310fcbe0a6ce6a6ac2
---

## Knowledge Drift Detection Documentation

This module is designed to identify changes in the knowledge base over time, a process known as drift. Drift can occur when new information is added that is semantically different from existing content, or when existing content is modified in a way that alters its meaning. This impacts the performance of Retrieval-Augmented Generation (RAG) systems.

**Module Responsibilities:**

The primary responsibility of this module is to compare newly processed document embeddings against a baseline (currently simulated) to detect drift. It generates a report detailing any identified conflicts or potential drift.

**Key Functions:**

*   **`cosine_similarity(a, b)`:** This function calculates the cosine similarity between two vectors, `a` and `b`. Cosine similarity measures the angle between the vectors, providing a value between -1 and 1, where 1 indicates perfect similarity and 0 indicates orthogonality (no similarity). It uses NumPy for efficient vector operations. Type hints are used to specify that both `a` and `b` are NumPy arrays.

*   **`detect_drift(embeddings, threshold=0.85)`:** This is the core function for drift detection. It accepts a list of tuples, where each tuple contains a document dictionary (`doc_dict`) and its corresponding embedding vector (`embedding_vector`). The `threshold` parameter (defaulting to 0.85) represents the minimum cosine similarity score considered acceptable.

    Currently, the drift detection logic is simulated. Instead of comparing against a stored snapshot of previous embeddings, it identifies near-duplicate documents within the current set. This is done by hashing document content and checking for collisions.

    The function generates a "Knowledge Base Health Report" in Markdown format, saved to `docs/generated/drift_report.md`. This report details any detected near-duplicate documents and a placeholder drift status.

    The function returns a list of document paths identified as having drifted. Currently, this list is always empty due to the simulated drift detection.

**Data Structures:**

*   **`embeddings`:** A list of tuples. Each tuple contains:
    *   `doc_dict`: A dictionary representing a document, expected to have a `"hash"` key (for content hashing) and a `"path"` key (for document location).
    *   `embedding_vector`: A NumPy array representing the document's embedding.

**Design Decisions and Patterns:**

*   **Report Generation:** The module employs a report-based approach to communicate drift information. This provides a human-readable summary of the knowledge base's health.
*   **Simulated Drift:** The current implementation uses a simulation for drift detection. This allows for testing and development without requiring a persistent storage of previous embeddings. The comment `TODO: load real previous embeddings snapshot for actual drift detection` indicates the intended future behavior.
*   **Near-Duplicate Detection:** The current simulation focuses on identifying near-duplicate documents as a proxy for content conflicts. This is a practical approach for detecting redundancy in the knowledge base.
*   **Type Hints:** Type hints are used to improve code readability and maintainability, and to enable static analysis.
*   **File Path Handling:** The module uses `os.path.join` to construct file paths, ensuring cross-platform compatibility. `os.makedirs(..., exist_ok=True)` is used to create directories if they do not exist, preventing errors.

**Future Considerations:**

The module is designed to be extended with real drift detection logic. This will involve loading previous embeddings from storage, calculating cosine similarity scores, and identifying documents that fall below the specified threshold. The report generation will be updated to reflect the actual drift detected. You will need to provide a mechanism for storing and retrieving embedding snapshots.