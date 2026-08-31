# Kjernekraft — what the program is, and what everything is called

*31 August 2026. Written against the working tree, not against intent.*

This document does two jobs. First it specifies the program: what it does, what
the pieces are, how a request moves through them, and which parts of Go and chi
it stands on. Second it names things — it collects every domain word the
codebase uses, says what each one means, and marks the places where one thing
has two names or one name means two things.

It is deliberately called `ARCHITECTURE.md` and not something from the house
metaphor (`ARKET`, `MASKINEN`, `KORREKTUREN`, `HUSET`). Those names are the
subject of this document, not a style to extend. Someone arriving at this
repository should be able to find this file without knowing the vocabulary
first — which is the whole problem being described here.

**Status of the code it describes:** the prototype is right. What it does and
how it looks is the specification; there is no better statement of intent than
the running program. How it is *built* is what needs work, and the largest part
of that work is naming.

### How to read this

- **§1** — what the program does. Read first.
- **§2** — the Go and chi machinery, from first principles. If you know Go, skim
  it; §2.13 lists what this codebase does unusually.
- **§3–§7** — the specification proper: the request path, the packages, the
  data, the routes, the view.
- **§8–§11** — the vocabulary, where it breaks, and what to do about it. This is
  the part that needs a decision, not a reading.

---

## 1. What the program is

### 1.1 The domain

A management system for a single yoga and pilates studio. The studio sells two
things and tracks one:

- **Memberships** — recurring access, with a binding period, a renewal date,
  and the ability to be frozen.
- **Klippekort** — punch cards. You buy *N* clips, each clip pays for one
  class, and the card expires on a date.
- **Classes** — what you spend the membership or the clip on. A class happens
  in a room, has a teacher, a kind, and a number of seats.

Everything else in the schema exists to support those three.

### 1.2 The three audiences

Every screen belongs to exactly one audience, and that ownership is enforced by
which route group it sits in (§6).

| Audience | Reaches | Does |
|---|---|---|
| **Member** (`elev`) | `/elev/*` | sees the schedule, signs up, holds a membership and punch cards, pays |
| **Administration** | `/admin`, `/api/admin/*` | creates classes and series, sets prices and rules, handles requests and notices |
| **Front desk** (`innsjekk`) | `/innsjekk`, `/api/innsjekk/*` | a kiosk screen at the door — ticks people off, sells drop-ins |

The front desk is the interesting one: it has **no login**. It is a screen
standing in a room the public can walk into, so it gets its own lock
(`innsjekkLås`, a shared code) rather than a user identity. It still gets a
session cookie and a CSRF token like every other request, so the general
protections apply without a special case.

### 1.3 What it is not, and why that matters

- **Not a single-page application.** There is no JavaScript framework, no
  bundler, no `node_modules`. Go renders complete HTML on every request.
- **Not built.** There is no build step of any kind. The CSS is concatenated at
  request time; the templates are parsed at request time; the translations are
  read from disk. `go run .` is the whole toolchain.
- **Not multi-tenant.** One studio. There is no `studio_id` anywhere, and
  adding one later would touch every table.
- **Not connected to payments.** `payment_methods` and `charges` are modelled
  and rendered, but no provider is wired up. `payment_api.go` returns
  hand-built demo charges.

The first two are deliberate and load-bearing. They are why the codebase can be
understood by reading it in file order, and why a change is visible on reload
without a toolchain in between.

### 1.4 External systems

Exactly one. The studio currently books in **Yogo**, a Norwegian booking
platform. `cmd/hent-timeplan` calls its API and writes the schedule into our
`events` table. `yogo/` is the client.

The import is one-directional and one-shot: Yogo does not know we exist, and
nothing syncs back. The relevant impedance mismatch is recorded in
`cmd/hent-timeplan/main.go:122` — Yogo has no concept of room capacity, only
seats per class, and those are two different things in our model (§5.3).

---

## 2. The Go foundations

*This section assumes no Go. If you know it, skip to §2.13, which lists what
this codebase does unusually.*

The useful mental model for Go's web stack is a Lego system, and the metaphor is
tighter than most: Go's whole approach to composition is **one stud pattern that
everything conforms to**. Learn that one shape and the rest of this section is
consequences of it.

### 2.1 One shape, and everything is that shape

The entire Go web world is built on a single interface, from the standard
library:

```go
type Handler interface {
    ServeHTTP(w http.ResponseWriter, r *http.Request)
}
```

That is the stud pattern. **Anything** that has a `ServeHTTP` method with that
signature is a `Handler` and can be plugged in anywhere a `Handler` is expected.
Nothing declares "I implement Handler" — in Go, having the method *is*
implementing the interface. This is structural typing, and in Lego terms: a
brick fits because its bumps are the right size and spacing, not because it came
in the same box.

The two arguments are the whole contract:

- `r *http.Request` — everything that came in. URL, method, headers, cookies,
  body, and a per-request "pocket" called the context (§2.6).
- `w http.ResponseWriter` — everything going out. You call `w.Header().Set(...)`
  to set headers, then `w.Write(...)` to send bytes. **Order matters:** the
  first `Write` sends the headers and the status line, and any header set after
  that is silently discarded. Almost every "why isn't my header showing up" bug
  in Go is this.

There is no return value. A handler that writes nothing produces `200 OK` with
an empty body.

### 2.2 Functions as bricks

Writing a struct with a `ServeHTTP` method for every route would be tedious, so
the standard library provides an adapter:

```go
type HandlerFunc func(http.ResponseWriter, *http.Request)

func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    f(w, r)
}
```

Read that carefully, because the whole ecosystem leans on it: `HandlerFunc` is
*a type whose underlying type is a function*, and it has a `ServeHTTP` method
that just calls itself. So `http.HandlerFunc(myFunc)` turns any plain function
of the right shape into a `Handler`.

Every route handler here is a plain function of that shape, and chi does the
wrapping. That is why `func (a *App) ElevDashboard(w, r)` can be handed straight
to `r.Get(...)`.

### 2.3 Middleware is a brick that swallows another brick

Middleware in Go is not a framework feature. It is a naming convention for a
function of this shape:

```go
func(next http.Handler) http.Handler
```

Take a brick, return a brick. The returned brick usually does something, calls
the one it swallowed, and possibly does something after. Here is this
codebase's simplest real one, `RequireAuth` (`middleware.go:52`):

```go
func RequireAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if GetUserFromSession(r) == nil {
            // don't call next — the chain stops here
            http.Redirect(w, r, "/innlogging", http.StatusSeeOther)
            return
        }
        next.ServeHTTP(w, r)   // hand off to whatever was wrapped
    })
}
```

