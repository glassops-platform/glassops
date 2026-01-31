---
type: Documentation
domain: native
origin: packages/adapters/native/Dockerfile
last_modified: 2026-01-31
generated: true
source: packages/adapters/native/Dockerfile
generated_at: 2026-01-31T11:03:17.407185
hash: 84f9b60160c1b3ff063fac221639ae76919d61d1e0a8761fefbca2fe254fd551
---

# Dockerfile Documentation: Native Adapter

This document details the Dockerfile located at `F:\Github\glassops\packages\adapters\native\Dockerfile`. It describes the build process, image layers, and instructions for building and running the container.

## Base Image and Rationale

The Dockerfile employs a multi-stage build. The first stage uses `golang:1.21-alpine` as its base image. This image is selected because it provides a lightweight environment with the Go toolchain pre-installed. Alpine Linux is a security-focused distribution known for its small size, reducing the final image size and attack surface. The specific version, `1.21`, ensures a consistent and reproducible build environment. The second stage uses `alpine:3.19` for the runtime environment, further minimizing the final image size.

## Stages

The Dockerfile consists of two distinct stages:

1.  **Builder Stage:** This stage is responsible for compiling the Go application.
2.  **Runtime Stage:** This stage creates the final, minimal image containing only the compiled binary and necessary runtime dependencies.

## Key Instructions and Purpose

### Builder Stage (`golang:1.21-alpine`)

*   `FROM golang:1.21-alpine AS builder`: Defines the base image for the build stage and assigns it the alias "builder".
*   `WORKDIR /app`: Sets the working directory inside the container to `/app`. Subsequent commands will be executed from this directory.
*   `COPY go.mod go.sum ./`: Copies the `go.mod` and `go.sum` files into the `/app` directory. These files define the project's dependencies.
*   `RUN go mod download`: Downloads the Go dependencies specified in `go.mod` and `go.sum`. This step is performed before copying the source code to leverage Docker's caching mechanism. If the dependency files haven't changed, Docker will reuse the cached layer, speeding up the build process.
*   `COPY . .`: Copies the entire project source code into the `/app` directory.
*   `RUN go build -o /adapter ./cmd/main.go`: Compiles the Go application located in `cmd/main.go` and outputs the executable binary to `/adapter`.

### Runtime Stage (`alpine:3.19`)

*   `FROM alpine:3.19`: Defines the base image for the runtime stage.
*   `RUN apk add --no-cache ca-certificates git`: Installs necessary runtime dependencies using the Alpine package manager (`apk`). `ca-certificates` are required for secure HTTPS connections. `git` is included as a potential dependency for the adapter's functionality. The `--no-cache` flag minimizes image size by preventing the package manager from storing the package index locally.
*   `RUN addgroup -g 1000 glassops && adduser -D -u 1000 -G glassops glassops`: Creates a non-root user and group named `glassops` with UID and GID 1000. This enhances security by preventing the application from running as root. The `-D` flag creates a user without a home directory or shell.
*   `COPY --chown=glassops:glassops --from=builder /adapter /usr/local/bin/adapter`: Copies the compiled binary from the "builder" stage to `/usr/local/bin/adapter` in the runtime image. The `--chown=glassops:glassops` flag ensures that the binary is owned by the `glassops` user and group.
*   `USER glassops`: Specifies that subsequent commands should be executed as the `glassops` user.
*   `ENTRYPOINT ["/usr/local/bin/adapter"]`: Defines the entrypoint for the container. When the container starts, this command will be executed.

## Security Considerations

*   **Non-Root User:** The Dockerfile creates and uses a non-root user (`glassops`) to run the application, reducing the potential impact of security vulnerabilities.
*   **Minimal Base Images:** The use of Alpine Linux as both the build and runtime base images minimizes the image size and reduces the attack surface.
*   **Dependency Caching:** Caching Go dependencies speeds up builds and ensures consistent dependency versions.
*   **HTTPS Certificates:** Installation of `ca-certificates` enables secure communication over HTTPS.
*   **`--no-cache` with `apk`:** Prevents caching of package lists, reducing image size.

## Building and Running the Container

**Building the Image:**

You can build the Docker image using the following command:

```bash
docker build -t native-adapter .
```

This command builds the image and tags it as `native-adapter`. The `.` specifies that the Dockerfile is located in the current directory.

**Running the Container:**

You can run the container using the following command:

```bash
docker run -d --name my-adapter native-adapter
```

This command runs the container in detached mode (`-d`) and names it `my-adapter`. The `native-adapter` argument specifies the image to use. You may need to expose ports or mount volumes depending on the adapter's requirements.