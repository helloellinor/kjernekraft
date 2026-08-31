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
	// StudentSenior says the user has a student or senior card. It decides
	// 	// which memberships they get to see.
	StudentSenior bool `json:"student_senior"`
	// When the student discount expires. Zero means no expiry — senior does
	// 	// not expire, and a discount never granted has no date.
	StudentSeniorGjengUt time.Time `json:"student_senior_gjeng_ut"`
	// The user's permissions: what they may do with the studio. Not the same
	// 	// as the groups they are in — those say who they belong to.
	Løyve []string `json:"loyve"`
}

// KvalifisertFor says whether the user gets the student or senior rate.
//
// The rule lives here and not in the handler, because two places ask for
// it: the surface, which decides which prices you see, and
// CanChangeMembership, which decides whether you may switch to them. In two
// places, the list could show you a price you were not allowed to buy.
//
// Two ways in, different in nature. Senior follows from the birth date — a
// number the system already has — and never expires. The student card is
// something someone at the studio has seen, and it expires, because a card
// does.
func (u User) KvalifisertFor(nå time.Time) bool {
	if u.StudentSenior {
		if u.StudentSeniorGjengUt.IsZero() || nå.Before(u.StudentSeniorGjengUt) {
			return true
		}
	}
	if fodd, err := time.Parse("2006-01-02", u.Birthdate); err == nil {
		if nå.Sub(fodd).Hours()/24/365.25 >= 67 {
			return true
		}
	}
	return false
}
