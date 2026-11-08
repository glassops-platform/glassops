#!/usr/bin/env bash
# Proves the license check fails on an AGPL package and passes on a clean image. It builds two
# throwaway images from the alpine base, one with an npm package that declares AGPL-3.0-only, one
# with an MIT package, scans both, and requires the check to fail on the first and pass on the
# second even when an allowlist tries to cover the AGPL package.
#   scripts/notices/test-planted.sh
set -euo pipefail
here="$(cd "$(dirname "$0")" && pwd)"
py=""
for cand in python3 python; do
  if "$cand" -c 'import sys' >/dev/null 2>&1; then py="$cand"; break; fi
done
[ -n "$py" ] || { echo "no working python3 or python" >&2; exit 2; }
work="$(mktemp -d)"
trap 'rm -rf "$work"; docker rmi -f glassops-planted:agpl glassops-planted:clean >/dev/null 2>&1 || true' EXIT

native() { if command -v cygpath >/dev/null 2>&1; then cygpath -m "$1"; else echo "$1"; fi; }

build() { # tag license
  mkdir -p "$work/$1"
  cat > "$work/$1/Dockerfile" <<DOCKER
FROM alpine:3.24
RUN mkdir -p /app/node_modules/planted-pkg \
 && printf '{"name":"planted-pkg","version":"1.0.0","license":"$2"}' > /app/node_modules/planted-pkg/package.json
DOCKER
  docker build -q -t "glassops-planted:$1" "$(native "$work/$1")" >/dev/null
}

check() { # tag extra-args...
  tag="$1"; shift
  "$here/sbom.sh" "glassops-planted:$tag" "$work/$tag.json"
  "$py" -X utf8 "$here/licenses.py" check "$(native "$work/$tag.json")" --allow "$(native "$here/allowlist.txt")" "$@"
}

build agpl AGPL-3.0-only
build clean MIT

fail=0
expect() { # want-exit description command...
  want="$1"; what="$2"; shift 2
  echo "--- $what"
  set +e; "$@"; got=$?; set -e
  if [ "$got" != "$want" ]; then echo "FAIL: $what: exit $got, want $want"; fail=1; fi
}
printf 'npm:planted-pkg = AGPL-3.0-only
' > "$work/try.txt"
expect 0 "a clean image passes" check clean
expect 1 "an AGPL image fails" check agpl
expect 1 "an AGPL image fails even when an allowlist names the package" check agpl "$(native "$work/try.txt")"
[ "$fail" = 0 ] && echo "ok: the license check behaves on planted images"
exit "$fail"