Three things fall out of this shape, and all three matter downstream:

1. **Not calling `next` ends the request.** That is how a guard works. There is
   no "abort" mechanism; you simply do not hand off.
2. **Middleware nests, it does not queue.** `A(B(C(handler)))` means A's
   "before" code runs first and A's "after" code runs *last*. It is an onion,
   not a pipeline. This is why `Recoverer` must be outermost of ours — its
   `defer recover()` can only catch panics raised inside it.
3. **Middleware can replace the request.** `next.ServeHTTP(w, r)` may be passed
   a *different* `r`. That is how `WithUser` injects the user (§2.6).

### 2.4 chi is a baseplate, and almost nothing else

`github.com/go-chi/chi/v5` is this project's only web dependency, and it is
about a thousand lines. It exists because the standard library's own router
(`http.ServeMux`) historically could not match on HTTP method or group routes
under shared middleware.

What chi gives you:

- **`chi.NewRouter()`** returns a `*chi.Mux`, which is *itself* an
  `http.Handler`. The baseplate is a brick. That is why the last line of `main`
  can be `http.ListenAndServe(addr, r)` — the whole application is one brick
  handed to the standard library.
- **Method-aware routing.** `r.Get(path, h)` and `r.Post(path, h)` register
  different handlers for the same path. `server.go:153-154` does exactly that
  for `/innlogging`: GET draws the form, POST processes it.
- **`r.Use(mw)`** attaches middleware to every route registered on that router.
- **`r.Group(fn)`** creates an inline sub-router that **inherits** the parent's
  middleware and may add its own. The entire permission model rests on this
  (§2.5).
- **`r.Route(pattern, fn)`** mounts a sub-router under a path prefix. Used once,
  at `server.go:255`, for `/api/admin/settings`.
- **Wildcards.** `r.Put("/api/admin/class/*", h)` matches any suffix.

**A constraint worth knowing:** `r.Use()` must be called before any route is
registered on that router, or chi panics. That is enforced, not a style rule.
It is also why `server.go` reads top-to-bottom as middleware-then-routes, and
why the ordering there is not freely rearrangeable.

### 2.5 Groups are how permission works here

The single most important structural idea in the program, and it is four lines
of chi:

```go
r.Group(func(r chi.Router) {
    r.Use(RequireAdmin)
    r.Get("/admin", app.AdminPage)
    // ...every other admin route
})
```

The `r` inside the closure **shadows** the outer `r` — it is a new sub-router.
Every route registered inside inherits the outer middleware (`Recoverer`,
`WithUser`, `CSRF`) *and* gets `RequireAdmin` on top.

Stated plainly: **a handler's permissions are determined entirely by which
`r.Group` block its registration line sits in.** There is no annotation on the
handler, no check inside it, nothing in its name. Move
`r.Get("/admin", app.AdminPage)` up forty lines and the admin panel is public,
with no compiler error and no failing test.

Several handlers carry a comment saying so — `AdminPage`'s first line is
*"Tilgangen vert avgjord av RequireAdmin i rutaren, ikkje her"* (access is
decided by RequireAdmin in the router, not here). That comment is load-bearing
documentation, not decoration.

### 2.6 The request's pocket: `context`

A `*http.Request` carries a `context.Context` — an immutable key-value store
scoped to that one request. You cannot modify it; you derive a *new* request
with a new context:

```go
func withUser(r *http.Request, u *models.User) *http.Request {
    return r.WithContext(context.WithValue(r.Context(), userCtxKey, u))
}
```

The key idiom avoids collisions between packages:

```go
type ctxKey int                  // a private type nobody else can construct
const (
    userCtxKey ctxKey = iota
    csrfCtxKey
)
```

Because `ctxKey` is unexported, no other package can create a value of that
type, so no other package's key can collide with ours — even if it also uses
the integer `0`.

Reading back needs a type assertion, because the value comes out as `any`:

```go
u, _ := r.Context().Value(userCtxKey).(*models.User)   // u is nil if absent
```

**This is the mechanism behind rule 2 in §3.** `WithUser` runs once per request,
does one database read, and puts the user in the pocket. Every later
caller — `brukaren()`, `sessionIsAdmin()`, `sessionUserName()`,
`IsLoggedIn()` — reads the pocket and touches no database. Before this was
tightened, `GetUserFromSession` had a database fallback of its own, which gave
the whole session and middleware layer a database dependency it did not need.

### 2.7 Methods, receivers, and what `*App` actually is

Go has no classes. It has functions *attached* to a type by declaring a
receiver:

```go
func (a *App) ElevDashboard(w http.ResponseWriter, r *http.Request) { ... }
//   ^^^^^^^^ the receiver
```

`a` is an ordinary parameter that happens to be written before the name. The
`*` makes it a pointer, so all 80 methods share one `App` rather than 80 copies.

`App` itself is four lines (`handsamarar/app.go`):

```go
type App struct {
    DB *database.Database
    Nå func() time.Time
}
```

**Why it exists.** Before it, every handler was a plain function reaching for a
package-level `var DB *database.Database`. A global. It worked, but it meant
there was no way to construct a handler pointed at a *different* database — so
no handler could be tested. Twenty-seven test files in the package tested
everything *except* the handlers, all stayed green, and a live admin button sat
behind a `501 Not Implemented` stub without anyone noticing.

With the receiver, a test writes:

```go
a := &App{DB: testDB, Nå: func() time.Time { return fixedMoment }}
w := httptest.NewRecorder()
a.ApproveFreezeRequest(w, httptest.NewRequest("POST", "/...?user_id=1", nil))
```

`httptest.NewRecorder()` is a `ResponseWriter` that records instead of sending.
That is the entire testing story for handlers, and it was unavailable until the
receiver existed.

`Nå` ("now") is a function *field*, not a value — called per use, so a test can
freeze or move time. `func() time.Time` is an ordinary first-class value in Go,
which is what makes this work without an interface.

### 2.8 Interfaces as stud patterns

Go interfaces are declared by the *consumer*, not the producer, and they are
usually tiny. A good example lives in `database/timekolonnar.go`:

```go
type radskannar interface {
    Scan(dest ...any) error
}
```

`database/sql` has two row types: `*sql.Row` (exactly one) and `*sql.Rows`
(many). They are unrelated, but both have a `Scan` method with that signature.
The two-line interface lets one function read either:

```go
func skannTime(rad radskannar, e *models.Event) error { ... }
```

Neither `*sql.Row` nor `*sql.Rows` knows this interface exists. They fit because
the studs line up. That is the payoff of structural typing: you can define a
shape *after the fact* and existing types satisfy it retroactively.

### 2.9 `database/sql`: a socket, not a database

