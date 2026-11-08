#!/usr/bin/env bash
# Decides what a release tag publishes (plan T12). Prints key=value lines for GITHUB_OUTPUT.
#   scripts/release/tags.sh v1.2.3
# A tag is vMAJOR.MINOR.PATCH with an optional -prerelease suffix. Anything else exits 2.
#
#   prerelease   true when the tag has a - suffix or the major version is 0
#   version      the version without the v
#   image_tags   one image tag per line: 1.2.3 and v1.2.3, and for a final release of major 1
#                or more also v1 and latest
#   major_tag    the git tag to move to the release commit (v1), empty for a prerelease
set -euo pipefail

tag="${1:-}"
re='^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z]+([.-][0-9A-Za-z]+)*)?$'
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

tags="$version"$'\n'"v$version"
major_tag=""
if [ "$prerelease" = false ]; then
  major_tag="v$major"
  tags="$tags"$'\n'"$major_tag"$'\n'"latest"
fi

echo "prerelease=$prerelease"
echo "version=$version"
echo "major_tag=$major_tag"
echo "image_tags<<EOF"
echo "$tags"
echo "EOF"
