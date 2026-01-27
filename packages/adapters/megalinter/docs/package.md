---
type: Documentation
domain: megalinter
origin: packages/adapters/megalinter/package.json
last_modified: 2026-01-26
generated: true
source: packages/adapters/megalinter/package.json
generated_at: 2026-01-26T05:05:45.416Z
hash: c1f7a117215c40362d4d920eb79766b1ee165f443acb1d0c6cc773018e610d8c
---

# `@glassops/megalinter-adapter` Package Documentation

This document details the structure and purpose of the `@glassops/megalinter-adapter` package, as defined by its `package.json` file.

## Overview

This package serves as an adapter between the MegaLinter static analysis tool and the GlassOps platform. It facilitates the integration of MegaLinter’s findings into the GlassOps governance workflow.  Specifically, it transforms MegaLinter's output (likely in a format like JSON or YAML) into the SARIF (Static Analysis Results Interchange Format) standard, which GlassOps utilizes for vulnerability and code quality analysis.

## Schema Breakdown

The `package.json` file defines metadata about the package.  The following fields are present:

| Field          | Type     | Required | Description