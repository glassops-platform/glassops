---
type: Documentation
domain: global
origin: scripts/build-docs.py
last_modified: 2026-01-26
generated: true
source: scripts/build-docs.py
generated_at: 2026-01-26T14:04:41.143Z
hash: acab8701c114c56d2be9071c2277f0c5d188635fec2be7dd91374e42096b8fbf
---

# Documentation Build Process: `build-docs.py`

This document details the functionality of the `build-docs.py` script, which automates the generation of documentation for the Glassops project. The script prepares source files, configures the MkDocs documentation generator, and builds the final static site.

## Overview

The script’s primary function is to take source documentation files, configure them for MkDocs, and generate a static website hosted in the `glassops_site` directory. It handles file copying, configuration adjustments, link fixing, and the execution of the MkDocs build process.  The script is designed to be robust, handling potential file system issues and providing informative output.

## Core Functionality

1.  **Directory Setup:**
    *   Defines key directories: `root_dir` (project root), `config_dir`, `staging_dir`, and `output_dir`.
    *   Cleans existing `staging_dir` and `output_dir` to ensure a fresh build.  Includes error handling for read-only files on Windows.
    *   Creates the `staging_dir` and a `content` subdirectory within it.

2.  **File Staging:**
    *   Copies specified documentation source files and directories (defined in `copy_list`) from the `root_dir` to the `content` directory within `staging_dir`.
    *   Uses an allow-list approach for copying, ensuring only explicitly defined files are included.
    *   Ignores `.git`, `node_modules`, and `.DS_Store` directories during the copy process.
    *   Logs warnings if source items are not found.
    *   Copies the `mkdocs.yml` configuration file from `config_dir` to the root of `staging_dir`.

3.  **Configuration Modification:**
    *   Modifies the `mkdocs.yml` file within the `staging_dir` to:
        *   Set `docs_dir` to `content`.
        *   Set `site_dir` to `../glassops_site` (relative path to the output directory).

4.  **Post-Processing & Link Correction:**
    *   **ADR Index Generation:** If an `index.md` or `README.md` does not exist in the `docs/adr` directory, it automatically generates a `README.md` file containing an index of all Architectural Decision Records (ADRs).
    *   **README/index.md Collision Resolution:** If both `README.md` and `index.md` exist in the same directory, `README.md` is renamed to `overview.md` to avoid conflicts.
    *   **Link Fixing:** Iterates through all Markdown (`.md`) files in the `content` directory and performs the following:
        *   **GitHub Alert Conversion:** Converts GitHub-style alerts (e.g., `> [!NOTE]`) to MkDocs-compatible admonitions (e.g., `!!! note`).  Supports `note`, `tip`, `important`, `warning`, and `caution` alert types.
        *   **Directory Link Correction:**  Corrects links to ADRs to point to `README.md` within each ADR directory.
        *   **Legacy Link Fixes:** Replaces outdated links to `index.md` with `README.md`.

5.  **MkDocs Build Execution:**
    *   Executes the `mkdocs build` command within the `staging_dir` to generate the static site.
    *   Captures and prints the standard output and standard error from the `mkdocs build` process for debugging purposes.
    *   Exits with a non-zero exit code if the build fails.

## Dependencies

*   Python 3.6 or higher
*   MkDocs
*   `pathlib` (standard library)
*   `shutil` (standard library)
*   `subprocess` (standard library)
*   `re` (standard library)

## Error Handling

The script includes error handling for:

*   Read-only files on Windows during directory cleanup.
*   Missing source files during the staging process.
*   Errors during Markdown file processing.
*   Failures during the `mkdocs build` process.

## Output

The generated documentation website is located in the `glassops_site` directory. The script provides informative output to the console during each stage of the process, including warnings and error messages.