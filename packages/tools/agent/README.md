# GlassOps Agent 🚀

An extensible, AI-powered agent for automating documentation, metadata validation, and repository management across the Salesforce DevOps lifecycle.

## Overview

GlassOps Agent uses **Gemma 3 (gemma-3-4b-it)** to analyze repository assets and generate high-quality, intent-preserving documentation. It is built for resilience, handling large files via **Semantic Greedy Chunking** and respecting strict API rate limits.

## Key Features

- **Principal Architect Persona:** Generates professional, coherent documents suitable for both technical and non-technical stakeholders.
- **Resilient Rate Limiting:**
    - Calibrated for **30 RPM / 15k TPM**.
    - Automatic **Incremental Backoff** (10s, 30s, 60s) for `429 RESOURCE_EXHAUSTED` errors.
- **Semantic Greedy Chunking:** Automatically splits large files (>10KB) into semantic sections and joins them into a single cohesive output.
- **Structured Output:** Automatically organizes documentation into `markdown/<ext>/` subdirectories at the project root.
- **Extensible Adapters:** Pluggable logic for different file formats.

## Supported Formats

- **TypeScript/JavaScript (`.ts`, `.js`)**: Documentation of classes, functions, and types.
- **Markdown (`.md`)**: Intent-preserving refinement of technical documentation.
- **YAML (`.yml`)**: Validation and explanation of configuration files.
- **JSON (`.json`)**: Analysis of schemas and data structures.

## Setup

1.  Ensure you have a `GOOGLE_API_KEY` in your `.env` file.
2.  Install dependencies:
    ```bash
    npm install
    ```
3.  Build the agent:
    ```bash
    npm run build
    ```

## Usage

Run the agent via root convenience scripts or directly from this directory.

### Basic Generation (Root)

```bash
# Uses default patterns: packages/**/*.ts, docs/**/*.md, packages/**/*.yml, packages/**/*.json
npm run docs:ai
```

### Targeted Patterns (Positional)

```bash
# Generate docs for specific paths
npm run docs:ai -- src/llm-client.ts docs/architecture/
```

## Architecture

- **Core Orchestrator**: Manages file scanning (`Scanner`), sequential processing, and results aggregation.
- **Adapters**: Modules in `src/adapters/` that implement the `AgentAdapter` interface for parsing and prompt generation.
- **LLM Client**: Handles rate-limited communication and backoff logic.
