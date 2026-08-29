# Maskinen — ettersynet

*29 August 2026. All of `handlers/`, `database/`, `models/`, `server.go`,
`handlers/templates/`, `static/js/`, and the test suite. Three parallel
read-throughs, cross-checked against `server.go` routing. The stylesheet is
not covered here — ARKET.md already governs it and its audit is done.*

This is the ARKET-style audit of everything that is **not** the stylesheet.
The question it answers: why does the code feel bespoke, and what is the
smallest set of mechanisms it should collapse into?

**The headline: the machine is not missing abstractions — it stopped
adopting the ones it already grew.** `renderPage`, `RequireAuth`/
`RequireAdmin`, `teiknFragment`, `veggtekst`, `stadfest.js`, `endringar.js`,
`faner.js`, the `faner` partial, `lesSerieskjema` — every one of these is the
right mechanism, already written, already commented as the intended way. And
at half the call sites the old hand-rolled version still stands beside it.
The stylesheet had the same disease before the 2026-08-28 split, and the cure
is the same: name the mechanism, migrate every call site, write the rule
down, and add a check that keeps it true.

A second, quieter finding: **the half-adoption is hiding real bugs.** They
are listed in §7 and are worth triaging before, not after, the cleanup.

---

## 1. The vocabulary the machine should have

When the worklist below is done, there is exactly one way to do each of
these, and this section becomes the rulebook (the ARKET of the machinery):

| Job | Mechanism | Today |
|---|---|---|
| render a page | `renderPage` + one typed `PageData` | 8 use it, 6 hand-roll it byte-identically, 1 more variant in `admin.go` |
| render a fragment | `teiknFragment`/`teiknFragmentFraa` | 7 use it, 2 hand-roll it |
| return an error | `svarFeil(w, r, status, nykel)` — localized, htmx-aware (to write) | 248 `http.Error` sites in 6 styles and 3 languages |
| return JSON | `svarJSON(w, status, v)` (to write) | 30 hand-rolled encoder sites, 15 of them the identical `{success, message}` map |
| read a form value | `heiltal(r, felt)` / `tekst(r, felt)` in the shape of `lesSerieskjema` (to write) | ~14 hand-written ParseInt-and-400 blocks, 4 message styles |
| know who is logged in | `RequireAuth`/`RequireAdmin` + a `brukaren(r)` accessor | middleware runs, **and** 24 dead `user == nil` checks + ~34 dead method guards remain |
| run several statements | `db.tx(func(tx) error)` (to write) | 14 hand-rolled begin/rollback/commit blocks, 3 with multi-branch commits, 1 sibling missing its transaction entirely |
| scan an event | `eventKolonner` const + `scanEvent(rows)` (to write) | 15 scan sites, 9 drifted SELECT lists |
| write a timestamp | `veggtekst` | 31 sites use it, 4 bypass it → mixed formats in one column → a 5-layout fallback parse loop exists to cope |
| read a timestamp | `lesTid(s)` (to write, from the loop at `database.go:2606`) | 3 strict single-layout copies that silently yield the zero time |
| add a column/table | one `[]migrasjon` registry + `harKolonne`/`harTabell` | 6 coexisting migration mechanisms, 15 pragma probes with inconsistent error handling |
| load a fragment into a div | `hx-get` + `hx-trigger="load"` | also two divergent hand-written `hent()`s for the same endpoints |
| confirm a destructive press | `data-stadfest` (`stadfest.js`) | also `confirm()` ×2 and ad-hoc disable-on-click |
| show a server error to the user | one channel: `.svar[role=status]` + `.gale`, fed by `svarFeil` + `feil.js` | **nine** distinct client-side mechanisms, incl. `alert()` and an unstyled `.message` block |
| tabs | `faner` partial + `faner.js` | also 3 hand-written tab strips, 16 hand-written panels, and 2 more tab controllers (`dagfokus.js`, `folk.js`) |
| a small popover | `veljar(knapp, meny)` (to write) | 4 copies of the same open/outside-click/aria dance |
| a translated string in JS | `data-t-*` on the target element | 4 incompatible conventions (`toJS`, `data-t-*`, `ADMIN_TEXTS`, inline `'{{t}}'`) |
| load templates in a test | `lastMalane` (`loyvemerke_test.go:12`) | 3 competing idioms across 12 sites, one of which mutates cwd and the singleton |

