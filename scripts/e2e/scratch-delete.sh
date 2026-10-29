#!/bin/sh
# Deletes the scratch org that scratch-create.sh made, and nothing else. Runs inside the
# glassops image (entrypoint sh). Needs SF_JWT_KEY, SF_CLIENT_ID and SF_USERNAME, and reads
# OUTDIR/org-id.
set -eu
umask 077

OUTDIR=${1:?usage: scratch-delete.sh OUTDIR}
: "${SF_JWT_KEY:?SF_JWT_KEY is not set}"
: "${SF_CLIENT_ID:?SF_CLIENT_ID is not set}"
: "${SF_USERNAME:?SF_USERNAME is not set}"

ORG_ID=""
if [ -f "$OUTDIR/org-id" ]; then
  ORG_ID=$(tr -d ' \r\n' < "$OUTDIR/org-id")
fi
case "$ORG_ID" in
  00D*) ;;
  *) echo "scratch-delete: org-id is empty or does not start with 00D, refusing" >&2; exit 1 ;;
esac
case "${#ORG_ID}" in
  15|18) ;;
  *) echo "scratch-delete: org-id is not 15 or 18 characters, refusing" >&2; exit 1 ;;
esac
case "$ORG_ID" in
  *[!A-Za-z0-9]*) echo "scratch-delete: org-id has unexpected characters, refusing" >&2; exit 1 ;;
esac
SHORT=$(printf '%s' "$ORG_ID" | cut -c1-15)

HOME=$(mktemp -d)
export HOME
chmod 700 "$HOME"
trap 'rm -rf "$HOME"' EXIT INT TERM

printf '%s\n' "$SF_JWT_KEY" > "$HOME/jwt.key"
if ! sf org login jwt --username "$SF_USERNAME" --client-id "$SF_CLIENT_ID" \
  --jwt-key-file "$HOME/jwt.key" --instance-url https://login.salesforce.com \
  --alias hub --json > "$HOME/login.json" 2>/dev/null; then
  echo "scratch-delete: the Dev Hub login failed" >&2
  exit 1
fi

if ! sf data query --query "SELECT Id FROM ActiveScratchOrg WHERE ScratchOrg = '$SHORT'" \
  --target-org hub --json > "$HOME/query.json" 2>/dev/null; then
  echo "scratch-delete: the ActiveScratchOrg query failed" >&2
  exit 1
fi
REC=$(node -e 'const j=JSON.parse(require("fs").readFileSync(process.argv[1],"utf8"));const r=(j.result&&j.result.records||[]);process.stdout.write(r.length===1?String(r[0].Id):(r.length?"MANY":""))' "$HOME/query.json")
case "$REC" in
  "") echo "scratch-delete: notice, no active scratch org $SHORT, nothing to delete"; exit 0 ;;
  MANY) echo "scratch-delete: more than one record matched, refusing" >&2; exit 1 ;;
  *[!A-Za-z0-9]*) echo "scratch-delete: unexpected record Id, refusing" >&2; exit 1 ;;
esac

if ! sf data delete record --sobject ActiveScratchOrg --record-id "$REC" \
  --target-org hub --json > "$HOME/delete.json" 2>/dev/null; then
  echo "scratch-delete: the delete failed" >&2
  exit 1
fi
echo "scratch-delete: deleted scratch org $SHORT"
