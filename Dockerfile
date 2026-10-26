# The glassops image: the CLI plus a pinned Salesforce CLI and sfdx-git-delta.

FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY schemas ./schemas
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" -o /out/glassops ./cmd/glassops

FROM alpine:3.22
# Alpine 3.22 ships Node 22, which the current Salesforce CLI requires.
RUN apk add --no-cache nodejs npm git coreutils ca-certificates \
    && git config --system --add safe.directory '*'
# safe.directory: CI systems mount the checkout owned by a different user than the
# container's, and git refuses to read such a repository without it.

# Keep sf's plugins and data outside $HOME. CI systems often replace $HOME inside a
# container step, which would hide plugins installed under the build user's home.
ENV SF_DATA_DIR=/usr/local/share/sf \
    SF_CACHE_DIR=/usr/local/share/sf-cache \
    SF_AUTOUPDATE_DISABLE=true \
    SF_DISABLE_TELEMETRY=true

ARG SF_CLI_VERSION=2.155.7
ARG SGD_VERSION=6.45.1
RUN npm install -g @salesforce/cli@${SF_CLI_VERSION} \
    && echo y | sf plugins install sfdx-git-delta@${SGD_VERSION} \
    && sf plugins --json > /dev/null \
    && npm cache clean --force

COPY --from=build /out/glassops /usr/local/bin/glassops
ENTRYPOINT ["glassops"]