`database/sql` is in the standard library and contains **no database code at
all**. It is a socket that drivers plug into. The driver here is
`github.com/mattn/go-sqlite3`, imported for its side effect only:

```go
import _ "github.com/mattn/go-sqlite3"
```

The `_` means "compile this in, but I am not naming it". The package's `init()`
registers it under the name `"sqlite3"`, which is what makes
`sql.Open("sqlite3", path)` work. Every `_test.go` that opens a database needs
this import, which is why it appears in so many of them.

Four properties of the socket that matter here:

- **`*sql.DB` is a pool, not a connection.** Safe for concurrent use, opened
  once and shared. `sql.Open` does not actually connect; the first query does.
- **`Query` returns many rows and must be closed.** `defer rows.Close()`,
  always — a leaked `*sql.Rows` holds a connection out of the pool forever. The
  `timane()` helper exists partly to make that unforgettable.
- **`QueryRow` defers its error to `Scan`.** It never returns an error itself.
  `sql.ErrNoRows` comes out of `Scan`, which is why `SisteITimeserie` checks
  for it there.
- **Placeholders are `?`, and they are not string interpolation.** The driver
  sends query and values separately, which is what makes SQL injection
  impossible. There are no string-concatenated values anywhere in `database/`,
  and there must not be.

**SQLite settings this project sets deliberately** (`database.go:36-46`):
`_journal_mode=WAL` so readers do not block the writer; `_busy_timeout=5000` so
a contended write waits five seconds instead of failing instantly;
`_foreign_keys=on` because **SQLite does not enforce foreign keys unless asked**
— without it the `REFERENCES` clauses in the schema are decorative;
`_synchronous=NORMAL`, safe with WAL and much faster than `FULL`.

**Transactions** appear at 46 sites, always in the same shape: `Begin()`,
`defer tx.Rollback()`, work, `tx.Commit()`. The deferred rollback after a
successful commit is a no-op, so the shape is safe and means no path can leave
a transaction open. **Prepared statements are used nowhere**; every query is
sent as text with arguments. At this scale that is fine.

### 2.10 `html/template`: escaping is the point

`html/template` looks like a string templating language but is not one. It
parses the HTML, understands where in the document each `{{...}}` sits, and
escapes accordingly — differently inside an attribute, inside a `<script>`,
inside a URL. That is why it is `html/template` and not `text/template`, and it
is why there is no XSS-mitigation code anywhere in this codebase: the protection
is structural.

The parts this project uses:

- **`template.New(name).Funcs(fm).ParseFiles(paths...)`** — a template *set* is
  several files parsed into one namespace. `{{define "x"}}` in any of them
  registers a block callable as `{{template "x" .}}` from any other.
  `template_manager.go` assembles a set per page: base layout, then every
  component and module, then the page file.
- **`FuncMap`** — a map of names to Go functions callable from templates.
  `getTemplateFuncs()` registers about a dozen: `slagklasse` (turn a raw
  `class_type` into a CSS hook), the mark geometry (`markViewBox`, `markWidth`,
  `markHeight`), `asset` for cache-busting URLs, and øre→kroner money
  formatting. Anything needing arithmetic or lookup belongs here, not in the
  template.
- **`ExecuteTemplate(w, "base", data)`** — renders the block named `base` and
  streams it to the writer.

**Two consequences of streaming.** First, output begins before the template
finishes, so an error halfway through leaves a half-written page with the status
already sent — precisely why `Recoverer` must produce something sensible.
Second, **`.Field` on a `map[string]any` renders empty rather than failing if
the key is absent.** A typo in a data-map key is silent. That is the failure
mode §9.7 is about.

### 2.11 `sync.Once` and the read-heavy lock

Ten places use `sync`, in two patterns:

- **`sync.Once`** — `locOnce`, `tmplOnce`, `termsOnce`, `assetVersionOnce`,
  `osloEinGong`. Each guards a "build this expensive thing exactly once, no
  matter how many requests arrive simultaneously" initialisation: translations,
  the template set, the parsed terms markdown, the asset version hash, the Oslo
  timezone.
- **`sync.RWMutex`** — `localization.mu`, `templateManager.mu`, `stilarkLås`.
  Many readers, rare writer. `RLock()` lets any number of requests read at once;
  `Lock()` is for the rare rebuild. This matters because in development the
  templates and stylesheet are re-read from disk while other requests are still
  rendering.

### 2.12 The dependency list, in full

Six direct dependencies. That is the whole supply chain.

| Module | Why |
|---|---|
| `go-chi/chi/v5` | the router (§2.4) |
| `gorilla/sessions` | signed, encrypted cookie sessions — `sessionStore` in `session.go` |
| `mattn/go-sqlite3` | the SQLite driver. **Uses cgo**, so builds need a C toolchain |
| `russross/blackfriday/v2` | renders the terms-and-conditions markdown to HTML |
| `golang.org/x/crypto` | `bcrypt`, for password hashing and checking |
| `leanovate/gopter` | property-based testing, used in two files under `test/` |

`gorilla/securecookie` is indirect, pulled in by `gorilla/sessions`.

The cgo dependency in `go-sqlite3` is the one with real consequences: it makes
cross-compilation awkward and slows clean builds. A pure-Go SQLite driver
exists (`modernc.org/sqlite`) if that ever becomes painful.

### 2.13 What this codebase does unusually

If you already know Go, these are the things worth flagging:

1. **chi is used as a mux and little else.** No `chi.URLParam` anywhere. The
   wildcard route `/api/admin/class/*` has its ID extracted by string slicing:
   `path[len("/api/admin/class/"):]` (`admin_classes.go:324`). Registering
   `/api/admin/class/{id}` and calling `chi.URLParam(r, "id")` is what chi is
   for. See R8.
2. **Templates are parsed per request in development and cached in production**,
   rather than embedded with `//go:embed`. This is what makes the no-build-step
   property real, at the cost of the binary not being self-contained — it needs
   `templates/`, `static/` and `mål/` on disk beside it.
3. **No `context` propagation into the database layer.** No `QueryContext`, no
   cancellation. A client that disconnects mid-request leaves the query running
   to completion.
4. **`database.Database` is a one-field struct** wrapping `*sql.DB`. It carries
   no state of its own, so it is effectively a namespace for 109 methods.
5. **Errors are returned but rarely wrapped.** No `fmt.Errorf("...: %w", err)`
   chains, no `errors.Is`/`errors.As` beyond a couple of `== sql.ErrNoRows`
   comparisons. A failure deep in the database layer arrives at the handler
   with no trail.

---

## 3. The one path everything follows

