package models

import "time"

// Membership represents a membership type/plan
type Membership struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	Price            int    `json:"price"`             // Price in Norwegian øre (1000 = 10.00 kr)
	CommitmentMonths int    `json:"commitment_months"` // 0 for no commitment, 1, 6, 12 etc.
	IsStudentSenior  bool   `json:"is_student_senior"`
	IsSpecialOffer   bool   `json:"is_special_offer"`
	Description      string `json:"description"`
	Features         string `json:"features"` // JSON string of features array
	Active           bool   `json:"active"`
	// Skjult = medlemskapet stend *ikkje* i lista der folk vel. Black er
	// skjult; alt anna er ikkje.
	//
	// Flagget er skrive som «skjult» og ikkje «synleg» med vilje. Eit
	// bool i Go er false naar ingen sa noko, og eit medlemskap som ingen
	// hev sagt noko om skal vera *synleg*. Med «synleg» var kvart
	// Membership{} som vart bygd i kode usynleg og ukjøpeleg utan at
	// nokon bad um det. Sjaa database/svartmedlem.go.
	Skjult bool `json:"skjult"`
}

// UserMembership represents a user's active membership
type UserMembership struct {
	ID           int        `json:"id"`
	UserID       int        `json:"user_id"`
	MembershipID int        `json:"membership_id"`
	Status       string     `json:"status"` // "active", "paused", "cancelled", "freeze_requested"
	StartDate    time.Time  `json:"start_date"`
	RenewalDate  time.Time  `json:"renewal_date"`
	EndDate      *time.Time `json:"end_date"`    // NULL if ongoing
	BindingEnd   *time.Time `json:"binding_end"` // When binding period ends
	LastBilled   time.Time  `json:"last_billed"` // When user was last billed
	// FrozenAt = naar frysingi tok til; NULL naar medlemskapet gjeng.
	// Utlaupet tel ikkje medan han er sett. Sjaa GjeldTil.
	FrozenAt  *time.Time `json:"frozen_at"`
	CreatedAt time.Time  `json:"created_at"`
}

// MembershipWithDetails combines membership info with user-specific data
type MembershipWithDetails struct {
	Membership
	UserMembership
	DaysUntilRenewal      int  `json:"days_until_renewal"`
	MonthsUntilBindingEnd int  `json:"months_until_binding_end"`
	CanCancel             bool `json:"can_cancel"`
	CanPause              bool `json:"can_pause"`
	// Tildelt = medlemskapet fylgjer eit løyve og er ikkje kjøpt. Det
	// finst ingi rad aa frysa eller seia upp, so malen skal gøyma dei
	// knappane. Sjaa database/svartmedlem.go.
	Tildelt bool `json:"tildelt"`
}

// GjeldTil svarar paa kor lenge kortet gjeld — utlaupsdatoen.
//
// nil tyder *utan utlaup*: eit tildelt medlemskap fylgjer ei rolla og
// hev ingen dato aa syna. Kortet teiknar ∞ der.
//
// Fire greiner, og kvar av dei hev ein grunn:
//
//	tildelt        ingen dato — rolla avgjer, ikkje kalenderen
//	sagt upp       fornyingi — du hev kjøpt ut den perioden og ikkje meir.
//	               Uppseiingi skriv berre `status` og rører korkje
//	               `end_date` eller `binding_end`, so bindingi stend att
//	               i basen som um ho gjaldt; utan denne greini hadde eit
//	               uppsagt aarskort sagt «2027» medan sida attmed sa
//	               «ut den tidi du hev betalt for».
//	binding > 0    slutten paa bindingi — eit aarskort varer tolv maanader
//	elles          fornyingi — eit maanadskort varer ein maanad
//
// Og so klokka: er medlemskapet frose *no*, hev ikkje utlaupet talt
// sidan `FrozenAt`, so den lagra datoen er for tidleg. Me legg til den
// tidi frysingi hev vart, som er nettupp det UnfreezeMembership skriv
// naar han set medlemskapet i gang att. Kortet syner difor det same
// fyre og etter, og talet er sant kvar dag frysingi varer.
func (m MembershipWithDetails) GjeldTil(naa time.Time) *time.Time {
	if m.Tildelt {
		return nil
	}

	var ut time.Time
	switch {
	case m.Status == "cancelled":
		ut = m.RenewalDate
	case m.CommitmentMonths > 0 && m.BindingEnd != nil:
		ut = *m.BindingEnd
	default:
		ut = m.RenewalDate
	}

	if m.Status == "paused" && m.FrozenAt != nil {
		if stod := naa.Sub(*m.FrozenAt); stod > 0 {
			ut = ut.Add(stod)
		}
	}
	return &ut
}
