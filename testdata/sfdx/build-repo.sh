#!/bin/sh
# Builds a git repo from the test Salesforce project with exactly two commits on main.
#
#   sh build-repo.sh DEST
#
# DEST must not exist or must be empty. Commit 1 is commit1/ in full. Commit 2 copies the
# files under commit2/ over it and deletes every path listed in commit2.remove. The
# identity is set in DEST only. No glassops.yaml is written.
set -eu

if [ $# -ne 1 ] || [ -z "$1" ]; then
  echo "usage: sh build-repo.sh DEST" >&2
  exit 2
fi
dest=$1
here=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)

if [ -e "$dest" ] && [ -n "$(ls -A "$dest" 2>/dev/null)" ]; then
  echo "build-repo: $dest exists and is not empty" >&2
  exit 2
fi
mkdir -p "$dest"

git -C "$dest" init -q -b main
git -C "$dest" config user.name "GlassOps Test"
git -C "$dest" config user.email "test@example.invalid"
git -C "$dest" config commit.gpgsign false

cp -R "$here/commit1/." "$dest/"
git -C "$dest" add -A
git -C "$dest" commit -q -m "Add Greeter, Account.Tier__c and Greeter_User"

cp -R "$here/commit2/." "$dest/"
while IFS= read -r path || [ -n "$path" ]; do
  [ -n "$path" ] || continue
  rm -f "$dest/$path"
done < "$here/commit2.remove"
git -C "$dest" add -A
git -C "$dest" commit -q -m "Change Greeter, remove Account.Tier__c"