Every request takes the same route. This is the most important diagram in the
document, because almost every naming decision below sits somewhere on it.

```
  browser
    │
    ▼
  server.go ─── chi router, the ONLY place access is decided
    │
    ├─ middleware.Logger      chi's own; logs method, path, status, duration
    ├─ Recoverer              panic → 500 with a body (outermost of ours)
    ├─ [dev only] no-store    stops the browser caching a half-broken page
    ├─ WithUser               ONE database read → user into the request context
    ├─ CSRF                   token required on anything that changes state
    │
    ├─ RequireAuth            ┐
    ├─ RequireAdmin           ├─ route groups; being in the group IS the permission
    ├─ RequireDevelopment     ┘
    ▼
  handsamarar ─── *App methods, one per route
    │              reads the user from the context, never from the database
    │
    ├──▶ database ─── *Database methods; every query in the program lives here
    │       │
    │       └──▶ SQLite (WAL, foreign keys on, 5s busy timeout)
    │
    ├──▶ models ───── the shapes that travel between database and template
    │
    └──▶ renderPage ── template_manager → templates/layouts/base.html
                         │
                         ├─ sidedata() supplies CurrentPage/Lang/CSRFToken/IsAdmin/UserName
                         └─ base.html picks which JS to load by matching CurrentPage
```

Three rules this path enforces. Each was learned the hard way, and each is
written in a comment at the place it applies.

### 3.1 Access is decided in `server.go`, never in a handler

The mechanism is §2.5. The consequences:

- A handler contains no permission check, so you cannot read a handler and learn
  who may call it. You must find its registration line.
- Moving a registration line between groups silently changes its security.
  Nothing catches this — not the compiler, not the tests.
- The kiosk group is the exception that proves the rule. It has no auth
  middleware; instead each of its handlers calls `innsjekkKrav`, a shared
  preamble that checks the lock. That is a *deliberate* in-handler check,
  because there is no user to authenticate.

### 3.2 The user is read once

- `WithUser` is `r.Use`'d at `server.go:94`, before any route is registered, so
  it covers every route without exception.
- It skips anything under `/static/`. The browser sends the session cookie with
  every CSS, font and image request; without the guard each static file cost a
  database query for a logged-in user.
- In development it also re-reads the translations per request, so a new key
  appears without a restart.
- Because it always runs first, `GetUserFromSession` only reads the context. It
  has no database access, which keeps the whole session and middleware layer
  free of one.

### 3.3 The clock is injected

- `App.Nå` is the studio's clock. Beneath it sits
  `config.GetInstance().GetCurrentTime()`, which can be moved in development to
  answer "what does the schedule look like next Tuesday".
- User-facing time decisions go through it: the two-hour signup and cancellation
  windows in `events.go`, and the membership page in `medlemskapet.go`.
- Three places still call `time.Now()` directly and are *correct* to: the login
  rate-limiter in `middleware.go` (wall-clock is right for a security control),
  the demo charge dates in `payment_api.go`, and the RNG seed in `test_data.go`.

---

## 4. The pieces

| Package | Size | Holds |
|---|---:|---|
| `server.go` | ~300 lines | the router, and therefore the access model |
| `handsamarar/` | ~12,300 lines | 68 HTTP handlers as `*App` methods, plus view-model building |
| `handsamarar/templates/` | 53 files | the HTML |
| `handsamarar/modules/` | 2 files | view-model structs for two reusable blocks |
| `handsamarar/config/` | — | settings, notably the injectable clock and the timezone |
| `database/` | ~6,200 lines | 109 `*Database` methods, the schema, and the migrations |
| `models/` | ~470 lines | the shapes that cross the database↔template boundary |
| `yogo/` | ~1,030 lines | the Yogo API client |
| `mål/` | 3 × 708 keys | translations: `nn` (default), `nb`, `en` |
| `static/css/deler/` | 35 files | the stylesheet, served as one file |
| `static/js/` | 17 files | small behaviours; htmx plus one script per concern |
| `cmd/hent-timeplan` | — | imports the schedule from Yogo |
| `cmd/tryggjing` | — | takes a safe backup while the server runs |
| `test/`, `scripts/` | — | test-data seeding, property tests, one-off scripts |

### 4.1 `handsamarar` is one flat package

81 files (54 source, 27 test) in a single namespace. Three consequences:

- **Every symbol is reachable from every other.** There are no internal
  boundaries; `admin_classes.go` can call anything in `innsjekk.go`.
- **Name collisions are possible and have happened.** §9.1 is a direct result.
- **The `*App` receiver gives it a seam but not a boundary.** It is now
  testable. It is still flat. §11.4 is the open question.

The files divide into three unlabelled kinds, which would be the natural seams:

1. **Route handlers** — `dashboard.go`, `admin_*.go`, `events.go`,
   `membership_api.go`. They speak HTTP.
2. **View-model builders** — `merkeform.go`, `aktivitet.go`, `timeplanbolk.go`,
   `helsing.go`, `timeserie.go`. Pure functions producing shapes for templates.
   These have no HTTP in them at all, and they are where nearly all the existing
   tests are, precisely *because* they were the only testable part.
3. **Infrastructure** — `middleware.go`, `session.go`, `template_manager.go`,
   `localization.go`, `stilark.go`, `berging.go`, `fragment.go`.

### 4.2 `database` is a god object with a good habit

- 109 methods on one `*Database` type wrapping a single `*sql.DB`.
- **The good habit:** every query in the program lives in this package. There is
  no SQL in `handsamarar`. That boundary is real and unbroken, and it is the
  best structural property the codebase has.
- **The bad habit:** the type knows about memberships, klippekort, events,
  groups, permissions, discounts, private classes, attendance and messages.
  There is no domain decomposition, only file-level grouping.
- `database.go` alone is ~2,600 lines and is the older, English-named layer. The
  per-feature files (`gruppe.go`, `frammote.go`, `rabattkrav.go`,
  `privattime.go`) are the newer Nynorsk layer. Nearly every inconsistency in §9
  splits along that seam.

### 4.3 `models` is the one consistent thing

Fourteen plain-data types with no behaviour beyond a handful of derived
accessors: `Event.Ledige()`, `Event.Full()`, `Event.LengdMin()`,
`MembershipWithDetails.GjeldTil()`, `User.KvalifisertFor()`.

The `WithDetails` suffix means "joined with the thing it points at, for
display". It is applied consistently across `KlippekortWithDetails`,
`MembershipWithDetails` and `ChargeWithDetails` — the only naming convention in
the codebase with no exceptions.

---

## 5. The data

Eighteen tables, created by `database.Migrate`. One more — `produktnamn` — is
created separately by `MigrerProduktnamn`, called from `main` immediately after
`Migrate`. That split is historical, not principled.

