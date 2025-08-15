package models

import "time"

// PaymentMethod represents a user's payment method (Stripe)
type PaymentMethod struct {
	ID                int       `json:"id"`
	UserID            int       `json:"user_id"`
	StripePaymentMethodID string `json:"stripe_payment_method_id"`
	Type              string    `json:"type"`          // "card", "bank_account", etc.
	Last4             string    `json:"last4"`         // Last 4 digits
	Brand             string    `json:"brand"`         // "visa", "mastercard", etc.
	ExpiryMonth       int       `json:"expiry_month"`
	ExpiryYear        int       `json:"expiry_year"`
	IsDefault         bool      `json:"is_default"`
	Active            bool      `json:"active"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// Charge represents a billing charge/transaction
type Charge struct {
	ID                int       `json:"id"`
	UserID            int       `json:"user_id"`
	PaymentMethodID   *int      `json:"payment_method_id"` // NULL if payment method was deleted
	StripeChargeID    string    `json:"stripe_charge_id"`
	Amount            int       `json:"amount"`            // Amount in øre
	Currency          string    `json:"currency"`          // "NOK"
	Status            string    `json:"status"`            // "succeeded", "failed", "pending"
	Description       string    `json:"description"`       // What the charge was for
	Type              string    `json:"type"`              // "medlemskap", "klippekort", "utdanninger"
	ChargeDate        time.Time `json:"charge_date"`
	FailureReason     *string   `json:"failure_reason"`    // NULL if successful
	CreatedAt         time.Time `json:"created_at"`
}

// ChargeWithDetails includes payment method information for display
type ChargeWithDetails struct {
	Charge
	PaymentMethodLast4 *string `json:"payment_method_last4"`
	PaymentMethodBrand *string `json:"payment_method_brand"`
}