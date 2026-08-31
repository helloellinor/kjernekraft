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
	// Skjult = the membership does *not* stand in the list where people
	// 	// choose. Black is hidden; nothing else is.
	// 	//
	// 	// The flag is written as "hidden" and not "visible" deliberately. A
	// 	// bool in Go is false when nobody said anything, and a membership
	// 	// nobody has said anything about should be *visible*. With "visible",
	// 	// every Membership{} built in code would be invisible and unbuyable
	// 	// without anyone asking for it. See database/svartmedlem.go.
	Skjult bool `json:"skjult"`
}

// UserMembership represents a user's active membership
type UserMembership struct {
	// RadID is this row's own key in user_memberships.
	//
	// Not ID: MembershipWithDetails embeds both Membership and this one,
	// and two embedded fields with the same name means Go promotes
	// neither and encoding/json drops the field without saying so. With
	// distinct names, .ID is the product — which is what the templates
	// read.
	RadID        int        `json:"rad_id"`
	UserID       int        `json:"user_id"`
	MembershipID int        `json:"membership_id"`
	Status       string     `json:"status"` // "active", "paused", "cancelled", "freeze_requested"
	StartDate    time.Time  `json:"start_date"`
	RenewalDate  time.Time  `json:"renewal_date"`
	EndDate      *time.Time `json:"end_date"`    // NULL if ongoing
	BindingEnd   *time.Time `json:"binding_end"` // When binding period ends
	LastBilled   time.Time  `json:"last_billed"` // When user was last billed
	// FrozenAt = when the freeze began; NULL while the membership runs. The
	// 	// expiry does not count while it is set. See GjeldTil.
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
	// Tildelt = the membership follows a permission and was not bought. There
	// 	// is no row to freeze or cancel, so the template hides those buttons.
	// 	// See database/svartmedlem.go.
	Tildelt bool `json:"tildelt"`
}

// GjeldTil answers how long the card is valid — the expiry date.
//
// nil means *no expiry*: an assigned membership follows a role and has no
// date to show. The card draws ∞ there.
//
// Four branches, each with a reason:
//
//	assigned       no date — the role decides, not the calendar
//	cancelled      the renewal — you have bought out that period and no
//	               more. Cancelling writes only `status` and touches
//	               neither end_date nor binding_end, so the binding stands
//	               in the database as though it applied; without this
//	               branch a cancelled yearly card said "2027" while the
//	               page beside it said "until the time you have paid for".
//	binding > 0    the end of the binding — a yearly card lasts twelve months
//	otherwise      the renewal — a monthly card lasts a month
//
// And then the clock: if the membership is frozen *now*, the expiry has not
// counted since FrozenAt, so the stored date is too early. We add the time
// the freeze has lasted, which is exactly what UnfreezeMembership writes
// when it restarts the membership. The card therefore shows the same before
// and after, and the number is true every day the freeze lasts.
func (m MembershipWithDetails) GjeldTil(nå time.Time) *time.Time {
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
		if stod := nå.Sub(*m.FrozenAt); stod > 0 {
			ut = ut.Add(stod)
		}
	}
	return &ut
}
