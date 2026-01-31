---
type: Documentation
domain: runtime-ts
origin: packages/runtime-ts/Dockerfile
last_modified: 2026-01-31
generated: true
source: packages/runtime-ts/Dockerfile
generated_at: 2026-01-31T11:09:52.887664
hash: 49486a08faa7bfdf010281f10987d601013eccb90345e92416ebac0b837b9a2b
---

# GlassOps Runtime Dockerfile Documentation

This document details the Dockerfile for the GlassOps Runtime, a TypeScript-based application. It explains each instruction, security considerations, and provides build/run instructions.

## Base Image

The base image is `node:20-alpine`. This image was selected because it provides a lightweight Node.js 20 runtime environment based on Alpine Linux. Alpine Linux is known for its small size, which results in smaller container images and reduced attack surface. The specific version, Node.js 20, ensures compatibility with the application's requirements.

## Stages

This Dockerfile employs a single-stage build. All instructions are executed within the same stage, simplifying the build process. Multi-stage builds were not deemed necessary for this application's complexity.

## Key Instructions

*   `FROM node:20-alpine`: Specifies the base image for the container.
*   `RUN apk add --no-cache ...`: Installs necessary system dependencies using the Alpine package manager (`apk`). The `--no-cache` flag minimizes image size by preventing the caching of package lists. Dependencies include `git`, `curl`, `bash`, `ca-certificates`, and `openssh-client`. `rm -rf /var/cache/apk/*` further reduces image size by removing the apk cache.
*   `WORKDIR /app`: Sets the working directory inside the container to `/app`. Subsequent commands will be executed relative to this directory.
*   `COPY package*.json ./`: Copies `package.json` and `package-lock.json` (or `package-lock.yaml`) to the working directory. This is done before copying the source code to leverage Docker's layer caching. Changes to dependencies will invalidate the cache for the dependency installation step, but changes to the source code will not.
*   `COPY tsconfig.json ./`: Copies the `tsconfig.json` file to the working directory. This file is required for the TypeScript build process.
*   `RUN npm ci --production=false`: Installs Node.js dependencies using `npm ci`. `npm ci` is preferred over `npm install` for its deterministic behavior and faster installation times. The `--production=false` flag ensures that all dependencies, including development dependencies, are installed. This is necessary for the build step.
*   `COPY src/ ./src/`: Copies the application's source code from the `src` directory on the host machine to the `src` directory inside the container.
*   `RUN npm run build`: Executes the build script defined in `package.json`. This script is expected to compile the TypeScript code into JavaScript.
*   `RUN addgroup -S glassops && adduser -S glassops -G glassops`: Creates a non-root user named `glassops` and a group also named `glassops`. The `-S` flag creates a system account, and `-G glassops` adds the user to the `glassops` group.
*   `RUN chown -R glassops:glassops /app`: Changes the ownership of the `/app` directory and all its contents to the `glassops` user and group.
*   `USER glassops`: Switches the user context to `glassops` for subsequent commands. This enhances security by preventing the application from running as root.
*   `ENV NODE_ENV=production`: Sets the `NODE_ENV` environment variable to `production`. This is a common practice to optimize the application for production environments.
*   `ENV PATH="/app/node_modules/.bin:${PATH}"`: Modifies the `PATH` environment variable to include the `node_modules/.bin` directory. This allows the application to execute locally installed binaries without specifying their full path.
*   `ENTRYPOINT ["node", "/app/dist/index.js"]`: Defines the entry point for the container. When the container starts, this command will be executed. It runs the compiled JavaScript application located at `/app/dist/index.js` using Node.js.

## Security Considerations

*   **Non-Root User:** The Dockerfile creates and switches to a non-root user (`glassops`) to minimize the impact of potential security vulnerabilities. Running the application as a non-root user reduces the risk of privilege escalation.
*   **Minimal Base Image:** Using `node:20-alpine` provides a smaller attack surface compared to larger base images.
*   **Package Cache Removal:** Removing the package cache (`/var/cache/apk/*`) after installing dependencies reduces the image size and potential exposure of sensitive information.
*   **Dependency Management:** Using `npm ci` ensures a consistent and reproducible build process.
*   **Regular Updates:** We recommend regularly updating the base image and dependencies to address security vulnerabilities.

## Build and Run Instructions

**Build:**

You can build the container image using the following command:

```bash
docker build -t glassops-runtime .
```

This command builds an image tagged as `glassops-runtime` from the Dockerfile in the current directory.

**Run:**

You can run the container using the following command:

```bash
docker run -d -p 3000:3000 glassops-runtime
```

This command runs the `glassops-runtime` image in detached mode (`-d`) and maps port 3000 on the host machine to port 3000 inside the container (`-p 3000:3000`). Adjust the port mapping as needed for your application.