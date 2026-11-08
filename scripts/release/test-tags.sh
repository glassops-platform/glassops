#!/usr/bin/env bash
# Table test of the T12 tag rules in tags.sh.   scripts/release/test-tags.sh
set -uo pipefail
here="$(cd "$(dirname "$0")" && pwd)"
fail=0

# tag | prerelease | major_tag | image tags, space separated | (invalid tags have "invalid")
table=(
  "v1.0.0|false|v1|1.0.0 v1.0.0 v1 latest"
  "v1.2.3|false|v1|1.2.3 v1.2.3 v1 latest"
  "v2.0.0|false|v2|2.0.0 v2.0.0 v2 latest"
  "v10.3.1|false|v10|10.3.1 v10.3.1 v10 latest"
  "v1.0.0-rc1|true||1.0.0-rc1 v1.0.0-rc1"
  "v1.0.0-rc.2|true||1.0.0-rc.2 v1.0.0-rc.2"
  "v2.1.0-beta|true||2.1.0-beta v2.1.0-beta"
  "v0.4.0|true||0.4.0 v0.4.0"
  "v0.4.0-rc1|true||0.4.0-rc1 v0.4.0-rc1"
  "v0.0.1|true||0.0.1 v0.0.1"
  "v1|invalid"
  "v1.0|invalid"
  "1.0.0|invalid"
  "v01.0.0|invalid"
  "v1.0.0+build5|invalid"
  "v1.0.0-|invalid"
  "vX.Y.Z|invalid"
  "|invalid"
)

for row in "${table[@]}"; do
  IFS='|' read -r tag want_pre want_major want_tags <<< "$row"
  out="$("$here/tags.sh" "$tag" 2>/dev/null)"
  code=$?
  if [ "$want_pre" = "invalid" ]; then
    if [ "$code" != 2 ]; then echo "FAIL $tag: exit $code, want 2"; fail=1; fi
    continue
  fi
  if [ "$code" != 0 ]; then echo "FAIL $tag: exit $code"; fail=1; continue; fi
  got_pre="$(sed -n 's/^prerelease=//p' <<< "$out")"
  got_major="$(sed -n 's/^major_tag=//p' <<< "$out")"
  got_tags="$(sed -n '/^image_tags<<EOF$/,/^EOF$/p' <<< "$out" | sed '1d;$d' | tr '\n' ' ' | sed 's/ $//')"
  [ "$got_pre" = "$want_pre" ] || { echo "FAIL $tag: prerelease=$got_pre, want $want_pre"; fail=1; }
  [ "$got_major" = "$want_major" ] || { echo "FAIL $tag: major_tag='$got_major', want '$want_major'"; fail=1; }
  [ "$got_tags" = "$want_tags" ] || { echo "FAIL $tag: image tags '$got_tags', want '$want_tags'"; fail=1; }
done
[ "$fail" = 0 ] && echo "ok: ${#table[@]} tags"
exit "$fail"
