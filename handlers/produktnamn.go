package handlers

import (
	"fmt"
	"strings"

	"kjernekraft/models"
)

// Namnet på eit produkt.
//
// Studioet skriv fakta, ikkje ord. «12-måneder», «Ingen binding» og «5
// Klipp - Gruppetimer» var namn nokon hadde skrive inn på bokmål — og
// det som er skrive inn, kan ingen omsetjing nå. Det er dessutan noko
// basen alt visste: bindinga står i `commitment_months`, klippa i
// `klipp_count`.
//
// So namnet vert skrive av systemet, på det språket den som les har
// valt. Studioet kan overstyre — eit haustkampanjekort heiter
// «Hausttilbod» og ikkje «Årskort» — og ei overstyring gjeld det språket
// ho vart skriven i. Tømer du feltet, fell namnet attende til det
// genererte. Det er slik ein angrar eit namn utan ein eigen knapp.

// MedlemskapNamn gjev det genererte namnet på eit medlemskap.
func MedlemskapNamn(lang string, m models.Membership) string {
	// Eit usynleg medlemskap er ikkje eit bindingsprodukt, og namnet
	// vert difor ikkje rekna ut or bindingi. Black hev null maanader
	// binding — som «Månadskort» — so den utrekningi gav honom namnet
	// til eit produkt han ikkje er.
	//
	// Namnet i basen er kjennemerket, ikkje det synlege ordet: han heiter
	// «Black» der, «Svart» paa norsk og «Black Card» paa engelsk. Difor
	// eit uppslag og ikkje ein fast streng — og eit uppslag som fell
	// attende paa raanamnet, so eit nytt usynleg medlemskap syner seg
	// med ein gong og fær sitt eige ord naar nokon skriv nykelen.
	// Sjaa database/svartmedlem.go.
	if m.Skjult {
		nykel := "produkt.medlemskap_" + strings.ToLower(strings.ReplaceAll(m.Name, " ", "_"))
		if namn := t(lang, nykel); namn != nykel {
			return namn
		}
		return m.Name
	}

	var grunn string
	switch m.CommitmentMonths {
	case 0:
		grunn = t(lang, "produkt.manadskort")
	case 6:
		grunn = t(lang, "produkt.halvaarskort")
	case 12:
		grunn = t(lang, "produkt.aarskort")
	default:
		grunn = fmt.Sprintf(t(lang, "produkt.binding_md"), m.CommitmentMonths)
	}
	if m.IsStudentSenior {
		return fmt.Sprintf(t(lang, "produkt.student_suffix"), grunn)
	}
	return grunn
}

// KlippekortNamn gjev det genererte namnet på ein klippekortpakke.
func KlippekortNamn(lang string, p models.KlippekortPackage) string {
	return fmt.Sprintf(t(lang, "produkt.klipp"), p.KlippCount)
}

// Namn vel overstyringa om ho finst, elles det genererte.
func Namn(overstyrt map[string]string, lang, generert string) string {
	if overstyrt != nil {
		if n, ok := overstyrt[lang]; ok && n != "" {
			return n
		}
	}
	return generert
}
