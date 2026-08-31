# Go, for someone who knows C++

*A primer written against this codebase. Every example is real code from this
repository, with a file reference.*

The fastest way into Go from C++ is to learn what was **deliberately removed**.
Go is not a small C++; it is a language designed by people who thought C++ had
too many ways to do things and cut most of them. Almost every "where is X?"
question has the answer "there isn't one, on purpose."

Read §1 first — those five differences cause most early confusion. §2–§10 are
the language, and **§7 (zero values) is the one that matters for the change we
are about to make**. §11 reads a whole real file line by line.

---

## 1. The five things that will trip you first

| C++ | Go |
|---|---|
| `int x = 5;` | `x := 5` — type inferred, `:=` declares |
| `class Foo { void bar(); };` | no classes; a method is a free function with a receiver |
| `class Foo : public Bar` | no inheritance at all; embedding and interfaces only |
| `throw`/`catch` | no exceptions; errors are ordinary return values |
| `private:` / `public:` | **capitalisation**: `Name` is exported, `name` is not |

Two more that bite quietly:

- **Unused variables and imports are compile errors**, not warnings. This feels
  hostile for about a day and then you stop writing dead code.
- **Formatting is not a preference.** `gofmt` has one style, it is not
  configurable, and every Go codebase looks the same. Never argue about braces
  again.

---

## 2. Declarations, and reading types backwards

Go puts the type *after* the name, and reads left to right:

```go
var count int           // C++: int count;
var names []string      // C++: std::vector<std::string> names;
var byId map[int]User   // C++: std::map<int, User> byId;
var p *User             // C++: User* p;
```

Inside a function you almost always use `:=` instead, which declares and infers:

```go
user := brukaren(r)              // handsamarar/dashboard_components.go
lang := GetLanguageFromRequest(r)
```

`var` at package level, `:=` inside functions. That's the whole rule.

**Functions** put the return type last, and can return several values:

```go
func skannTime(rad radskannar, e *models.Event) error
func sessionUserID(r *http.Request) (int64, bool)
func (db *Database) GetEventByID(eventID int64) (*models.Event, error)
```

Named parameters of the same type collapse: `func f(a, b int)`.

---

## 3. No classes — structs and receivers

A struct is a plain aggregate, like a C++ `struct` with no access control:

```go
// handsamarar/app.go
type App struct {
    DB *database.Database
    Nå func() time.Time
}
```

A "method" is a free function with an extra parameter written *before* the name.
That parameter is called the receiver:

```go
// handsamarar/dashboard.go
func (a *App) ElevDashboard(w http.ResponseWriter, r *http.Request) {
    // `a` is just a parameter. There is no implicit `this`.
}
```

`a *App` is a **pointer receiver** — all 80 methods share one `App`. `a App`
would be a value receiver and copy the struct on every call. Rule of thumb: use
a pointer receiver unless the type is tiny and immutable. This codebase uses
pointer receivers for `*App` and `*Database`, and value receivers for the small
model helpers:

```go
// models/event.go
func (e Event) Full() bool { return e.Capacity > 0 && e.CurrentEnrolment >= e.Capacity }
```

**There is no inheritance.** No virtual, no base classes, no `override`. If you
want polymorphism you use an interface (§4). If you want code reuse you use
composition — embed a struct and its methods are promoted.

**There are no constructors.** A struct is always creatable as `App{}`, fully
zeroed (§7). The convention is a package-level function returning one:

```go
// handsamarar/app.go
func NyApp(db *database.Database) *App {
    return &App{
        DB: db,
        Nå: func() time.Time { return config.GetInstance().GetCurrentTime() },
    }
}
```

`&App{...}` is "allocate one and give me a pointer". There is no `new`/`delete`
pairing to think about — it's garbage collected, and returning a pointer to a
local is *safe and normal* in Go, unlike C++.

---

## 4. Interfaces are implicit — this is the big idea

In C++ you declare intent: `class Foo : public Drawable`. In Go you never do.
An interface is a list of method signatures, and any type that happens to have
those methods satisfies it **retroactively, without knowing the interface
exists**.

Your own code has the clearest possible example:

```go
// database/timekolonnar.go
type radskannar interface {
    Scan(dest ...any) error
}
```

