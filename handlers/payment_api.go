package handlers

import (
	"html/template"
	"kjernekraft/models"
	"net/http"
	"strconv"
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

	data := struct {
		PaymentMethods    []models.PaymentMethod
		HasPaymentMethods bool
	}{
		PaymentMethods:    paymentMethods,
		HasPaymentMethods: len(paymentMethods) > 0,
	}

	tmpl := `{{if .HasPaymentMethods}}
<div class="payment-methods-list">
    {{range .PaymentMethods}}
    <div class="payment-method-card {{if .IsDefault}}default{{end}}">
        <div class="payment-method-info">
            <div class="payment-method-icon">
                {{if eq .Brand "visa"}}VISA
                {{else if eq .Brand "mastercard"}}MC
                {{else if eq .Brand "amex"}}AMEX
                {{else}}CARD
                {{end}}
            </div>
            <div class="payment-method-details">
                <div class="payment-method-brand">{{.Brand}}</div>
                <div class="payment-method-last4">•••• •••• •••• {{.Last4}}</div>
                <div class="payment-method-expiry">Utløper {{.ExpiryMonth}}/{{.ExpiryYear}}</div>
            </div>
        </div>
        <div class="payment-method-actions">
            {{if .IsDefault}}
            <span class="default-badge">Standard</span>
            {{else}}
            <button class="payment-method-btn set-default-btn" onclick="setDefaultPaymentMethod({{.ID}})">
                Sett som standard
            </button>
            {{end}}
            <button class="payment-method-btn remove-btn" onclick="removePaymentMethod({{.ID}})">
                Fjern
            </button>
        </div>
    </div>
    {{end}}
</div>
{{else}}
<div class="no-data">
    Du har ingen betalingsmetoder registrert.
</div>
{{end}}`

	t, err := template.New("payment-methods").Parse(tmpl)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
	}
}

// ChargesHandler provides HTMX endpoint for user's charges/billing history
func ChargesHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from session
	user := GetUserFromSession(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// For now, return mock data
	charges := []models.ChargeWithDetails{
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
			},
			PaymentMethodLast4: stringPtr("5555"),
			PaymentMethodBrand: stringPtr("mastercard"),
		},
	}

	data := struct {
		Charges    []models.ChargeWithDetails
		HasCharges bool
	}{
		Charges:    charges,
		HasCharges: len(charges) > 0,
	}

	tmpl := `{{if .HasCharges}}
<div class="charges-list">
    {{range .Charges}}
    <div class="charge-item">
        <div class="charge-info">
            <div class="charge-description">{{.Description}}</div>
            <div class="charge-date">{{.ChargeDate.Format "2. January 2006"}}</div>
            {{if and .PaymentMethodBrand .PaymentMethodLast4}}
            <div class="charge-payment-method">{{.PaymentMethodBrand | title}} •••• {{.PaymentMethodLast4}}</div>
            {{else}}
            <div class="charge-payment-method">Betalingsmetode fjernet</div>
            {{end}}
        </div>
        <div class="charge-amount">{{printf "%.0f" (divf .Amount 100)}} kr</div>
        <div class="charge-status {{.Status}}">
            {{if eq .Status "succeeded"}}Vellykket
            {{else if eq .Status "failed"}}Mislykket
            {{else if eq .Status "pending"}}Venter
            {{else}}{{.Status}}
            {{end}}
        </div>
    </div>
    {{end}}
</div>
{{else}}
<div class="no-data">
    Ingen belastninger funnet.
</div>
{{end}}`

	// Parse template with custom functions
	tmplFuncs := template.FuncMap{
		"divf": func(a, b interface{}) float64 {
			var aFloat, bFloat float64

			switch v := a.(type) {
			case int:
				aFloat = float64(v)
			case float64:
				aFloat = v
			default:
				return 0
			}

			switch v := b.(type) {
			case int:
				bFloat = float64(v)
			case float64:
				bFloat = v
			default:
				return 0
			}

			if bFloat == 0 {
				return 0
			}
			return aFloat / bFloat
		},
		"title": func(s string) string {
			if len(s) == 0 {
				return s
			}
			return string(s[0]-32) + s[1:]
		},
	}

	t, err := template.New("charges").Funcs(tmplFuncs).Parse(tmpl)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
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