---
type: Documentation
domain: knowledge
origin: packages/knowledge/drift/detect_drift.py
last_modified: 2026-01-31
generated: true
source: packages/knowledge/drift/detect_drift.py
generated_at: 2026-01-31T09:47:39.302149
hash: dee44b64dd3a391ae28d74379a8cb4f62cf1733a64945e310fcbe0a6ce6a6ac2
---

## Knowledge Drift Detection Documentation

This module is responsible for comparing newly ingested document embeddings against a baseline to identify potential drift in the knowledge base. Drift refers to changes in the underlying data distribution, which can negatively impact the performance of retrieval-augmented generation (RAG) systems. Currently, the implementation simulates drift detection by identifying near-duplicate documents, indicating potential redundancy or conflicts.

### Key Components

**1. `cosine_similarity(a, b)` Function:**

   - **Purpose:** Calculates the cosine similarity between two vectors `a` and `b`.
   - **Parameters:**
     - `a`: A NumPy array representing the first vector.
     - `b`: A NumPy array representing the second vector.
   - **Return Value:** A float representing the cosine similarity score between the two vectors.  A higher score indicates greater similarity.
   - **Type Hints:** `a: np.ndarray`, `b: np.ndarray` -> `float`

**2. `detect_drift(embeddings, threshold=0.85)` Function:**

   - **Purpose:**  Analyzes a list of document embeddings to detect potential drift.  Currently, it focuses on identifying near-duplicate documents as a proxy for drift.
   - **Parameters:**
     - `embeddings`: A list of tuples, where each tuple contains a document dictionary (`doc_dict`) and its corresponding embedding vector.
     - `threshold`: (Optional) A float representing the similarity threshold for drift detection. This parameter is not currently used in the near-duplicate detection logic but is reserved for future semantic drift comparison. Defaults to 0.85.
   - **Return Value:** A list of document paths that are identified as having drifted (currently, this list is empty as drift is simulated).
   - **Type Hints:** `embeddings: list[tuple[dict, np.ndarray]]`, `threshold: float = 0.85` -> `list[str]`

### Design and Operation

The `detect_drift` function operates as follows:

1. **Report Generation:** It creates a markdown report file (`drift_report.md`) located in the `docs/generated` directory. This report summarizes the findings of the drift detection process.

2. **Near-Duplicate Detection:** It iterates through the provided `embeddings`, using document hashes to identify potential duplicates.  A dictionary `seen_hashes` stores the path of the first document encountered for each unique hash. If a document with the same hash is encountered again, it's flagged as a potential conflict.

3. **Report Content:** The generated report includes:
   - A header indicating it's a Knowledge Base Health Report.
   - The name of the script generating the report.
   - A section detailing any detected near-duplicate documents, listing the paths of the conflicting files.
   - A section stating that no significant semantic drift was detected (as the current implementation focuses on duplicates).

4. **Future Enhancements:** The `threshold` parameter and the `cosine_similarity` function are included to facilitate future implementation of true semantic drift detection, where embeddings are compared against a baseline snapshot.  The current implementation serves as a placeholder and provides a basic conflict report.

### Notable Patterns

- **Report-Driven Approach:** The module generates a human-readable report to communicate the results of the drift detection process. This is intended to make the findings accessible to both technical and non-technical stakeholders.
- **Hash-Based Duplicate Detection:** The use of document hashes provides an efficient way to identify near-duplicate content.
- **Type Hints:** The use of type hints improves code readability and maintainability, and allows for static analysis.
- **Modular Design:** The `cosine_similarity` function is separated from the `detect_drift` function to promote code reuse and testability.