#!/usr/bin/env python3
"""Rewrite the marketplace manifest so every plugin resolves to a LOCAL PATH.

The shipped manifest pins three of the four plugins to a published release tag via a `git-subdir`
source. That is correct for consumers and wrong for a universe built to test a checkout: install
would fetch the release and the working tree would never be consulted — silently, with the build
reporting success.

Only the `source` changes. Names, descriptions and the marketplace's own identity are the
manifest's, so the universe installs the same four plugins under the same names.
"""
import json
import sys

src, dst = sys.argv[1], sys.argv[2]
d = json.load(open(src, encoding="utf-8"))
for p in d["plugins"]:
    p["source"] = "./plugins/" + p["name"]
with open(dst, "w", encoding="utf-8") as f:
    json.dump(d, f, indent=2)
print(f"[universe] staged a local-path manifest: {len(d['plugins'])} plugin(s) resolve to the checkout")
