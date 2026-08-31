#!/usr/bin/env bash
# Klassar i stilarket som ingen mal, skript eller handsamar skriv.
#
# Ein klasse som ingen skriv er ikkje berre daud kode — han er ei felle.
# Neste gong nokon leitar etter ein komponent finn dei honom, byggjer paa
# honom, og oppdagar fyrst etterpaa at han aldri hev vore teikna. Difor
# er dette ei prøve og ikkje eit raad.
#
# Skriv ut kvar klasse som er definert i `kjernekraft.css` men aldri
# nemnd i handsamarar/, static/js/, models/ eller database/. Gjeng ut med 1
# um det finst nokon, so han kann standa i ein pre-commit-krok.
#
# Kommentarane i stilarket vert skrella av fyrst: prosa som nemner
# `.noko` er ikkje ein definisjon.
set -euo pipefail
cd "$(dirname "$0")/.."

python3 - "$@" <<'PY'
import os, re, sys

# Stilarket er mange filer og éi adresse. Sjekken lyt lesa *summen*,
# som nettlesaren gjer — ein klasse kann vera definert i ei fil og
# brukt frå ei anna.
import glob
css = "\n".join(open(f, encoding="utf-8").read()
                 for f in sorted(glob.glob("static/css/deler/*.css")))
defined = set(re.findall(r"\.(-?[_a-zA-Z][_a-zA-Z0-9-]*)",
                         re.sub(r"/\*.*?\*/", "", css, flags=re.S)))

words = set()
for base in ("handsamarar", "static/js", "models", "database"):
    for dirpath, dirnames, filenames in os.walk(base):
        dirnames[:] = [d for d in dirnames if d not in (".git", "worktrees", "node_modules")]
        for fn in filenames:
            if fn.endswith((".html", ".js", ".go", ".json")):
                with open(os.path.join(dirpath, fn), encoding="utf-8", errors="ignore") as f:
                    words |= set(re.findall(r"[-_a-zA-Z][-_a-zA-Z0-9]*", f.read()))

dead = sorted(defined - words)
for c in dead:
    print(f".{c}")
if dead:
    print(f"\n{len(dead)} klassar er definerte men aldri skrivne.", file=sys.stderr)
    sys.exit(1)
print("ingen daude klassar", file=sys.stderr)
PY