### 5.1 The spine

```
users ──┬── user_memberships ──── memberships
        ├── user_klippekort ───── klippekort_packages
        ├── event_signups ─────── events ──┬── rooms
        ├── brukarloyve ───────── loyve     ├── grupper (via gruppe_id)
        ├── gruppemedlem ──────── grupper   └── events (via serie_id, self)
        ├── rabattkrav
        ├── user_payment_methods ─ payment_methods
        └── charges
                                  meldingar
                                  membership_rules
                                  produktnamn
```

### 5.2 `events` is overloaded

One row is one class on one day. But three separate concepts also ride on it:

- **`serie_id`** points at other rows in the same table, making a recurring
  class. A series is not a table; it is a shared ID among sibling rows. That is
  why "edit the series" means "update *N* rows in one transaction"
  (`admin_serielagring.go` exists entirely for this), and why a half-applied
  update was a real bug worth a long comment: three classes moved, the rest did
  not, and the interface said only "could not save".
- **`gruppe_id`** restricts the class to members of a group. Absent means open.
- **`private_user_id`** makes it a private session for exactly one person.

The visibility rule combining the last two lives in one SQL fragment,
`synlegFor`, concatenated into several queries. That fragment is the single
source of truth for "may this person see this class" — and it is a string, so
nothing checks that its two `?` placeholders are passed the same argument twice
at every call site.

### 5.3 Capacity is computed and must never be read raw

A class's own `capacity` of `0` means *"no capacity of its own"* — and then the
room's number applies. The canonical expression:

```sql
COALESCE(NULLIF(e.capacity, 0), r.capacity, 0)
```

- It now lives in exactly one place, `eventKolonnar` in
  `database/timekolonnar.go`. It was written out eleven times before that.
