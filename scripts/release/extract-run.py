#!/usr/bin/env python3
"""Prints the run: script of one step of a workflow or composite action file.

  extract-run.py FILE STEP

STEP is matched against a step's id or name. The script is printed as the runner would write it,
with the YAML block indent removed. Standard library only, and only for the "run: |" block
style these files use. Exit 2 when the step is not found.
"""
import re
import sys


def main(path, step):
    with open(path, encoding="utf-8") as f:
        lines = f.read().split("\n")
    start = None
    for i, line in enumerate(lines):
        m = re.match(r"^(\s*)- (?:id|name): (.*)$", line)
        if m and m.group(2).strip().strip("'\"") == step:
            start = i
            indent = len(m.group(1))
            break
    if start is None:
        print("step not found: " + step, file=sys.stderr)
        return 2
    i = start + 1
    while i < len(lines):
        line = lines[i]
        if line.strip() and len(line) - len(line.lstrip()) <= indent and line.lstrip().startswith("- "):
            break
        m = re.match(r"^(\s*)run: \|\s*$", line)
        if m:
            base = None
            out = []
            i += 1
            while i < len(lines):
                body = lines[i]
                if body.strip() == "":
                    out.append("")
                    i += 1
                    continue
                cur = len(body) - len(body.lstrip())
                if cur <= len(m.group(1)):
                    break
                if base is None:
                    base = cur
                out.append(body[base:])
                i += 1
            sys.stdout.write("\n".join(out).rstrip("\n") + "\n")
            return 0
        i += 1
    print("step has no run block: " + step, file=sys.stderr)
    return 2


if __name__ == "__main__":
    if len(sys.argv) != 3:
        print(__doc__, file=sys.stderr)
        sys.exit(2)
    sys.exit(main(sys.argv[1], sys.argv[2]))
