package models

import "time"

type User struct {
	ID                     int    `json:"id"`
	Name                   string `json:"name"`
	Birthdate              string `json:"birthdate"`
	Email                  string `json:"email"`
	Phone                  string `json:"phone"`
	Address                string `json:"address"`
	PostalCode             string `json:"postal_code"`
	City                   string `json:"city"`
	Country                string `json:"country"`
	Password               string `json:"password"`
	NewsletterSubscription bool   `json:"newsletter_subscription"`
	TermsAccepted          bool   `json:"terms_accepted"`
	// StudentSenior fortel at brukaren hev student- eller honnørbevis.
	// Det avgjer kva medlemskap han fær sjaa.
	StudentSenior bool `json:"student_senior"`
	// Naar studentrabatten gjeng ut. Null tyder «ingen utgang» — honnør
	// gjeng ikkje ut, og ein rabatt som aldri vart gjeven hev ingen dato.
	StudentSeniorGjengUt time.Time `json:"student_senior_gjeng_ut"`
	// Løyvi brukaren hev: kva han fær gjera med studioet. Ikkje det
	// same som gruppone han er med i — dei segjer kven han høyrer til.
	Loyve []string `json:"loyve"`
}

// KvalifisertFor segjer um brukaren fær student- eller honnørprisen.
//
// Regelen bur her og ikkje i handsamaren, av di tvo stader spør um
// honom: flata, som avgjer kva prisar du fær sjaa, og
// `CanChangeMembership`, som avgjer um du fær byta til deim. Stod han
// tvo stader, kunde lista syna deg ein pris du ikkje fekk kjøpa.
//
// Tvo vegar inn, og dei er ulike av natur. Honnør kjem av fødselsdagen —
// eit tal systemet alt hev — og gjeng aldri ut. Studentbeviset er noko
// nokon i studioet hev sett, og det gjeng ut, av di eit bevis gjer det.
func (u User) KvalifisertFor(naa time.Time) bool {
	if u.StudentSenior {
		if u.StudentSeniorGjengUt.IsZero() || naa.Before(u.StudentSeniorGjengUt) {
			return true
		}
	}
	if fodd, err := time.Parse("2006-01-02", u.Birthdate); err == nil {
		if naa.Sub(fodd).Hours()/24/365.25 >= 67 {
			return true
		}
	}
	return false
}