---

## 2. Handlers (`handlers/*.go`, `server.go`)

Routing itself is healthy: four chi groups, one middleware each. The rot is
inside the handlers.

- **F1 — dead auth guards.** 24 `GetUserFromSession(r) == nil` checks
  (19 → 401, 5 → redirect) all sit inside `RequireAuth`/`RequireAdmin`
  groups and cannot fire. The redirect variants also disagree with the
  middleware on status code (303 vs 307). Several admin files already carry
  the comment «Tilgangen vert avgjord av RequireAdmin i rutaren, ikkje her» —
  the convention exists; it was never applied backwards. ~70 lines.
- **F2 — dead method guards.** 37 `Method not allowed` blocks; chi registers
  per-method, so ~34 are unreachable. The live ones are the dual-registered
  handlers (`admin_priser.go`, `admin_settings.go`, `users.go`, `innsjekk.go`,
  `membership.go`). ~100 lines.
- **F3 — the render epilogue, copied.** Six byte-identical hand-rolled page
  renders (`dashboard.go:147`, `betaling.go:44`, `membership.go:82,506`,
  `users.go:42`, `admin.go:135` as a seventh variant) — all drop
  `charset=utf-8` and the reload-retry that `renderPage` has. Two hand-rolled
  fragment renders (`dashboard_components.go:62`, `payment_api.go:181`).
  `renderPage`'s own doc comment says it was written to end exactly this.
- **F4 — the data map, rebuilt 14 times.** `CurrentPage`/`Lang`/`CSRFToken`/
  `IsAdmin`/`UserName` hand-assembled per handler; `base.html:76-100`
  string-matches `CurrentPage` to decide which scripts load, so a typo ships
  a page with no JS, silently. Worst: `membership.go:465-503`, a 24-key map
  with 9 keys that re-spell `user.X` beside `"User": user`.
- **F5 — errors in six styles.** English constants, Nynorsk constants,
  Bokmål, raw `err.Error()` leaked to the client (8 sites), two localized
  sites, and one `http.StatusTeapot`. No error path below 500 produces
  htmx-visible output — see bug B1.
- **F6 — htmx is barely spoken.** `HX-Request` read in 3 places; no
  `HX-Redirect`/`HX-Trigger`/`HX-Refresh` anywhere. Post-mutation redirects
  swap whole pages into fragment targets; `HeimehovudHandler` exists only to
  work around the missing `HX-Trigger`.
- **F7 — near-duplicate handler families.**
  `AdminPriserHandler` ≈ `AdminKlippeprisarHandler` (~250 lines, one generic
  price-table editor); the 8 membership handlers across
  `membership_api.go`/`membership_management.go` (each should be ~6 lines);
  the 4 innsjekk kiosk handlers sharing a ~20-line preamble (→ `RequireKiosk`
  middleware); signup/cancel in `events.go` sharing 42 lines; 3 shuffle
  handlers in `test_data.go`.
- **F8 — the 300-line outlier.** `MembershipRecommendationsHandler`
  (`membership.go:96-395`): two inline template strings with inline CSS and a
  locally redefined func map, bypassing `TemplateManager` entirely. The
  single largest copy-paste in the tree.
- **F9 — two clocks.** `config.GetInstance().GetCurrentTime()` at 15+ sites
  vs raw `time.Now()` at 4 (`events.go:97,137` sit in user-facing window
  checks). The injected-clock design is undermined exactly where it matters.
- **F10 — leftovers.** `EventHandler`/`NewEventHandler` (an entire second
  handler style, never instantiated); `handlers.DB` and `handlers.AdminDB`
  are the same object assigned twice; `romkonflikt.go:62` has the comment
  for a type that was never created.

## 3. Database (`database/`, `models/`)

