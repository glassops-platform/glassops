---
type: Documentation
domain: runtime
last_modified: 2026-02-02
generated: true
source: packages/runtime/Dockerfile
generated_at: 2026-02-02T22:34:30.264859
hash: 335afd8b950d1cbbdaa8c07079af1c61cb68b2480aa25df0e978ec233dcde8fd
---

# Dockerfile Documentation: glassops Runtime

## Overview

This Dockerfile defines the build and runtime environment for the `glassops` application, a Go-based tool. It employs a multi-stage build to minimize the final image size and improve security. We aim to provide a self-contained environment for running `glassops` with the Salesforce CLI.

## Base Images

*   **`golang:1.25-alpine` (Builder Stage):** This image is based on Alpine Linux and includes the Go 1.25 toolchain. Alpine is chosen for its small size, reducing the build image footprint.
*   **`alpine:3.19` (Runtime Stage):**  This image is also based on Alpine Linux, providing a minimal base for the runtime environment. Its small size contributes to a smaller final image.

## Stages

The Dockerfile consists of two stages:

1.  **Builder Stage:** Responsible for compiling the Go application.
2.  **Runtime Stage:**  Responsible for creating the final image containing only the compiled binary and necessary runtime dependencies.

## Key Instructions and Purpose

*   **`FROM <image> AS <name>`:** Defines the base image for a stage and assigns it a name for referencing in later stages.
*   **`WORKDIR /app`:** Sets the working directory inside the container. Subsequent commands will be executed from this directory.
*   **`COPY <src> <dest>`:** Copies files or directories from the host machine to the container's filesystem.
*   **`RUN <command>`:** Executes a command inside the container during the build process.
*   **`go mod download`:** Downloads the Go module dependencies defined in `go.mod` and `go.sum`. The `|| true` ensures the build doesn't fail if the download fails (e.g., due to network issues) during initial setup.
*   **`CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /glassops ./cmd/glassops`:** Compiles the Go application.
    *   `CGO_ENABLED=0`: Disables CGO, resulting in a more portable binary.
    *   `GOOS=linux`: Sets the target operating system to Linux.
    *   `-ldflags="-s -w"`: Strips debug information and symbol table from the binary, reducing its size.
    *   `-o /glassops`: Specifies the output path and name of the compiled binary.
    *   `./cmd/glassops`: Specifies the main package to build.
*   **`apk add --no-cache nodejs npm git coreutils`:** Installs Node.js, npm, git, and coreutils within the runtime stage. `--no-cache` minimizes image size by not storing the package cache. Coreutils provides the `env -S` command, required by the Salesforce CLI.
*   **`npm install -g @salesforce/cli`:** Installs the Salesforce CLI globally using npm.
*   **`COPY --from=builder /glassops /usr/local/bin/glassops`:** Copies the compiled binary from the builder stage to the runtime stage.
*   **`ENTRYPOINT ["/usr/local/bin/glassops"]`:** Defines the command that will be executed when the container starts.

## Security Considerations

*   **Minimal Base Images:** Using Alpine Linux as the base image reduces the attack surface due to its small size and limited number of installed packages.
*   **Stripped Binary:** Removing debug information from the binary using `-ldflags="-s -w"` reduces the amount of potentially sensitive information included in the image.
*   **No Root User:** The Dockerfile does not explicitly create or switch to a root user. The application will run as the default user within the Alpine image.
*   **Package Management:** Using `apk add --no-cache` avoids caching package lists, reducing the risk of outdated or vulnerable packages.
*   **Dependency Management:** The `go.mod` and `go.sum` files ensure reproducible builds and help prevent dependency-related vulnerabilities.

## Building and Running the Container

1.  **Build the image:**

    You can build the Docker image using the following command from the directory containing the Dockerfile:

    ```bash
    docker build -t glassops-runtime .
    ```

2.  **Run the container:**

    You can run the container using the following command:

    ```bash
    docker run -it --rm glassops-runtime
    ```

    *   `-it`: Allocates a pseudo-TTY and keeps STDIN open, allowing you to interact with the container.
    *   `--rm`: Automatically removes the container when it exits.
    *   `glassops-runtime`: The name of the image to run.

You may need to configure the Salesforce CLI within the container after it starts, using `sf config set ...`.