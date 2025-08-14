package handlers

import (
	"html/template"
	"kjernekraft/handlers/config"
	"kjernekraft/models"
	"log"
	"net/http"
	"strconv"
)

// UserKlippekortHandler provides HTMX endpoint for user's klippekort display
func UserKlippekortHandler(w http.ResponseWriter, r *http.Request) {
	// For now, use a test user ID. In a real app, this would come from session/auth
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		userIDStr = "1" // Default test user
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	klippekort, err := DB.GetUserKlippekort(userID)
	if err != nil {
		http.Error(w, "Could not fetch user klippekort", http.StatusInternalServerError)
		log.Printf("Error fetching klippekort for user %d: %v", userID, err)
		return
	}

	// Calculate additional fields for display
	for i := range klippekort {
		k := &klippekort[i]

		// Calculate progress percentage (remaining klipps)
		if k.TotalKlipp > 0 {
			k.ProgressPercentage = (k.RemainingKlipp * 100) / k.TotalKlipp
		}

		// Calculate days until expiry
		settings := config.GetInstance()
		now := settings.GetCurrentTime()
		k.DaysUntilExpiry = int(k.ExpiryDate.Sub(now).Hours() / 24)
		k.IsExpiring = k.DaysUntilExpiry <= 30 && k.DaysUntilExpiry > 0
	}

	data := struct {
		Klippekort    []models.KlippekortWithDetails
		HasKlippekort bool
	}{
		Klippekort:    klippekort,
		HasKlippekort: len(klippekort) > 0,
	}

	tmpl := `{{if .HasKlippekort}}
<div class="klippekort-cards">
    {{range .Klippekort}}
    <div class="klippekort-card {{if .IsExpiring}}expiring{{end}}">
        {{if .IsExpiring}}
        <div class="expiring-badge">Utløper snart</div>
        {{end}}
        
        <div class="card-header">
            <h4 class="card-title">{{.Name}}</h4>
            <span class="card-category">{{.Category}}</span>
        </div>
        
        <div class="card-content">
            <div class="klipp-count">
                <span class="remaining">{{.RemainingKlipp}}</span>
                <span class="total">/ {{.TotalKlipp}} klipp</span>
            </div>
            
            <div class="progress-bar">
                <div class="progress-fill" style="width: {{.ProgressPercentage}}%;"></div>
            </div>
            
            <div class="expiry-info">
                {{if gt .DaysUntilExpiry 0}}
                    Utløper om {{.DaysUntilExpiry}} dager
                {{else if eq .DaysUntilExpiry 0}}
                    Utløper i dag
                {{else}}
                    Utløpt
                {{end}}
            </div>
        </div>
    </div>
    {{end}}
</div>
{{else}}
<div class="no-klippekort">
    <p>Du har ingen aktive klippekort</p>
    <a href="/klippekort" class="buy-klippekort-btn">Kjøp klippekort</a>
</div>
{{end}}

<style>
.klippekort-cards {
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.klippekort-card {
    background: white;
    border-radius: 8px;
    padding: 1.5rem;
    box-shadow: 0 2px 8px rgba(0,0,0,0.1);
    border: 2px solid #e0e0e0;
    position: relative;
    transition: transform 0.2s, box-shadow 0.2s;
}

.klippekort-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 15px rgba(0,0,0,0.15);
}

.klippekort-card.expiring {
    border-color: #ff6b35;
    background: linear-gradient(135deg, #fff5f0, #ffffff);
}

.expiring-badge {
    position: absolute;
    top: -8px;
    right: 12px;
    background: #ff6b35;
    color: white;
    padding: 0.25rem 0.75rem;
    border-radius: 12px;
    font-size: 0.75rem;
    font-weight: 600;
}

.card-header {
    margin-bottom: 1rem;
}

.card-title {
    font-size: 1.1rem;
    font-weight: 600;
    color: #333;
    margin-bottom: 0.25rem;
}

.card-category {
    font-size: 0.85rem;
    color: #666;
    background: #f0f0f0;
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
}

.klipp-count {
    margin-bottom: 0.75rem;
}

.klipp-count .remaining {
    font-size: 1.5rem;
    font-weight: 700;
    color: #007cba;
}

.klipp-count .total {
    font-size: 1rem;
    color: #666;
}

.progress-bar {
    width: 100%;
    height: 8px;
    background: #e0e0e0;
    border-radius: 4px;
    overflow: hidden;
    margin-bottom: 0.75rem;
}

.progress-fill {
    height: 100%;
    background: linear-gradient(90deg, #007cba, #4a9fd1);
    border-radius: 4px;
    transition: width 0.3s ease;
}

.expiry-info {
    font-size: 0.85rem;
    color: #666;
}

.no-klippekort {
    text-align: center;
    padding: 2rem;
    color: #666;
}

.buy-klippekort-btn {
    display: inline-block;
    margin-top: 1rem;
    padding: 0.75rem 1.5rem;
    background: #007cba;
    color: white;
    text-decoration: none;
    border-radius: 6px;
    font-weight: 600;
    transition: background-color 0.2s;
}

.buy-klippekort-btn:hover {
    background: #005a87;
}
</style>`

	t, err := template.New("klippekort").Parse(tmpl)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
	}
}

