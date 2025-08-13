package models

import "time"

// KlippekortPackage represents a klippekort package for purchase
type KlippekortPackage struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`    // "Personlig Trening", "Reformer", etc.
	KlippCount  int    `json:"klipp_count"` // Number of clips in package
	Price       int    `json:"price"`       // Price in Norwegian øre
	PricePerSession int `json:"price_per_session"` // Calculated price per session
	Description string `json:"description"`
	ValidDays   int    `json:"valid_days"`  // How many days the package is valid for
	Active      bool   `json:"active"`
	IsPopular   bool   `json:"is_popular"`  // For highlighting best value
}

// UserKlippekort represents a user's purchased klippekort
type UserKlippekort struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	PackageID     int       `json:"package_id"`
	TotalKlipp    int       `json:"total_klipp"`
	RemainingKlipp int      `json:"remaining_klipp"`
	ExpiryDate    time.Time `json:"expiry_date"`
	PurchaseDate  time.Time `json:"purchase_date"`
	IsActive      bool      `json:"is_active"`
}

// KlippekortWithDetails combines package info with user's klippekort data
type KlippekortWithDetails struct {
	KlippekortPackage
	UserKlippekort
	ProgressPercentage int  `json:"progress_percentage"` // How much has been used
	DaysUntilExpiry    int  `json:"days_until_expiry"`
	IsExpiring         bool `json:"is_expiring"`         // True if expires within 30 days
}