#!/usr/bin/env python3
"""License checks and third party notices from a syft JSON SBOM.

  licenses.py check SBOM... --allow FILE...   fail on AGPL, GPL and LGPL packages
  licenses.py notices SBOM... -o FILE         write THIRD_PARTY_NOTICES
        [--licenses FILE...]                   licenses for packages that declare none
  licenses.py notices SBOM... --check FILE    fail when FILE lists different packages

Standard library only. Exit codes: 0 ok, 1 a finding (a copyleft package, or stale notices),
2 a bad input.

The check reads each package's declared license expression and asks whether the package can be
used without taking a copyleft license. In "A OR B" one side being free is enough, because the
user may choose it. In "A AND B" both sides must be free. AGPL can never be allowlisted.

An allowlist line is "type:name-glob = license-text", for example
  apk:openjdk17-jre-headless = GPL-2.0-only WITH Classpath-exception-2.0
and matches only a package whose declared expression equals the text, so a change of license
brings the package back into the report. Text after # is a comment.
"""

import argparse
import fnmatch
import json
import re
import sys

COPYLEFT = re.compile(r"^(A|L)?GPL(-|$)", re.IGNORECASE)
AGPL = re.compile(r"^AGPL(-|$)", re.IGNORECASE)
TOKEN = re.compile(r"\(|\)|[^\s()]+")


class ParseError(Exception):
    pass


def tokens(expr):
    return TOKEN.findall(expr)


def violations(expr):
    """Return the copyleft ids that the expression forces on the user, empty when none."""
    toks = tokens(expr)
    if not toks:
        raise ParseError("empty expression")
    pos = 0

    def peek():
        return toks[pos] if pos < len(toks) else None

    def take():
        nonlocal pos
        tok = peek()
        pos += 1
        return tok

    def is_op(tok, name):
        return tok is not None and tok.upper() == name

    def parse_or():
        parts = [parse_and()]
        while is_op(peek(), "OR"):
            take()
            parts.append(parse_and())
        if any(not p for p in parts):
            return []
        return [v for p in parts for v in p]

    def parse_and():
        out = parse_atom()
        while is_op(peek(), "AND"):
            take()
            out = out + parse_atom()
        return out

    def parse_atom():
        tok = take()
        if tok is None:
            raise ParseError("unexpected end in " + repr(expr))
        if tok == "(":
            inner = parse_or()
            if take() != ")":
                raise ParseError("missing ) in " + repr(expr))
            return inner
        if tok == ")" or tok.upper() in ("AND", "OR", "WITH"):
            raise ParseError("unexpected %r in %r" % (tok, expr))
        ident = tok
        if is_op(peek(), "WITH"):
            take()
            exc = take()
            if exc is None:
                raise ParseError("WITH without an exception in " + repr(expr))
            ident = tok + " WITH " + exc
        return [ident] if COPYLEFT.match(tok) else []

    out = parse_or()
    if pos != len(toks):
        raise ParseError("trailing %r in %r" % (toks[pos], expr))
    return out


def load_sbom(path):
    try:
        with open(path, encoding="utf-8") as f:
            doc = json.load(f)
    except (OSError, ValueError) as err:
        raise SystemExit("error: cannot read SBOM %s: %s" % (path, err))
    if not isinstance(doc, dict) or "artifacts" not in doc:
        raise SystemExit("error: %s is not a syft JSON SBOM (no artifacts)" % path)
    return doc["artifacts"]


def expressions(pkg):
    """The declared license strings of a package, in syft's order, without duplicates."""
    seen = []
    for lic in pkg.get("licenses") or []:
        text = " ".join((lic.get("spdxExpression") or lic.get("value") or "").split())
        # syft splits "A AND B" into entries and sometimes keeps the operator as an entry.
        if text and text.upper() not in ("AND", "OR", "WITH") and text not in seen:
            seen.append(text)
    return seen


def load_allowlist(paths):
    rules = []
    for path in paths:
        try:
            with open(path, encoding="utf-8") as f:
                lines = f.read().splitlines()
        except OSError as err:
            raise SystemExit("error: cannot read allowlist %s: %s" % (path, err))
        for n, line in enumerate(lines, 1):
            line = line.split("#", 1)[0].strip()
            if not line:
                continue
            if "=" not in line or ":" not in line.split("=", 1)[0]:
                raise SystemExit("error: %s:%d: want type:name = license" % (path, n))
            left, lic = line.split("=", 1)
            typ, name = left.split(":", 1)
            rules.append((typ.strip(), name.strip(), " ".join(lic.split())))
    return rules


def allowed(pkg, expr, rules):
    for typ, name, lic in rules:
        if fnmatch.fnmatchcase(pkg["type"], typ) and fnmatch.fnmatchcase(pkg["name"], name) and expr == lic:
            return True
    return False