// UserMembershipHandler provides HTMX endpoint for user's membership display
func UserMembershipHandler(w http.ResponseWriter, r *http.Request) {
	// For now, use a test user ID. In a real app, this would come from session/auth
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		userIDStr = "1" // Default test user
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	membership, err := DB.GetUserMembership(userID)
	if err != nil {
		http.Error(w, "Could not fetch user membership", http.StatusInternalServerError)
		log.Printf("Error fetching membership for user %d: %v", userID, err)
		return
	}

	// Calculate additional fields if membership exists
	if membership != nil {
		settings := config.GetInstance()
		now := settings.GetCurrentTime()
		membership.DaysUntilRenewal = int(membership.RenewalDate.Sub(now).Hours() / 24)

		// Business logic for what actions are available
		membership.CanPause = membership.Status == "active"

		// Can cancel if no binding period OR if binding period has ended
		if membership.BindingEnd == nil {
			membership.CanCancel = true
		} else {
			membership.CanCancel = now.After(*membership.BindingEnd)
		}
	}

	data := struct {
		Membership    *models.MembershipWithDetails
		HasMembership bool
	}{
		Membership:    membership,
		HasMembership: membership != nil,
	}

	tmpl := `{{if .HasMembership}}
<div class="membership-card {{.Membership.Status}}">
    {{if .Membership.IsSpecialOffer}}
    <div class="special-offer-badge">Spesialtilbud</div>
    {{end}}
    
    <div class="membership-header">
        <h3 class="membership-name">{{.Membership.Name}}</h3>
        <div class="membership-status status-{{.Membership.Status}}">
            {{if eq .Membership.Status "active"}}Aktiv
            {{else if eq .Membership.Status "paused"}}Pause
            {{else if eq .Membership.Status "cancelled"}}Kansellert
            {{end}}
        </div>
    </div>
    
    <div class="membership-details">
        <div class="price-info">
            <span class="price">{{printf "%.0f" (divf .Membership.Price 100)}} kr/mnd</span>
            {{if gt .Membership.CommitmentMonths 0}}
            <span class="commitment">{{.Membership.CommitmentMonths}} mnd binding</span>
            {{else}}
            <span class="commitment">Ingen binding</span>
            {{end}}
        </div>
        
        <div class="dates-info">
            {{if gt .Membership.DaysUntilRenewal 0}}
            <div class="renewal-date">
                <strong>Neste fornyelse:</strong> {{.Membership.RenewalDate.Format "2. January 2006"}}
                <span class="days-until">(om {{.Membership.DaysUntilRenewal}} dager)</span>
            </div>
            {{else if eq .Membership.DaysUntilRenewal 0}}
            <div class="renewal-date urgent">
                <strong>Fornyes i dag!</strong>
            </div>
            {{end}}
            
            {{if .Membership.BindingEnd}}
            <div class="binding-end">
                <strong>Binding utløper:</strong> {{.Membership.BindingEnd.Format "2. January 2006"}}
            </div>
            {{end}}
        </div>
    </div>
    
    {{if or .Membership.CanPause .Membership.CanCancel}}
    <div class="membership-actions">
        {{if .Membership.CanPause}}
        <button class="action-btn pause-btn" onclick="pauseMembership()">
            Sett på pause
        </button>
        {{end}}
        
        {{if .Membership.CanCancel}}
        <button class="action-btn cancel-btn" onclick="cancelMembership()">
            Si opp medlemskap
        </button>
        {{end}}
    </div>
    {{end}}
</div>
{{else}}
<div class="no-membership">
    <h3>Intet aktivt medlemskap</h3>
    <p>Du har ikke et aktivt medlemskap for øyeblikket.</p>
    <a href="/medlemskap" class="get-membership-btn">Finn medlemskap</a>
</div>
{{end}}

<style>
.membership-card {
    background: white;
    border-radius: 12px;
    padding: 2rem;
    box-shadow: 0 4px 12px rgba(0,0,0,0.1);
    border: 2px solid #e0e0e0;
    position: relative;
    max-width: 500px;
}

.membership-card.active {
    border-color: #28a745;
}

.membership-card.paused {
    border-color: #ffc107;
    background: linear-gradient(135deg, #fff9c4, #ffffff);
}

.membership-card.cancelled {
    border-color: #dc3545;
    background: linear-gradient(135deg, #f8d7da, #ffffff);
}

.special-offer-badge {
    position: absolute;
    top: -10px;
    right: 15px;
    background: #ff6b35;
    color: white;
    padding: 0.25rem 0.75rem;
    border-radius: 12px;
    font-size: 0.8rem;
    font-weight: 600;
}

.membership-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.5rem;
}

.membership-name {
    font-size: 1.5rem;
    font-weight: 600;
    color: #333;
    margin: 0;
}

.membership-status {
    padding: 0.5rem 1rem;
    border-radius: 20px;
    font-size: 0.85rem;
    font-weight: 600;
    text-transform: uppercase;
}

.status-active {
    background: #d4edda;
    color: #155724;
}

.status-paused {
    background: #fff3cd;
    color: #856404;
}

.status-cancelled {
    background: #f8d7da;
    color: #721c24;
}

.membership-details {
    margin-bottom: 1.5rem;
}

.price-info {
    margin-bottom: 1rem;
}

.price {
    font-size: 1.8rem;
    font-weight: 700;
    color: #007cba;
    margin-right: 0.5rem;
}

.commitment {
    font-size: 0.9rem;
    color: #666;
    background: #f0f0f0;
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
}

.dates-info {
    font-size: 0.9rem;
    color: #666;
}

.renewal-date, .binding-end {
    margin-bottom: 0.5rem;
}

.renewal-date.urgent {
    color: #dc3545;
    font-weight: 600;
}

.days-until {
    color: #999;
}

.membership-actions {
    display: flex;
    gap: 1rem;
    margin-top: 1.5rem;
    padding-top: 1.5rem;
    border-top: 1px solid #e0e0e0;
}

.action-btn {
    padding: 0.75rem 1.5rem;
    border: none;
    border-radius: 6px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
    flex: 1;
}

.pause-btn {
    background: #ffc107;
    color: #000;
}

.pause-btn:hover {
    background: #e0a800;
}

.cancel-btn {
    background: #dc3545;
    color: white;
}

.cancel-btn:hover {
    background: #c82333;
}

.no-membership {
    text-align: center;
    padding: 2rem;
    background: white;
    border-radius: 12px;
    box-shadow: 0 4px 12px rgba(0,0,0,0.1);
    max-width: 500px;
}

.no-membership h3 {
    color: #333;
    margin-bottom: 0.5rem;
}

.no-membership p {
    color: #666;
    margin-bottom: 1.5rem;
}

.get-membership-btn {
    display: inline-block;
    padding: 0.75rem 1.5rem;
    background: #007cba;
    color: white;
    text-decoration: none;
    border-radius: 6px;
    font-weight: 600;
    transition: background-color 0.2s;
}

.get-membership-btn:hover {
    background: #005a87;
}
</style>

<script>
function pauseMembership() {
    if (confirm('Er du sikker på at du vil sette medlemskapet på pause?')) {
        // TODO: Implement pause functionality
        alert('Pause-funksjonalitet kommer snart!');
    }
}

function cancelMembership() {
    if (confirm('Er du sikker på at du vil si opp medlemskapet? Dette kan ikke angres.')) {
        // TODO: Implement cancellation functionality
        alert('Oppsigelse-funksjonalitet kommer snart!');
    }
}
</script>`

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
	}

	t, err := template.New("membership").Funcs(tmplFuncs).Parse(tmpl)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := t.Execute(w, data); err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
	}
}
