#!/bin/sh
# The exec.describe command of the exec provider end-to-end run. The target is the directory
# EXEC_E2E_TARGET_DIR, and its Id is the one line in its .target-id file, so the guard compares
# what the target says about itself with targets.<t>.id, never the config with itself.
set -eu

: "${EXEC_E2E_TARGET_DIR:?EXEC_E2E_TARGET_DIR is not set}"
id=$(tr -d ' \r\n' < "$EXEC_E2E_TARGET_DIR/.target-id")
printf '{"id": "%s", "kind": "%s"}\n' "$id" "${EXEC_E2E_TARGET_KIND:-production}"