Two strata are visible: an older English layer inside `database.go` (2,629
lines) and a newer Nynorsk per-feature layer. Nearly every finding splits
along that seam.

- **F11 — five ways to hold the handle.** `*Database` methods (125), free
  funcs on `*sql.DB` (14), free funcs on `*sql.Tx` (3), two narrow one-off
  interfaces, and one method that ignores its receiver. Converge on one
  `querier` interface so no function needs a transactional twin
  (`brukKlipp` vs `BokPrivatTime` is the standing casualty — the klipp-
  consumption logic is duplicated because one side runs in a tx).
- **F12 — 37 hand-written scan loops.** `models.Event` alone: 15 scan
  sites, 9 SELECT variants, and the drift is live (see B4).
  `COALESCE(NULLIF(e.capacity,0), r.capacity, 0)` is written 12 times.
  Memberships: 4 variants (see B5). Users: 3 variants whose past drift
  already caused two documented NULL-scan outages.
- **F13 — transactions by hand.** 14 sites, consistent shape, but 3
  functions commit from multiple branches, and
  `CancelUserSignupForEvent` runs un-transactionally the exact race its
  sibling `SignupUserForEvent` guards against (B3). A `db.tx(fn)` helper
  removes ~70 lines and the multi-commit branches.
- **F14 — time chaos.** `veggtekst` is the intended single writer (31
  sites, 15-line comment explaining why) — and 4 write paths bypass it,
  which is why `FolkOversyn` needs a 5-format fallback loop, and why the 3
  strict `time.Parse` copies with discarded errors yield zero times (B6).
  Membership dates are a fourth convention (`"2006-01-02"` strings, built
  inline 6×).
- **F15 — six migration mechanisms**, no version table, idempotence
  re-derived differently each time; the pragma-probe error is swallowed in
  4 places and returned in 3, and the swallowing variant already caused one
  silently-skipped migration (comment at `database.go:380`). And the
  `charges` table is never created at all (B2).
- **F16 — duplicate function families.** 7 `UpdateSerie*` (one SET column
  apart), 3 `RoomConflict*` (one AND term), 3 `*Fleire` tx-loops,
  `GetTodaysEvents`/`GetThisWeeksEvents` (WHERE clause apart), and ~8 more
  pairs listed in the survey. Each family is one parameterized function.
- **F17 — errors unwrapped.** 20 `== sql.ErrNoRows` comparisons, `errors.Is`
  never used, one sentinel total, `sql.ErrNoRows` abused as a domain error
  twice, messages in three languages. One `ErrIkkjeFunne` + `%w` wrapping.
- **F18 — the model boundary leaks both ways.** 6 entity types live in
  `database/` instead of `models/` (`Person` even parses role strings in its
  scan loop); `DaysUntilExpiry`/`IsExpiring` are computed identically in two
  handlers while the model method that reads them silently no-ops if a
  caller forgot; `CanCancel`/`CanPause` are set exactly once, to `false`,
  and a handler comment already flags it; `models.FreezeRequest` duplicates
  `UserMembership` field-for-field with different types to suit one scan.
- **F19 — the visibility rule exists three times** (`synlegFor` in SQL,
  `KannSjaaTimen` in Go, and inline in `SignupUserForEvent`). It is a
  security predicate; three copies of it is how one goes stale (B7).

## 4. Templates (`handlers/templates/`)

i18n coverage is excellent (539 `{{t}}` calls; one hardcoded label outside
the specimen sheet). The problems are structural.

- **F20 — only 4 partials are real vocabulary** (`setning`, `faner`, `ljos`,
  `dagmerke`/`timeliste`). Everything else has one caller; four
  "components" are page-specific script dumps; `profile-scripts.html` has
  **zero** callers and toggles DOM that no longer exists — delete.
- **F21 — the unsaved-changes dock is copied verbatim five times**
  (`admin-klippekort-prisar:52`, `admin-membership-rules:59`,
  `admin-medlemskapsprisar:58`, `admin-class-management:743`,
  `min-profil:93`). → `{{template "endringsdokk"}}`.