`*sql.Row` and `*sql.Rows` are unrelated types from the standard library. Both
have a `Scan` method with that signature. Neither has ever heard of
`radskannar`. Both satisfy it, so one function reads either:

```go
func skannTime(rad radskannar, e *models.Event) error {
    return rad.Scan(&e.ID, &e.Title, /* ... */)
}
```

This is closest to C++20 concepts / duck-typed templates, except it's checked at
compile time *and* dispatched at runtime, with no template instantiation.

The convention is **small interfaces, declared by the consumer**. One or two
methods is normal. `io.Reader` — the most-used interface in Go — has exactly
one. You will not find a `IUserRepository` with fourteen methods; that's a Java
habit, not a Go one.

The whole web stack is one such interface:

```go
type Handler interface {
    ServeHTTP(w http.ResponseWriter, r *http.Request)
}
```

See `docs/ARCHITECTURE.md` §2 for how everything hangs off that.

---

## 5. Pointers, minus the sharp edges

```go
p := &user        // address-of, same as C++
p.Name            // NOT p->Name — Go auto-dereferences
*p                // explicit dereference, rarely needed
```

What's gone:

- **No pointer arithmetic.** You cannot do `p + 1`.
- **No `delete`.** Garbage collected.
- **No dangling pointers.** Returning `&localVar` is safe — the compiler moves
  it to the heap.
- **No references (`&`) as a separate concept.** Pointer or value, that's it.
- **No `const`.** This is a genuine loss coming from C++. Any function holding
  a `*models.Event` can modify it, and nothing in the signature says whether it
  will. Value receivers are the closest thing to `const` — `func (e Event)`
  gets a copy, so it *can't* modify the original.

`nil` is the null pointer, and unlike C++ it's also the zero value for slices,
maps, interfaces, channels and function types. A `nil` map is readable (returns
zero values) but panics if you write to it — a common early surprise.

---

## 6. Slices and maps

A **slice** is a view onto an array: pointer, length, capacity. It is the
`std::vector` of Go, but it's a *view*, so slicing shares memory.

```go
// handsamarar/slag.go
func slagUtanKrok() []string {
    ut := make([]string, len(slagKlassane))   // length 4, all ""
    for i, k := range slagKlassane {
        ut[i] = strings.TrimPrefix(k, "slag-")
    }
    return ut
}
```

`make([]T, n)` allocates *n* zeroed elements. `append(s, x)` grows it, returning
a new slice header — **you must use the return value**: `s = append(s, x)`.
Forgetting that is the classic beginner bug.

`range` gives index and value. If you only want the value: `for _, k := range xs`.
`_` is the blank identifier, and it's how you discard anything Go would
otherwise force you to use.

**Careful:** in `for i, e := range events`, `e` is a *copy*. Mutating it does
nothing. That's why `klargjerKlippekort` in `dashboard_components.go` indexes
instead:

```go
for i := range klippekort {
    k := &klippekort[i]      // pointer into the slice, so writes stick
    k.DaysUntilExpiry = int(k.ExpiryDate.Sub(now).Hours() / 24)
}
```

A **map** is `std::unordered_map`. The two-value read is the idiom for "is it
there":

```go
tmpl, ok := tm.GetTemplate(name)
if !ok { /* absent */ }
```

Iteration order is **randomised on purpose**, so you can't accidentally depend
on it. Sort the keys if you need order.

---

## 7. Zero values — the thing to actually understand

Every type has a zero value, and every variable is *always* initialised to it.
There is no uninitialised memory in Go, ever.

| Type | Zero |
|---|---|
| `int`, `float64` | `0` |
| `string` | `""` |
| `bool` | `false` |
| pointer, slice, map, interface, func | `nil` |
| struct | every field zeroed, recursively |

This removes a whole class of C++ bugs — no reading uninitialised memory. But it
creates a different one, and **this codebase has it badly**, which is why we're
about to do the work in §11.

`models.Event.Capacity` is an `int`. `0` is a legitimate stored value meaning
"this class sets no capacity of its own, use the room's". `0` is *also* what you
get if the query never fetched the column. The type cannot tell you which
happened, and neither can the caller:

```go
// Before we fixed it — same class, two answers:
ein, _  := db.GetEventByID(id)   // Capacity == 0   (never joined rooms)
alle, _ := db.GetAllEvents()     // Capacity == 18  (computed from the room)
```