def check(sboms, rules):
    """Return (findings, unknown, used) over the SBOMs."""
    findings, unknown, used = [], [], set()
    seen = set()
    for pkgs in sboms:
        for pkg in pkgs:
            exprs = expressions(pkg)
            key = (pkg["type"], pkg["name"], pkg.get("version", ""), tuple(exprs))
            if key in seen:
                continue
            seen.add(key)
            if not exprs:
                unknown.append((pkg["type"], pkg["name"], pkg.get("version", ""), "no license declared"))
                continue
            bad = []
            for expr in exprs:
                try:
                    v = violations(expr)
                except ParseError:
                    unknown.append((pkg["type"], pkg["name"], pkg.get("version", ""), "not an SPDX expression: " + expr))
                    continue
                if not v:
                    continue
                if any(AGPL.match(x) for x in v):
                    bad.append((expr, v, True))
                elif allowed(pkg, expr, rules):
                    used.add((pkg["type"], pkg["name"], expr))
                else:
                    bad.append((expr, v, False))
            for expr, v, hard in bad:
                findings.append((pkg["type"], pkg["name"], pkg.get("version", ""), expr, hard))
    return findings, unknown, used


def cmd_check(args):
    sboms = [load_sbom(p) for p in args.sbom]
    rules = load_allowlist(args.allow)
    findings, unknown, used = check(sboms, rules)
    total = sum(len(s) for s in sboms)
    print("checked %d packages from %d SBOM(s) with %d allowlist rule(s)" % (total, len(sboms), len(rules)))
    for typ, name, ver, why in sorted(unknown):
        print("warning: %s %s %s: %s" % (typ, name, ver, why))
    for typ, name, ver, expr, hard in sorted(findings):
        tag = "AGPL, never allowlisted" if hard else "copyleft"
        print("FAIL: %s %s %s declares %s (%s)" % (typ, name, ver, expr, tag))
    if findings:
        print("%d package(s) take a copyleft license" % len(findings))
        return 1
    print("no AGPL, GPL or LGPL package outside the allowlist")
    return 0


def known_license(pkg, known):
    for typ, name, lic in known:
        if fnmatch.fnmatchcase(pkg["type"], typ) and fnmatch.fnmatchcase(pkg["name"], name):
            return lic
    return None


def package_lines(sboms, known=()):
    """One line per distinct package name, version and license, for the notices file.

    A package with no declared license (Go modules carry none) takes its license from the known
    list, which a person filled in from the module's own LICENSE file."""
    rows = {}
    for pkgs in sboms:
        for pkg in pkgs:
            exprs = expressions(pkg)
            if not exprs:
                lic = known_license(pkg, known)
                exprs = [lic] if lic else ["UNKNOWN"]
            ver = pkg.get("version") or ""
            if ver == "UNKNOWN":
                ver = ""
            rows[(pkg["type"], pkg["name"], ver, " AND ".join(exprs))] = True
    return sorted(rows)


HEADER = """THIRD PARTY NOTICES

GlassOps is licensed under the Apache License 2.0 (see LICENSE). The container images and the
release archives include the third party software listed below. The list is generated from the
software bill of materials (SBOM) of the images with scripts/notices/licenses.py, so it names
what is installed and the license each package declares. The license texts ship with each
package in the image, for example under /usr/share/licenses and in node_modules.

Format: <type> <name> <version> : <declared license>
"""


def render_notices(sboms, known=()):
    out = [HEADER]
    for typ, name, ver, lic in package_lines(sboms, known):
        out.append("%s %s%s : %s" % (typ, name, (" " + ver) if ver else "", lic))
    return "\n".join(out) + "\n"


def names_in(text):
    """The (type, name, license) of each package line, ignoring versions."""
    out = set()
    for line in text.splitlines():
        m = re.match(r"^(\S+) (\S+)(?: \S+)? : (.*)$", line)
        if m and not line.startswith("Format:"):
            out.add((m.group(1), m.group(2), m.group(3)))
    return out


def cmd_notices(args):
    sboms = [load_sbom(p) for p in args.sbom]
    text = render_notices(sboms, load_allowlist(args.licenses))
    if args.check:
        try:
            with open(args.check, encoding="utf-8") as f:
                have = f.read()
        except OSError as err:
            print("error: cannot read %s: %s" % (args.check, err), file=sys.stderr)
            return 2
        want, got = names_in(text), names_in(have)
        missing, extra = sorted(want - got), sorted(got - want)
        for item in missing:
            print("missing from %s: %s %s : %s" % (args.check, *item))
        for item in extra:
            print("not in the images any more: %s %s : %s" % item)
        if missing or extra:
            print("regenerate it: python3 scripts/notices/licenses.py notices <sbom>... -o %s" % args.check)
            return 1
        print("%s matches the images (%d packages)" % (args.check, len(want)))
        return 0
    if args.output:
        with open(args.output, "w", encoding="utf-8", newline="\n") as f:
            f.write(text)
    else:
        sys.stdout.write(text)
    return 0


def main(argv):
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    sub = ap.add_subparsers(dest="cmd", required=True)
    c = sub.add_parser("check")
    c.add_argument("sbom", nargs="+")
    c.add_argument("--allow", nargs="*", default=[])
    c.set_defaults(fn=cmd_check)
    n = sub.add_parser("notices")
    n.add_argument("sbom", nargs="+")
    n.add_argument("--licenses", nargs="*", default=[])
    n.add_argument("-o", "--output")
    n.add_argument("--check")
    n.set_defaults(fn=cmd_notices)
    args = ap.parse_args(argv)
    return args.fn(args)


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
