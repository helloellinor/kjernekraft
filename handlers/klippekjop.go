package handlers

import (
	"kjernekraft/models"
)

// Kjøpsbolken paa klippekortsida.
//
// Pakkorne stod skrivne inn i HTML-en fyrr — tvo og tjue kort med namn,
// pris og «spar so mykje» — medan basen hadde sine eigne fem. Det var
// tvo sanningar um det same, og dei var ikkje samde: HTML-en baud fem
// klipp for 1100 kroner der basen sa 499. Ein som endra prisen i
// Prising endra ingen ting av det ein kunde saag.
//
// Difor kjem alt herifraa no, og kategoriane er dei som *finst*: HTML-en
// hadde seks, basen hev tvo, og dei fire som ikkje fanst kunde ikkje
// kjøpast i det heile — korti deira mangla pakke-id, so knappen enda i
// «Kunne ikke finne pakke-ID».
type Kategori struct {
	Namn   string
	Nykel  string // trygg for id-ar og spurnadsstrengen
	Pakkar []models.KlippekortPackage
}

// Kategoriar grupperer pakkorne i den rekkjefylgdi basen gjev deim.
// GetAllKlippekortPackages sorterer alt paa kategori og so pris, so
// grupperingi treng berre fylgja rekkja.
func Kategoriar(pakkar []models.KlippekortPackage) []Kategori {
	var ut []Kategori
	for _, p := range pakkar {
		if len(ut) == 0 || ut[len(ut)-1].Namn != p.Category {
			ut = append(ut, Kategori{Namn: p.Category, Nykel: nykel(p.Category)})
		}
		i := len(ut) - 1
		ut[i].Pakkar = append(ut[i].Pakkar, p)
	}
	return ut
}

// nykel gjer eit kategorinamn um til noko som toler aa staa i ein id og
// i ein spurnadsstreng. «Reformer/Apparatus» vart elles tvo stigsteg i
// ei adressa.
func nykel(s string) string {
	ut := make([]rune, 0, len(s))
	fyrre := '-'
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			ut = append(ut, r)
			fyrre = r
		case r >= 'A' && r <= 'Z':
			ut = append(ut, r+32)
			fyrre = r
		case r == 'æ' || r == 'Æ':
			ut = append(ut, 'a', 'e')
			fyrre = 'e'
		case r == 'ø' || r == 'Ø':
			ut = append(ut, 'o', 'e')
			fyrre = 'e'
		case r == 'å' || r == 'Å':
			ut = append(ut, 'a', 'a')
			fyrre = 'a'
		default:
			if fyrre != '-' {
				ut = append(ut, '-')
				fyrre = '-'
			}
		}
	}
	for len(ut) > 0 && ut[len(ut)-1] == '-' {
		ut = ut[:len(ut)-1]
	}
	return string(ut)
}
