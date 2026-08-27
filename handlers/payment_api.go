package handlers

import (
	"html/template"
	"kjernekraft/models"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// PaymentMethodsHandler provides HTMX endpoint for user's payment methods
func PaymentMethodsHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from session
	user := GetUserFromSession(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// For now, return mock data since we don't have Stripe integration yet
	paymentMethods := []models.PaymentMethod{
		{
			ID:          1,
			UserID:      user.ID,
			Type:        "card",
			Last4:       "4242",
			Brand:       "visa",
			ExpiryMonth: 12,
			ExpiryYear:  2025,
			IsDefault:   true,
			Active:      true,
		},
		{
			ID:          2,
			UserID:      user.ID,
			Type:        "card",
			Last4:       "5555",
			Brand:       "mastercard",
			ExpiryMonth: 8,
			ExpiryYear:  2026,
			IsDefault:   false,
			Active:      true,
		},
	}

	// Merket vert skrive som folk kjenner det att — «Visa», ikkje
	// «visa», og ikkje bokstavane i ein boks. Er merket ukjent, står det
	// som det står.
	merke := map[string]string{
		"visa": "Visa", "mastercard": "Mastercard", "amex": "American Express",
	}
	type kort struct {
		models.PaymentMethod
		Merke string
	}
	rader := make([]kort, 0, len(paymentMethods))
	for _, p := range paymentMethods {
		namn, finst := merke[p.Brand]
		if !finst {
			namn = p.Brand
		}
		rader = append(rader, kort{p, namn})
	}

	teiknFragment(w, "betalingsmaatar", map[string]interface{}{
		"Lang": GetLanguageFromRequest(r),
		"Kort": rader,
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

	// Malen nyttar `t` til umsetjingi, og ho hadde si eigi funksjons-
	// samling med berre `divf` og `title` i. Daa feila parsingi paa
	// «function "t" not defined», og /api/charges svara 500 kvar gong.
	// Ho fær den same samlingi som alle hine malarne no.
	tmplFuncs := getTemplateFuncs()
	tmplFuncs["title"] = func(s string) string {
		if len(s) == 0 {
			return s
		}
		return strings.ToUpper(s[:1]) + s[1:]
	}

	// Malen ligg i modules/membership/, ikkje i components/.
	templatePath := filepath.Join("handlers", "templates", "modules", "membership", "charges.html")
	t, err := template.New("charges.html").Funcs(tmplFuncs).ParseFiles(templatePath)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error parsing charges template: %v", err)
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