- `models.Event` carries **both**: `Capacity` (computed) and `EigenPlassar`
  (the class's own). The administration needs the difference, because a room of
  12 with an override of 12 is indistinguishable in the computed value.
- `EigenPlassar` is a `*int`, and `nil` means "sets none of its own". It was an
  `int` where `0` carried that meaning — the same shape of mistake as reading
  `e.capacity` raw, because `0` cannot say whether it is an answer or an
  absence. The column is selected with `NULLIF(e.capacity, 0)` so that both
  `NULL` and legacy `0` rows arrive as `nil`.
- Reading `e.capacity` directly is a bug, and has been one twice: once making
  every signup answer "class is full", and once making `GetEventByID` report 0
  seats where `GetAllEvents` reported 18 for the same class.
- The query must include the `rooms` join, or half the columns come back NULL.
  That requirement is documented on `eventKolonnar` and is the reason `eventFrå`
  exists as a separate constant.

### 5.4 The Black membership is not a row

`database/svartmedlem.go`. Teachers and developers get full access without
buying anything and without appearing in the price list where members choose.

- It is **derived from a permission**, not stored in `user_memberships`.
- That is the whole design decision. A stored row would keep giving a former
  teacher free access until someone cleared it by hand; a permission removed
  takes the membership with it in the same instant, because it never was
  anything else.
- Consequently it cannot be frozen, cancelled or changed — there is nothing to
  change. The template hides those buttons, keyed on a flag called `Tildelt`.

### 5.5 A freeze stops a clock

- `user_memberships.frozen_at` records when the freeze began; `NULL` means not
  frozen.
- On unfreeze the expiry is pushed forward by the elapsed time, so a year card
  gives twelve months of *use*, not twelve months of calendar.
- The status flow is `active → freeze_requested → paused → active`. The middle
  step needs an administrator, which is why `ApproveFreezeRequest` and
  `RejectFreezeRequest` exist.
- `UnfreezeMembership` is the **only** way out of `paused`.
  `UpdateMembershipStatus` accepts any status and knows nothing about the clock;
  called with `"active"` on a frozen membership it leaves `frozen_at` behind and
  the expiry drifts permanently. That is documented on the method and is a real
  trap.

### 5.6 Time is Oslo wall-clock, stored without a zone

- Values are stored as `"2026-08-27 16:30:00"` — the clock on the studio wall,
  no offset.
- `handsamarar/tid.go` owns the conversion; `veggtekst()` formats a `time.Time`
  for storage.
- Do not do zone arithmetic across this boundary without reading that file. A
  query comparing a stored value against a zone-aware `time.Time` will be wrong
  by one or two hours depending on the season.
- The week is ISO, Monday-start. `vike.go` owns that reckoning, extracted from
  the schedule handler because inside it, it could not be tested.
- Related performance note: `GetEventsForWeek` uses a half-open range
  (`>= monday AND < next_monday`) rather than `DATE(e.start_time)`, because the
  function call defeated the index and forced a table scan.

---

## 6. The surface

Sixty-eight handlers, fifteen of which render a full page. The grouping in
`server.go` *is* the permission model, so it reads as a table:

| Group | Middleware | Routes |
|---|---|---|
| open | — | `/`, `/signup`, `/terms`, `/innlogging`, `/glemt-passord`, `/logout`, `/static/*` |
| kiosk | own lock | `/innsjekk`, `/innsjekk/laas`, `/api/innsjekk/*` |
| member | `RequireAuth` | `/elev/*`, `/api/user/*`, `/api/membership/*`, `/api/klippekort/*`, `/api/events/*`, `/api/payment-methods/*` |
| admin | `RequireAdmin` | `/admin`, `/api/admin/*`, `POST /api/events` |
| development | `RequireDevelopment` | `/api/shuffle-*`, `/api/setup-test-data`, `/elev/testdata` |
| workshop | — | `/arket` |

### 6.1 Two kinds of route, with nothing in the names to tell them apart

- **Page routes** return a full document through `renderPage`. Fifteen of them.
- **Fragment routes** return an HTML snippet for htmx to swap into an existing
  page. These use `fragment.go`, not `renderPage`, because a fragment must not
  carry the base layout. `UserKlippekort`, `Heimehovud`, `LedigPlass`, `Charges`
  and `PaymentMethods` are of this kind.

Getting this wrong is not a compile error — it renders a whole page inside a
`<div>`. §9.6 and R5 are about giving the distinction a lexical marker.

### 6.2 `Heimehovud` is a workaround with a good reason

The greeting and the week summary on the home page are the only two things that
say something about your signups *in words*. The lists beneath them refresh
themselves via htmx; these two did not, and so they lied at exactly the moment
you signed up. `Heimehovud` is one endpoint returning both, refreshed together,
so the page and the refresh cannot disagree.

It exists because htmx's `HX-Trigger` response header is used nowhere in this
codebase. With it, the signup response could tell the page to refresh those two
elements, and this endpoint would be unnecessary.

More broadly, htmx is spoken thinly here. The request header `HX-Request` is
read in three production sites — `berging.go:50`, `middleware.go:92`,
`feilsvar.go:24`, all of them answering "is this a fragment request, so should
the error be a fragment too?" The response headers `HX-Redirect`, `HX-Trigger`
and `HX-Refresh` appear **nowhere**. htmx is used as a fetch-and-swap library
and little more, and `Heimehovud` is the cost of that.

---

## 7. The view

### 7.1 One stylesheet, many files

`static/css/deler/*.css` are numbered by layer and concatenated by
`handsamarar/stilark.go` into a single response at
`/static/css/kjernekraft.css`.

- One `<link>`, one fetch, no `@import` chain loading in rounds, no build step.
- They are split because a single 4,000-line file once lost its entire dark
  theme to one bad find-and-replace. A file per section bounds the damage to
  that section.
- The numbering *is* the cascade: `00` tokens, `05–25` chrome and primitives,
  `30–45` the big surfaces, `50–99` per-page and per-component.

### 7.2 Values and meaning live apart

- `00-palett.css` holds hex values and nothing else.
- `00-token.css` gives them meaning and maps them into four theme blocks: light,
  dark-by-preference, dark-by-choice, light-by-choice.
- `TestVerdiOgTydingBurKvarForSeg` enforces the split;
  `TestDeiTvoMyrkeBlokkaneErSamde` enforces that every dark token has a light
  counterpart.
- `TestSlagfargarKjennerDeiSameSlagi` holds the stylesheet's class colours and
  Go's `Slagi` list in agreement — one of the few Go↔CSS contracts that is
  actually checked.

### 7.3 Templates are three tiers

- `layouts/base.html` — the page shell.
- `pages/*.html` — one per route.
- `components/`, `modules/` — reusable blocks, `modules/` being the larger
  feature-scoped ones (`modules/admin/`, `modules/dashboard/`,
  `modules/membership/`).

`template_manager.go` assembles a set per page by parsing base, then every
component and module, then the page file (§2.10).

### 7.4 `base.html` picks the JavaScript by string-matching

Seven conditionals at `base.html:77-100` compare `.CurrentPage` against literal
strings to decide which scripts load. The Go side now uses typed `Sidenykel`
constants, and `TestMalenKjennerBerreSidenyklarSomFinst` walks the templates and
fails if any literal has no matching constant.

The coupling is wider than `base.html`: `navigation.html` uses `.CurrentPage`
too, for highlighting the active nav item.

### 7.5 Every visible string is a translation key

- `{{t .Lang "bolk.nykel"}}`, present in all three files under `mål/`, 708 keys
  in each.
- `TestMalaneBedBerreUmNyklarSomFinst` fails if a template asks for a key that
  does not exist. `TestIngenUmsetjingErTom` fails if a key exists but is empty.
- **Translations are cached per process** via `sync.Once`. In production a new
  key needs a restart; in development `WithUser` re-reads them per request. If
  labels render as raw keys after adding one, restart before debugging anything
  else.

---

## 8. The vocabulary

This is the part an outside reader needs and cannot currently get.

### 8.1 Domain nouns

| Word | Means | Appears as |
|---|---|---|
| **time** (pl. *timar*) | a class — one session on one day | `events` table, `models.Event`, `TimarIVindauget`, `InnsjekkTime` |
| **timeserie**, **serie** | a recurring class: "yoga with Leon, Monday 18:00" | `events.serie_id`, `timeserie.go`, `LagraSerie` |
| **timeplan** | the schedule (the week view) | `timeplan.go`, `timeplanbolk.go`, `/elev/timeplan` |
| **slag** | kind of training — fascia, yoga, pilates, reformer | `events.class_type`, `slag.go`, `--slag-*`, `.slag-yoga` |
| **elev** | a member (literally "pupil") | `/elev/*`, `ElevDashboard` |
| **medlemskap** | membership | `memberships`, `user_memberships`, `medlemskapet.go` |
| **klippekort** | punch card; **klipp** = one punch | `user_klippekort`, `klippekort_packages` |
| **løyve** | permission | `loyve`, `brukarloyve`, `SettLøyve` |
| **gruppe** (pl. *grupper*) | a group of members; restricts a class | `grupper`, `gruppemedlem`, `events.gruppe_id` |
| **rom** | room — a resource with a capacity | `rooms`, `romkonflikt.go`, `RoomConflict` |
| **frammøte** | attendance — actually turning up | `event_signups.attended_at`, `frammote.go` |
| **innsjekk** | check-in; the front-desk kiosk | `innsjekk.go`, `/innsjekk` |
| **frysing** | freezing a membership, which stops its clock | `user_memberships.frozen_at` |
| **rabattkrav** | a claim for the student/senior discount | `rabattkrav`, `admin_rabattkrav.go` |
| **melding** (pl. *meldingar*) | a notice for the administration to handle | `meldingar`, `HandsamaMelding` |
| **privattime** | a class reserved for one person | `events.private_user_id`, `privattime.go` |
| **svart medlemskap** | the invisible "Black" membership, derived from a permission | `svartmedlem.go`, `--svartkort` |
| **vike** | week | `vike.go`, `veketal.js`, `VikorIAaret` |
| **mål** | language | the `mål/` directory, `localization.go` |
| **produktnamn** | a product's display name, per language | `produktnamn` table |
| **ledig** | available (a free seat) | `ledige.go`, `LedigPlass`, `Event.Ledige()` |

### 8.2 Interface nouns

| Word | Means |
|---|---|
| **arket** | "the sheet" — the design system, and `/arket`, the page that renders every component outside its work |
| **verkstaden** | "the workshop" — the CSS for that page |
| **merke** | the mark: both the brand mark and the SVG class badge drawn as a clock (`merkeform.go`) |
| **dagmerke** | the day badge on a schedule row |
| **briefing** | a line of prose telling you how things stand before you open anything — *not* `setning` |
| **setning** | "the sentence" — the class-creation form written as a sentence rather than twelve stacked fields |
| **dokka** | "the dock" — holds the thing you are looking at still while the rest scrolls |
| **faner** | tabs |
| **hovudet** / **botnlina** | header / footer |
| **kortet** | the card component |
| **folk** | "people" — the member lookup screen (a search, not a table) |
| **skissone** | the teacher portraits in settings |
| **ljos** | light — the theme toggle |
| **blekk** / **ark** | ink / sheet — the two ends of the colour system |
| **hårlina** | hairline — the 1px rule that separates a card from the sheet, in place of a shadow |
| **glans**, **lippe** | gloss, lip — surface treatments in the token file |
| **berging** | rescue — the panic recoverer |
| **tryggjing** | a backup |

### 8.3 Words that only look like domain terms

`handsamar` = handler. `stilark` = stylesheet. `mal` = template. `skjema` =
form. `nykel` = key. `spurning` = query. `base` = database. `bolk` = section.
`stoda` = status. `prøve` = test. `lås` = lock. These are engineering words in
Nynorsk and carry no studio meaning.

---

## 9. Where the names break

Eight problems, worst first.

### 9.1 `Session` means two different things in one package

- `handsamarar/timeplanbolk.go:27` — `type Session struct` is **a class as it
  runs on one particular day**: `Event`, `Day`, `Date`, `DayName`, `IsToday`,
  `IsPast`.
- `handsamarar/session.go` — `sessionStore`, `sessionName`, `sessionUserID`,
  `GetUserFromSession`, `SetUserInSession`, and the user-visible error string
  `"Session error"` are **the HTTP login session**.

Both live in package `handsamarar`, so both are in scope everywhere. This is the
most dangerous name in the codebase: both readings are plausible at every site,
and Go will not warn you. `AvailableSessions` returns the first kind;
`"Session error"` refers to the second; they sit forty lines apart in
`membership.go`.

Note this is a *direct consequence* of §4.1 — in two packages, both names would
be fine.

### 9.2 The central noun has two names

The thing the studio sells is a **class**. The code calls it:

- `Event` — 123 references. The table, `models.Event`, `EventSignup`,
  `GetEventByID`, `CreateEvent`, `events.go`.
- *time / timar / timen* — 336 references. `TimarIVindauget`, `InnsjekkTime`,
  `timeplan`, `timeserie`, `privattime`, `LedigPlass`.

Neither is wrong; having both is. And `Event` is independently inaccurate —
nothing here is an event in the ordinary English sense. It is a scheduled,
recurring, bookable class.

### 9.3 One colour, two names

The palette carries the four class colours twice:

| Hex (light) | Meaningful name | Legacy name |
|---|---|---|
| `#00908c` | `--l-slag-yoga` | `--l-klas` |
| `#6d2e9e` | `--l-slag-pilates` | `--l-tuneup` |
| `#d6006e` | `--l-slag-reformer` | `--l-togu` |
| `#c8901a` | `--l-slag-fascia` | *(was `--l-fjellraeven`)* |

- `--togu` is still used 23 times, `--tuneup` 7 times.
- `--tuneup-svak` is used **zero** times — dead.
- `--merke`, the brand colour, is defined as `var(--klas)`, so the brand mark
  and the yoga class share a hex. Whether that is intentional cannot be
  recovered from the code, which is the real cost.
- `togu`, `tuneup` and `fjellräven` are equipment and clothing brands. They were
  mnemonics for "the pink one", "the purple one". No outside reader recovers
  that; `--slag-*` says what the colour is *for*.

### 9.4 The names are in two languages, and ten are in both at once

Of the 70 exported `*App` methods: **40 English-only**, **20 Norwegian-only**,
**10 mixed inside a single identifier** — `ElevDashboard`, `KlippekortPage`,
`PurchaseKlippekort`, `AdminPrisar`, `AdminReglar`, `AdminKlippeprisar`,
`UserKlippekort`, `ShuffleUserKlippekort`, `Heimehovud`, `Gruppemedlem`.

The same split runs through the comments — 22 files in `handsamarar` mix Nynorsk
and English prose, 29 are Nynorsk-only, 17 English-only — and through the
database, where `database.go` is an older English layer and the per-feature files
are the newer Nynorsk one.

### 9.5 `mål` means two things

`mål` is the translations directory and the Nynorsk word for *language*. It is
also the ordinary word for *target*, used that way as a parameter:
`vekedagssteg(frå time.Weekday, mål int)` in `admin_classes.go:285`.

### 9.6 `-Page` is applied to a third of the pages

Five carry it — `KlippekortPage`, `MembershipPage`, `SignUpPage`, `AdminPage`,
`TestDataPage`. Ten do not — `Arket`, `Betaling`, `Innlogging`, `MinProfil`,
`ElevDashboard`, `ElevTimeplan`, `Terms`, `GløymtPassord`, `Innsjekk`,
`InnsjekkLås`. All fifteen render a full page.

Nor does anything distinguish a page handler from a fragment handler (§6.1) —
which is the distinction that would actually be useful.

The terms page shows how far this spreads: the URL is `/terms`, the handler is
`Terms`, the template is `pages/vilkaar.html`, and the page key is
`SidaVilkaar`. Four names, two languages, one page.

### 9.7 Verb prefixes are inconsistent, and the failure is silent

`GetAllEvents`, `GetMembershipRules`, `GetUserPaymentMethods` use `Get`.
`PaymentMethods`, `Charges`, `UserKlippekort`, `UserSignups` are the same kind
of thing without it. In `database/`, `Get*` and bare nouns alternate freely.

This matters more than it looks, because of §2.10: a template asking `.Foo` on a
`map[string]any` with no `Foo` renders empty rather than failing. `sidedata()`
now guarantees the five keys every page needs, but every page-specific key is
still an unchecked string on both sides of the boundary.

### 9.8 One effect, three vocabularies *(fixed 31.8.2026)*

The mark's bevel — the thing that makes a shape read as pressed into the
sheet — was named three different ways in three layers:

| layer | called |
|---|---|
| the colour | `--lippe-djup` / `--lippe-ljos` |
| the SVG filter that shapes it | `#merkedjup` / `#merkeljos` |
| the class that wears it | `.form-djup`, `.skivedjup`, `.rutedjup`, `.dauddjup`, and `.skiveskugge` / `.ruteskugge` for the soft layer |

Every element filtered by `#merkedjup` is filled with `var(--lippe-djup)`.
They are the same thing, and the palette had already settled the name: the
**lip**. The filter called it *depth*; the class called it *depth* or *shadow*
depending on the variant.

Worse, *djup*/*ljos* name the metaphor rather than the effect. Nothing in
`#merkedjup` says it is half of a matched pair, that its partner goes on the
opposite edge, or that `-lite`/`-mjuk` pick a different *element* (the window
versus the dial) rather than a different *intensity*.

**Resolved:** the filters now carry the palette's name — `#lippe-djup`,
`#lippe-ljos`, `#lippe-djup-smal`, `#lippe-ljos-smal`, `#lippe-djup-fall`.
`#merkekant` and `#merkering` keep their names: they are genuinely edges
(`feMorphology` outlines), not lips, and the names say so. Twelve references
across two files; verified pixel-identical.

The CSS classes (`.skivedjup`, `.ruteskugge`, …) still carry the third
vocabulary. They are a separate pass.

This is §9.3 — one thing, two names — one level down, and it is the reason
this document exists.

---

## 10. Proposed rules

Proposals, not decisions. Stated as rules so they can be argued with concretely.

### R1 — One language per layer, and the layer boundary is the language boundary

| Layer | Language | Reason |
|---|---|---|
| SQL schema — tables, columns | **English** | already English; renaming columns is a migration with no user-visible benefit |
| `models/` field names | **English** | serialised to JSON and read by templates |
| Go identifiers | **Nynorsk** | house language; the domain words have no clean English equivalents (`klippekort`, `løyve`, `slag`) |
| Comments and docs | **Nynorsk** | pick one; the current 50/50 split is the worst available state |
| CSS class names and tokens | **Nynorsk** | already consistent |
| URLs | **Nynorsk** | already mostly so |
| Translation keys | **English** | they are identifiers, not prose |

Fix the ten mixed-inside-one-identifier names (§9.4) first — they are
indefensible under any rule, including the current absence of one.

### R2 — The class is a `Time`, not an `Event`

Rename `models.Event` → `models.Time`. The `events` table stays, per R1. This is
the highest-value rename in the codebase and it removes §9.2.

If `Time` colliding with `time.Time` is judged too risky — a real concern, since
`models.Time` and `time.Time` would appear in the same signatures — the fallback
is `models.Klasse`. What must not happen is keeping both `Event` and *time*.

### R3 — `Session` becomes `Dagstime`; the login session keeps `session`

`Session` in `timeplanbolk.go` is a class on a particular day; `Dagstime` says
that. The auth concept matches every reader's expectation of the word "session",
so it keeps it. `AvailableSessions`/`EnrolledSessions` become
`LedigeDagstimar`/`PaameldeDagstimar`. Resolves §9.1.

### R4 — Colours are named for what they are for

Delete `--togu`, `--klas`, `--tuneup` and their `-svak` derivatives; use
`--slag-*` throughout. Give `--merke` its own value rather than aliasing
`--klas`, so the question of whether the brand shares the yoga colour is
answered in the file instead of being unanswerable. Resolves §9.3.

### R5 — Page handlers are named for the page; fragments are marked

`Klippekort`, `Medlemskap`, `Registrering`, `Admin`, `Testdata`, `Arket`,
`Betaling`, `Innlogging`, `MinProfil`, `Heim`, `Timeplan`, `Vilkår` — no
suffix, and the `Sidenykel` constant, the template file and the handler all
carry the same word.

Fragment handlers take a marker (`Bit`-, or a `-bit` suffix) so §6.1's
distinction becomes visible at the call site and in `server.go`. Resolves §9.6.

### R6 — `Get` only when there is a matching setter

`GetMembershipRules`/`SaveMembershipRules` keeps the prefix; the pair is the
point. `GetAllEvents` becomes `AlleTimar`. Resolves §9.7.

### R7 — `mål` is only ever the language

Rename the parameter in `vekedagssteg` to `til`. Resolves §9.5.

### R8 — Use chi for what chi is

Replace `/api/admin/class/*` with `/api/admin/class/{id}` and
`chi.URLParam(r, "id")`. Not a naming rule, but the same family: the current
code hand-parses a path because the router's own facility was never reached for
(§2.13.1).

---

## 11. What is decided, and what is open

### Decided by the running program — treat as fixed

1. Server-rendered, no build step, htmx for partials.
2. Access is decided in `server.go` by route group.
3. One stylesheet, split by layer, values separate from meaning.
4. Every visible string is a translation key.
5. Nynorsk is the default language; the domain vocabulary is Nynorsk.
6. The clock is injectable; times are Oslo wall-clock.
7. All SQL lives in `database/`.

### Open, and needing a decision before renaming starts

1. **`Event` → `Time` or `Klasse`?** (R2) Everything else is smaller than this,
   and the choice constrains R3.
2. **Which language for comments?** (R1) The half-migration is worse than either
   endpoint. One decision, then a mechanical pass.
3. **Is `events` one table or three?** `serie_id`, `gruppe_id` and
   `private_user_id` are three concepts sharing a row (§5.2). Splitting them is
   a design question, not a rename — and the only item here that changes the
   schema.
4. **Does `handsamarar` stay one package?** 81 files in one namespace (§4.1).
   Splitting by audience — `elev`, `admin`, `innsjekk`, plus a shared `vising`
   for the view-model builders — would give it boundaries, and would have made
   §9.1 impossible.
5. ~~**Should the narrow database probes return `models.Event` at all?**~~
   **Resolved.** `RoomConflict` and `RoomConflictUtanSerie` now return
   `*Romkollisjon`, `PaameldeYver` returns `*Overfylt`, `GetFutureEventsBySerie`
   returns `[]SerieTime`, and `NettFrammott` returns a plain `bool` — every
   caller only ever asked *whether*, never *which*. The types are declared in
   `database/timekolonnar.go` beside `eventKolonnar`, with a comment explaining
   why a half-filled `models.Event` is the same failure as a sentinel zero.
   `TimarIVindauget` still returns `[]models.Event` with nine of twenty fields
   set; it feeds the kiosk list and is the one left.
6. **Does the fragment/page distinction get a name?** (R5) It is a real
   architectural split with no lexical marker, and getting it wrong renders a
   whole page inside a `<div>`.

---

## Appendix A — where to start reading

1. `server.go` — the whole surface and the whole access model, in one file.
2. `handsamarar/app.go` and `handsamarar/sidedata.go` — the seam and the page
   contract. Both are short.
3. `database/timekolonnar.go` — the canonical class query, and the comment
   explaining what drifted before it existed.
4. `handsamarar/middleware.go` — every middleware in the program, in one file,
   in the order they wrap.
5. `handsamarar/timeplanbolk.go` — the densest domain logic and the best example
   of the house comment style.
6. `docs/ARKET.md` — binding for anything visual.

## Appendix B — related documents

- `docs/ARKET.md` — the design system. **Binding**; read before touching any
  template, stylesheet or UI handler.
- `docs/MASKINEN.md` — a machine audit, findings F1–F13+. History, not current
  state: several are now fixed (the copied render epilogue, the duplicated
  `DB`/`AdminDB`, the 300-line recommendations handler).
- `docs/STRUCTURE_AUDIT.md` — an earlier structural pass. Also history.
- `docs/LOCALIZATION.md` — how to add a language.
- `docs/DESIGN_GUIDELINES.md`, `docs/KORREKTUREN.md` — visual and prose
  standards.
