#!/bin/sh
# The Sigstore end-to-end run. It signs a PERMIT keyless with the job's own OIDC token and
# checks the four things that matter:
#
#   E2E_IDENTITY=https://github.com/<workflow_ref> sh scripts/e2e/sigstore-e2e.sh [glassops]
#
#   1. gate signs and permit verify passes
#   2. a tampered PERMIT is refused
#   3. a PERMIT checked against another identity is refused
#   4. a PERMIT checked against another repository is refused
#
# It must run in a GitHub Actions job with id-token: write, with GITHUB_ACTIONS,
# GITHUB_REPOSITORY, ACTIONS_ID_TOKEN_REQUEST_URL and ACTIONS_ID_TOKEN_REQUEST_TOKEN in the
# environment. E2E_IDENTITY is the certificate identity this workflow gets, which is
# https://github.com/ followed by the workflow_ref context value. It needs sh, git, base64,
# head, node and cosign on the PATH, and network access to Fulcio and Rekor.
#
# Every signing writes a public Rekor transparency log entry: the hash of the PERMIT and the
# signing certificate, which names this repository, workflow and commit. The project is a
# dummy one, so the PERMIT holds nothing but made-up target names.
#
# A refusal must be exit 1. Exit 2 is a tool error, and it fails here, so a Rekor or Fulcio
# outage cannot pass for a refusal. Each refusal must also say the signature is not valid, so
# a refusal for some other reason (a changed config hash, say) cannot pass for the check.
set -eu

: "${E2E_IDENTITY:?E2E_IDENTITY is not set}"
: "${GITHUB_REPOSITORY:?GITHUB_REPOSITORY is not set}"
ISSUER=${E2E_ISSUER:-https://token.actions.githubusercontent.com}

# shellcheck source=/dev/null
. "$(dirname "$0")/lib-exec-project.sh"
e2e_init "${1:-glassops}"

# shellcheck disable=SC2089
E2E_EXTRA_YAML="policy:
  permit:
    signing: sigstore
    identity: \"$E2E_IDENTITY\"
    issuer: \"$ISSUER\""
# shellcheck disable=SC2090
export E2E_EXTRA_YAML
e2e_project

# The HMAC key is not used by a sigstore run. Dropping it proves nothing silently falls back.
unset GLASSOPS_PERMIT_KEY

# expect_signature_refusal checks a refusal that comes from the signature.
expect_signature_refusal() {
  what=$1
  shift
  expect_refusal "$what" "$@"
  if grep -q "signature is not valid" "$WORK/out.txt"; then
    pass "$what: the refusal is about the signature"
  else
    fail "$what: the refusal does not say the signature is not valid"
    cat "$WORK/out.txt" >&2
  fi
}

expect_exit 0 "plan" plan --from HEAD~1 --to HEAD
expect_exit 0 "gate signs keyless" gate --target prod
# shellcheck disable=SC2154
if [ "$failures" -ne 0 ]; then
  # Nothing after this means anything without a signed PERMIT.
  e2e_done "sigstore e2e"
fi
if [ "$(permit_json 'p.signature.alg')" = '"sigstore-bundle"' ]; then
  pass "the PERMIT is signed with sigstore-bundle"
else
  fail "the PERMIT algorithm is $(permit_json 'p.signature.alg'), want sigstore-bundle"
fi
expect_exit 0 "permit verify" permit verify --target prod

# A copy of the config with the same bytes verifies, so the refusals below come from the one
# change each makes and not from reading the config another way.
cp "$REPO/glassops.yaml" "$WORK/same.yaml"
expect_exit 0 "permit verify with a copy of the config" permit verify --target prod --config "$WORK/same.yaml"

# 2. A tampered PERMIT: the target kind is edited, so only the signature can refuse it.
cp "$REPO/.glassops/permit.json" "$WORK/permit.json"
sed 's/"kind": "production"/"kind": "sandbox"/' "$WORK/permit.json" > "$REPO/.glassops/permit.json"
if cmp -s "$WORK/permit.json" "$REPO/.glassops/permit.json"; then
  fail "the PERMIT tamper changed nothing"
fi
expect_signature_refusal "permit verify of a tampered PERMIT" permit verify --target prod
cp "$WORK/permit.json" "$REPO/.glassops/permit.json"

# 3. Another identity: the same config with a different workflow in the identity.
sed "s|cli-e2e.yml|other-workflow.yml|" "$REPO/glassops.yaml" > "$WORK/wrong-identity.yaml"
if cmp -s "$REPO/glassops.yaml" "$WORK/wrong-identity.yaml"; then
  fail "the wrong identity config changed nothing"
fi
expect_signature_refusal "permit verify against a wrong identity" permit verify --target prod --config "$WORK/wrong-identity.yaml"

# 4. Another repository: the config pins a repository that did not sign.
sed '/^    issuer:/a\
    repository: "glassops-platform/not-this-repository"' "$REPO/glassops.yaml" > "$WORK/wrong-repository.yaml"
if [ "$GITHUB_REPOSITORY" = glassops-platform/not-this-repository ] || ! grep -q "not-this-repository" "$WORK/wrong-repository.yaml"; then
  fail "the wrong repository config is not a wrong repository"
fi
expect_signature_refusal "permit verify against a wrong repository" permit verify --target prod --config "$WORK/wrong-repository.yaml"

# The PERMIT is intact after all of it.
expect_exit 0 "permit verify after the refusals" permit verify --target prod

e2e_done "sigstore e2e"
