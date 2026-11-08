#!/usr/bin/env python3
"""Tests for licenses.py. Run: python3 scripts/notices/test_licenses.py

The SBOM packages here are synthetic, written to the shape syft's JSON output has [confirmed
against syft 1.55.0 output: artifacts[] with name, version, type and licenses[].value and
.spdxExpression]."""

import os
import sys
import tempfile
import unittest

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import licenses as L  # noqa: E402


def pkg(typ, name, *lics):
    return {"type": typ, "name": name, "version": "1", "licenses": [{"value": x, "spdxExpression": x} for x in lics]}


class Expressions(unittest.TestCase):
    def test_table(self):
        cases = [
            ("MIT", False),
            ("Apache-2.0 OR MIT", False),
            ("(MIT OR GPL-3.0-or-later)", False),
            ("GPL-3.0-or-later OR MIT", False),
            ("GPL-2.0-or-later OR LGPL-3.0-or-later", True),
            ("MIT AND GPL-2.0-or-later", True),
            ("(MIT AND BSD-2-Clause) OR GPL-2.0-only", False),
            ("( Apache-2.0 OR MIT ) AND MPL-2.0", False),
            ("GPL-2.0-only", True),
            ("gpl-2.0-only", True),
            ("LGPL-2.1-or-later", True),
            ("LGPL-2.1", True),
            ("AGPL-3.0-only", True),
            ("GPL-2.0-only WITH Classpath-exception-2.0", True),
            ("GPL-2.0-with-classpath-exception", True),
            ("MIT AND (Apache-2.0 OR LGPL-2.1-only)", False),
            ("BSD-3-Clause", False),
            ("GPLv3-lookalike-but-not", False),
        ]
        for expr, bad in cases:
            with self.subTest(expr=expr):
                self.assertEqual(bool(L.violations(expr)), bad)

    def test_malformed(self):
        for expr in ["", "(MIT", "MIT AND", "MIT OR OR GPL-2.0-only", "MIT)", "WITH"]:
            with self.subTest(expr=expr):
                with self.assertRaises(L.ParseError):
                    L.violations(expr)


class Check(unittest.TestCase):
    def run_check(self, pkgs, rules=()):
        return L.check([pkgs], list(rules))

    def test_clean(self):
        f, u, _ = self.run_check([pkg("npm", "a", "MIT"), pkg("apk", "b", "Apache-2.0 OR GPL-2.0-only")])
        self.assertEqual((f, u), ([], []))

    def test_copyleft_fails_and_allowlist_clears_it(self):
        p = pkg("apk", "jre", "GPL-2.0-with-classpath-exception")
        self.assertEqual(len(self.run_check([p])[0]), 1)
        rule = ("apk", "jre", "GPL-2.0-with-classpath-exception")
        self.assertEqual(self.run_check([p], [rule])[0], [])

    def test_allowlist_is_exact(self):
        p = pkg("apk", "jre", "GPL-3.0-only")
        rule = ("apk", "jre", "GPL-2.0-with-classpath-exception")
        self.assertEqual(len(self.run_check([p], [rule])[0]), 1)
        other_type = ("npm", "jre", "GPL-3.0-only")
        self.assertEqual(len(self.run_check([p], [other_type])[0]), 1)

    def test_agpl_is_never_allowlisted(self):
        p = pkg("npm", "x", "AGPL-3.0-only")
        findings, _, _ = self.run_check([p], [("*", "*", "AGPL-3.0-only")])
        self.assertEqual(len(findings), 1)
        self.assertTrue(findings[0][4])

    def test_agpl_branch_with_free_choice_passes(self):
        findings, _, _ = self.run_check([pkg("npm", "x", "MIT OR AGPL-3.0-only")])
        self.assertEqual(findings, [])

    def test_missing_and_odd_licenses_warn_not_fail(self):
        f, u, _ = self.run_check([pkg("go-module", "m"), pkg("npm", "n", "SEE LICENSE IN LICENSE.txt")])
        self.assertEqual(f, [])
        self.assertEqual(len(u), 2)

    def test_operator_entries_from_syft_are_ignored(self):
        p = pkg("apk", "xz", "0BSD", "AND", "Public-Domain")
        self.assertEqual(L.expressions(p), ["0BSD", "Public-Domain"])


class Notices(unittest.TestCase):
    def test_render_and_check(self):
        sbom = [pkg("npm", "a", "MIT"), pkg("go-module", "m"), pkg("apk", "b", "BSD-3-Clause", "MIT")]
        text = L.render_notices([sbom], [("go-module", "m", "Apache-2.0")])
        self.assertIn("npm a 1 : MIT", text)
        self.assertIn("go-module m 1 : Apache-2.0", text)
        self.assertIn("apk b 1 : BSD-3-Clause AND MIT", text)
        self.assertNotIn("UNKNOWN", text)
        with tempfile.TemporaryDirectory() as d:
            path = os.path.join(d, "NOTICES")
            with open(path, "w", encoding="utf-8") as f:
                f.write(text.replace(" 1 :", " 2 :"))  # versions may drift
            with open(path, encoding="utf-8") as f:
                self.assertEqual(L.names_in(text), L.names_in(f.read()))

    def test_unknown_stays_visible(self):
        text = L.render_notices([[pkg("npm", "z")]])
        self.assertIn("npm z 1 : UNKNOWN", text)


if __name__ == "__main__":
    unittest.main()