- **F22 — tabs half-adopted.** The `faner` partial has 5 callers **and** 3
  hand-written tab strips **and** 16 hand-written tabpanel quadruples.
  → `{{template "fanerom"}}` + route the 3 strips through `faner`.
- **F23 — the price-row editor is two near-identical modules** with a real
  contract bug between them: `admin-klippekort-prisar.html:29` converts
  øre→kr in the template, `admin-medlemskapsprisar.html:26` receives kroner
  from Go. Two money conventions on adjacent tabs (B8).
- **F24 — three form-row vocabularies** (`.form-group`, `.lappfelt`,
  `.felt`), two empty-state classes (`tomt` ×11, `no-data` ×6), three
  spellings of the loading shell. Pick one of each.
- **F25 — `admin-class-management.html` is 1,372 lines** — 24 % of all
  template code, with 5 embedded script controllers.
- **F26 — the func map carries 42 helpers**: 7 dead (`translate` is a
  byte-for-byte duplicate of `t` with 0 uses vs 539), 13 used once, five
  time-formatters where one would do, and 7 date-name helpers split between
  hardcoded-Norwegian tables and the key-based `vekedagnykel`/`maanadnykel`
  regime — `norskDatoAar` emits Norwegian month names on pages rendered in
  three languages. `toJS` is a hand-rolled escaper that misses `<`, `&`,
  U+2028/9.
- **F27 — `arket.html` duplicates production markup as frozen copies**
  (status rows, faner, empty states) and inlines the dead-square variant of
  `dagmerke`, which is the only reason the four `mark*` func-map helpers
  exist. The workshop should call the same partials the pages do — that is
  its stated purpose.

## 5. JavaScript (`static/js/` + inline scripts)

- **F28 — htmx vs hand-rolled fetch: 22 fetch sites in 13 files against 9
  htmx attributes.** The same endpoint (`/api/charges`) is loaded three
  ways, two of them divergent `hent()` copies that disagree on the failure
  path (one XSS-prone `innerHTML` concat — B9). The
  `if (!r.ok) throw new Error(t)` idiom is pasted 6×, and 3 more sites
  discard the message.
- **F29 — nine error-display mechanisms** (§1 table). Six `alert()` calls;
  one block styling classes that don't exist in the stylesheet; verbatim
  display of `http.Error` bodies that are not localized and not even in one
  language.
- **F30 — three tab controllers, four popover implementations** — merge into
  `faner.js` (pluggable "what a tab shows") and one `veljar()` helper.
- **F31 — CSRF has one mechanism and one bypass** (`innsjekk.html` reads a
  meta tag by hand; `csrf.js` already covers fetch globally).
- **F32 — re-binding after swaps is a lottery**: 2 files re-bind on
  `htmx:afterSwap`, the rest rely on document-level delegation or break
  silently when swapped. Delegation style itself comes in four flavours.
- **F33 — `stadfest.js` loads only on `medlemskap|betaling|admin`**
  (`base.html:82-84`), so a `data-stadfest` button anywhere else is a
  one-click destroyer (B10). Load it everywhere; it's 92 lines.

## 6. Tests

- **F34 — template loading in 3 idioms across 12 sites**; `lastMalane`
  (`loyvemerke_test.go:12`) already does it right without touching cwd or
  the singleton. Standardize.
- **F35 — the stylesheet-building helper exists twice plus one inline copy**
  (`lesStilarket` vs `stilarkFraaProva`), and the walk-all-templates loop
  exists three times with different roots and inconsistent zero-match
  guards.
- **F36 — three hand-rolled string dedups** (`unike`, 2× `ulike`) with
  subtly different semantics; Go 1.24 has `slices.Sort`+`slices.Compact`.

## 7. Bugs surfaced by the audit — triage these first

