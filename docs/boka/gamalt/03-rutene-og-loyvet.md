# 3. Rutene, og kvar løyvet bur

*Klossen som slukar ein annan kloss.*

---

## 3.1 Mellomvare er ikkje ein rammeverksfunksjon

I dei fleste språk er «middleware» noko rammeverket gjev deg, med ein
grunnklasse å arva frå eller ein dekoratør å setja yver metoden.

I Go er det ein **namneskikk**. Ein mellomvare er kva som helst funksjon med
denne formi:

```go
func(next http.Handler) http.Handler
```

Tak ein kloss, gjev ein kloss attende. Det er alt. Det finst ingen type som
heiter `Middleware`, ingi lista å melda seg inn i, og ingi klasse å arva.

Den einfelte i huset stend i `handsamarar/middleware.go`:

```go
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if GetUserFromSession(r) == nil {
			denyUnauthenticated(w, r)
			return          // næraste klossen vert aldri kalla
		}
		next.ServeHTTP(w, r)
	})
}
```

Legg merke til korleis han er sett saman: han gjev attende ein
`http.HandlerFunc` — som me lærde i §2.4 er ein funksjon som *er* ein
`Handler` — og inni han ligg ei **lukking** (closure) som hev fanga `next`.
Det er same greia som ei C++-lambda som fangar med `[next]`, berre at Go gjer
det utan at du segjer frå.

## 3.2 Tri ting som fylgjer av formi

### Å *ikkje* kalla `next` er avbrotet

Det finst ingen «abort»-mekanisme. Ein vakt stoggar soknaden med å lata vera å
gjeva han vidare. `return`-en i `RequireAuth` er heile stoppet.

Dette er verdt å halda fast ved: i C++ ville du kasta eit unnatak eller gjeve
attende ein feilkode som nokon lyt sjekka. Her er avbrotet **strukturelt** —
kjeda gjeng ikkje vidare av di ingen bad henne um det.

### Det er ein lauk, ikkje eit røyr

`A(B(C(handsamar)))` tyder at A sitt «fyre»-arbeid gjeng fyrst — og A sitt
«etter»-arbeid gjeng **sist**. Kvar mellomvare ligg heilt kring den neste.

Dette er ikkje pedanteri. Det er grunnen til at `Recoverer` lyt liggja
ytst av våre: `recover()` i ein `defer` kann berre taka panikk som kjem frå
noko som ligg *inni* honom. Ligg han innanfor noko anna, ser han ikkje det som
brest utanfor.

### Mellomvara kann byta ut soknaden

```go
next.ServeHTTP(w, r)
```

Den `r`-en treng ikkje vera den same `r`-en som kom inn. Det er slik `WithUser`
legg brukaren inn i soknaden — meir um det i §3.5.

## 3.3 chi er ei byggjeplate, og lite anna

`github.com/go-chi/chi/v5` er den einaste vevavhengnaden i programmet — kring
tusen liner. Han finst av di rutaren i standardbiblioteket lenge ikkje kunde
skilja på HTTP-metode eller gruppera rutor under sams mellomvare.

Det du nyttar av honom:

| | |
|---|---|
| `chi.NewRouter()` | gjev ein `*chi.Mux` som **sjølv er ein `Handler`** |
| `r.Get`, `r.Post` | same stigen kann ha ulike handsamarar per metode |
| `r.Use(mw)` | heng mellomvare på kvar rute som er meldt inn på denne rutaren |
| `r.Group(fn)` | ein underrutar som **ervar** foreldra si mellomvare og kann leggja til si eigi |
| `r.Route(sti, fn)` | ein underrutar under eit stiprefiks — nytta éin gong, på `/api/admin/settings` |
| `/*` | jokerteikn i stigen |

Sjå `server.go:153-154`: `/innlogging` er meld inn tvo gonger, GET og POST.
GET teiknar skjemaet, POST tek imot det. Same handsamaren, som sjølv ser på
`r.Method`.

**Ein regel du lyt vita:** `r.Use()` *lyt* kallast fyre den fyrste ruta vert
meld inn på den rutaren, elles panikkar chi. Det er ikkje ein stilregel — det
er handheva. Difor les `server.go` ovanfrå og ned som mellomvare-so-rutor, og
difor kann du ikkje flytta fritt på linone der.

## 3.4 Fire liner som er heile løyvemodellen

Dette er den viktigaste strukturelle ideen i programmet:

```go
r.Group(func(r chi.Router) {
	r.Use(RequireAdmin)
	r.Get("/admin", app.AdminPage)
	// ... alle dei hine admin-rutone
})
```

`r`-en inne i lukkingi **skuggar** `r`-en utanfor. Det er ein ny underrutar.
Kvar rute som vert meld inn der inne ervar mellomvara utanfrå — `Recoverer`,
`WithUser`, `CSRF` — og fær `RequireAdmin` på toppen.

Sagt beint ut:

> **Løyvet til ein handsamar er avgjort *berre* av kva `r.Group`-blokk
> innmeldingslina hans stend i.**

Det stend ingen ting på handsamaren. Ingen merkelapp, ingen sjekk inni, ingen
ting i namnet. Flytt `r.Get("/admin", app.AdminPage)` førti liner upp, og
administrasjonen er open for alle — utan ein kompilatorfeil, utan ei prøve som
brest.

Difor ber `AdminPage` denne kommentaren på fyrste lina:

```go
// Tilgangen vert avgjord av RequireAdmin i rutaren, ikkje her.
```

Det er ikkje pynt. Det er den einaste staden den kunnskapen finst, sedd frå
fila du les.

Grupperingi i `server.go` er difor verdt å lesa som ein tabell:

