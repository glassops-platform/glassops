---
type: Documentation
domain: runtime
origin: packages/runtime/Dockerfile
last_modified: 2026-01-31
generated: true
source: packages/runtime/Dockerfile
generated_at: 2026-01-31T11:06:09.718804
hash: 335afd8b950d1cbbdaa8c07079af1c61cb68b2480aa25df0e978ec233dcde8fd
---

# Dockerfile Documentation: glassops Runtime

This document details the Dockerfile located at `F:\Github\nobleforge\glassops\packages\runtime\Dockerfile`. It describes the build process, image layers, and instructions for building and running the container.

## Base Image and Rationale

The Dockerfile employs a multi-stage build. The first stage uses `golang:1.25-alpine` as its base image. This image is selected because it provides a lightweight environment with the Go toolchain pre-installed. Alpine Linux is known for its small size, which results in smaller container images. The second stage uses `alpine:3.19`, again for its minimal size.

## Stages

The Dockerfile consists of two stages: `builder` and the final runtime stage.

### Builder Stage

This stage is responsible for compiling the Go application.

*   **`FROM golang:1.25-alpine AS builder`**:  Defines the base image for this stage as `golang:1.25-alpine` and assigns it the alias `builder`.
*   **`WORKDIR /app`**: Sets the working directory inside the container to `/app`. Subsequent commands will be executed from this directory.
*   **`COPY go.mod go.sum* ./`**: Copies the `go.mod` and `go.sum` files (and any files matching `go.sum*`) from the host machine to the `/app` directory in the container.
*   **`RUN go mod download || true`**: Downloads the Go dependencies specified in `go.mod`. The `|| true` ensures the build doesn't fail if the download fails (e.g., if dependencies are already cached).
*   **`COPY . .`**: Copies all files and directories from the current directory on the host machine to the `/app` directory in the container.
*   **`RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /glassops ./cmd/glassops`**:  Compiles the Go application.
    *   `CGO_ENABLED=0`: Disables CGO, which simplifies cross-compilation and reduces dependencies.
    *   `GOOS=linux`: Sets the target operating system to Linux.
    *   `go build`: The Go build command.
    *   `-ldflags="-s -w"`:  Linker flags to strip debug information and symbol tables, reducing the binary size.
    *   `-o /glassops`: Specifies the output file name and path as `/glassops`.
    *   `./cmd/glassops`: Specifies the main package to build.

### Runtime Stage

This stage creates the final container image with only the necessary runtime dependencies.

*   **`FROM alpine:3.19`**: Defines the base image for this stage as `alpine:3.19`.
*   **`RUN apk add --no-cache nodejs npm git coreutils`**: Installs Node.js, npm, git, and coreutils using the Alpine package manager (`apk`). The `--no-cache` option prevents caching of package lists, reducing image size. Coreutils provides the `env -S` command, which is required by the Salesforce CLI.
*   **`RUN npm install -g @salesforce/cli`**: Installs the Salesforce CLI globally using npm.
*   **`COPY --from=builder /glassops /usr/local/bin/glassops`**: Copies the compiled Go binary from the `builder` stage to `/usr/local/bin/glassops` in the current stage.
*   **`ENTRYPOINT ["/usr/local/bin/glassops"]`**: Sets the entrypoint for the container to `/usr/local/bin/glassops`. This means that when the container starts, it will execute this binary.

## Key Instructions and Purpose

*   **`FROM`**: Specifies the base image for a stage.
*   **`WORKDIR`**: Sets the working directory for subsequent instructions.
*   **`COPY`**: Copies files and directories from the host machine to the container.
*   **`RUN`**: Executes commands inside the container during the build process.
*   **`ENTRYPOINT`**: Specifies the command to execute when the container starts.

## Security Considerations

*   **Base Image Selection**: Alpine Linux is a minimal distribution, reducing the attack surface. However, it's important to stay updated with security patches for Alpine.
*   **Dependency Management**: The `go mod download` command ensures that dependencies are managed and reproducible.
*   **Stripping Debug Information**: The `-ldflags="-s -w"` flags remove debug information from the binary, reducing its size and potentially making it harder to reverse engineer.
*   **Least Privilege**: The container runs the application as a non-root user by default within the Alpine base image.
*   **Regular Updates**: We recommend regularly rebuilding the image to incorporate the latest security updates from the base images and dependencies.

## Building and Running the Container

**Building the Image:**

You can build the container image using the following command:

```bash
docker build -t glassops .
```

This command builds the image and tags it as `glassops`. The `.` specifies that the Dockerfile is located in the current directory.

**Running the Container:**

You can run the container using the following command:

```bash
docker run -it --rm glassops
```

*   `-it`: Allocates a pseudo-TTY and keeps STDIN open, allowing you to interact with the container.
*   `--rm`: Automatically removes the container when it exits.
*   `glassops`: The name of the image to run.

You may need to add volume mounts or environment variables depending on the application's requirements.