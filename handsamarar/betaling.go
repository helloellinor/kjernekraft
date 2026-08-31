package handsamarar

import (
	"net/http"
)

// Betaling handles the payment methods page
func (a *App) Betaling(w http.ResponseWriter, r *http.Request) {
	user := brukaren(r)

	// Get language from cookies/request (using new system)
	lang := GetLanguageFromRequest(r)

	// Same forma som klippekort- og medlemskapssida: tittel, ei lina som
	// seier kva som stend til, og so fanone. Korta kjem frå den same
	// funksjonen som brotstykket teiknar, so talet og lista ikkje kann
	// segja kvar sitt.
	kort := betalingskorta(user)

	faner := []Tab{
		{Key: "korta", Name: t(lang, "payments.payment_methods")},
		{Key: "nytt-kort", Name: t(lang, "payments.add_payment_method")},
	}

	data := sidedata(r, SidaBetaling, "Betaling", map[string]any{
		"Faner": fanerekkje(r, "faneark-betaling", "fane",
			t(lang, "payments.title"), faner),
		"KortTal": len(kort),
		"User":    user,
	})

	renderPage(w, r, "pages/betaling", data)
}
