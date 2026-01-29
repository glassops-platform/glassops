---
type: Documentation
domain: knowledge
origin: packages/knowledge/drift/detect_drift.py
last_modified: 2026-01-28
generated: true
source: packages/knowledge/drift/detect_drift.py
generated_at: 2026-01-28T22:39:00.075917
hash: dee44b64dd3a391ae28d74379a8cb4f62cf1733a64945e310fcbe0a6ce6a6ac2
---

## Knowledge Drift Detection Documentation

This module is designed to identify changes in the knowledge base over time, a process known as drift. Drift can occur when new information is added that is semantically different from existing content, or when existing content is modified. This impacts the performance of Retrieval-Augmented Generation (RAG) systems.

**Module Responsibilities:**

The primary responsibility of this module is to compare newly processed document embeddings against a baseline (currently simulated) to detect drift. It generates a report detailing any identified conflicts or potential drift.

**Key Functions:**

*   **`cosine_similarity(a, b)`:** This function calculates the cosine similarity between two vectors, `a` and `b`. Cosine similarity measures the angle between the vectors, providing a value between -1 and 1, where 1 indicates perfect similarity and 0 indicates orthogonality (no similarity). It uses NumPy for efficient vector operations. Type hints are used to specify that both `a` and `b` are NumPy arrays.

*   **`detect_drift(embeddings, threshold=0.85)`:** This is the core function for drift detection. It accepts a list of tuples, where each tuple contains a document dictionary (`doc_dict`) and its corresponding embedding vector (`embedding_vector`). The `threshold` parameter (defaulting to 0.85) represents the minimum cosine similarity score considered acceptable.

    Currently, the drift detection logic is simulated. Instead of comparing against a stored snapshot of previous embeddings, it identifies near-duplicate documents within the current set. This is done by hashing document content and checking for collisions.

    The function generates a markdown report (`drift_report.md`) summarizing the findings. The report includes sections for:
    *   Knowledge Base Health Report header
    *   Generation source file
    *   Conflicting/Duplicate Documentation (if any are found)
    *   Drift Status (currently always reports "No significant semantic drift detected")

    The report is written to the `docs/generated` directory. The function returns a list of document paths identified as drifted (currently an empty list due to the simulated nature of the drift detection).

**Key Data Structures:**

*   **`embeddings`:** A list of tuples. Each tuple contains a document dictionary (`doc_dict`) and its embedding vector. The `doc_dict` is expected to have a `"hash"` key representing the document's content hash and a `"path"` key representing the document's file path.
*   **`seen_hashes`:** A dictionary used to track document hashes and their corresponding paths. This is used to identify near-duplicate documents.

**Design Decisions and Patterns:**

*   **Simulated Drift Detection:** The current implementation simulates drift detection due to the absence of a persistent storage mechanism for previous embeddings. This allows for testing and demonstration of the reporting functionality.
*   **Markdown Reporting:** The use of markdown for the drift report provides a human-readable and easily maintainable format for communicating drift information.
*   **Hashing for Near-Duplicate Detection:** Employing document hashing is an efficient way to identify potential content conflicts or redundancy.
*   **Type Hints:** Type hints are used throughout the code to improve readability and maintainability, and to enable static analysis.
*   **File Path Handling:** The module uses `os.path.join` to construct file paths, ensuring cross-platform compatibility. `os.makedirs(..., exist_ok=True)` is used to create the report directory if it does not exist, preventing errors.

**Future Improvements:**

We plan to implement the following improvements:

*   Load previous embeddings from a persistent storage location (e.g., a vector database) for accurate drift detection.
*   Implement a more sophisticated drift detection algorithm based on cosine similarity or other distance metrics.
*   Allow users to configure the drift threshold.
*   Provide more detailed information in the drift report, such as the magnitude of the drift and the affected documents.