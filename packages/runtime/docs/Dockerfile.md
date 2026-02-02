---
type: Documentation
domain: runtime
origin: packages/runtime/Dockerfile
last_modified: 2026-02-01
generated: true
source: packages/runtime/Dockerfile
generated_at: 2026-02-01T19:38:24.441058
hash: 335afd8b950d1cbbdaa8c07079af1c61cb68b2480aa25df0e978ec233dcde8fd
---

# Dockerfile Documentation: glassops Runtime

This document details the Dockerfile located at `F:\Github\nobleforge\glassops\packages\runtime\Dockerfile`. It describes the build process, image layers, and instructions for building and running the container.

## Base Image and Rationale

The Dockerfile employs a multi-stage build. The first stage uses `golang:1.25-alpine` as its base image. This image is selected because it provides a lightweight environment with the Go toolchain pre-installed. Alpine Linux is a security-focused distribution known for its small size, reducing the final image size and attack surface. The version `1.25` specifies a particular Go version, ensuring build reproducibility.

The second stage uses `alpine:3.19`. This image provides an even smaller base for the runtime environment, containing only the necessary dependencies to execute the compiled Go binary.

## Stages

The Dockerfile consists of two stages: `builder` and the final runtime stage.

### Builder Stage

This stage is responsible for compiling the Go application.

*   **`FROM golang:1.25-alpine AS builder`**: Defines the base image for this stage as `golang:1.25-alpine` and assigns it the alias `builder`.
*   **`WORKDIR /app`**: Sets the working directory inside the container to `/app`. Subsequent commands will be executed from this directory.
*   **`COPY go.mod go.sum* ./`**: Copies the `go.mod` and `go.sum` files to the `/app` directory. These files define the project's dependencies.
*   **`RUN go mod download || true`**: Downloads the Go dependencies specified in `go.mod`. The `|| true` ensures the build doesn't fail if the download fails (e.g., if dependencies are already cached).
*   **`COPY . .`**: Copies the entire project source code to the `/app` directory.
*   **`RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /glassops ./cmd/glassops`**: Compiles the Go application.
    *   `CGO_ENABLED=0`: Disables CGO, resulting in a statically linked binary and reducing dependencies.
    *   `GOOS=linux`: Sets the target operating system to Linux.
    *   `-ldflags="-s -w"`:  Removes debugging information and symbol table from the binary, further reducing its size.
    *   `-o /glassops`: Specifies the output path and name of the compiled binary as `/glassops`.
    *   `./cmd/glassops`: Specifies the main package to build.

### Runtime Stage

This stage creates the final container image with only the necessary runtime dependencies.

*   **`FROM alpine:3.19`**: Defines the base image for this stage as `alpine:3.19`.
*   **`RUN apk add --no-cache nodejs npm git coreutils`**: Installs required packages using the Alpine package manager (`apk`).
    *   `nodejs` and `npm`: Required for installing the Salesforce CLI.
    *   `git`: May be required by the Salesforce CLI or the application itself.
    *   `coreutils`: Provides essential utilities, including `env` with `-S` support, which is needed by the Salesforce CLI. The `--no-cache` flag minimizes image size by preventing apk from storing package caches.
*   **`RUN npm install -g @salesforce/cli`**: Installs the Salesforce CLI globally using npm.
*   **`COPY --from=builder /glassops /usr/local/bin/glassops`**: Copies the compiled Go binary from the `builder` stage to `/usr/local/bin/glassops` in the final image.
*   **`ENTRYPOINT ["/usr/local/bin/glassops"]`**: Sets the entrypoint for the container. When the container starts, it will execute the `/usr/local/bin/glassops` binary.

## Key Instructions and Purpose

*   **`FROM`**: Specifies the base image for a stage.
*   **`WORKDIR`**: Sets the working directory for subsequent instructions.
*   **`COPY`**: Copies files or directories from the host machine to the container.
*   **`RUN`**: Executes commands inside the container during the build process.
*   **`ENTRYPOINT`**: Configures the container to run as an executable.

## Security Considerations

*   **Alpine Linux**: Using Alpine Linux as the base image reduces the attack surface due to its minimal package set.
*   **Static Linking**: Disabling CGO and statically linking the Go binary reduces external dependencies and potential vulnerabilities.
*   **Package Management**: Using `apk --no-cache` minimizes the image size and reduces the risk of outdated packages.
*   **Least Privilege**: The runtime stage only includes the necessary dependencies for running the application, adhering to the principle of least privilege.
*   **Regular Updates**: We recommend regularly rebuilding the image to incorporate the latest security patches from the base images and dependencies.

## Building and Running the Container

**Building the Image:**

You can build the Docker image using the following command:

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
*   `glassops`: Specifies the image to run.

You may need to add volume mounts or environment variables to the `docker run` command depending on the application's requirements.