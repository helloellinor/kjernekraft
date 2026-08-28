package handlers

import (
	"kjernekraft/models"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Betalingskort er eit kort slik sida syner det: sjølve kortet, og
// merket skrive som folk kjenner det att.
type Betalingskort struct {
	models.PaymentMethod
	Merke string
}

// betalingskorta gjev korta brukaren hev.
//
// Dei er paadikta enno — Stripe er ikkje kopla paa — og det er nettupp
// difor dei lyt koma fraa éin stad. Briefingen paa sida tel dei og
// brotstykket teiknar dei; tvo kjeldor hadde sagt kvar sitt tal, og
// «Du har 2 kort» yver ei tom lista er verre enn ingen briefing.
func betalingskorta(user *models.User) []Betalingskort {
	if user == nil {
		return nil
	}
	raa := []models.PaymentMethod{
		{ID: 1, UserID: user.ID, Type: "card", Last4: "4242", Brand: "visa",
			ExpiryMonth: 12, ExpiryYear: 2025, IsDefault: true, Active: true},
		{ID: 2, UserID: user.ID, Type: "card", Last4: "5555", Brand: "mastercard",
			ExpiryMonth: 8, ExpiryYear: 2026, IsDefault: false, Active: true},
	}

	// Merket vert skrive som folk kjenner det att — «Visa», ikkje
	// «visa», og ikkje bokstavane i ein boks. Er merket ukjent, står det
	// som det står.
	merke := map[string]string{
		"visa": "Visa", "mastercard": "Mastercard", "amex": "American Express",
	}
	ut := make([]Betalingskort, 0, len(raa))
	for _, p := range raa {
		namn, finst := merke[p.Brand]
		if !finst {
			namn = p.Brand
		}
		ut = append(ut, Betalingskort{p, namn})
	}
	return ut
}

// PaymentMethodsHandler provides HTMX endpoint for user's payment methods
func PaymentMethodsHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	teiknFragment(w, "betalingsmaatar", map[string]interface{}{
		"Lang": GetLanguageFromRequest(r),
		"Kort": betalingskorta(user),
	})
}

// ChargesHandler provides HTMX endpoint for user's charges/billing history
func ChargesHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from session
	user := GetUserFromSession(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get filter type from query parameter
	filterType := r.URL.Query().Get("type")

	// For now, return mock data
	allCharges := []models.ChargeWithDetails{
		{
			Charge: models.Charge{
				ID:            1,
				UserID:        user.ID,
				Amount:        104000, // 1040 kr
				Currency:      "NOK",
				Status:        "succeeded",
				Description:   "12-måneder medlemskap",
				ChargeDate:    time.Now().AddDate(0, 0, -15), // 15 days ago
				FailureReason: nil,
				Type:          "medlemskap",
			},
			PaymentMethodLast4: stringPtr("4242"),
			PaymentMethodBrand: stringPtr("visa"),
		},
		{
			Charge: models.Charge{
				ID:            2,
				UserID:        user.ID,
				Amount:        104000, // 1040 kr
				Currency:      "NOK",
				Status:        "succeeded",
				Description:   "12-måneder medlemskap",
				ChargeDate:    time.Now().AddDate(0, -1, -15), // 1 month and 15 days ago
				FailureReason: nil,
				Type:          "medlemskap",
			},
			PaymentMethodLast4: stringPtr("4242"),
			PaymentMethodBrand: stringPtr("visa"),
		},
		{
			Charge: models.Charge{
				ID:            3,
				UserID:        user.ID,
				Amount:        104000, // 1040 kr
				Currency:      "NOK",
				Status:        "failed",
				Description:   "12-måneder medlemskap",
				ChargeDate:    time.Now().AddDate(0, -2, -10), // 2 months and 10 days ago
				FailureReason: stringPtr("Insufficient funds"),
				Type:          "medlemskap",
			},
			PaymentMethodLast4: stringPtr("5555"),
			PaymentMethodBrand: stringPtr("mastercard"),
		},
		{
			Charge: models.Charge{
				ID:            4,
				UserID:        user.ID,
				Amount:        220000, // 2200 kr
				Currency:      "NOK",
				Status:        "succeeded",
				Description:   "10 klipp Gruppetimer Sal",
				ChargeDate:    time.Now().AddDate(0, 0, -5), // 5 days ago
				FailureReason: nil,
				Type:          "klippekort",
			},
			PaymentMethodLast4: stringPtr("4242"),
			PaymentMethodBrand: stringPtr("visa"),
		},
		{
			Charge: models.Charge{
				ID:            5,
				UserID:        user.ID,
				Amount:        280000, // 2800 kr
				Currency:      "NOK",
				Status:        "succeeded",
				Description:   "10 klipp Reformer/Apparatus",
				ChargeDate:    time.Now().AddDate(0, 0, -20), // 20 days ago
				FailureReason: nil,
				Type:          "klippekort",
			},
			PaymentMethodLast4: stringPtr("4242"),
			PaymentMethodBrand: stringPtr("visa"),
		},
	}

	// Filter charges by type if specified
	var charges []models.ChargeWithDetails
	if filterType != "" {
		for _, charge := range allCharges {
			if charge.Type == filterType {
				charges = append(charges, charge)
			}
		}
	} else {
		charges = allCharges
	}

	data := struct {
		Charges    []models.ChargeWithDetails
		HasCharges bool
		Lang       string
	}{
		Charges:    charges,
		HasCharges: len(charges) > 0,
		Lang:       GetLanguageFromRequest(r),
	}

	// TemplateManager lastar alle modular éin gong i drift, og lastar
	// dei på nytt når ei fil er endra i utvikling. Det sparer parsing på
	// kvar htmx-forespurnad utan å gjera malendringar trege å sjå.
	t, ok := GetTemplateManager().GetTemplate("modules/membership/charges")
	if !ok {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Charges template not found")
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := t.ExecuteTemplate(w, "charges_module", data); err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		log.Printf("Error executing charges template: %v", err)
	}
}

// SetDefaultPaymentMethodHandler handles setting a payment method as default
func SetDefaultPaymentMethodHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	user := GetUserFromSession(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	paymentMethodIDStr := r.FormValue("payment_method_id")
	_, err := strconv.Atoi(paymentMethodIDStr)
	if err != nil {
		http.Error(w, "Invalid payment method ID", http.StatusBadRequest)
		return
	}

	// TODO: Implement actual database update
	// For now, just return success
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Payment method set as default"))
}

// RemovePaymentMethodHandler handles removing a payment method
func RemovePaymentMethodHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from session
	user := GetUserFromSession(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	paymentMethodIDStr := r.FormValue("payment_method_id")
	_, err := strconv.Atoi(paymentMethodIDStr)
	if err != nil {
		http.Error(w, "Invalid payment method ID", http.StatusBadRequest)
		return
	}

	// TODO: Implement actual database removal and Stripe detachment
	// For now, just return success
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Payment method removed"))
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
