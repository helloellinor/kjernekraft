# *Kjernekraft, Rebuilt* — outline

**Decisions taken (31 Aug 2026):**

| §A | Decision |
|---|---|
| 1. Slice | Timeplan + signup |
| 2. Chapters | Thirteen, four parts |
| 3. Format | **Read-alongside** — chapters explain the real files in place; no second app is built |
| 4. Language | **Nynorsk**, matching R1 and the house style. (I had recommended English; overridden, and the book follows the code's language.) |
| 5. On disk | `docs/boka/`, one file per chapter, with an index |
| 6. Pace | **Part I first** (ch. 1–4), read and corrected, then the rest |
| 7. Publish | Deferred |

The chapter plan below stands as written, with one adjustment for the
read-alongside decision: chapters no longer end with "type this in", but with
*"open this file and you can now read it"*. The motivating order is kept — a
problem is felt before its fix is named.

---

## The premise

Three commitments, taken from what makes the Blazor/Web Forms book work:

1. **Anchored on C++.** Every new idea opens with the thing you already do in
   C++, then shows Go's answer and *why it was made different*. Never Go in
   isolation.
2. **One running example, and it is your own app.** Not a toy. We build the
   **class schedule and signup flow** — `/elev/timeplan` and
   `POST /api/events/signup` — from an empty file to what is on disk today.
3. **Every chapter ends with something that runs**, and with a named capability:
   *"you can now do X without me."*

### Why that slice

The timeplan + signup path is the spine of the program. Verified against the
real code, it touches every mechanism in the codebase:

| Layer | What the slice uses |
|---|---|
| routing & access | `RequireAuth` group; `brukaren(r)` |
| the App seam | `a.DB`, `a.Nå()` |
| database | four queries incl. `GetEventsForWeek`, `GetEventByID` |
| the zero-value trap | capacity: room vs class's own |
| view models | `BuildWeekRows`, `weekOptions` |
| templates | `renderPage`, a page + a fragment |
| errors | 404, 400-with-reason, 500 |
| tests | pure helpers *and* an httptest handler |

Nothing important in Kjernekraft is missing from it, and nothing in it is
padding.

---

## Part I — The machine

### 1. Orientation
- What you have: 68 handlers, 109 queries, 53 templates, and why the prototype
  being *correct* is the thing worth protecting.
- The three commands: `go build ./...`, `go test ./...`, `gofmt -w .`. No CMake,
  no headers, no include guards.
- **C++ anchor:** what you give up (build system, headers, `const`) and what you
  get (compile in seconds, one formatter, one test runner).
- **Ends with:** the server running on :18108 and you knowing why it's not 8080.

### 2. A server that answers
- `func main`, `http.ListenAndServe`, one handler function.
- **The one stud pattern:** `http.Handler`, and `HandlerFunc` as the adapter that
  turns a function into one. This is *the* idea; everything later is a
  consequence.
- **C++ anchor:** no framework needed — the stdlib is the server. Compare to
  picking Beast/Drogon/Crow.
- `w.Header().Set` before `w.Write`, and why the order is not negotiable.
- **Build:** a handler that returns "timeplan goes here".
- **Ends with:** you can add a route that returns text.

### 3. Routing, and where permission actually lives
- chi: `NewRouter`, `Get`/`Post`, and the fact that the router is *itself* a
  Handler.
- **Middleware as a brick that swallows a brick:** `func(next http.Handler) http.Handler`.
  Write `RequireAuth` by hand.
- The onion, not the pipeline — why `Recoverer` must be outermost.
- `r.Group` and shadowing: **build the access model**, then state the rule it
  earns — *"a handler's permissions are which group its registration line is
  in."*
- **C++ anchor:** decorators without inheritance; composition over a base class.
- **Ends with:** `/elev/timeplan` returns 303 to login when logged out, 200 when
  in — and you can explain why with no annotation on the handler.

### 4. Structs, receivers, and the App
- Structs as plain aggregates. Methods as free functions with a receiver.
  **There is no `this`.**
- Pointer vs value receivers, and when each is right.
- No constructors → the `NyApp` convention. No inheritance → composition.
- **The chapter's argument:** start with the global `var DB`, feel it work, then
  hit the wall — *you cannot test this* — and introduce `*App` as the fix. The
  seam is motivated, not decreed.
- `Nå func() time.Time` as a field: functions are values.
- **C++ anchor:** class → struct+receiver; constructor → convention;
  dependency injection without a DI container.
- **Ends with:** your handler is a method, and the clock is something a test can
  move.

## Part II — The data

### 5. Talking to SQLite
- `database/sql` is a socket with no database in it. The `_` blank import and
  what `init()` does.
- `*sql.DB` is a pool, not a connection.
- `Query` / `QueryRow` / `Scan`; `defer rows.Close()`; why `QueryRow` defers its
  error to `Scan`.
- Placeholders are `?` and are *not* interpolation — SQL injection made
  structurally impossible.
- **Errors are values.** `if err != nil` for the first time, honestly: it's
  noisy, here's why it exists, here's what you get back (control, no hidden
  exits).
- **Build:** `TimarForVika` from scratch — the real `GetEventsForWeek`.
- **C++ anchor:** no RAII → `defer`, function-scoped, written at the call site.
- **Ends with:** the schedule page shows real classes from the database.

### 6. The zero value, and when it lies ★
- *The chapter that vindicates your C++ instinct.*
- Go zeroes everything, always. No uninitialised memory — a whole C++ bug class
  gone.
- **But:** `0` is a legitimate capacity *and* the value you get when nothing was
  fetched. `SerieID == 0` means "no series" and "didn't ask for it".
- The real bug, reproduced: `GetEventByID` said 0 seats where `GetAllEvents`
  said 18, for the same class. Write the failing test first.
- `std::optional<T>` → `*T` / `sql.NullT`, and why Go made them uncomfortable.
- `COALESCE` vs `NULLIF`: where absence gets destroyed, in SQL, before Go ever
  sees it.
- **Half-filled structs are the same lie one level up** — a `*models.Event` with
  four of twenty fields set. Narrow types (`Romkollisjon`, `Overfylt`) as the
  answer; `NettFrammott` returning `bool` because every caller only asked
  *whether*.
- **Ends with:** you can spot a sentinel and know the two ways to kill it.

### 7. Interfaces, small and declared late
- Implicit satisfaction: no `: public Drawable`, ever.
- **Build:** you wrote two nearly identical scan loops in Ch. 5 — one for
  `*sql.Row`, one for `*sql.Rows`. Now write `radskannar` (two lines) and delete
  one of them.
- Consumer-declared, one or two methods. Why `io.Reader` has exactly one and
  `IUserRepository` is a Java habit.
- **C++ anchor:** closest to C++20 concepts, but checked at compile time *and*
  dispatched at runtime, with no template instantiation and no code bloat.
- **Ends with:** you can collapse duplicate code by naming the shape it shares.

## Part III — The view

### 8. Rendering HTML
- `html/template` is not a string templater — it parses the HTML and escapes by
  *position*: differently in an attribute, a `<script>`, a URL.
- **Therefore:** there is no XSS code in this codebase, and that is not an
  oversight. Contrast with what you'd own yourself in C++ with inja/Mustache.
- Template *sets*: `{{define}}` / `{{template}}`, `ParseFiles`, `FuncMap`.
- **The silent failure:** `.Foo` on a map with no `Foo` renders empty, never
  errors. This is why `sidedata()` and typed `Sidenykel` constants exist.
- **Build:** the timeplan page, then extract `sidedata`.
- **Ends with:** you can add a page and know what will silently render blank.

### 9. Fragments, htmx, and the page/fragment split
- Page vs fragment: `renderPage` vs `fragment.go`, and what happens when you
  confuse them (a whole document inside a `<div>` — no compile error).
- `CurrentPage` string-matching in `base.html`, the typed-constant fix, and the
  test that keeps Go and the template in agreement.
- Where htmx is under-used here: `HX-Trigger` is unused, and `Heimehovud` is the
  price.
- **Ends with:** signing up updates the list without a page reload.

### 10. Concurrency you never wrote
- There are **zero** `go` statements, channels or `select` blocks in Kjernekraft
  — and your handlers are concurrent anyway, because `net/http` runs each
  request in its own goroutine.
- Therefore: shared mutable state needs a lock even though you never spawned a
  thread. `sync.Once` for the template set and translations; `RWMutex` for
  many-readers-one-writer.
- Deliberately short. Channels get a page saying "you don't need these yet, here
  is when you would."
- **C++ anchor:** `std::mutex`/`shared_mutex`, `call_once` — the concepts
  transfer directly.
- **Ends with:** you can tell when a package-level variable needs protecting.

## Part IV — Confidence

### 11. Testing, and why it was impossible
- The cautionary tale, true and from this repo: **27 test files, all green, and
  a live admin button wired to `501 Not Implemented`.** Everything *except* the
  handlers was tested, because handlers reaching a global cannot be tested.
- `httptest.NewRecorder()` and `httptest.NewRequest()`; the Ch. 4 seam paying
  off.
- Table-driven tests — the Go idiom, and it suits your C++ instincts.
- Testing against a real temp SQLite, not a mock. Why mocks are rare in Go.
- **Build:** the freeze-approval test, and one for the signup window using a
  frozen `a.Nå`.
- **Ends with:** you can test any handler in the program.

### 12. Errors, properly
- What the codebase does: 451 bare `if err != nil`, no wrapping, no trail.
- What it should do: `fmt.Errorf("...: %w", err)`, `errors.Is`, `errors.As`, and
  sentinel errors.
- `panic` is not an exception — it's an assert. `brukaren()` panics on purpose.
  `Recoverer` and `defer recover()`.
- **C++ anchor:** exceptions vs values, honestly compared — what each costs.
- **Ends with:** an error from the database layer arrives at the handler with a
  trail.

### 13. Adding a feature alone
- Capstone. One real feature, end to end, with nothing new taught — only the
  checklist:
  route → group → handler method → query → narrow type → view model →
  template → translation keys in all three files → typed page key → test.
- **Ends with:** the point of the book.

## Appendices
- **A.** C++ → Go cheat sheet (lift from `GO.md` §12).
- **B.** Compiler errors you will hit, and what they actually mean —
  *declared and not used*, *cannot use X as Y*, nil map assignment, `append`
  without reassignment.
- **C.** The vocabulary — pointer to `ARCHITECTURE.md` §8, not a duplicate.
- **D.** Where the real code differs from the book's version, and why.

---

## §A — Decisions I need from you

1. **Slice.** Timeplan + signup, as argued above? The alternative is the
   membership/freeze flow — more business logic, less view work. I recommend
   timeplan; freeze appears in Ch. 11 as the test subject either way.

2. **Chapter count.** Thirteen is faithful to the eBook's depth. I could
   compress Parts III–IV into four chapters and lose the htmx and errors
   material. **I'd keep thirteen** — the errors chapter is where the codebase is
   genuinely weakest.

3. **Build-alongside, or read-alongside?** Does the reader type the code into a
   scratch package and end with a second working app, or read while pointing at
   the real files? Build-alongside teaches far better and is more work to write.
   **I recommend build-alongside**, with each chapter ending in a diff against
   the real file.

4. **Language.** English, like `ARCHITECTURE.md` and `GO.md` — or Nynorsk, like
   the code's house style? R1 in `ARCHITECTURE.md` says docs should be Nynorsk.
   **I recommend English for this one** and would note the exception: it's a
   teaching text for a C++ audience, and the Go/HTTP vocabulary it teaches is
   English regardless.

5. **Shape on disk.** One file per chapter under `docs/boka/` with an index, or
   one long `docs/BOOK.md`? **I recommend a directory** — thirteen chapters is
   3,000–5,000 lines and one file is unnavigable.

6. **Pace.** This is several sessions, not one. I'd write **Part I first**
   (chapters 1–4), you read it, and we correct the pitch before I continue.
   Writing all thirteen before you've read any risks getting the level wrong
   thirteen times.

7. **Publish?** I can also put each part up as an Artifact — a private page with
   proper typography, rather than markdown in a terminal. Your call; the files
   stay in the repo either way.
