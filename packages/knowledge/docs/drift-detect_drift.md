---
type: Documentation
domain: knowledge
origin: packages/knowledge/drift/detect_drift.py
last_modified: 2026-02-01
generated: true
source: packages/knowledge/drift/detect_drift.py
generated_at: 2026-02-01T19:28:19.477427
hash: dee44b64dd3a391ae28d74379a8cb4f62cf1733a64945e310fcbe0a6ce6a6ac2
---

## Knowledge Drift Detection Documentation

This module is responsible for comparing newly ingested document embeddings against previously established embeddings to identify potential drift in the knowledge base. Drift, in this context, refers to changes in the content or meaning of documents that could impact the performance of retrieval-augmented generation (RAG) systems.

**Key Responsibilities:**

*   Analyzing document embeddings for similarity to existing embeddings.
*   Generating a report detailing any detected conflicts or potential drift.
*   Providing a mechanism to alert users to potential knowledge base issues.

**Functions:**

1.  **`cosine_similarity(a, b)`**

    *   **Purpose:** Calculates the cosine similarity between two vectors `a` and `b`.
    *   **Parameters:**
        *   `a`: A NumPy array representing the first vector.
        *   `b`: A NumPy array representing the second vector.
    *   **Return Value:** A float representing the cosine similarity between the two vectors.  A value closer to 1 indicates higher similarity.
    *   **Type Hints:** `a: numpy.ndarray`, `b: numpy.ndarray` -> `float`

2.  **`detect_drift(embeddings, threshold=0.85)`**

    *   **Purpose:** Detects drift by comparing new document embeddings to a baseline. Currently, it simulates drift detection by identifying near-duplicate documents.
    *   **Parameters:**
        *   `embeddings`: A list of tuples, where each tuple contains a document dictionary (`doc_dict`) and its corresponding embedding vector. The `doc_dict` is expected to have a "hash" key for content identification and a "path" key for document location.
        *   `threshold`: (Optional) A float representing the similarity threshold.  Currently unused in the near-duplicate detection logic, but reserved for future semantic drift comparison. Defaults to 0.85.
    *   **Return Value:** A list of document paths that are identified as having drifted (currently, this list is empty as drift detection is simulated).
    *   **Type Hints:** `embeddings: list[tuple[dict, numpy.ndarray]]`, `threshold: float = 0.85` -> `list[str]`
    *   **Behavior:**
        *   The function iterates through the provided `embeddings`.
        *   It identifies potential conflicts by checking for documents with identical content hashes.
        *   A "Knowledge Base Health Report" is generated as a Markdown file (`drift_report.md`) located in the `docs/generated` directory.
        *   The report details any detected near-duplicate documents. If no duplicates are found, it indicates that all documents appear unique.
        *   Currently, the drift status section of the report always states that no significant semantic drift has been detected.

**Design Decisions and Patterns:**

*   **Report Generation:** The module employs a report-based approach to communicate drift information. This provides a human-readable summary of potential issues.
*   **Simulated Drift Detection:**  The current implementation simulates drift detection by identifying near-duplicate documents. This is a placeholder for more sophisticated semantic drift analysis that would compare embeddings directly.
*   **Hashing for Content Identification:** Document content is identified using a hash value stored in the `doc_dict`. This allows for efficient detection of identical documents.
*   **Type Hints:** Type hints are used throughout the code to improve readability and maintainability. They also enable static analysis tools to catch potential errors.
*   **File Path Handling:** The module uses `os.path.join` to construct file paths, ensuring cross-platform compatibility. `os.makedirs(..., exist_ok=True)` is used to create the report directory if it doesn't exist, preventing errors.

**Future Considerations:**

*   Implement actual semantic drift detection by comparing new embeddings to a snapshot of previous embeddings using the `cosine_similarity` function and the provided `threshold`.
*   Integrate with a version control system to track changes to the knowledge base over time.
*   Provide more detailed drift analysis, including the magnitude and direction of drift.
*   Allow users to configure the drift detection parameters, such as the similarity threshold.