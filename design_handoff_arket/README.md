# Handoff: Arket — the Kjernekraft design system

## Overview

Arket is the design system behind the Kjernekraft studio app (Go templates + one
stylesheet + htmx, no build step). This bundle is the documentation of that
system, written so an agent can work inside the real repository without breaking
it, plus four open proposals that need a human decision.

The system already exists and ships. **Nothing here asks you to restyle the app.**

## How to actually use this with Claude Code

1. Copy `ARKET.md` into the repository — `docs/ARKET.md` is the natural home,
   next to `DESIGN_GUIDELINES.md`.
2. Add a pointer at the top of the repo's `CLAUDE.md` (create it if absent) so
   every session picks the rules up without being asked:

   ```markdown
   ## Design system
   Before changing any template, stylesheet or UI handler, read `docs/ARKET.md`.
   It is binding. `static/css/kjernekraft.css` is the only stylesheet; all colour
   lives in `:root`. Run the checks in ARKET.md §11 before committing.
   ```
3. Optional, and worth it: put the same checks in a pre-commit hook or a
   `make sjekk` target. Rules that are only written down drift; rules that fail a
   command do not.
4. When you want the *reasoning* rather than the rules, open `Arket.dc.html` in a
   browser (see below) or read `docs/DESIGN_GUIDELINES.md` in the repo.

`ARKET.md` is deliberately terse and imperative — it is written to be read by an
agent, in English, with every token and class name kept verbatim in Norwegian so
that grep still works. The repo's own docs are in nynorsk; translate it if you
want one voice.

## About the design files

`Arket.dc.html` and `Dagmerke.dc.html` are **design references**, not production
code. They are a browsable, live-rendered specification: the tokens, the type and
space ladders, the light/depth model, and every component drawn from the same
token values as the stylesheet, in both themes.

Do not port them into the app. The app already has the real implementation in
`static/css/kjernekraft.css` and `handlers/templates/`. Use these files to *see*
what a rule means, then implement in the repo's own idiom: Go `html/template`
partials, one stylesheet, htmx for interaction, no client framework, no build
step.

They open directly in a browser. They need the sibling `fonts/` and `img/`
folders, which are copies of `static/fonts/` and `static/img/` from the repo.
`Arket.dc.html` imports `Dagmerke.dc.html`; keep them side by side.

## Fidelity

**High-fidelity.** Every colour, font, size, radius and shadow in the reference
pages is the literal token value from `static/css/kjernekraft.css`. The one place
the pages deliberately diverge from shipped code is the punch-card klippekort on
the *Klippet* page, which is a proposal and labelled as such; the shipped light
band remains documented on the *Komponentar* page.

The day-mark in `Dagmerke.dc.html` is rebuilt from the constants in
`handlers/merkeform.go` (case 58×73, dial at 29,34, index track r=17.6, pie
r=14.7, hands 10.1 and 14.7). It is faithful to those numbers, but
`merkeform.go` is the authority — the marks in `/arket` are drawn by the code
itself.

## What's in the bundle

| File | What it is |
|---|---|
| `ARKET.md` | The rules. This is the file Claude Code should read. Eleven sections: colour, light and depth, nothing shifts, no horizontal scrolling, type, space, words, the mark, component contracts, open questions, pre-commit checks. |
| `Arket.dc.html` | The browsable documentation — ten pages, live components, light/dark toggle, and a `skala` switch that shows the tightened ladder in place. |
| `Dagmerke.dc.html` | The day-mark, rebuilt from `merkeform.go`, imported by the page above. |
| `fonts/`, `img/` | Ubuntu, Ubuntu Condensed, Ubuntu Mono, Union Gothic (temporary, CC BY-NC-ND), and the mark. Copies of the repo's own. |

## The four open questions

These are in `ARKET.md` §10 with full reasoning. They need a decision from the
owner, not from an agent:

1. `.btn-primary` — outline (per `DESIGN_GUIDELINES.md` §18) or fill (per shipped
   CSS). The book and the code contradict each other.
2. The klippekort readout — the light band that ships, or the punch card with the
   count as the hero.
3. The tightened space/type ladder (step e^¼, topping out at e rem).
4. Dead code: `--fordjuping`/`--fasett`, the twice-written dark block, the
   duplicated card box.

## Constraints an agent must not break

- No network requests for fonts, icons or anything else. The page must render
  fully with the studio's connection down.
- No build step, no npm, no client framework. One stylesheet.
- No emoji, no hand-drawn decorative SVG, no drop shadows except
  `--skugge-flytande`, no horizontal scrollbars, no colour outside `:root`.
- Every user-facing string is a key in all three files under `locales/`.
- Union Gothic is licence-restricted and gitignored. Do not commit it, and do not
  add a rule that assumes it is present — the fallback chain must keep working.

## Assets

All fonts and the mark are already in the repository under `static/`. The copies
in this bundle exist only so the reference pages render standalone. There are no
photographs or icon sets in this system; the six category drawings called for in
`DESIGN_GUIDELINES.md` §14 do not exist yet.
