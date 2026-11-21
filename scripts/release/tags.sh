#!/usr/bin/env bash
# Decides what a release tag publishes (plan T12). Prints key=value lines for GITHUB_OUTPUT.
#   RELEASE_TAGS="$(git tag -l 'v*')" scripts/release/tags.sh v1.2.3
# A tag is vMAJOR.MINOR.PATCH with an optional -prerelease suffix. Anything else exits 2.
#
#   prerelease   true when the tag has a - suffix or the major version is 0
#   version      the version without the v
#   image_tags   one image tag per line: 1.2.3 and v1.2.3, and for a final release of major 1
#                or more also v1 and latest, each only when the release is the newest of its kind
#   major_tag    the git tag to move to the release commit (v1), empty for a prerelease
#   latest       true when this release should be GitHub's latest release
#
# RELEASE_TAGS lists the repository's existing tags, one per line. A final release that is
# older than another final release keeps the moving tags where they are: v1 and its image tag
# move only for the highest final version of major 1, and latest only for the highest final
# version of all. So a fix released as v1.0.2 after v1.1.0 never moves v1 or latest back.
# With RELEASE_TAGS unset, the release is taken to be the newest, as before.
set -euo pipefail
tag="${1:-}"
re='^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z]+([.-][0-9A-Za-z]+)*)?$'
final_re='^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$'
if ! [[ "$tag" =~ $re ]]; then
  echo "error: '$tag' is not a release tag such as v1.2.3 or v1.2.3-rc1" >&2
  exit 2
fi
major="${BASH_REMATCH[1]}"
suffix="${BASH_REMATCH[4]}"
version="${tag#v}"
prerelease=false
if [ -n "$suffix" ] || [ "$major" = "0" ]; then
  prerelease=true
fi

# newer_than prints the final tags in RELEASE_TAGS, optionally of one major, that sort after
# this version. sort -V orders 1.10.0 after 1.9.0.
newer_than() {
  local only_major="${1:-}"
  local t
  while IFS= read -r t; do
    [[ "$t" =~ $final_re ]] || continue
    [ -z "$only_major" ] || [ "${BASH_REMATCH[1]}" = "$only_major" ] || continue
    [ "$t" = "$tag" ] && continue
    if [ "$(printf '%s\n%s\n' "${t#v}" "$version" | sort -V | tail -n 1)" = "${t#v}" ]; then
      echo "$t"
    fi
  done <<< "${RELEASE_TAGS:-}"
}

tags="$version"$'\n'"v$version"
major_tag=""
latest=false
if [ "$prerelease" = false ]; then
  if [ -z "$(newer_than "$major")" ]; then
    major_tag="v$major"
    tags="$tags"$'\n'"$major_tag"
  fi
  if [ -z "$(newer_than)" ]; then
    latest=true
    tags="$tags"$'\n'"latest"
  fi
fi
echo "prerelease=$prerelease"
echo "version=$version"
echo "major_tag=$major_tag"
echo "latest=$latest"
echo "image_tags<<EOF"
echo "$tags"
echo "EOF"
