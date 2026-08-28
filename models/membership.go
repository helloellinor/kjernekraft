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
	CreatedAt    time.Time  `json:"created_at"`
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
