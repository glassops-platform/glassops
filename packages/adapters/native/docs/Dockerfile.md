---
type: Documentation
domain: native
origin: packages/adapters/native/Dockerfile
last_modified: 2026-02-01
generated: true
source: packages/adapters/native/Dockerfile
generated_at: 2026-02-01T19:25:39.183180
hash: 551d4cd83700be770a2e81d208944752fc3cb9bc93606dadbcfd16f84ec8655e
---

# Dockerfile Documentation: Native Adapter

This document details the Dockerfile located at `F:\Github\glassops\packages\adapters\native\Dockerfile`. It describes the build process, image layers, and instructions for building and running the container.

## Base Image and Rationale

The Dockerfile employs a multi-stage build pattern. The initial build stage is based on `golang:1.21-alpine`. This image is selected because it provides a lightweight environment with the Go toolchain pre-installed. Alpine Linux is a security-focused distribution known for its small size, reducing the final image size and attack surface. The second stage uses `alpine:3.19` for its minimal footprint and security benefits.

## Build Stages

The Dockerfile consists of two distinct stages: `builder` and the runtime stage (unnamed).

### Stage 1: `builder`

This stage is responsible for compiling the Go application.

*   `FROM golang:1.21-alpine AS builder`: Defines the base image for this stage as `golang:1.21-alpine` and assigns it the alias `builder`.
*   `WORKDIR /app`: Sets the working directory inside the container to `/app`. All subsequent commands will be executed from this directory.
*   `COPY go.mod go.sum ./`: Copies the `go.mod` and `go.sum` files into the `/app` directory. These files define the project's dependencies.
*   `RUN go mod download`: Downloads the Go dependencies specified in `go.mod` and `go.sum`. This step is performed before copying the source code to leverage Docker's caching mechanism. If the dependency files haven't changed, Docker will reuse the cached layer, speeding up the build process.
*   `COPY . .`: Copies the entire source code into the `/app` directory.
*   `RUN go build -o /adapter ./cmd/main.go`: Compiles the Go application located in `cmd/main.go` and outputs the executable to `/adapter`.

### Stage 2: Runtime Stage

This stage creates the final runtime image.

*   `FROM alpine:3.19`: Defines the base image for this stage as `alpine:3.19`.
*   `RUN apk add --no-cache ca-certificates git nodejs npm coreutils`: Installs necessary runtime dependencies using the Alpine package manager (`apk`). `--no-cache` prevents caching of package lists, further reducing image size. `ca-certificates` are required for secure HTTPS connections. `git`, `nodejs`, `npm`, and `coreutils` are included for Salesforce CLI functionality and general utility.
*   `RUN npm install -g @salesforce/cli`: Installs the Salesforce CLI globally using npm.
*   `RUN addgroup -g 1000 glassops && adduser -D -u 1000 -G glassops glassops`: Creates a non-root user and group named `glassops` with UID and GID 1000. The `-D` flag prevents the creation of a home directory, further minimizing image size.
*   `COPY --chown=glassops:glassops --from=builder /adapter /usr/local/bin/adapter`: Copies the compiled binary from the `builder` stage to `/usr/local/bin/adapter` and changes ownership to the `glassops` user and group.
*   `USER glassops`: Specifies that subsequent commands should be executed as the `glassops` user.
*   `ENTRYPOINT ["/usr/local/bin/adapter"]`: Defines the entrypoint for the container, which is the compiled Go application.

## Key Instructions and Purpose

*   `FROM`: Specifies the base image for a stage.
*   `WORKDIR`: Sets the working directory inside the container.
*   `COPY`: Copies files or directories from the host machine to the container.
*   `RUN`: Executes commands inside the container.
*   `go mod download`: Downloads Go module dependencies.
*   `go build`: Compiles a Go application.
*   `apk add`: Installs packages using the Alpine package manager.
*   `addgroup`, `adduser`: Creates a user and group.
*   `chown`: Changes the ownership of files or directories.
*   `USER`: Specifies the user to run subsequent commands as.
*   `ENTRYPOINT`: Defines the command to be executed when the container starts.

## Security Considerations

*   **Non-Root User:** The Dockerfile creates and uses a non-root user (`glassops`) to run the application. This limits the potential damage if the application is compromised.
*   **Minimal Base Images:** Using Alpine Linux as the base image reduces the attack surface due to its small size and limited number of installed packages.
*   **Dependency Caching:** Caching Go dependencies speeds up builds and reduces the need to download dependencies repeatedly.
*   **`--no-cache` with `apk`:** Prevents caching of package lists during installation, reducing image size and ensuring the latest package versions are used.
*   **HTTPS:** The inclusion of `ca-certificates` ensures secure HTTPS connections.

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
docker run -it --rm native-adapter
```

*   `-it`: Allocates a pseudo-TTY and keeps STDIN open, allowing you to interact with the container.
*   `--rm`: Automatically removes the container when it exits.
*   `native-adapter`: Specifies the image to run.

The application will then execute as the `glassops` user within the container. You may need to provide environment variables or mount volumes depending on the application's requirements.