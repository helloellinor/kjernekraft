package handsamarar

import "net/http"

// Fanone.
//
// Ei fana er ei lenkja, og valet stend i adressa.
//
// Det var 139 liner JavaScript fyrr: alle bolkarne stod i dokumentet,
// eit skript sette `hidden` paa dei ein ikkje saag, og ein hash var det
// einaste som fortalde kva ein stod paa. Tenaren visste ingen ting — han
// teikna alle fem bolkarne i administrasjonen kvar einaste gong, og
// fire av deim fekk ingen augo.
//
// No er det tenaren som veit. `?fane=` er valet, handsamaren les det
// her, og malen teiknar den eine bolken. htmx byter ut `.faneark` og
// skriv adressa; er JavaScript-en burte, er lenkja framleis ei lenkja og
// sida kjem heil.

// Tab er ei fana: nykelen som stend i adressa, og namnet som stend paa
// henne. `Href` og `Vald` vert fyllte av fanerekkje() — ein kallar skriv
// dei tvo fyrste og ikkje meir.
type Tab struct {
	Key  string
	Name string
	Href string
	Vald bool
}

// Fanerekkje er alt malen treng for aa teikna rekkja.
//
// `ID` er id-en paa `<div class="faneark">`. htmx byter nett det
// elementet — `hx-target` og `hx-select` peikar baae paa honom — so han
// lyt vera den same i svaret som i sida ein stend paa.
type Fanerekkje struct {
	ID    string
	Label string
	Tabs  []Tab
	Vald  string
}

// fanerekkje les valet or adressa, fell attende paa fyrste fana, og
// skriv lenkja til kvar av deim.
//
// `param` er namnet i spurnadsstrengen. Ei rekkje inni ei onnor lyt ha
// sitt eige namn: prisrekkja i administrasjonen stend inne i fana
// «Prisar», og skreiv dei tvo det same namnet, tok eit klikk i den indre
// den ytre med seg.
//
// `nullstill` er dei namni denne rekkja skal taka burt naar ho byggjer
// lenkjone sine. Ei ytre rekkje nullstiller dei indre: `?fane=folk` skal
// ikkje slæpa med seg `&prisfane=reglar` fraa ei fana ein hev gjenge ut
// or.
func fanerekkje(r *http.Request, id, param, label string, faner []Tab, nullstill ...string) Fanerekkje {
	vald := ""
	bede := r.URL.Query().Get(param)
	for _, f := range faner {
		if f.Key == bede {
			vald = f.Key
			break
		}
	}
	// Ein nykel som ikkje finst er ikkje ein feil — det er ei gamal
	// lenkja, eller ei fana som er burte for denne brukaren (byt-fana
	// finst ikkje for eit tildelt medlemskap). Fyrste fana er svaret.
	if vald == "" && len(faner) > 0 {
		vald = faner[0].Key
	}

	ut := make([]Tab, len(faner))
	for i, f := range faner {
		f.Href = adresseMed(r, param, f.Key, nullstill...)
		f.Vald = f.Key == vald
		ut[i] = f
	}

	return Fanerekkje{ID: id, Label: label, Tabs: ut, Vald: vald}
}

// adresseMed er den same adressa med eitt spurnadsledd endra.
//
// Ei fana og ei vikepil gjer det same: dei peikar paa sida du stend paa,
// med éin ting annleis, og alt anna urørt. Sikti yver vika skal fylgja
// med naar du bladar, og vika skal fylgja med naar du siktar — held ein
// ikkje resten av strengen, misser ein det andre kvar gong.
//
// `nullstill` er dei ledi som ikkje hev meining lenger naar dette eine
// skifter. Sjaa fanerekkje().
func adresseMed(r *http.Request, param, verd string, nullstill ...string) string {
	q := r.URL.Query()
	for _, namn := range nullstill {
		q.Del(namn)
	}
	q.Set(param, verd)
	return r.URL.Path + "?" + q.Encode()
}