- ~~**B1 — every 4xx error is invisible to the user.**~~ **Fixed
  2026-08-29.** `feil.js` swaps on ≥ 400 now, and `svarFeil`
  (`handlers/feilsvar.go`) is the one door an error leaves through:
  localized, htmx gets `berging.go`'s div, everyone else plain text. The
  middleware denials (401/403/CSRF) and the whole registration flow go
  through it, with `ErrEpostIBruk`/`ErrTelefonIBruk` sentinels so the two
  conflicts the user can act on get their own words. The remaining ~240
  `http.Error` sites are step 2 of §8 — they are *visible* now, just not
  all localized.
- ~~**B2 — the `charges` table is never created.**~~ **Fixed 2026-08-29.**
  `Migrate` creates it (columns exactly as `SimulateBilling` writes them),
  and the INSERT's timestamps now go through `veggtekst` so the new column
  is born with one format.
- ~~**B3 — `CancelUserSignupForEvent` is un-transactional**~~ **Fixed
  2026-08-29.** One transaction; the DELETE's `RowsAffected` is the
  membership check (race-free), and the decrement is clamped at zero.
- **B4 — event scan drift is live**: `GetEventsForWeek` selects `e.color`
  without the COALESCE the others use (a NULL color kills that query's
  Scan), and it alone got the NULLIF capacity fix — today/week/all views can
  disagree on the same event.
- **B5 — `GetMembershipByID` omits `skjult`**, so a membership fetched by ID
  always claims to be visible.
- **B6 — zero-time dates**: `meldingar.go:111` and `privattime.go:295` parse
  timestamps with one strict layout and a discarded error; values with
  timezone suffixes (which the 5-layout loop proves exist) become
  0001-01-01.
- **B7 — the visibility predicate has three copies** (F19); a fix to one
  does not reach the others.
- **B8 — two money conventions on adjacent admin tabs** (F23).
- **B9 — `klippekjop-scripts.html:96` interpolates into `innerHTML`**
  unescaped.
- **B10 — `data-stadfest` outside medlemskap/betaling/admin is a one-click
  button** (F33).
- **B11 — three routed handlers return 501 «Not implemented»**
  (`admin.go:156-164`), two payment handlers parse an ID and return a fake
  200, and `payment_api.go:78-154` serves 76 lines of hardcoded mock charges
  to real users. Decide: implement or unroute.
- **B12 — raw `time.Now()` in the signup/cancel window checks**
  (`events.go:97,137`) escapes the injected clock the rest of the house
  uses — untestable and wrong under simulated time.

## 7b. Fraa attersynet av f50e471 (2026-08-29) — det som stend att

*Later addition, same day:* the user's eye found what no reviewer did —
the lips **never rendered as lips on any mark**. The band filters
subtract the shape from a shifted copy of its own alpha, which assumes
an opaque fill; the lips are 22–40 % translucent, so the subtraction
never cancelled the interior and every "band" rendered as a uniform
film. Fixed by saturating alpha (`feComponentTransfer`, slope 255)
before the band is built, plus a second soft shadow layer for the time
window (`#merkedjup-mjuk`, the two-layer shadow idiom). Verified by
rendering with headless Chrome and looking, both themes. The lesson for
§8 step 7's checks: reviewers read code and confirmed intent; only
rendered pixels showed the truth.

The multi-agent review of the window-lip commit confirmed a handful of
findings beyond this audit. Fixed same day: the filter test's blind
spots (ids were harvested from raw files, so dropping `merke_defs` from
`base.html` stayed green — proven caught now; quoted `url("#…")` forms;
the `id=` regex's missing left boundary; the `ILDetHeile` typo),
`.form-djup`'s hand-mixed lip (inverted in dark; now 55 %/73 % of
`--lippe-djup`), the missing `#merkeljos-lite` (the window's light lip
was bigger *and* softer than its dark one), the three contradicting
lip-doctrine comments, and the dead cell pasted in two files (now
`{{template "daudmerke"}}`).

Still open, needs a decision or its own pass:

- **`models/klippekort.go:34` + `membership.go:47`**: embedded structs
  both promote `json:"id"` at equal depth, so `encoding/json` silently
  drops the field — the two standing `go vet` warnings. Which id the
  wrapper means is a semantic choice, not a mechanical fix.
