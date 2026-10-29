#!/bin/sh
# Creates one scratch org for the end-to-end run and writes how to reach it into OUTDIR.
# Runs inside the glassops image (entrypoint sh). Needs SF_JWT_KEY (PEM text), SF_CLIENT_ID and
# SF_USERNAME for the Dev Hub user. Writes OUTDIR/scratch-info-id, OUTDIR/org-id and
# OUTDIR/auth-url (mode 0600). It prints nothing secret.
#
# It never runs "sf org create scratch". That command leaves a live org behind when its own
# follow-up JWT login to the new org fails. The org is requested with a ScratchOrgInfo record and
# reached by exchanging its auth code for a refresh token.
set -eu
umask 077

OUTDIR=${1:?usage: scratch-create.sh OUTDIR}
: "${SF_JWT_KEY:?SF_JWT_KEY is not set}"
: "${SF_CLIENT_ID:?SF_CLIENT_ID is not set}"
: "${SF_USERNAME:?SF_USERNAME is not set}"

mkdir -p "$OUTDIR"
chmod 700 "$OUTDIR"

HOME=$(mktemp -d)
export HOME
chmod 700 "$HOME"
cleanup() {
  rc=$?
  if [ -n "${OUT_OWNER:-}" ]; then
    chown -R "$OUT_OWNER" "$OUTDIR" 2>/dev/null || true
  fi
  rm -rf "$HOME"
  exit "$rc"
}
trap cleanup EXIT INT TERM

CALLBACK='http://localhost:1717/OauthRedirect'
Q="$HOME/query.json"

printf '%s\n' "$SF_JWT_KEY" > "$HOME/jwt.key"
if ! sf org login jwt --username "$SF_USERNAME" --client-id "$SF_CLIENT_ID" \
  --jwt-key-file "$HOME/jwt.key" --instance-url https://login.salesforce.com \
  --alias hub --json > "$HOME/login.json" 2>/dev/null; then
  echo "scratch-create: the Dev Hub login failed" >&2
  exit 1
fi

if ! sf data create record --sobject ScratchOrgInfo \
  --values "Edition=Developer OrgName='GlassOps E2E' DurationDays=1 ConnectedAppConsumerKey=PlatformCLI ConnectedAppCallbackUrl=$CALLBACK" \
  --target-org hub --json > "$HOME/create.json" 2>/dev/null; then
  echo "scratch-create: could not create the ScratchOrgInfo record" >&2
  exit 1
fi
INFO_ID=$(node -e 'const j=JSON.parse(require("fs").readFileSync(process.argv[1],"utf8"));process.stdout.write(String(j.result&&j.result.id||""))' "$HOME/create.json")
# Save the record Id before checking it, so scratch-delete can still find the org if anything
# after this point fails. scratch-delete checks the Id again before it uses it.
if [ -z "$INFO_ID" ]; then
  echo "scratch-create: no record Id came back" >&2
  exit 1
fi
printf '%s\n' "$INFO_ID" > "$OUTDIR/scratch-info-id"
case "$INFO_ID" in
  *[!A-Za-z0-9]*) echo "scratch-create: the record Id has unexpected characters" >&2; exit 1 ;;
esac
case "${#INFO_ID}" in
  15|18) ;;
  *) echo "scratch-create: the record Id is not 15 or 18 characters" >&2; exit 1 ;;
esac
echo "scratch-create: requested scratch org (record $INFO_ID)"

# Poll every 10 seconds for up to 15 minutes. The query output holds the auth code, so it goes
# to a 0600 file and only the status, error code and org Id are read back.
STATUS=""
n=0
while [ "$n" -lt 90 ]; do
  if sf data query --query "SELECT Status, ErrorCode, ScratchOrg, LoginUrl, AuthCode FROM ScratchOrgInfo WHERE Id = '$INFO_ID'" \
    --target-org hub --json > "$Q" 2>/dev/null; then
    STATUS=$(node -e 'const j=JSON.parse(require("fs").readFileSync(process.argv[1],"utf8"));const r=(j.result&&j.result.records||[])[0]||{};process.stdout.write((r.Status||"")+"|"+(r.ErrorCode||"")+"|"+(r.ScratchOrg||""))' "$Q")
  fi
  case "$STATUS" in
    Active\|*) break ;;
    Error\|*)
      code=${STATUS#Error|}
      code=${code%%|*}
      echo "scratch-create: signup failed with ErrorCode ${code:-unknown}" >&2
      exit 1 ;;
  esac
  n=$((n + 1))
  sleep 10
done
case "$STATUS" in
  Active\|*) ;;
  *) echo "scratch-create: the scratch org was not active after 15 minutes" >&2; exit 1 ;;
esac

# Write the org Id before the token exchange, so scratch-delete can find the org even when the
# exchange fails.
ORG_ID=${STATUS##*|}
printf '%s\n' "$ORG_ID" > "$OUTDIR/org-id"

# Exchange the auth code. The code is read from the query file and never reaches argv or stdout.
if ! Q="$Q" CALLBACK="$CALLBACK" OUTDIR="$OUTDIR" node -e '
const fs = require("fs");
const j = JSON.parse(fs.readFileSync(process.env.Q, "utf8"));
const r = (j.result && j.result.records || [])[0] || {};
if (!r.AuthCode || !r.LoginUrl) { console.error("no auth code on the record"); process.exit(1); }
const body = new URLSearchParams({
  grant_type: "authorization_code",
  code: r.AuthCode,
  client_id: "PlatformCLI",
  redirect_uri: process.env.CALLBACK,
});
fetch(r.LoginUrl.replace(/\/+$/, "") + "/services/oauth2/token", {
  method: "POST",
  headers: { "content-type": "application/x-www-form-urlencoded" },
  body,
}).then(async (res) => {
  const t = await res.json().catch(() => ({}));
  if (!res.ok || !t.refresh_token || !t.instance_url) {
    console.error("token exchange failed, status " + res.status + (t.error ? ", " + t.error : ""));
    process.exit(1);
  }
  const host = t.instance_url.replace(/^https?:\/\//, "").replace(/\/+$/, "");
  fs.writeFileSync(process.env.OUTDIR + "/auth-url", "force://PlatformCLI::" + t.refresh_token + "@" + host + "\n", { mode: 0o600 });
}).catch((e) => { console.error("token exchange failed: " + e.message); process.exit(1); });
'; then
  exit 1
fi
chmod 600 "$OUTDIR/auth-url"
echo "scratch-create: scratch org $ORG_ID is ready"
