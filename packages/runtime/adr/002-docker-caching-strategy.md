# ADR 002: Caching Strategy in Docker Runtime

## Status

Accepted

## Context

In GitHub Actions pipelines, caching dependencies (like `node_modules` or `~/.local/share/sfdx`) is a common pattern to improve performance. The legacy TypeScript runtime (`runtime-ts`) attempted to use `@actions/cache` for this purpose.

However, the GlassOps Runtime executes as a **Docker Container Action**. This introduces significant challenges for traditional file-based caching:

1.  **Isolation**: The runtime executes inside a container, isolated from the runner's file system except for mounted volumes (typically only `GITHUB_WORKSPACE`).
2.  **API Access**: To access the GitHub Actions Cache API, the container needs the `ACTIONS_CACHE_URL` and `ACTIONS_RUNTIME_TOKEN` environment variables, which are not automatically passed to container actions for security reasons.
3.  **Ineffectiveness**: Investigation revealed that the legacy runtime often failed to access the cache API when running in Docker, effectively rendering the cache logic dead code.

## Decision

We will **NOT** implement dynamic file-based caching (e.g., caching plugins or CLI updates) within the Go Runtime at this time.

Instead, we will rely on **Immutable Docker Images** as the primary caching mechanism.

### Rationale

1.  **Immutability & Determinism**: The `Dockerfile` already pre-installs the Salesforce CLI and necessary tools. By baking dependencies into the image, we ensure that every run uses the exact same version of tools, eliminating "it works on my machine" issues caused by drift in cached plugins.
2.  **Simplicity**: Implementing a custom cache client in Go that talks to the undocumented GitHub Actions Cache API—while managing token passing—adds significant complexity to the codebase for marginal gain.
3.  **Correctness**: If a user needs specific plugins, they should be using a specific version of the GlassOps Runtime image (or a derivative image) that has those plugins installed, rather than relying on runtime installation and caching.

## Consequences

- **Performance**: Users needing heavy custom plugins that are not in the base image will incur installation time on every run.
- **Mitigation**: We will recommend users with heavy customization needs to build their own Docker image `FROM glassops/runtime` and install plugins there.
- **Cleanliness**: The runtime code remains focused on governance and orchestration, not package management.

## Alternatives

- **Use @actions/cache**: Attempted in `runtime-ts` but failed reliably in Docker contexts.
- **Pass Secrets**: We could pass `ACTIONS_RUNTIME_TOKEN` to the container, but this increases the security surface area.
- **Volume Mounting**: We could mount a host directory for caching, but this requires user configuration and breaks the "zero-config" goal.
