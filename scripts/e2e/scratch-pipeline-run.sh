#!/bin/sh
# Runs the end-to-end pipeline once, on the runner, in one container. One scratch org at most.
#
#   sh scratch-pipeline-run.sh REPO IMAGE
#
# It builds the two-commit test repository in REPO (which must not exist), installs the e2e
# config, the scratch definition and the settings check, then runs "glassops run e2e" in IMAGE.
# The pipeline plans a delta that deletes a field, creates a scratch org from the definition,
# checks the org's settings, validates with the minimal seed, gates, deploys, verifies the
# deploy, tests with the coverage gate, runs an exec step, reports and deletes the org.
# Needs SF_JWT_KEY, SF_CLIENT_ID, SF_USERNAME, SF_DEVHUB_ORG_ID and GLASSOPS_PERMIT_KEY in the
# environment, passed to the container by name. Exits with the pipeline's exit code. Run it
# from the repository root. Run scratch-pipeline-check.sh on REPO/.glassops/results afterwards.
set -eu

REPO=${1:?usage: scratch-pipeline-run.sh REPO IMAGE}
IMAGE=${2:?usage: scratch-pipeline-run.sh REPO IMAGE}
HERE=$(CDPATH="" cd -- "$(dirname -- "$0")" && pwd)
ROOT=$(CDPATH="" cd -- "$HERE/../.." && pwd)

sh "$ROOT/testdata/sfdx/build-repo.sh" "$REPO"
mkdir -p "$REPO/config"
cp "$ROOT/testdata/e2e/glassops-run.yaml" "$REPO/glassops.yaml"
cp "$ROOT/testdata/e2e/scratch-def.json" "$REPO/config/e2e-scratch-def.json"
cp "$HERE/scratch-settings-check.sh" "$REPO/scratch-settings-check.sh"
sed -i 's|^    - step: report$|    - step: report --summary .glassops/summary.md|' "$REPO/glassops.yaml"

set +e
docker run --rm -e SF_JWT_KEY -e SF_CLIENT_ID -e SF_USERNAME -e SF_DEVHUB_ORG_ID -e GLASSOPS_PERMIT_KEY \
  -v "$REPO:/work" -w /work \
  "$IMAGE" run e2e --ci local
code=$?
set -e
exit "$code"