| Gruppa | Mellomvare | Kva ligg der |
|---|---|---|
| open | — | `/`, `/signup`, `/terms`, `/innlogging`, `/glemt-passord`, `/logout` |
| kiosk | eigen lås | `/innsjekk`, `/api/innsjekk/*` |
| medlem | `RequireAuth` | `/elev/*`, `/api/user/*`, `/api/membership/*` |
| administrasjon | `RequireAdmin` | `/admin`, `/api/admin/*` |
| utvikling | `RequireDevelopment` | `/api/shuffle-*`, `/elev/testdata` |

Kiosken er unnataket som stadfester regelen: han hev ingen innlogging, av di
det stend ein skjerm i eit rom folk kann gå inn i. I staden kallar kvar av
handsamarane hans `innsjekkKrav`, som ser etter låsen. Det er ein *medviten*
sjekk inni handsamaren, av di det ikkje finst nokon brukar å kjenna att.

## 3.5 Brukaren vert lesen éin gong

Opna `WithUser` i `handsamarar/middleware.go`. Han er meld inn på
`server.go:94`, **fyre kvar einaste rute**, so han gjeng alltid.

Tri ting han gjer, og alle tri er verdt grunnen:

**Han hoppar yver `/static/`.** Lesaren sender øktkaka med kvar CSS-, skrift-
og biletsoknad. Utan denne vakti kosta kvar einskild fil eit basesøk for ein
innlogga brukar.

**Han les brukaren éin gong og legg honom i lomma.** Soknaden ber ein
`context.Context` — eit uforanderleg nykel–verd-lager som fylgjer nett denne
soknaden. Du kann ikkje endra han; du lagar ein ny soknad med eit nytt innhald:

```go
func withUser(r *http.Request, u *models.User) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), userCtxKey, u))
}
```

Nykelen er av ein **privat type**:

```go
type ctxKey int
const (
	userCtxKey ctxKey = iota
	csrfCtxKey
)
```

Sidan `ctxKey` ikkje er eksportert, kann ingen annan pakke laga ein verd av
den typen. Difor kann ingen annan pakke sin nykel kollidera med vår — jamvel um
han òg nyttar talet 0. Det er verdt å merkja seg som mønster: **ein privat type
som nykel** er Go sin måte å laga eit namnerom utan eit namnerom.

**I utvikling les han umsetjingane på nytt** for kvar soknad, so ein ny nykel
syner seg utan at du lyt starta tenaren att.

Fylgja av at han alltid gjeng fyrst: `GetUserFromSession` treng berre å lesa
lomma.

```go
func GetUserFromSession(r *http.Request) *models.User {
	u, _ := r.Context().Value(userCtxKey).(*models.User)
	return u
}
```

`.(*models.User)` er ei **typepåstand** — verdet kjem ut av lomma som `any`, og
dette hentar den konkrete typen ut att. Tvo-verd-forma panikkar aldri; `u` vert
`nil` um det ikkje låg noko der.

Fyrr dette vart stramma inn, fall funksjonen attende til eit basesøk på eigi
hand. Det gav honom ei basetilkopling — og av di han vert kalla frå
`brukaren()`, `sessionIsAdmin()`, `sessionUserName()` og `IsLoggedIn()`, drog
den tilkoplingi seg gjenom heile økt- og mellomvarelaget. Å taka henne burt var
ikkje ein optimalisering; det var å fjerna ei avhengnad som ikkje trongst.

## 3.6 `Recoverer` er meir enn ein `recover()`

Opna `handsamarar/berging.go`. Han gjer det du ventar — tek panikk, loggar
stakken, svarar 500 — men han gjer noko meir, og det knyter seg beint til
regelen i §2.3.

Han **bufrar svaret**:

```go
b := &bufra{w: w}
defer func() {
	p := recover()
	if p == nil {
		b.send()
		return
	}
	...
```

Grunnen: handsamarane strøymer malar beint ut. Kjem panikken midt i teikningi,
er statuslina alt send, og då kann ein ikkje lenger svara 500 — sida stend
halvskriven med `200 OK` yver seg. Ved å halda att bytene til handsamaren er
ferdig, kann `Recoverer` kasta det halve svaret og senda ei heil feilmelding i
staden.

Legg merke til kommentaren nedst i honom: *«No template here: the template
system may be the very thing that fell.»* Feilsida er skriven for hand av di
malsystemet kann vera det som brast.

Og at han sender `http.ErrAbortHandler` vidare i staden for å gjera han til ein
500: det er `net/http` sin eigen måte å bryta eit svar på, og han skal upp til
tenaren, ikkje verta ein feil.

## 3.7 Sjå det sjølv

Med tenaren i gang:

```
curl -s -o /dev/null -w '%{http_code}\n' http://localhost:18108/elev/timeplan
curl -s -o /dev/null -w '%{http_code}\n' http://localhost:18108/innlogging
```

Den fyrste gjev **303** — `RequireAuth` såg ingen brukar og sende deg til
innloggingi. Den andre gjev **200**.

Skilnaden stend ikkje i `ElevTimeplan`. Opna honom og sjå: der er ingen sjekk.
Skilnaden er kva blokk lina hans stend i.

---

## Du kann no

- lesa og skriva ein mellomvare, og vita at forma er heile avtala
- segja kvifor `Recoverer` lyt liggja ytst
- finna ut kva løyve ein handsamar krev — med å slå upp innmeldingslina hans
- skjøna korleis brukaren kjem frå økti til handsamaren utan eit basesøk per
  kall
- opna `handsamarar/middleware.go` og lesa heile fila

**Neste:** kapittel 4 spør kvifor handsamarane er *metodar* i det heile — og
byrjar med å lata deg kjenna problemet dei løyser.
