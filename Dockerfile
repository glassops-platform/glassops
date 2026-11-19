# The glassops image: the CLI plus a pinned Salesforce CLI and sfdx-git-delta.

FROM golang:1.27.4-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY schemas ./schemas
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" -o /out/glassops ./cmd/glassops

FROM alpine:3.24
# Alpine 3.24 ships Node 24, which the current Salesforce CLI supports.
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

# cosign signs and verifies PERMITs keyless (policy.permit.signing: sigstore), and opa runs
# opt-in Rego rules (policy.rego). Both are separate Apache-2.0 programs, copied from their
# official images pinned by digest.
COPY --from=gcr.io/projectsigstore/cosign:v2.5.3@sha256:f1946d0f30fc8e3777b02f2201e02efdba9fe38f4918162f937052fac98e083f /ko-app/cosign /usr/local/bin/cosign
COPY --from=openpolicyagent/opa:1.22.0-static@sha256:906bfc9904b2693733e7959d71773dd073e6c135cf06f26d955abe94603e82b3 /opa /usr/local/bin/opa
COPY --from=build /out/glassops /usr/local/bin/glassops
COPY LICENSE THIRD_PARTY_NOTICES /usr/share/doc/glassops/
ENTRYPOINT ["glassops"]
