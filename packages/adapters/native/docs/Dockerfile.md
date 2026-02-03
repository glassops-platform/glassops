---
type: Documentation
domain: native
last_modified: 2026-02-02
generated: true
source: packages/adapters/native/Dockerfile
generated_at: 2026-02-02T22:22:25.186378
hash: 551d4cd83700be770a2e81d208944752fc3cb9bc93606dadbcfd16f84ec8655e
---

# Dockerfile Documentation: Native Adapter

This document details the Dockerfile located at `F:\Github\glassops\packages\adapters\native\Dockerfile`. It describes the build process, image layers, and instructions for building and running the container.

## Base Image and Rationale

The Dockerfile employs a multi-stage build. The initial build stage is based on `golang:1.21-alpine`. This image is selected because it provides a lightweight environment with the Go toolchain pre-installed. Alpine Linux is a security-focused distribution known for its small size, reducing the final image size and attack surface. The second stage uses `alpine:3.19` for its minimal footprint and security benefits.

## Build Stages

The Dockerfile consists of two stages: `builder` and the final runtime stage.

### Stage 1: `builder`

This stage is responsible for compiling the Go application.

*   `FROM golang:1.21-alpine AS builder`: Defines the base image for this stage as `golang:1.21-alpine` and assigns it the alias `builder`.
*   `WORKDIR /app`: Sets the working directory inside the container to `/app`.
*   `COPY go.mod go.sum ./`: Copies the `go.mod` and `go.sum` files to the working directory. These files define the project's dependencies.
*   `RUN go mod download`: Downloads the Go dependencies specified in `go.mod` and `go.sum`. This step is performed before copying the source code to leverage Docker's caching mechanism. If the dependency files haven't changed, this layer will be cached, speeding up subsequent builds.
*   `COPY . .`: Copies the entire source code to the working directory.
*   `RUN go build -o /adapter ./cmd/main.go`: Compiles the Go application located in `cmd/main.go` and outputs the executable to `/adapter`.

### Stage 2: Runtime Stage

This stage creates the final container image with only the necessary runtime dependencies.

*   `FROM alpine:3.19`: Defines the base image for this stage as `alpine:3.19`.
*   `RUN apk add --no-cache ca-certificates git nodejs npm coreutils`: Installs required runtime dependencies using the Alpine package manager (`apk`). `--no-cache` prevents caching of package lists, further reducing image size. `ca-certificates` are essential for secure HTTPS connections. `git`, `nodejs`, `npm`, and `coreutils` are added to support the Salesforce CLI and any related functionalities.
*   `RUN npm install -g @salesforce/cli`: Installs the Salesforce CLI globally using npm.
*   `RUN addgroup -g 1000 glassops && adduser -D -u 1000 -G glassops glassops`: Creates a non-root user named `glassops` with user ID (UID) and group ID (GID) of 1000. The `-D` flag prevents the creation of a home directory, further minimizing the image size. This enhances security by avoiding running the application as root.
*   `COPY --chown=glassops:glassops --from=builder /adapter /usr/local/bin/adapter`: Copies the compiled binary from the `builder` stage to `/usr/local/bin/adapter` and changes the ownership to the `glassops` user and group.
*   `USER glassops`: Specifies that the container should run as the `glassops` user.
*   `ENTRYPOINT ["/usr/local/bin/adapter"]`: Defines the entry point for the container, which is the compiled Go application.

## Key Instructions and Purpose

*   `FROM`: Specifies the base image for each stage.
*   `WORKDIR`: Sets the working directory inside the container.
*   `COPY`: Copies files and directories from the host machine to the container.
*   `RUN`: Executes commands inside the container.
*   `go mod download`: Downloads Go dependencies.
*   `go build`: Compiles the Go application.
*   `apk add`: Installs packages using the Alpine package manager.
*   `addgroup`, `adduser`: Creates a user and group.
*   `chown`: Changes the ownership of files and directories.
*   `USER`: Specifies the user to run the container as.
*   `ENTRYPOINT`: Defines the command to execute when the container starts.

## Security Considerations

*   **Non-Root User:** The application runs as a non-root user (`glassops`), reducing the potential impact of security vulnerabilities.
*   **Minimal Base Images:** Alpine Linux is used as the base image due to its small size and security focus.
*   **Dependency Caching:** Caching Go dependencies speeds up builds and reduces the risk of downloading malicious packages.
*   **`--no-cache` with `apk`:** Prevents caching of package lists during installation, reducing image size and potential attack surface.
*   **HTTPS:** The inclusion of `ca-certificates` ensures secure HTTPS connections.

## Building and Running the Container

**Building the image:**

You can build the Docker image using the following command:

```bash
docker build -t native-adapter .
```

This command builds the image and tags it as `native-adapter`.

**Running the container:**

You can run the container using the following command:

```bash
docker run -it --rm native-adapter
```

This command runs the container in interactive mode (`-it`), removes the container after it exits (`--rm`), and uses the `native-adapter` image. You may need to add volume mounts or environment variables depending on the application's requirements.