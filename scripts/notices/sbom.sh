#!/usr/bin/env bash
# Writes a syft JSON SBOM of a container image.
#   scripts/notices/sbom.sh <image> <out.json>
# <image> is a local image name or a registry reference. The syft image is pinned, and a local
# image is read through the Docker socket.
set -euo pipefail
image="${1:?usage: sbom.sh <image> <out.json>}"
out="${2:?usage: sbom.sh <image> <out.json>}"
SYFT_IMAGE="${SYFT_IMAGE:-anchore/syft:v1.55.0}"
export MSYS_NO_PATHCONV=1
if docker image inspect "$image" >/dev/null 2>&1; then
  source="docker:$image"
else
  source="registry:$image"
fi
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock "$SYFT_IMAGE" "$source" -o syft-json -q > "$out"
