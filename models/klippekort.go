package models

import "time"

// KlippekortPackage represents a klippekort package for purchase
type KlippekortPackage struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Category        string `json:"category"`          // "Personlig Trening", "Reformer", etc.
	KlippCount      int    `json:"klipp_count"`       // Number of clips in package
	Price           int    `json:"price"`             // Price in Norwegian øre
	PricePerSession int    `json:"price_per_session"` // Calculated price per session
	Description     string `json:"description"`
	ValidDays       int    `json:"valid_days"` // How many days the package is valid for
	Active          bool   `json:"active"`
	IsPopular       bool   `json:"is_popular"` // For highlighting best value
}

// UserKlippekort represents a user's purchased klippekort
type UserKlippekort struct {
	ID             int       `json:"id"`
	UserID         int       `json:"user_id"`
	PackageID      int       `json:"package_id"`
	TotalKlipp     int       `json:"total_klipp"`
	RemainingKlipp int       `json:"remaining_klipp"`
	ExpiryDate     time.Time `json:"expiry_date"`
	PurchaseDate   time.Time `json:"purchase_date"`
	IsActive       bool      `json:"is_active"`
}

// KlippekortWithDetails combines package info with user's klippekort data
type KlippekortWithDetails struct {
	KlippekortPackage
	UserKlippekort
	ProgressPercentage int  `json:"progress_percentage"` // How much has been used
	DaysUntilExpiry    int  `json:"days_until_expiry"`
	IsExpiring         bool `json:"is_expiring"` // True if expires within 30 days
	KlipteHol          int  `json:"klipte_hol"`  // Hol som er klipte, av HolPerKort
}

// HolPerKort er kor mange klippehol eit kort hev — alltid det same
// talet, same kor mange klipp pakka inneheldt. Stilboki teiknar ti, og
// det er ei *form* og ikkje ei mengd: hakkrekkja svarar paa «er kortet
// brukt», ikkje «kor mange gonger». Talet attmed svarar paa det andre.
//
// Fylgja er at eit kort med tjuge klipp fær eit hol for kvart andre
// klipp. Det er meiningi: elles laut anten kortet verta dobbelt so
// høgt, eller hola so smaa at dei ikkje er hol lenger.
const HolPerKort = 10

// KlipteHolAv reknar ut kor mange av dei ti hola som er klipte.
// Avrunding til nærmaste, so eit kort som er halvvegs brukt syner fem.
// Eit ubrukt kort syner null — vengen stend heil — og eit tomt kort
// syner alle ti.
func KlipteHolAv(brukte, alle int) int {
	if alle <= 0 || brukte <= 0 {
		return 0
	}
	if brukte >= alle {
		return HolPerKort
	}
	return (brukte*HolPerKort + alle/2) / alle
}

// NaermastUtlop finn det kortet som gjeng ut fyrst av dei som framleis
// hev klipp att.
//
// Kort som er tome eller alt utgjengne tel ikkje med: eit tomt kort som
// gjeng ut er ikkje ein frist, det er ei kvittering.
//
// Han budde i klippekort-brotstykket fyrr. Sida treng den same
// utrekningi til briefingen sin, og tvo utgaavor av den same regelen
// driv frå kvarandre — difor bur han her, der baae kann naa honom.
func NaermastUtlop(kort []KlippekortWithDetails) *KlippekortWithDetails {
	var naermast *KlippekortWithDetails
	for i := range kort {
		k := &kort[i]
		if k.RemainingKlipp <= 0 || k.DaysUntilExpiry < 0 {
			continue
		}
		if naermast == nil || k.DaysUntilExpiry < naermast.DaysUntilExpiry {
			naermast = k
		}
	}
	return naermast
}
