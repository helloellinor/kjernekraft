# Arket — rules for agents working in kjernekraft

Read this before touching any template, stylesheet or Go handler that renders UI.
These are rules, not preferences. Each one has a reason; if the reason does not
hold in your case, say so and propose rewriting the rule — do not quietly break it.

Canonical sources, in this order:

1. `static/css/deler/*.css` — the tokens and every component. What ships.
   Many files on disk, **one** stylesheet to the browser: Go concatenates them
   in filename order and serves the result at `/static/css/kjernekraft.css`.
   One `<link>`, no `@import` chain, no build step. Order matters and lives in
   the number prefix — `00-token` before anything that reads a token,
   `45-smale-glas` after what it overrides. Adding a section means picking the
   number that puts it in the right place, not the next free one.
   (Split 2026-08-28: it was one 4,000-line file, and one bad edit across the
   token block took the whole house with it, dark theme included.)
2. `docs/DESIGN_GUIDELINES.md` — the reasoning in long form. Also `FASETTEN.md`,
   `SKRAAKANTEN.md`, `KORREKTUREN.md`.
3. `/arket` — the workshop: components drawn by the same code the pages use.

Where 1 and 2 disagree, 1 is what users see and 2 is what we meant. Flag the
drift; do not pick silently. (One known case: `DESIGN_GUIDELINES.md` §18,
the button's radius.)

---

## 1. Colour lives in `:root`, nowhere else

A hex value in a component works in one theme and lies in the other.

```sh
grep -rn 'style="[^"]*#' handlers/templates/    # must print nothing
```

Three theme states, and every token must be written in all three:

| State | Selector |
|---|---|
| light | `:root, [data-theme="light"]` |
| dark by system | `@media (prefers-color-scheme: dark) { :root }` |
| light by choice | `[data-theme="light"]` |
| dark by choice | `[data-theme="dark"]` |

Four blocks, not three, and **the order is the logic** (changed 2026-08-28).
The system-dark block is plain `:root` now rather than
`:root:not([data-theme="light"])`: the two `[data-theme]` blocks come *after*
the media query and carry the same weight, so an explicit choice wins by source
order. Same effect, and the tokens live in a selector you can look up instead of
inside an expression nobody can grep for.

Every flat colour is written **once**, as `--l-*` for light and `--m-*` for dark
at the top of `00-token.css`; the four blocks only assign meaning to them. The
*derived* tokens — `--merke`, `--merke-svak`, `--togu-svak`, `--tuneup-svak`,
`--glodkjerne` — still appear in each block and must: the percentage differs in
the dark (18 % against 12 %), and `--glodkjerne` inverts, because light on paper
is saturation and light in the dark is blown out.

A colour written only inside a media or `[data-theme]` block does not apply in the
unmarked state — which is the state most people see.

A derived token is written once **only if its recipe is the same in both
themes** — `--faneflate` is, and follows the theme by itself. `--merke-svak`,
`--togu-svak`, `--tuneup-svak` and `--glodkjerne` are not, per the paragraph
above, and belong in all four blocks. (This paragraph used to name
`--merke-svak` and `--glodkjerne` as write-once, flatly contradicting the
paragraph above it. The code has always followed the paragraph above:
`00-token.css` writes `--merke-svak` four times, 12/18/12/18.)

Selectors take `[data-theme]` on **any** element, not just the root, because
`/arket` puts light and dark side by side on one page.

### The browser paints some of it, and it must be told which theme

Every one of the tokens above is ours. Some of what lands on the page is not:
scrollbars, a `<select>`'s drop-down list, the date and time pickers, the clear
`×` in a search field, and `buttontext` — the ink a `<button>` gets when nobody
says otherwise. None of those read a token, and **without `color-scheme` they
stay in the light rendering no matter what the palette does.**

So each of the four blocks declares it beside its tokens — `color-scheme: light`
in the two light blocks, `dark` in the two dark ones. It inherits, so it follows
`[data-theme]` on any element exactly as the tokens do, and `/arket` still works.

It was missing until 2026-08-28, and the way it showed was the clock in the class
list (`.serienaar`) standing in **black on the dark sheet** — that column had no
colour of its own, so it inherited `buttontext` from the row's `<button>`, and
`buttontext` is black in a browser that has not been told the page is dark. See
§9 on classed buttons for the other half of that bug.

### Meaning

| Token | Means |
|---|---|
| `--klas` turquoise | membership — the thing that keeps running. Also `--merke`. |
| `--togu` pink | klippekort — the thing that is counted |
| `--tuneup` purple | courses, PT, guidance — the thing that happens once |
| `--student` yellow | the student/senior rate — not a product, a *rate*. Only the membership card wears it. |
| `--svartkort` | Black — the membership that follows a role. Not a product, so it wears no product colour. Black ink on light paper, foil on dark; it must differ per theme, because true black on a dark sheet has no edge at all. |
| `--aatvaring` | went wrong, or cannot be undone. **Never** a product colour. |

**Colour says *what*; form says *how*.** Filled mark = active, open ring =
waiting, rule = frozen, dotted = expired. Two channels, so they never collide,
and the row still reads in grey. A component takes the product colour through
`--produkt`, never directly:

```css
.membership-card { --produkt: var(--klas); }
.klippekort-card { --produkt: var(--togu); }
```

`--student` is the one entry that is not a product. A student membership is
the same product at another rate, and the rate is what the card mostly *is*
to whoever carries it — so it takes the wing and the mark, and nothing else
in the house may use the colour. Measured against the card surface: 4.67:1
light, 9.40:1 dark, the same band as the product colours.

The wing (`border-left: 3px solid var(--produkt)`) says which product a card is.
Never recolour a wing with `--aatvaring` — an expiring klippekort must still look
like a klippekort. Put the warning on the number and the expiry line.

### The kind of training — and a collision that is still open

A *class* is not a product. You do not buy a yoga class; you spend a klipp on it
or walk in on a membership. So a class carries a second, independent thing:
which kind of training it is. It travels on `--slagfarge`, set by a `.slag-*`
class, and the map lives in `00-token.css`:

```css
.slag-fascia   { --slagfarge: var(--fjellraeven); }
.slag-yoga     { --slagfarge: var(--klas); }
.slag-pilates  { --slagfarge: var(--tuneup); }
.slag-reformer { --slagfarge: var(--togu); }
```

`slagklasse` in the Go func map washes the free-text `events.class_type` down to
that class name. An unknown type gets no `.slag-*` at all, and the wing falls
back to `--haarlina` — grey says "no colour", which is true. It must never fall
back to `--merke`, or an unknown class impersonates yoga.

**Three of the four are product colours, and they now appear on the same screen
as products.** This used to be safe: the class colour lived in a hole on the
activity board, the product colour on a card's wing, and a hole is never a card,
so they never met. Class cards broke that. The home page carries the class list
in the left column and the membership card in the right, so a turquoise wing
means "yoga" on one and "membership" on the other, four hundred pixels apart.
That is one channel carrying two meanings — exactly what this section forbids.

The fix is four values in one block: give the kinds their own hues, clear of the
three product colours, and the board and the wing stay in step by themselves
because both read `.slag-*`. It has not been done. It is a colour decision, not
a bug fix — the whole house changes hue with it, the wheel is already crowded
(`--fjellraeven` at 44° is one hue away from `--student` at 46°), and any new
hue must clear 4.5:1 on `--flate` in both themes, as `--student` does.

---

## 2. Paper takes no light of its own. Things on the paper do. Light comes from above.

Cards, sections, tables, tabs — anything that *is* paper — are separated by a
1px hairline (`--haarlina` draws a card, `--kant` separates two sections). No
gradients, no radius beyond `--rund: 2px`.

Paper used to take no shadow at all. Since 28.8.2026 a card takes a shallow one
(`--skugge-kort`) — it is still paper, it just lies one leaf up off the sheet
rather than printed onto it. Two layers do that: a tight one directly under the
bottom edge that draws the edge, and a soft one that carries the rest. Drop the
tight layer and the card floats; drop the soft one and it is glued down.

Keep the two shadows apart. `--skugge-kort` is for things that lie *on* the
page; `--skugge-flytande` is for things that lie *over* it — dropdowns and
dialogs — and is much deeper. A card must never take the floating one.

Every card in the house reads the token and nothing else, so
`--skugge-kort: none` in the token block turns the whole idea off again without
touching another file. It was added on the understanding that it might come
back out.

Anything that is an *object* catches light, always from above:

| Thing | What light does | Token |
|---|---|---|
| field | falls into a groove | `--skraakant` |
| button, peg, bar | falls on a dome — glim on top, darker to the foot | `--knappedjup` |
| empty day, punched clip, pegboard hole | falls into a hole | `--skraakant-djup` |
| the light band | is the source | `--glodkjerne` |
| card | flat paper, lying one leaf up | `--skugge-kort` |
| dropdown, dialog | floats above the page | `--skugge-flytande` |

### The wet look is the rim, not the sheen

Every raised object wears the **same two layers**, and both are needed. Miss the
second and the thing looks shaded rather than solid — this is the difference
people actually see:

1. **The sheen.** `--glans` at the top falling to transparent by 72 %. Light on
   a curved surface.
2. **The rim.** A 1px bright edge along the top, nothing down the sides, a
   fainter one along the foot — `inset 0 1px 0 var(--glans)` and
   `inset 0 -1px 1px` inside `--knappedjup`. *This is the one that makes it look
   wet.* A gradient alone is a shaded rectangle; the rim is what makes it an
   object with an edge that catches light.

The rim is **not** a uniform outline, and this is the part that keeps getting
misread. On a capsule `inset 0 1px 0` lands along the flat top and fades out on
the curved ends by itself, so it *looks* like it traces the shape — but a rim of
even weight all the way round is a frame, and a frame is a drawn border, not
light. Both the pegs and the bars were rebuilt in Aug 2026 because each had
copied the rim as an outline.

**`--knappedjup` is shaped for a capsule. Do not put it on other shapes.** On a
circle its two offset 1px lines touch only the top arc and only the bottom arc,
leaving two detached white crescents that miss on the sides — a halo sitting
beside the shape rather than on it. Round things take their light the way round
things do, and the way you would shade a sphere on paper: a broad soft specular
high and to the left, a smooth ramp through the terminator, and simply darker
and darker out to the rim that faces away.

**No bounce light on the round things.** A bright edge along the bottom was
tried twice — first as a straight band across the foot, then as a crescent along
the far rim — and both were wrong. A straight band does not follow a curve, so
it lights the bottom as a flat chord; and a reflected highlight is what wet
lacquer does over a white surface, which is not what a peg in a board is. It
falls off at the bottom and stops there. The bars follow the same rule: the
button's bright bottom lip is the paper bouncing under a small capsule, and a
bar standing on a baseline has no such underside.

### One sun, and two shapes (added 2026-08-28)

**`--sol: 34% 22%` is the light, and it is the same for everything.** Two things
side by side may not each have their own. The bars set it — their highlight peaks
at 34 % across the width — so every round or raised thing in the same picture
takes its specular from that point.

There are therefore exactly **two** gloss tokens, not seven:

- `--glim-rund` — the sphere. Light caught at a *point*: a broad soft specular,
  a smooth falloff, then simply darker out to the rim facing away. Size the
  gradient explicitly (`ellipse 82% 82%`); left to `farthest-corner` the darkest
  tone sits in the corner of the box, outside the circle, and never appears.
- `--glim-kapsel` — the lying cylinder. Light caught in a *line*, so the
  specular is a wide flat oval, plus two short ends curving away from it. It
  replaced a band straight across the full width, which was flat precisely
  because it knew nothing about the shape beneath it.

Buttons, the selected tab and the seat mark read the capsule; the pegs on the
board read the sphere. **Selects do not** — they took the capsule for two days in
Aug 2026 and it is described in §9, fields. **Change the gloss and you change it everywhere at
once** — which is the point. The rim (the thin bright line along the top) is not
here: it lives in `--knappedjup`, and it is the rim, not the sheen, that makes a
raised thing look wet.

Build it in `background-image`, not `box-shadow`:
gradient stops are in **percent** and so scale with the object, and the same peg
is drawn at 11px in the legend and 20px on the board. Only the shadow the object
casts on the paper stays in px — the paper is the same distance below it at any
size.

### Paper and a coloured body do not take the same amount of white (added 2026-08-28)

`--glans` is measured **against paper** — 55 % white on a light sheet and 20 %
on a near-black one give the same lift. That calibration is right for everything
whose body is paper-coloured: the button, the select, the selected tab, the glass
on the day mark.

It is wrong for anything carrying a product or training colour, and the reason is
headroom. In the light theme a peg is *darker* than the sheet it sits on
(`--klas` at L\* 54 on L\* 93 paper); in the dark theme it is far *brighter*
(L\* 78 on L\* 8), so white has almost nowhere left to go. Measured on a full
peg, the specular drew ΔL **+17** in light and **+2.5** in dark. The shading side
was fine either way (−17 and −23) — so the dark ball was not a dimmer ball, it
was a coloured disc with a shadow on it, and the bars were flat colour with a lit
line on top.

So the gloss on a **coloured body** is its own token, `--glans-farga`: 55 % in
light — the same number as before, so the light theme is untouched — and 40 % in
dark, which brings the full peg to ΔL +5.5 and the bars back to a ramp. Only two
things read it, and both carry `--slagfarge`: `--glim-rund` (the pegs, the taken
seat) and the bar gradients in `90-aktiviteten.css`. Everything else keeps
`--glans`.

A dim peg now takes *more* light in dark than in light (+19 against +7). That is
the same fact seen from the other end — a dark body has more room above it than a
bright one — and it is why one number could not serve both grounds. The `--ned`
side of `--glim-rund` is unchanged: what was missing was the light, not the shade.

This does not add a third *shape*. There are still exactly two, `--glim-rund` and
`--glim-kapsel`; what splits here is how much white the sun is worth on the
surface it lands on. If you add a depth effect on a coloured body, measure it in
ΔL against **that colour**, not against the sheet.

**Tall things need their cross-section too.** The sheen ramp is a fraction of
the object's own height, so on a button it is 30px and crisp and on a 200px bar
it is a slow wash — flat colour bands with a lit top. A standing rod needs two
more things: it darkens toward both its left and right edges (`.bjelkerunding`,
a gradient across the bar, painted *under* the sheen — form first, then light),
and it keeps darkening toward its foot. Neither is a second light source; both
are the one light from above falling on a curved face.

SVG has no `inset`, so there the rim is a `stroke` on the same shape
(`.bjelkeglim`). Two things about that stroke, both of which have bitten:

- `stroke-width: 1` is **one user unit, not one pixel**. The bar row is 80 units
  stretched across the column, so a "1px" rim rendered four or five pixels wide —
  a thick translucent border around every bar. Bind it to the screen with
  `vector-effect: non-scaling-stroke`.
- Paint it with a **gradient, not a colour** (`#bjelkekant`), so it is bright on
  top, transparent down the sides and faint at the foot. A flat `stroke` colour
  is exactly the uniform outline this section warns against.

Put the gradients **inside the figure**, never in a shared `<defs>`: a
`stop-color: var(--glans)` resolves where it is written, so one shared set gives
every figure the root theme's colours (§9, the same trap as the day-mark).

Write the light with `--glans`, `--lippe-ljos` and `--lippe-djup` and never with
a literal `#fff`. Those tokens already differ per theme (55 % white in light,
26 % in dark, measured in ΔL below); a hardcoded white is three times too loud
in light mode, which is the tell that a raised object was built in dark mode and
never checked in the other.

`--ned: #000` / `--upp: #fff` are a *direction*, not colours: gravity does not
flip when the palette does. The lip and glans percentages differ per theme
because they were **measured in ΔL**, not guessed (55 % white on near-black gives
+0.318 vs +0.073 on light; 26 % gives +0.093 — the same lip, not a stronger one).
If you add a depth effect, measure it the same way.

**Light always diffuses fully outside its div.** No `overflow: hidden`, no
isolation layer, no `z-index: -1` under the card, and every glow layer must fall
to transparent *inside* its own box — a gradient at 150 % of its box reaches the
edge with colour left and stands there as a straight line. The glow box is
deliberately larger than the band: 13px above and below a 2px line.

---

## 3. Nothing may shift

Click something and you must still be looking at the place you were looking at.

- Borders are always present and only change colour. A border that appears moves
  everything around it by a pixel.
- Sections are never swapped out. All tab panels stay in the document;
  `hidden` decides what is visible.
- If a thing changes size, change the padding, not the border. The open tab rises
  by adding `padding-top: var(--loft)`; the text does not move.
- A cell that empties is still a cell.
- A field that switches to editing keeps its width. The number lives in
  `data-oere`; the formatted text is never read back.

The only thing in the house that moves is a pressed button: glim out, shadow
inward, one pixel down, `transition: … .08s ease-out` inside
`@media (prefers-reduced-motion: no-preference)`. All glow is off under `reduce`.

---

## 4. No horizontal scrolling. Ever.

```sh
grep -rn 'overflow-x' static/css/                          # nothing
document.documentElement.scrollWidth > window.innerWidth   # false on every page
```

Ask before adding a scrollbar anywhere and assume the answer is no.

| Instead of | Do |
|---|---|
| grid that overflows | columns that *divide* the width, `minmax(0, 1fr)`, stacking below a breakpoint |
| table that overflows | `table-layout: fixed` and text that wraps |
| tabs that overflow | they wrap to a second row |
| a long lead | it wraps |

Two rules make text stay inside, and you need both:

```css
:where(body *) { min-width: 0; }   /* the forgotten one: grid/flex children default to min-width:auto */
:where(p, li, td, th, label, h1, h2, h3, h4, h5, h6) { overflow-wrap: break-word; }
```

Norwegian compounds are long («medlemskapsadministrasjon»); they must break, not
push the column out of the window.

---

## 5. Type: width carries rank

Union Gothic for headings, Ubuntu for reading text, Ubuntu Mono for numbers that
stand in a column. All self-hosted in `static/fonts/`. **No network calls.**
Union Gothic is temporary (CC BY-NC-ND, gitignored); if missing everything falls
back to Ubuntu Condensed by itself, and swapping in a licensed face is one
`@font-face` block.

| Token | Width | Where |
|---|---|---|
| `--vidd-tittel` | 125 % | page title — where you are |
| `--vidd-yver` | 105 % | section headings |
| `--vidd-leiding` | 90 % | the lead |
| `--vidd-merkelapp` | 68 % | labels, tabs, table heads, `dt` |

Widest on top, narrowest at the bottom, so you can see what is what before you
have read a word. A label must not take more room than it is worth. Uppercase +
tracking + 68 % is the language of things that **name a category** — never a
button, never a table of contents, never a nav group heading.

Numbers use `--skrift-tal` and `font-variant-numeric: tabular-nums` wherever they
stand as figures or columns, or the column dances.

### Document rank

The step follows depth in the document; the appearance comes from the role class.
Never choose the two separately.

| Step | Role | Class |
|---|---|---|
| `h1` | page title. One per page, **always one**. | `.page-title` |
| `h2` | a section, with the rule under it | `.section-title` |
| `h3` | a card, or a group without a rule | card titles, `.undertittel` |
| `h4` | a row or field inside a card — rare; `dt` usually does the work | — |

Need `h5`? The page carries too much and should be split.

```sh
grep -rnE '<h[1-6]( [^>]*)?>' handlers/templates/ | grep -v 'class='   # nothing
for f in handlers/templates/pages/*.html; do echo "$(grep -c '<h1' $f) $f"; done  # max 1
```

---

## 6. Space: one ladder, and each step means something

| Step | | Means |
|---|---|---|
| `--rom-7` | 2rem | between sections on a page |
| `--rom-6` | 1.5rem | title/lead down to the first section; above an `.undertittel` |
| `--rom-5` | 1.25rem | padding inside a box; gap between columns |
| `--rom-4` | 1rem | between cards in a grid; heading to its content; form rows |
| `--rom-3` | 0.75rem | inside a group: label to field, buttons that belong together |
| `--rom-2` | 0.5rem | things that read as one: icon and word, number and unit |
| `--rom-1` | 0.25rem | the smallest air, where zero would stick |

The law behind the table: **the space outside a thing is always larger than the
space inside it.** That is how the eye sees grouping without walls.

- **At most one painted box between the sheet and the content.** A box inside a
  box means neither is a thing; they are packaging. What separates two sections
  is space, not walls.
- **A div either paints, lays out, or is a handle.** No background, border,
  padding, flex, grid or width, and no class the stylesheet knows? It is left
  over — delete it. **The exception is a div that exists for a script**: a swap
  target a fetch writes into, or a box that establishes a container-query
  context. Both are invisible and carry no CSS of their own, and both would read
  as leftover by the sentence above. Name them so the job is visible, or the next
  cleanup deletes them:

  | Job | Kind | Name |
  |---|---|---|
  | a script replaces what is inside it | `id` | `<what>-plass` |
  | a card is measured against it (`cqi`) | class | `.kortplass` |

  This got away from us once (fixed 2026-08-30). The swap targets were
  `#user-klippekort-container`, `#klippekort-charges-container`,
  `#payment-methods-container` and `#klippekort-container` — three languages, id
  and class mixed, and `-container` never used for the *other* kind of container
  at all. The measuring boxes were five classes listed in one `container-type`
  rule, so every new stack of cards had to be added to that list by hand, and a
  stack that was forgotten silently fell back to measuring the viewport. One word
  per job fixes both: `.kortplass` is now a class of its own, written beside the
  class that does the layout (`class="klippekort-cards kortplass"`), so a new
  stack only has to say that it is one.
- **A component carries no space of its own; the parent sets the gap.** A
  component with its own margin knows where it stands, and then it cannot move.

Layout uses flex/grid with `gap`. Never space siblings with source whitespace or
per-element margins.

### A box cannot query its own width (added 2026-08-28)

`container-type: inline-size` on a component looks like it makes that component
the yardstick for itself. It does not: `cqi` resolves against the nearest
container **above** the element using it. So the declaration splits in two, and
the split is silent.

- On *descendants* of the component, `cqi` is the component. Correct.
- On the component's **own** properties — its `padding`, `gap`, `min-height` —
  there is no container above, and `cqi` falls back to the **small viewport**.

The card was its own container for exactly this reason and paid for it: the type
scaled with the column while `padding` sat pinned at its clamp bound —
19.2 / 28.8 / 28.8 / 38.4px at *every* width from 18rem to 42rem, 38px of left
padding on an 18rem card. The comment beside it said the padding scaled with the
card. It never had. A component that scales itself must put the container one
step out — `.kortplass` (written by the component itself, so no call site can
forget it), `.klippekort-cards`, `.verkstad-stabel` — and then `cqi` means the
same thing on the box and inside it.

Measure the padding at two widths before believing a scale works. Same discipline
as the ΔL numbers in §2 and the notch offset: the browser knows, the stylesheet
cannot ask.

---

## 7. Words

Every string is a translation key: `{{t .Lang "bolk.nykel"}}`, and the key exists
in all three files under `locales/`. A hardcoded string exists in one language and
nobody notices until someone switches.

- A button is a **verb in sentence case**: «Lag timen», not «LAG TIMEN».
- A label names a category: 68 % width, uppercase, tracked.
- Numbers read as a sentence — «6 medlemer kjem paa 3 timar» — not as three
  stat tiles. The number is wide and mono; the word is narrow beside it.
- A warning says what will happen *before* you press, on the page, not in a
  dialog that appears after. This bans the **surprise**, not the confirmation:
  an action that cannot be undone may still ask for a second press, as long as
  the warning was already standing on the page before the first one. The second
  press happens *in place* — the button swaps its own word («Sei opp» → «Ja,
  sei opp»), and the room for the longer word is taken before anything is
  clicked, so nothing moves (§3). `static/js/stadfest.js` does this for any
  button carrying `data-stadfest`; it never needs a dialog and the flow's own
  script never needs to know it is there.

  Read the rule the strict way once — no confirmation at all — and «Sei opp»
  became a single click on a live membership. The rule is about *when the
  words arrive*, not about how many presses it takes.
- **A number that can go down is a delete button without a warning.**
  «Serien gjeng i [8] veker» reads as a setting, so typing 4 reads as an
  adjustment — but it destroys four classes, signups and all, with no press
  that said so. Split the two directions: the count is a *fact* in text, and
  what you can type is how many weeks to **add**. Removing happens by selecting
  the days and cancelling them, where you see what goes as it goes. Same
  principle as the surprise ban above: the direction that cannot be undone must
  be the one you look at, not the one you fall into. (admin › Timeplan,
  2026-08-28.)
- **When one control edits many rows, the rows must change with it.** The rule
  sentence sets teacher, clock, length and weekday on every future class, and
  those classes are listed right beside it. Write only to the server and the
  list beneath still shows the old values until a reload — the page then holds
  two answers to one question and one of them is false. Mirror the change into
  the visible rows in the same success handler that clears the error line.

- **No emoji.** They are drawn by three companies, differ per machine, carry skin
  tone and gender we did not ask for, and cannot take the product colour. A
  drawing is: SVG, `stroke="currentColor"`, `fill="none"`, square
  `viewBox="0 0 64 64"`, one stroke width across the whole set (~2.7 % of width,
  matching the mark), `stroke-linecap="butt"`, no text in the file. The box
  (`.teikning`, `--teikningstorleik`) is already fixed, so nothing jumps when the
  drawing lands. The nine emoji `DESIGN_GUIDELINES.md` §14 lists are gone
  (checked 2026-08-28). What is left is three `✓` in `innsjekk.html`, all
  `aria-hidden`, and one `∞` on the membership card; a check mark and an
  infinity sign are typographic glyphs in the reading face, not pictures drawn
  by a vendor, so the reasons above do not reach them. **But a glyph that
  *carries* the value, rather than decorating one, needs the word beside it**
  (added 2026-08-28). The three checks are ornament and are hidden from the
  reader; the `∞` **is** the answer to "Gjeld til", and what a screen reader
  says about U+221E is not ours to decide — some say "infinity", some say
  nothing. So it ships `aria-hidden` with `.skjermlesar` carrying the word.
  That class is the house's only visually-hidden utility (`07-sida.css`):
  `clip-path: inset(50%)` on a 1px box, never `display: none`, which takes the
  text out of the tree so nobody reads it at all.

---

## 8. The mark

`static/img/kjernekraft.svg`. Eight K's in a ring, 45° apart, drawn as **one**
glyph rotated eight times. `.merke` is a **mask** in `currentColor`, not an
image — one file on disk, follows the theme, can carry the product colour.

1.7rem beside the name in the header, 4rem on the login/registration doors. It
grows nowhere else: a mark standing large on every page is a background.

---

## 9. Components — the contracts

| Class family | Rule |
|---|---|
| `.kort` (and `.charge-item`, `.no-data`) | one definition: `--flate`, 1px `--haarlina`, `--rund`, `--rom-5` padding. **A card in a list is denser**: `.timerad` takes `.kort` and overrides padding only (`--rom-4` across, and vertical padding tuned so the seat badge — which is centred on the case corner and so hangs past two edges — clears the card border). Override the padding, never redraw the box. Selected/highlighted adds a 3px line in `--produkt` down the left edge — one line, not a frame, and the card does not move. **Draw that line as `border-left-color` plus `box-shadow: inset 2px 0 0`, never as a `border-left` shorthand** (fixed 2026-08-28) — the hairline that is already there gets recoloured and the inset adds the other two, so the wing still reads 3px and the edge changes colour without changing width. A border was what shipped, and it sat *outside* the 1px hairline: the left edge went from 1px to 3px, every box in the house is `border-box`, so the content box lost two pixels and the text moved two pixels in. On `klippekort.html` the popular package stood in a row beside its neighbours with its text two pixels deeper than theirs, and any card with `aria-pressed` twitched at the moment you clicked it — §3, in the one place §3 is quoted. An inset shadow costs no layout, so the line can come and go under any padding the variant sets; `.timerad` retunes its own padding, so compensating with `padding-left` here would have broken it. Keep `var(--skugge-kort)` in the same list or the wing switches the card's own shadow off. §2 otherwise keeps shadow for light; this is the one place an inset draws a line, because §3 outranks it. `.medlemskort` keeps a real `border-left`: it is unconditional, so it moves nothing, and it has to be able to go `dashed` when a freeze is pending. **A card's state indicator — lamp, status word, expiry — sits at the right end of the card's top line, always.** The membership card puts it there with `space-between` in `.card-header`; the klippekort card puts the expiry there; the payment card uses `margin-left: auto`. An indicator that follows straight after the text reads as one more word in the row; at the right end it reads as a *state*, and the eye finds it in the same place on every card. |
| `.medlemskort` / `.klippekort-card` | **The two things people paid for hold one shape, and it is `e^½` = 1.649** (added 2026-08-28). They are cards; a card has proportions. Nothing pinned them before, so height was whatever the content happened to need, and measured from an 18rem column to a 42rem one the membership card ran 1.15 → 1.75 and the punch card **1.25 → 3.63** — at the wide end not a card but a strip. The ratio is the second step of the same ladder the type stands on (§10.3: 1 → 1.284 → 1.649 → 2.118 → 2.718), so shape and sizes come out of one system; it also lands within 4 % of a bank card (ISO ID-1, 1.586) and 2 % of the golden section, which is why it reads as something from a wallet. It lives in `--kortform`; `--kortbreidd` (28rem) is the **natural size — the width it stops growing at**. **The card says how long it is good for, and the answer comes out of the membership, not the caller** (added 2026-08-28). «Gjeld til» used to show the binding end and only when the caller passed `Bunden`, so a monthly card had no such line at all — a card that does not say how long it lasts fails to say the main thing about itself. It is the **expiration date**: `∞` when the membership is `Tildelt`, because one that follows a role has no end to state; `renewal_date` when it is `cancelled`; `binding_end` when `commitment_months > 0` (a year card runs a year); otherwise `renewal_date` (a month card runs a month — `renewal_date` is always one month out, since billing is monthly for everything). Four branches, not one, and they live in `MembershipWithDetails.GjeldTil` rather than in the template — they are rules, they carry a clock, and there they can be tested (`models/membership_test.go`, `database/frysing_test.go`). The year/month split is needed because `renewal_date` is right for the month card and wrong for the year card. **The cancelled branch has to come first**, and that one is a correctness fix, not a refinement: cancelling writes only `status` and never touches `end_date` or `binding_end`, so a cancelled year card would have claimed «Gjeld til 2027» while `cancel_note`, in prose a few centimetres below it on the same page, says the membership lasts only as long as you have paid for. Two answers to one question, one of them false — §7 names that failure and it applies to a date as much as to a list. **And a freeze stops the clock.** Approving a freeze stamps `frozen_at`; `UnfreezeMembership` — the only way out of `paused`, deliberately, because `UpdateMembershipStatus` takes any status and knows nothing about the clock — pushes `renewal_date` and `binding_end` forward by however long it stood. `GjeldTil` adds the same offset live while it is still frozen, so the card shows the same date the day before you resume and the day after. Without this a year card frozen for two months bought ten months of access, and twelve months would not mean twelve months. **The three sizes are a whole, a half and a third**, all off one `aspect-ratio: calc(--kortform / --korthogd)`: the membership card is whole, the punch card is `.5` — two of them stack into one membership — and `.charge-item` is `calc(1 / 3)`. That is a set of meanings, not three arbitrary heights, and it is also what the content wants: the punch card carries two lines against the membership card's five, and a full-height one would have stood with an empty middle just to reach the shape. Measured at 28rem: 271.7 / 136.0 / 90.5px — 1 : 1.998 : 3.000. A charge whose description wraps, or a failed one carrying a status and a reason, grows past its third; the ratio is a preferred size, never a clip. A receipt is not worth truncating to keep a rhythm. Below the natural width the card shrinks *whole*: type, padding and notches go down together, because each is a multiple of `--kortgrunn` and `--kortgrunn` is a fraction of the width. **It may get smaller; it may not become another shape.** There is deliberately **no lower clamp** on `--kortgrunn` — a floor is exactly what breaks the shape, since the type would stop shrinking while the box kept going. Two things still hold a floor and are the whole of the residual: the button (a press target under eleven pixels is not a button) and `@container (max-width: 20rem)`, which wraps the header rather than let it run out of the box, because §4 outranks the shape. Measured: exact at 28rem and 34rem (1.649 / 3.295 / 4.948), 1.4 % out at 23rem, 3 % at 20rem — a phone — and only below any column in the house does it visibly give. **The space between two cards is one number, `--kortrom`, and it is `--rom-4`** — the step §6 already gives to "between cards in a grid". It was three numbers doing one job (`--rom-4` for the punch-card stack, `--rom-3` for the charges list, `--rom-4` in the workshop), each inherited rather than chosen. Deriving it from the ratio was tried and rejected: padding × `--kortform` is 1.814 units = 26.7px, and a stack with that much air between the cards stops reading as a stack. **One thing is left open and must not be settled alone:** §6 says the space outside a thing is always larger than the space inside it, and here they are nearly equal — 16px between against 16.2px of padding. Two ways out: `--kortfyll: 1` (14.7px inside), or `--rom-5` (20px between). The first leaves the rhythm alone, the second leaves the card alone. Height being known before the content is, the slack has to be *placed* rather than left to fall at the bottom: `justify-content: space-between` puts what names the card at the top and what you came for at the bottom, the way a number sits on a card in your pocket. |
| `.btn` / `.btn-primary` / `.btn-danger` | a capsule (`--rund-knapp: 999px`), reading font, sentence case, `--knapp-loddrett` × `--rom-5` padding, `--knappedjup`. One primary per screen. Disabled has no depth at all. |
| `button:not([class])` | carries the surrounding font and nothing else. Most buttons in the house are this. Do not put base styling on bare `<button>` — twenty places had to strip it off again. **Mind the gap between this rule and the one above.** The `.btn` family style is written as an explicit selector list (`.btn, .btn-primary, .btn-danger, .plettmerke, a.byteveg`) and this one requires *no* class at all — so a `<button class="my-hook">` that carries only a behaviour or layout class matches neither, falls through every rule in the house, and renders as a raw browser button: system font, grey fill, wrong corner radius. **And, the one that actually shipped: the browser's own ink.** A button nobody gives a `color` inherits `buttontext`, which is black — so any child span without a colour of its own is black text, and in the dark theme that is black on a dark sheet. `.seriehovud` in the class list stripped `background` and `border` but never stated its ink, `.serienaar` inside it had no colour, and the clock was unreadable in dark mode. `.folk-hovud` had the same hole and only escaped because every one of its children happened to carry a colour — luck, not a rule. **A component that strips the browser's look must state its ink in the same breath.** `color-scheme` in the token block (§1) fixes the root, but a component must not depend on a browser default to be legible. Give it a family class beside its hook (`class="btn-danger val-avlys"`), never the hook alone. And a button inside a `.setning` needs `.setning :is(.btn, .btn-primary, .btn-danger)` — the rule was scoped to `.btn-primary` while that was the only button ever put in a sentence, so the first `.btn-danger` in one inherited the 1.9 line-height and stood twice as tall as the words beside it. |
| `.bolkhovud` + `.sikt` | **a compartment that filters puts the controls on the heading line, right-justified against its `<h2>`** (added 2026-08-28). Not a row of their own underneath: a filter is a setting on something you are already looking at, not a question to answer before you get to it — a row underneath makes three lines to read (title, question, answer) before the content. `.bolkhovud` is the line; it takes the hairline off `.section-title` and carries it itself, because a rule that stops at the width of the word is not a section rule. Right-justify with `justify-content: space-between`, never `margin-left: auto` — the difference only shows when the line wraps on a narrow screen, where an auto margin keeps the control pinned right while it sits alone on its own line, floating in space. **Two things will silently push the heading off the baseline of the section beside it, and both must be handled:** (1) `:where(h2,h3,h4) + *` in the token block gives 12px of top margin to whatever follows a heading — correct in normal flow, wrong here, where the control sits *beside* the heading and `gap` already owns the spacing; `.bolkhovud > * { margin: 0 }` clears it. (2) Align the row `center`, not `baseline` — Chrome takes a `<select>`'s baseline from the **bottom edge of its box**, not its text, so a baseline row drags the heading down to meet a baseline that is not where the type is. Either one alone leaves two equally-sized `<h2>`s in adjacent columns sitting at different heights, reading as two different ranks. Measure it: both headings' `getBoundingClientRect().top` must match, and so must the bottoms of their two hairlines. **A third thing breaks it, and it is the one that actually shipped:** two `.bolkhovud` side by side whose controls are *different kinds*. One column's heading carried naked `<select>`s and the other only a `<span>`, and a select's box is taller than a span's — so the two hairlines landed at different heights and every row beneath them in the two columns was out of register, which reads to the eye as one column having been nudged down. Give whatever sits opposite the same box metrics (font-size and vertical padding) as `.sikt select`, so the two heads are the same height by construction instead of by a measured constant someone has to maintain. The selects are naked (no border, no fill) and carry **no visible label**: the first option is the label — «Alle typar», «Alle klassar» — so a label beside it says the same thing twice; the name lives in `aria-label`. An active filter shows as *light* (`--merke-svak` fill, ink up from `--blekk-lys` to `--blekk`), never as a frame. The schedule and the charges list are the same component — `.timesikt` was a second copy of it under another name and was folded in. **A search field may stand in the row, and it keeps its groove** (added 2026-08-28, admin › Timeplan). The selects are naked because a select always shows what it stands on — it needs no frame to say it is there. An empty text field says nothing at all, and §9 gives depth exactly one meaning, on every field in the house, at rest and not only on focus: *you can write here*. Stripping it to match the selects would be matching the wrong thing. Give it a fixed width, though — a field that grows with its content pushes the selects beside it while you type (§3). |
| a list you act on | **One mechanism for one row and for many.** Do not build a per-row editor *and* a bulk bar: that is two ways to do one thing, stacked on the same screen, and it forces a third element — a hint line explaining which one you are standing in. A flat surface that has to explain itself is not finished. Put a checkbox on each row, keep the row itself as text, and put one action sentence below the list that acts on whatever is selected: «[3] valde: … [Avlys]». One selected or five, same controls. Fields that only make sense for a single row (a date — five classes moved to one date is a collision, not a move) stay in the sentence and go inactive unless exactly one is selected; the *words* never change, only the depth, so nothing shifts (§3). The sentence stands at rest, disabled, rather than appearing on first selection — a row that appears when you act is a row nobody knows exists, and it shoves everything below it when it arrives. (admin › Timeplan, 2026-08-28: the day list carried 7 controls per row plus a 5-control bulk bar — 60 controls for an eight-week class, and two hint lines whose only job was to say which of the three editors you were in. It is 22 now, and nothing needs explaining.) |
| a selected row in a list | **Light, never a wing.** A list you pick from — the rule list in admin › Timeplan — marks the selected row with `--merke-svak` in the surface and the ink up to `--blekk`, which is what an active select in `.sikt` already does. It must not take the `border-left: 3px` wing: that line is the channel that says *what kind* a thing is — the product on a card, the training on a class (§1) — and a page where the same 3px line means "membership" in one column and "this is the one you clicked" in the other is one channel carrying two meanings, which §1 forbids. Light is unclaimed, and light is already what the house uses for "this one is on". |
| a count beside a name | **Say the word, never `×`.** «8 ×» was meant as "eight times" and read as a close button — a × in the corner of a row is what the whole web uses for *remove this*. It is also, three files away in the same admin page, literally the delete control (`.slettmerke`). One glyph, two meanings, one screen: §1 forbids that for colour and the reason does not change for a glyph. It was not even inert — it sat inside the button that selects the row, so it promised "delete" and performed "select". Write «8 gonger», and pick singular or plural with the `gonger` func map helper: the number and the word were written separately, so the word never saw the number and «1 gonger» shipped. Anything that rewrites the count after load must make the same choice — the JS helper beside it exists for that. |
| `.status-*` | class names are the values the database uses (`active`, `paused`, `cancelled`, `freeze_requested`). No translation table. |
| `.faneark` / `.faner` / `.fane` / `.fanerom` | **a slider with positions, not a folder** (changed 2026-08-27). The track is a groove (`--skraakant`, the light of a field); the selected position is a dome sitting in it (`--knappedjup`, the light of a button). Same two-layer form as `.btn`, because it is the same thing — something raised out of the surface. The label is the reading font in sentence case: §5 reserves uppercase + tracking + 68 % for what *names a category* — «never a button» — and a tab is a `<button role="tab">`. `.fanerom` paints nothing: §6 allows one box between sheet and content, and the folder's room made it two. `.faneark` carries no rule at all and must stay — `faner.js` uses `closest('.faneark')` so two tab rows don't steal each other's tabs (admin has a row inside *Prising*, and one inside every class-series form). The choice lives in the URL, and only the **outermost** sheet owns it; an inner row opens on its first tab. **The `faner` partial's contract is `.Tabs`, each with `Key` and `Name`** — those names come from `handlers.Tab`, where a wrong field is a compile error you see. A template that builds the list with `dict` gets no such error: a missed lookup is the empty string and nothing else, so the row renders buttons with no label and `data-bolk=""`, no tab matches any panel, and clicking does nothing. That shipped in *Prising* (fixed 2026-08-28). Build the list with `Key`/`Name`, and if a tab row ever does nothing, look there first. |
| `.setning` (a sentence of controls) | **The line must make room for the tallest thing in it, and the sentence carries no reading measure.** Two failures, both recurring. (1) A sentence sets `line-height` for *text*, but a button is taller than text — it carries its own vertical padding and a border — so when the sentence wrapped, the button scraped the underside of the field on the line above. That is neither a button bug nor a field bug: the line was never asked to clear the tallest control it holds. Give controls in a sentence `margin-block: var(--rom-1)` off the space ladder rather than nudging one offender. (2) `max-width: 54ch` is the measure for *prose* — how far the eye tracks without losing the line — and a row of controls written as a sentence is not prose. The cap made the admin form wrap halfway across a wide screen with the other half standing empty. The column sets the width; the sentence breaks when the column makes it break and not before. |
| fields | `:where(input…, textarea)` with no border, `--skraakant`, and `--rund`. Depth means "you can write here", so it is on every field at rest, not only on focus. A `select` **has the field's form, and the arrow is the difference** (settled 2026-08-30, reversing the change of two days earlier). It briefly wore the button's capsule, dome and gloss, on the reasoning that it *is* a button with a list behind it. That is true of what it does and turned out to be wrong about what it is on a page, for three reasons, all of which were visible before anyone argued about them:

**(1) The capsule never met its dome.** `.setning select.felt` was written earlier and took the flat fill and hairline back — but not the radius — so in every sentence in the house the select rendered as a pill-shaped box with a thin frame and a gloss laid over a surface that was not domed. In the light theme `--glans` is 55 % white, so that is a white smear across an almost-white box. Two rules that each meant to settle the form met in a third nobody drew. **(2) Prising was the one place they looked right, and it broke the rule to get there** — its selects carry `.lappfelt`, the field's form. The house was already doing the opposite of what was written. **(3) A capsule in a row of boxes does not belong to the row.** Everything else on the sheet has a broken edge (§2); a pill among them reads as a foreign object, not as a button.

So: 3 % ink fill, `--rund`, `--skraakant` — the same box as every field — and **the arrow** says it is picked rather than typed. The arrow is always there, so the distinction is always visible. One form for everything you can touch, one mark for the one thing that differs; cheaper than two forms that must be kept in step by every rule that touches a field. **`FASETTEN.md` §3 still says the opposite** — that depth means "you can write here" and a select must not have it — and that is the known drift between source 1 and source 2 for this component. What ships is here. It is as wide as its widest option, `width: auto; max-width: 100%`. **The arrow and the room it needs are one number** (added 2026-08-30). They were written separately in four places — base 1.05/0.80rem against 1.5rem of padding, `.setning` 0.85/0.60 against 1.25, `.sikt` 0.55/0.35, the language picker 0.45/0.25 — four guesses at one thing, and the two tightest left **1.4px and 2.2px** between the last letter and the arrow. On a phone, where a sentence wraps and the select is squeezed, that reads as the arrow sitting on the text, which is what it was reported as. `--pil-inn` is where the arrow stands; `--pilrom` is the padding *computed* from it (`--pil-inn + .25rem + --pil-storleik + --pil-glipe`); no rule writes a position of its own any more, and every select in the house measures the same 29.8px of right padding. Verified by reading `padding-right` off all of them, in a sentence, a rule, a filter and a plain form, at 360/480/760px. **A class that sets padding on all four sides starves the arrow**, and there are exactly two: `.setning .felt` (`.1rem .4rem` with `field-sizing: content`, deliberately tight because a field in a sentence should be as wide as its word) and `.lappfelt`. Both must give the right side back as `var(--pilrom)` — never as a number typed at the call site, which is exactly how the four positions came to exist. That is the one narrow exception to the measure-only rule below: a class may restore what it took, in the token that computed it. **`.sikt select` is the one exception, and it is deliberate:** a filter beside a heading is a setting on what you are already looking at, not a form, so it stays naked. Taking the frame off is not enough — it must also clear `box-shadow` and reset `background-image` to `var(--pil)` alone, or it inherits a dome it never asked for and the arrow lands wrong, because the position list is written for five layers and not two. Base styling must be the easiest thing to override: use `:where()`. **That is only half a rule, and the missing half is what a class may override** (added 2026-08-28). `:where()` makes the house standard the *weakest* declaration in the house — deliberately, and it has earned that twice — but nothing then said what a local class is allowed to change, so every call site quietly redrew the box to match its own neighbours. **A class on a field sets measure and nothing else: `width`, `max-width`, `flex`. `font-size`, `font-weight`, `padding`, `background`, `border-radius` and `box-shadow` belong to the base.** `20-skjema.css` has said this in a comment since the fields were folded into one — "the classes that remain are only *measure*: how wide the field is, not how it looks" — and a comment in one file is not a rule. Three search fields broke it and are the reason it is written down — `.soke`, `.siktsok` and `.dropinnsok`, all `<input type="search">`, all filtering or looking up a list, standing ~50px, ~28px and ~47px tall against a base of ~41px, each sized against whatever happened to stand beside *it*. They are one class now; see `.sok` below. |
| `.sok` (the search field) | **One class, and the form is the base field's** (folded 2026-08-28). `21-soket.css` carries only what belongs to a search field and nothing else: a `max-width` so it cannot run the width of a wide column, and a `--blekk-svakt` placeholder. Everything visible — the groove, the fill, the font, the padding — comes from `:where(input[type="search"])` in `20-skjema.css`, which is the point. **Measure lives with the layout that owns it, not on the field** (§6): `.folk-sok .sok` takes `flex: 1 1 22rem` because it shares a line with the count, `.timestyring-seriar > .sok` takes the full column and the offset that lines its top up with the bench in the next column, `.dropinnfelt .sok` takes a fixed 16rem so it does not shove the sentence beside it while you type (§3). A field does not know what stands next to it; the column does. The one real difference between the three is size, and it is set **once, on the container**: `.dropinnfelt` is `1.25rem` because it is read across a counter, and the field and its result list both inherit it — `font: inherit` on the base does the work, so nobody writes a size on a search field. Draw it in `/arket` beside the text field (`prov_skjema`): it was absent from the workshop the whole time it was drifting, and side by side in both themes is the only way anyone was going to notice. |
| tables | no grid lines. Hairline under each row, `--kant` under the head, sticky head at `var(--sidehovud-hogd)`, `table-layout: fixed`. |
| sticky layers | The site header is sticky at `top: 0`, `z-index: 20` — a page you have scrolled must never make you scroll back up to navigate away. Anything else that sticks stacks *below* it, and the offsets are **measured, not written**: `klistrelag.js` writes `--sidehovud-hogd` and `--tittellinje-hogd` to the root from a `ResizeObserver`. They cannot be constants — the header is 56px wide and 96px narrow, where the nav wraps under the wordmark. Same move as `ljosband.js`: the browser knows, CSS cannot ask. Order is header (20) → page title (11) → controls (10) → dock (30, over everything). **A sticky bar must be opaque**, or the content passing under shows through it; if a bar should carry no surface of its own, it must carry no padding either — give the air to the opaque bar next to it, since padding on a transparent thing is just pixels a card can scroll into. |
| empty states | three different things: what does not exist, what has not arrived yet, and what is empty because you have not bought anything. Only the last one gets a way out of it. |
| `.klippemaalar` / `.progress-segments` | see §10. |
| `.plassmaalar` | the same light, other meaning: a lit cell is a **taken** seat, so a full class is fully lit. Correct as is. |
| `.dagmerke` | a clock in a tall case. Geometry is computed in `handlers/merkeform.go`; `kroppH` is the one number you tune, everything else derives from it. Size is one scale and nothing else. The gradient and filters live **inside each figure**, never in a shared `<defs>` — a `stop-color: var(--glans)` resolves where the gradient is written, so a shared set gives every mark the root theme's colours. The seat count is a **number**, not a pie: a fraction cannot be told apart (2 of 10 looks like 4 of 20). |
| the week grid | seven columns that divide the width; `--merkestorleik` (5.4rem) computes the ladder for both head and rows. The dead cell keeps the row at seven places, so you can see what does *not* run. Day names are buttons that **dim** the others; they do not filter — filtering removes the one thing a grid can do that a list cannot. Below 62rem each day becomes a section under the previous one, with the day as its heading. |
| `.dokk` | the thing that actually floats: hairline top edge with the light *above* it, `--skugge-flytande`, and the page adds `--dokkhogd` of padding under itself so nothing hides behind it. |

---

## 10. Decisions taken, and what is still open — do not decide these alone

1. ~~**`.btn-primary`: line or fill.**~~ **Decided — the line ships**
   (2026-08-28, from the Arket design-system project). The colour lives in the
   border and the word, never in the surface. The reasoning was already written
   two rules below, about the warning button — «a red lump pulls the eye away
   from everything else» — and it holds for a turquoise lump just as well, in a
   house that builds everything else out of hairlines. The body is now identical
   to `.btn`: same gloss, same dome, same shadow. The only thing that marks the
   primary action is that its line and its word carry `--merke`. It is stronger
   than the others because it is the only one with colour in it, not because it
   is heavier.

2. ~~**The klippekort readout.**~~ **Decided — the punch card ships.** The count
   is the hero (`.klipptal .att`, `--skrift-tal`, 4.482rem, tabular) and the spent
   clips are bites out of the wing (`.klipphakk.klipt`, `--skraakant-djup`, fill
   `--ark`). Reason: a percentage bar says "three quarters left"; you want to know
   "seven". The clips are the receipt, not the readout. When it is nearly out the
   **number** and the expiry line take `--aatvaring`; the notch row does not
   change, because a hole is an event that already happened.

   Two things the reference pages do not settle, decided here:

   - **The notch row is a fixed ten** (`models.HolPerKort`), not one notch per
     clip. It is a *form*, not a meter: it answers "has this card been used", and
     the number answers "how many". Packages run 5/10/20 and the cap is 20, so
     1:1 would need either a card twice as tall or notches too small to be
     notches. A 20-clip card gets one hole per two clips.
   - **The header wraps on a narrow *card*, not a narrow screen.** The reference
     sets `flex-wrap: nowrap` on a full-width card (34rem then, `--kortbreidd`
     = 28rem now); on a phone «Gruppetimar» plus
     «gjeng ut um 54 dagar» is wider than the card and pushed the expiry out of
     the box. §4 outranks fidelity to a width the reference never drew. This was
     a media query first, and that was the wrong question: the screen can be
     1280 while the card is 370. It is a container query now — the card answers
     for itself, so it behaves the same in a wide column, a narrow one, or a
     sidebar. The expiry's lift goes with the wrap, being an optical correction
     for sitting on the title's line, and it is no longer on that line.
   - **The expiry's lift is derived, not measured.** It was `translateY(-.489em)`,
     a number measured against the reading face back when the expiry was set in
     it, lowercase. The expiry now carries the label voice with the status and
     the lamp — the three are one indicator, so they are one voice — and its top
     edge is therefore a cap line rather than an ascender. Because title and
     expiry are now the *same* face, the lift needs no font metrics to state:
     x-height is the same fraction of both, so the title's x-height expressed in
     the expiry's own units is `(tittel ÷ lapp) × 1ex`, and the expiry's top is
     `1cap`. The lift is the difference. Change a step on the ladder and it
     follows by itself. A plain `em` value precedes it for browsers without
     `cap`, so an unsupported unit drops one declaration instead of the rule.

   The light band (`.progress-segments`) stays for meters where a lit cell is
   something that **exists** rather than something spent: `.plassmaalar` in the
   dock, and `.medlemskapsmaalar`. Its unused nynorsk aliases (`.klippemaalar`,
   `.klipp`, `.tend`, `.kursmaalar`) are gone.
3. ~~**The tightened ladder.**~~ **Decided — it ships** (2026-08-28, from the
   Arket design-system project). `--rom-5/6/7` are 1.25/1.5/2rem, and the
   heading grades sit on step e^¼ ≈ 1.284 topping out at e rem:
   1 → 1.284 → 1.649 → 2.118 → 2.718. **Tight means tighter space and
   *larger* type**, because space was what made the pages long. The steps
   below 1rem are untouched: they are distances *inside* a thing rather than
   between things, and that is fine motor control.

   Note the consequence, and it is intended: a bare `h2` is now 2.118rem while
   `.section-title` is 1.649rem. §5 already says the step follows depth in the
   document and the appearance comes from the role class — nearly every `h2` in
   the house carries `.section-title`, so it renders at the class's grade. Width
   carries the rank, not size, which is exactly why two headings can be almost
   the same size and still say plainly which sits under which.

4. **Dead code.** *Classes: done (2026-08-27).* 49 class selectors that no
   template, script or handler ever wrote are gone, `.tabell-rom` among them —
   mostly the leavings of `pages/membership.html` after `medlemskapet.html`
   replaced it, plus the light band once the punch card took over. Verified by
   rendering all eight pages against the old stylesheet in both themes at two
   widths: pixel-identical. `grep`-able check below.

   **Still open:** `--fordjuping` and `--fasett` have been unused since
   `--skraakant` took over. ~~The dark block is written twice, word for word…~~ **Done
   (2026-08-28).** Every flat colour is a `--l-*` or `--m-*` value written once,
   and the four theme blocks only read them. The two dark blocks are now
   word-for-word identical — if they ever differ again it is a bug, not a
   nuance — and `handlers/stilark_test.go` asserts it, plus the new invariant
   that light-by-choice agrees with the light default. And the box is still drawn by hand in
   seven places besides `.kort` (`.timeplan`, `.login-container`, `.bolk`,
   `.folk-liste`, `.regel-liste`, …); folding those into `.kort` is the next
   cleanup, and it is the thing that keeps generating new one-off classes.

   ~~**And the search field is drawn three times**…~~ **Done (2026-08-28).**
   `.soke`, `.siktsok` and `.dropinnsok` are one `.sok` (`21-soket.css`), the
   form is the base field's, the measure sits on the three columns that own it,
   and the one genuine difference — the counter reads a step larger — is set
   once on `.dropinnfelt` and inherited. `.siktsok`'s hand-derived `margin-top`
   moved onto `.timestyring-seriar`, where lining a field up with the bench in
   the next column is the column's business. A search field is in `prov_skjema`
   now; the workshop never drew one, which is why nothing caught the drift.

---

## 11. Before you commit

Every one of these prints nothing (or the stated number) on a clean tree. If
one is noisy, the check is broken and gets fixed — a block where two of six are
always red is a block nobody reads.

```sh
grep -rn 'style="[^"]*#' handlers/templates/         # no colour in attributes
grep -rn 'overflow-x' static/css/                    # no sideways scrolling
grep -rn 'font-stretch:[^;]*%' static/css/ | grep -v 00-token.css   # the ladder, not a literal
./scripts/daude-klassar.sh                           # no class nobody writes

# No naked heading, and exactly one h1 per page. Strip the comments first:
# a Norwegian comment that mentions <h1> is not a heading, and both of these
# used to trip on the same one in min-profil.html.
naken() { perl -0777 -pe 's/\{\{\/\*.*?\*\/\}\}//gs; s/<!--.*?-->//gs' "$1"; }
for f in $(find handlers/templates -name '*.html'); do
  naken $f | grep -nE '<h[1-6]( [^>]*)?>' | grep -v 'class=' | sed "s|^|$f:|"
done
for f in handlers/templates/pages/*.html; do echo "$(naken $f | grep -c '<h1') $f"; done
```

The width check was `grep -c 'font-stretch' static/css/` — no `-r`, so it errored
on a directory every time, and a count with no threshold was never a check
anyway. What it asks now is the actual invariant: a width comes off the ladder in
§5, never as a literal percentage.

`daude-klassar.sh` exits non-zero and lists them. A class nobody writes is not
just dead weight — the next person looking for a component finds it, builds on
it, and only then discovers it was never drawn. That is how 49 of them
accumulated. Delete on sight; git remembers.

In the browser, on every page you touched:

```js
document.documentElement.scrollWidth > window.innerWidth   // false
```

And read the page in both themes. Half the rules above exist because something
was only ever looked at in one of them.