- **ARKET §9's «filters never in a shared defs»** contradicts
  `merke-defs.html`, which legitimately solved the theming trap by
  making the filters colorless. The rule needs splitting (proposed to
  the user 2026-08-29): colour-bearing defs stay in-figure; colourless
  filters may be shared.
- **Dev-mode efficiency set**: `WithUser` re-reads all three locale
  files per request; CSRF middleware runs on `/static/` (token-mint
  race on a cookie-less first load); template loading is
  O(pages × components) and the test suite does ~34 full reloads.
- **`base.html:76-100`'s script matrix** string-matches `CurrentPage`,
  with a second hand-kept copy in `hovudskript_test.go` — folds into
  the typed `PageData` work in §8 step 3.
- **`design_handoff_arket/Arket.dc.html`** still specs the pre-fix
  two-layer window.

## 8. Arbeidslista

In order. Each step ships alone and the tests stay green after it. Steps
1–4 are mechanical; 5–7 need judgement.

1. ~~**Delete the dead.**~~ **Done 2026-08-29.** The 24 auth guards are
   `brukaren(r)` now — an accessor in `session.go` that assumes the
   middleware ran and panics if it did not (Recoverer turns that into a
   real 500). The ~34 unreachable method guards, the `EventHandler`
   trio, `handlers.AdminDB` (same object as `DB`, assigned twice),
   `profile-scripts.html`, `database/events.go`, `isColumnExistsError`,
   `stilarkTid`, the dead `var _ = models.Event{}` in `romkonflikt.go`,
   and the 7 dead func-map helpers incl. `translate` are gone. The dual
   GET/POST handlers (`InnloggingHandler`, `AdminSettingsHandler`,
   `MinProfilHandler`, `InnsjekkLaasHandler`) keep their method
   branches — those are reachable.
2. **One error channel, end to end.** Write `svarFeil` (localized,
   htmx-aware, `berging.go`'s div shape); migrate the 248 `http.Error`
   sites file by file; widen `feil.js` to ≥ 400; collapse the nine client
   mechanisms onto `.svar`/`.gale`; delete the `alert()`s. Fixes B1 and the
   three-language error soup in one motion.
3. **One render path.** `type PageData struct` + `sidedata(r, tittel,
   side)`; migrate the 6+2 hand-rolled renders onto
   `renderPage`/`teiknFragmentFraa`; make `CurrentPage` a typed constant so
   `base.html`'s script matching can't miss. Then the partials:
   `endringsdokk`, `fanerom`, the loading shell, one empty-state class.
4. **One DB idiom set.** `db.tx(fn)` (fixes B3); `eventKolonner` +
   `scanEvent` (fixes B4); `veggtekst` at the 4 bypasses + `lesTid` (fixes
   B6); `svarJSON`/`heiltal` on the handler side; `ErrIkkjeFunne` + `%w`.
5. **The migration registry.** `[]migrasjon` + `schema_versjon` table,
   `harTabell`/`harKolonne` promoted; create `charges` in it (B2).
6. **The judged consolidations.** One price-table editor for the two admin
   price modules (settle the øre/kroner contract, B8); merge the 8
   membership handlers; `RequireKiosk`; extract
   `MembershipRecommendationsHandler`'s inline templates; one source for
   `synlegFor` (B7); move the 6 entity types to `models/`; merge the tab
   controllers and popovers; one JS-i18n convention (`data-t-*`); replace
   the `hent()`s with `hx-get`; decide B11's stubs.
7. **Write the rulebook and arm the checks.** §1 of this file, rewritten in
   ARKET's voice once each row is true, plus greppable checks in the §11
   spirit: no `json.NewEncoder` outside `svarJSON`, no `http.Error` outside
   `svarFeil`, no `time.Now()` in `handlers/`, no `== sql.ErrNoRows`, no
   `rows.Next()` outside the scan helpers, no `fetch(` in templates.

The end state is the one the user asked for: a machine with a short list of
named parts, where the answer to «how do I X here» is one row in one table,
and a grep proves nobody went around it.
