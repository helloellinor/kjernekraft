# 2. Ein tenar som svarar

*Éi form, og alt er den formi.*

---

## 2.1 I C++ ville du valt eit rammeverk

Skulde du skriva ein HTTP-tenar i C++, er det fyrste spørsmålet kva du byggjer
på. Boost.Beast? Drogon? Crow? cpp-httplib? oat++? Ingen av dei er sjølvsagd,
alle hev sin eigen måte å skriva ein handsamar på, og valet bind deg i fleire
år.

I Go finst ikkje det spørsmålet. Tenaren er i standardbiblioteket, og han er
bygd på **éin** type — eit grensesnitt med éin metode:

```go
type Handler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}
```

Det er heile grunnlaget. Alt anna i dette kapittelet, og mesteparten av
kapittel 3, er fylgjor av desse tvo linone.

## 2.2 Knoppen på klossen

Tenk på det som ein legokloss. Grensesnittet er **knoppemynsteret**: kva som
helst som hev ein `ServeHTTP` med nett den signaturen, *er* ein `Handler`, og
kann setjast inn kvar som helst ein `Handler` skal vera.

Og no det som skil Go frå C++: **du erklærer det aldri.** Det finst ingen
`class Foo : public Handler`. Ein type stettar grensesnittet av di han hev
metoden, ikkje av di han hev sagt frå at han vil. Klossen passar av di
knoppane hev rett storleik og avstand — ikkje av di han låg i same øskja.

Det heiter strukturell typing, og du kjem til å møta det att i kapittel 7, der
det gjer noko meir enn å spara ei line.

## 2.3 Dei tvo argumenti

```go
func (a *App) Stilark(w http.ResponseWriter, r *http.Request)
```

`r *http.Request` er **alt som kom inn**: adressa, metoden, hovudlinone,
kakone, kroppen — og ei lomme som fylgjer soknaden, som me kjem attende til i
kapittel 3.

`w http.ResponseWriter` er **alt som skal ut**. Du set hovudliner med
`w.Header().Set(...)`, og so skriv du bytes med `w.Write(...)`.

**Rekkjefylgdi er ikkje forhandlingsspørsmål.** Fyrste `Write` sender
hovudlinone og statuslina av garde. Alt du set etterpå vert stillteiande kasta.
Dette er den vanlegaste «kvifor kjem ikkje hovudlina mi fram»-feilen i Go, og
det er ikkje ein feil kompilatoren kann taka.

Legg merke til at det ikkje finst noko returverd. Ein handsamar som ikkje
skriv noko gjev `200 OK` med tom kropp.

## 2.4 Funksjonar som klossar

Skulde du laga ein `struct` med ein `ServeHTTP`-metode for kvar rute, vart det
tungt. Difor hev standardbiblioteket ein overgang:

```go
type HandlerFunc func(http.ResponseWriter, *http.Request)

func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f(w, r)
}
```

Les det ein gong til, for heile økosystemet kviler på dette trikset.
`HandlerFunc` er **ein type som *er* ein funksjon**, og han hev ein
`ServeHTTP`-metode som berre kallar seg sjølv. So `http.HandlerFunc(minFunksjon)`
gjer kva som helst funksjon med rett form um til ein `Handler`.

I C++ finst det ikkje noko heilt som dette. Det næraste er å pakka ei lambda i
ein `std::function` og gjeva klassa ein `operator()` — men her er det typen
sjølv som er funksjonen, ikkje ein innpakning kring honom.

Dette er grunnen til at du kann skriva

```go
r.Get("/elev/timeplan", app.ElevTimeplan)
```

og gjeva ein metode beint til rutaren. Han hev rett form; rutaren pakkar
honom inn.

## 2.5 Ein handsamar du alt hev

Opna `handsamarar/dashboard_components.go` og finn `UserSignups`. Han er kort
nok til å lesast heilt, og han er heilt vanleg:

```go
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

	teiknFragmentFrå(w, "pages/dashboard", "signed_up_classes_module", map[string]interface{}{
		"EnrolledSessions": paamelde,
		"Lang":             lang,
		"CSRFToken":        CSRFToken(r),
		"IsAdmin":          sessionIsAdmin(r),
		"UserName":         sessionUserName(r),
	})
}
```

Fem ting å merkja seg:

**`user := brukaren(r)`** — `:=` erklærer og gissar typen. Inne i ein funksjon
er dette nesten alltid det du nyttar; `var` er for pakkenivå.

**`paamelde, err := ...`** — **fleire returverdar**. Go hev ingen unnatak; ein
funksjon som kann mislukkast gjev feilen attende som siste verd. Meir um dette
i kapittel 5, og skikkeleg i kapittel 12.

**`if err != nil { ... return }`** — og so er du ute. Ingi utrulling av stakken,
ingen `catch`, ingen skjult utgang. Det er meir å skriva enn `throw`, og til
gjengjeld finst det ingi line i programmet som kann hoppa ut utan at det stend
der.

**`map[string]interface{}{...}`** — eit kart frå streng til kva som helst. Dette
er dataene malen fær. Han er **utypa med vilje**, og han er òg ei felle:
spør malen etter ein nykel som ikkje er der, teiknar han tomt i staden for å
klaga. Kapittel 8 handlar um det.

**Ingen `return` til slutt.** Funksjonen gjev ikkje noko attende; han *skriv*.
Svaret gjeng ut gjenom `w`, ikkje attende gjenom typen.

## 2.6 Kva `main` gjer

Opna `server.go`. Sjå etter dei siste linone — mønsteret er alltid det same:

```go
r := chi.NewRouter()
// ... mellomvare og rutor ...
http.ListenAndServe(addr, r)
```

`http.ListenAndServe` tek ei adressa og **éin `Handler`**. Ikkje ei liste,
ikkje eit rammeverk — éin kloss.

Og her er poenget: `r`, heile rutaren med alle 68 rutone og all mellomvara, er
sjølv ein `Handler`. Han hev ein `ServeHTTP`. Byggjeplata er ein kloss.

Det er difor det ikkje finst noko «rammeverkets inngangspunkt» i Go. Du set
saman ein `Handler` — som gjerne er sett saman av tusen andre — og rekkjer
honom til standardbiblioteket.

## 2.7 Ein handsamar som ikkje rører basen

Vil du sjå den minste heile handsamaren i huset, opna
`handsamarar/stilark.go`. Han les 35 filer frå disken, set dei saman og skriv
dei ut. Han spør ikkje basen um noko.

Difor kann prøva i `handsamarar/prover/palett_test.go` skriva:

```go
handsamarar.NyApp(nil).Stilark(w, httptest.NewRequest(http.MethodGet, "/static/css/kjernekraft.css", nil))
```

`nil` som base. Handsamaren rører henne aldri, so det gjeng. `httptest.NewRecorder()`
er ein `ResponseWriter` som **skriv til minnet i staden for ut på nettet** — og
sidan `ResponseWriter` er eit grensesnitt (§2.2), veit ikkje handsamaren
skilnaden.

Det er ei lita sak, men det er heile prøvehistoria til Go i miniatyr, og
kapittel 11 byggjer på henne.

## 2.8 Kvifor rekkjefylgdi i §2.3 tyder noko

Fyrr me rydda i det, skreiv sju handsamarar sitt eige svar for hand og gløymde
teiknsettet:

```go
w.Header().Set("Content-Type", "text/html")   // ikkje "; charset=utf-8"
```

No gjeng dei alle gjenom `renderPage` i `handsamarar/middleware.go`, som set
det rett. Prøv sjølv:

```
curl -s -D- -o /dev/null http://localhost:18108/innlogging | grep -i content-type
```

Du skal få `text/html; charset=utf-8`.

At det ikkje synte seg som ein feil i lesaren skuldast at `base.html` ber
`<meta charset="UTF-8">`, og malen vinn når hovudlina ikkje segjer noko. Tvo
sanningar um det same, og den eine var gale i eit år utan at nokon såg det.
Det er eit betre argument for éin veg ut enn noko eg kunde skrive.

---

## Du kann no

- segja kva ein `Handler` er, og kvifor alt i Go-vev er den formi
- lesa ein handsamarsignatur og vita kva dei tvo argumenti ber
- hugsa at hovudliner lyt setjast fyre fyrste `Write`
- skjøna kvifor ein metode kann gjevast beint til `r.Get(...)`
- opna `handsamarar/dashboard_components.go` og lesa `UserSignups` heilt

**Neste:** kapittel 3 tek klossen som *slukar* ein annan kloss, og syner at
heile løyvemodellen i Kjernekraft er fire liner chi.
