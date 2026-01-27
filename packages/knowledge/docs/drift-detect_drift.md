---
type: Documentation
domain: knowledge
origin: packages/knowledge/drift/detect_drift.py
last_modified: 2026-01-26
generated: true
source: packages/knowledge/drift/detect_drift.py
generated_at: 2026-01-26T14:08:13.981Z
hash: d0dd56b3a7fe7c647aea0ef51d8bd8e42c2c528c93402eb42b8f8ce217be0232
---

## Knowledge Base Drift Detection Documentation

**1. Introduction**

This document details the functionality of the `detect_drift.py` module, responsible for identifying changes (drift) in a knowledge base. Drift detection is crucial for maintaining the relevance and accuracy of Retrieval-Augmented Generation (RAG) systems. This module currently focuses on identifying duplicate or near-duplicate content as a proxy for drift, and generates a health report summarizing its findings.

**2. Functionality**

The core function, `detect_drift`, analyzes a set of document embeddings to identify potential drift.  Currently, the implementation simulates drift detection by identifying documents with identical content (hashes).  Future iterations will incorporate comparison against historical embedding snapshots to detect semantic changes.

**3. Technical Details**

*   **Input:** The `detect_drift` function accepts a list of tuples. Each tuple contains a document dictionary (`doc_dict`) and its corresponding embedding vector (`embedding_vector`). The `doc_dict` is expected to contain at least a `"path"` key (representing the document's file path) and a `"hash"` key (representing a hash of the document's content).
*   **Drift Detection (Current Implementation):** The module calculates a hash of each document's content. If a hash is encountered more than once, the corresponding documents are flagged as potential conflicts.
*   **Cosine Similarity (Helper Function):** The `cosine_similarity` function calculates the cosine similarity between two vectors. While currently unused in the drift detection logic, it is included for potential use in future semantic drift comparison algorithms.
*   **Reporting:** A "Knowledge Base Health Report" is generated as a Markdown file (`drift_report.md`) located in the `docs/generated` directory. This report details:
    *   Generation timestamp and source file.
    *   Identified content conflicts (duplicate documents).
    *   A placeholder drift status indicating no significant semantic drift has been detected (due to the current implementation).
*   **Output:** The `detect_drift` function currently returns an empty list (`drifted`). In future versions, this will return a list of document paths identified as having drifted significantly.

**4. Future Enhancements**

The following enhancements are planned:

*   **Historical Snapshot Comparison:** Implement comparison of current embeddings against previously stored embedding snapshots to detect semantic drift.
*   **Similarity Thresholding:** Utilize cosine similarity (or other similarity metrics) to identify documents with embeddings that have diverged beyond a defined threshold.
*   **Drift Severity Assessment:**  Categorize drift based on its severity (e.g., minor, moderate, significant).
*   **Automated Remediation:**  Integrate with workflows to automatically address detected drift (e.g., flagging documents for review, triggering re-indexing).

**5. Dependencies**

*   `numpy`: For numerical operations, specifically vector calculations.
*   `os`: For file system operations, such as creating directories and writing files.