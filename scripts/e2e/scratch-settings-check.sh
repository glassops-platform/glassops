#!/bin/sh
# The settings proof. Runs as an exec step right after "env create" in the e2e pipeline, inside
# the glassops image, so it reaches the scratch org that env create made and applied
# testdata/e2e/scratch-def.json to. It retrieves Settings:Mobile with that org's own login and
# fails unless enableS1EncryptedStoragePref2 is false, the value the definition asked for.
#
#   sh scratch-settings-check.sh
#
# Needs GLASSOPS_RESULTS_DIR (exec steps set it), node and sf. It prints no secret: the auth
# file path is printed, its content never is. Nothing is written to the project directory.
set -eu
umask 077

: "${GLASSOPS_RESULTS_DIR:?GLASSOPS_RESULTS_DIR is not set, run this as an exec step}"
STATE="$GLASSOPS_RESULTS_DIR/env-create.json"
[ -s "$STATE" ] || { echo "settings-check: $STATE is missing, env create did not run" >&2; exit 1; }

AUTH=$(node -e 'const r=JSON.parse(require("fs").readFileSync(process.argv[1],"utf8"));process.stdout.write(String((r.facts||{})["env.auth_file"]||""))' "$STATE")
[ -n "$AUTH" ] && [ -f "$AUTH" ] || { echo "settings-check: env create recorded no usable auth file" >&2; exit 1; }

WORK=$(mktemp -d)
HOME=$WORK/home
export HOME
mkdir -p "$HOME" "$WORK/proj"
trap 'rm -rf "$WORK"' EXIT INT TERM

# The retrieve runs in a throwaway project, so the real project stays untouched.
printf '{"packageDirectories":[{"path":"force-app","default":true}]}\n' > "$WORK/proj/sfdx-project.json"
mkdir -p "$WORK/proj/force-app"

if ! sf org login sfdx-url --sfdx-url-file "$AUTH" --alias settings-check --json > "$WORK/login.json" 2>/dev/null; then
  echo "settings-check: login to the scratch org failed" >&2
  exit 1
fi
cd "$WORK/proj"
if ! sf project retrieve start --metadata Settings:Mobile --target-org settings-check --json > "$WORK/retrieve.json" 2>"$WORK/retrieve.err"; then
  echo "settings-check: the retrieve of Settings:Mobile failed" >&2
  cat "$WORK/retrieve.err" >&2 || true
  exit 1
fi

FILE=$(find "$WORK/proj" -name 'Mobile.settings*' -type f | head -n 1)
[ -n "$FILE" ] || { echo "settings-check: the retrieve returned no Mobile settings file" >&2; exit 1; }

VALUE=$(tr -d ' \t\r\n' < "$FILE" | sed -n 's|.*<enableS1EncryptedStoragePref2>\([a-z]*\)</enableS1EncryptedStoragePref2>.*|\1|p')
echo "settings-check: enableS1EncryptedStoragePref2 = ${VALUE:-<absent>}"
if [ "$VALUE" != false ]; then
  echo "settings-check: expected false, the definition's mobileSettings did not reach the org" >&2
  exit 1
fi
echo "settings-check: Mobile settings match the definition"