In C++ you'd reach for `std::optional<int>` and the compiler would force you to
ask. Go's equivalents exist but are deliberately less comfortable:

```go
var n *int              // nil means absent
var n sql.NullInt64     // {Int64: 0, Valid: false}
```

The codebase uses `COALESCE` in SQL **51 times** to flatten absence into a zero
before it ever reaches Go, against only 20 `sql.Null*` and 7 pointer fields.
That ratio *is* the bug class.

`EigenPlassar` is now a `*int` — `nil` means "no own capacity", and the query
uses `NULLIF` rather than `COALESCE` so the absence survives the trip. That is
Go's `std::optional`, and it is the shape to reach for whenever zero is a
legitimate value in its own right.

---

## 8. Errors are values, not exceptions

There is no `throw`. `error` is a built-in interface with one method:

```go
type error interface { Error() string }
```

Functions that can fail return it as the last value, and you check it
immediately:

```go
// handsamarar/dashboard_components.go — a whole real handler
func (a *App) UserSignups(w http.ResponseWriter, r *http.Request) {
    user := brukaren(r)
    lang := GetLanguageFromRequest(r)
    nå := config.GetInstance().GetCurrentTime()

    paamelde, err := a.EnrolledSessions(int64(user.ID), lang, nå)
    if err != nil {
        log.Printf("paameldingar for %d: %v", user.ID, err)
        http.Error(w, "Could not fetch user signups", http.StatusInternalServerError)
        return
    }
    // ... use paamelde
}
```

That `if err != nil` block appears **451 times** in this codebase, roughly once
every 34 lines. It is the single most-criticised thing about Go, and the
criticism is fair: it's noisy, and it carries no stack trace.

The mitigation Go offers is wrapping, and **this codebase does not use it**:

```go
return fmt.Errorf("henta timen %d: %w", id, err)   // %w preserves the cause
errors.Is(err, sql.ErrNoRows)                      // test through the chain
```

Adding that is a separate, worthwhile pass.

**`panic` exists** and is *not* an exception. It's for "this should be
impossible" — it unwinds and crashes the process. `recover()` inside a `defer`
catches it, which is exactly what `berging.go` does so a bug in one handler
returns a 500 instead of killing the server. You will see one deliberate panic:

```go
// handsamarar/session.go
func brukaren(r *http.Request) *models.User {
    user := GetUserFromSession(r)
    if user == nil {
        panic("brukaren() kalla utanfor RequireAuth/RequireAdmin")
    }
    return user
}
```

That's a programmer error — a handler in the wrong route group — so crashing
loudly in development is correct.

---

## 9. `defer` is not RAII

`defer` schedules a call to run when the **function** returns:

```go
rows, err := db.Conn.Query(...)
if err != nil { return nil, err }
defer rows.Close()          // runs however we leave this function
```

Differences from a C++ destructor, all of which matter:

- **Function-scoped, not block-scoped.** A `defer` inside a `for` loop does not
  run at the end of each iteration — it stacks up until the function ends. That
  is a real leak pattern.
- **You write it at the call site**, not once on the type. Forget it and nothing
  reminds you. There is no `~Rows()`.
- Deferred calls run **LIFO**, and they run even during a panic — which is what
  makes `recover()` work.

The standard transaction shape, used at 46 sites in `database/`:

```go
tx, err := db.Conn.Begin()
if err != nil { return err }
defer tx.Rollback()     // no-op after a successful commit
// ... work ...
return tx.Commit()
```

That's Go's answer to RAII: a convention you repeat, not a guarantee the type
provides.

---

## 10. Odds and ends you'll meet in this repo

**Visibility is capitalisation.** `func AdminPage` is exported from the package;
`func sidedata` is not. There is no `private`. The package is the unit of
encapsulation — which is exactly why `handsamarar` being one flat package of 81
files matters (`ARCHITECTURE.md` §4.1).

**Struct tags** are metadata strings read by reflection, mostly for JSON:

```go
// models/event.go
type Event struct {
    ID    int    `json:"id"`
    Title string `json:"title"`
}
```

**`iota`** is an auto-incrementing constant generator:

```go
// handsamarar/session.go
type ctxKey int
const (
    userCtxKey ctxKey = iota   // 0
    csrfCtxKey                 // 1
)
```

**Type assertions** pull a concrete type out of an `any`:

```go
u, _ := r.Context().Value(userCtxKey).(*models.User)   // u is nil if wrong type
```

The two-value form never panics; the one-value form `x.(T)` does.

**Functions are values.** `App.Nå` is a *field* of type `func() time.Time`, which
is why a test can swap the clock:

```go
a := &App{DB: testDB, Nå: func() time.Time { return fixedMoment }}
```

**Concurrency you do not write.** There are **zero** `go` statements, channels,
or `select` blocks in this codebase. But `net/http` runs every request in its
own goroutine, so your handlers *are* concurrent — which is why `sync.Once` and
`sync.RWMutex` guard the template set, the translations and the stylesheet.
Shared mutable state needs a lock even though you never spawned a thread.

**The toolchain is three commands.** `go build ./...`, `go test ./...`,
`gofmt -w .`. No CMake, no vcpkg, no header files, no forward declarations, no
include guards. Compilation is fast enough that the compiler is a usable
search tool — which is how we converted 80 functions to methods safely.

---

## 11. Reading one real function, line by line

`database/timekolonnar.go`, which we wrote today. Every concept above appears in
it.

```go
package database                          // one package per directory

import (
    "database/sql"
    "kjernekraft/models"                  // module-rooted, not relative
)

// A raw string literal — backticks, no escaping needed. Perfect for SQL.
const eventKolonnar = `e.id, e.title, COALESCE(e.description, ''), ...`

// A consumer-declared interface (§4). *sql.Row and *sql.Rows both fit.
type radskannar interface {
    Scan(dest ...any) error               // ...any is variadic
}

// Takes anything scannable; writes through a pointer so the caller sees it.
func skannTime(rad radskannar, e *models.Event) error {
    return rad.Scan(&e.ID, &e.Title, /* 16 more */)
}

func timane(rows *sql.Rows) ([]models.Event, error) {   // two return values
    defer rows.Close()                    // §9 — runs on every exit path
    var ut []models.Event                 // nil slice; append works on nil
    for rows.Next() {
        var e models.Event                // fresh, fully zeroed (§7)
        if err := skannTime(rows, &e); err != nil {   // scoped if-statement
            return nil, err               // §8 — error is a value
        }
        ut = append(ut, e)                // must reassign (§6)
    }
    return ut, rows.Err()
}
```

Two syntax details worth naming:

- `if err := f(); err != nil { }` declares `err` **scoped to the if**. Very
  common, and it keeps error variables from leaking into the function body.
- `var ut []models.Event` is a `nil` slice, and `append` on `nil` works fine.
  You do not need to initialise it.

---

## 12. C++ → Go cheat sheet

| C++ | Go |
|---|---|
| `std::vector<T>` | `[]T` |
| `std::unordered_map<K,V>` | `map[K]V` |
| `std::string` | `string` (immutable, UTF-8) |
| `std::optional<T>` | `*T` or `sql.NullT` — no first-class option |
| `std::unique_ptr` | just a pointer; GC handles it |
| `nullptr` | `nil` |
| `throw` / `catch` | `return err` / `if err != nil` |
| `assert` / invariant violation | `panic` |
| RAII destructor | `defer` at the call site |
| `class : public Base` | interfaces + embedding; no inheritance |
| templates | generics (`[T any]`), much weaker; or interfaces |
| `const T&` parameter | value receiver / value parameter (copies) |
| `namespace` | package (one per directory) |
| header + impl split | none; one file, exported by capitalisation |
| `#include` | `import` |
| CMake | `go build ./...` |

---

## Where to read next

1. `handsamarar/app.go` — 37 lines, and it's the type everything hangs off.
2. `database/timekolonnar.go` — §11 above, in full.
3. `handsamarar/sidedata.go` — a typed constant, a map, and a merge. Short.
4. `handsamarar/middleware.go` — every middleware in one file; read it after
   `ARCHITECTURE.md` §2.3.
5. `models/event.go` — structs, tags, and value receivers.

Then `docs/ARCHITECTURE.md` §2 for how Go's `http.Handler` and chi fit together,
and §5.3 for the capacity problem that §7 above is the fix for.
