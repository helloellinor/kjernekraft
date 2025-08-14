package handlers

import (
	"html/template"
	"kjernekraft/models"
	"net/http"
	"strconv"
)

// KlippekortPageHandler serves the klippekort two-step selection page
func KlippekortPageHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title":       "Klippekort",
		"CurrentPage": "klippekort",
	}

	// Use the new template system
	tm := GetTemplateManager()
	if tmpl, exists := tm.GetTemplate("pages/klippekort"); exists {
		w.Header().Set("Content-Type", "text/html")
		if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
			http.Error(w, "Template execution error", http.StatusInternalServerError)
		}
		return
	}

	// If template doesn't exist, return error
	http.Error(w, "Template not found", http.StatusInternalServerError)
}

// MembershipSelectorHandler serves the interactive membership selector page
func MembershipSelectorHandler(w http.ResponseWriter, r *http.Request) {
	// For now, use a test user. In a real app, this would come from session/auth
	userID := int64(1)
	
	// Check if user has a membership
	membership, err := DB.GetUserMembership(userID)
	hasCurrentMembership := membership != nil && err == nil
	
	// Check if user has ever had a membership (for hiding offers)
	// For now, we'll just use the current membership check
	hasHadMembership := hasCurrentMembership
	
	// Determine page title and show special offer
	pageTitle := "Finn ditt perfekte medlemskap"
	showSpecialOffer := true
	
	if hasCurrentMembership {
		pageTitle = "Bytt medlemskapet mitt"
	}
	
	if hasHadMembership {
		showSpecialOffer = false
	}
	
	data := map[string]interface{}{
		"Title":                "Medlemskap",
		"CurrentPage":          "medlemskap",
		"PageTitle":            pageTitle,
		"HasCurrentMembership": hasCurrentMembership,
		"HasHadMembership":     hasHadMembership,
		"ShowSpecialOffer":     showSpecialOffer,
		"UserMembership":       membership,
	}

	// Use the new template system
	tm := GetTemplateManager()
	if tmpl, exists := tm.GetTemplate("pages/membership"); exists {
		w.Header().Set("Content-Type", "text/html")
		if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
			http.Error(w, "Template execution error", http.StatusInternalServerError)
		}
		return
	}

	// If template doesn't exist, return error
	http.Error(w, "Template not found", http.StatusInternalServerError)
}

// MembershipRecommendationsHandler provides endpoint for membership filtering
func MembershipRecommendationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	isStudentSenior := r.FormValue("is_student_senior") == "true"
	commitment := r.FormValue("commitment")
	startTime := r.FormValue("start_time")

	// Get all memberships
	allMemberships, err := DB.GetAllMemberships()
	if err != nil {
		http.Error(w, "Could not fetch memberships", http.StatusInternalServerError)
		return
	}

	// Filter memberships based on criteria
	var recommendations []models.Membership
	for _, membership := range allMemberships {
		// Check student/senior eligibility
		if isStudentSenior != membership.IsStudentSenior {
			continue
		}

		// Check commitment preferences
		if commitment == "trial" {
			// Show trial options (2-week trial, monthly pass)
			if membership.ID == 7 || membership.ID == 8 {
				recommendations = append(recommendations, membership)
			}
		} else if commitment != "" {
			commitmentMonths, _ := strconv.Atoi(commitment)
			if membership.CommitmentMonths == commitmentMonths {
				recommendations = append(recommendations, membership)
			}
		}

		// Special handling for autumn offer
		if startTime == "august" && membership.IsSpecialOffer {
			recommendations = append(recommendations, membership)
		}
	}

	// If no specific matches, show some default options
	if len(recommendations) == 0 && commitment != "" {
		for _, membership := range allMemberships {
			if membership.IsStudentSenior == isStudentSenior && !membership.IsSpecialOffer {
				recommendations = append(recommendations, membership)
			}
		}
	}

	// Check if this is an HTMX request
	isHTMX := r.Header.Get("HX-Request") != ""
	
	if isHTMX {
		// Return HTML fragment for HTMX
		data := struct {
			Recommendations []models.Membership
			ShowAutumnOffer bool
		}{
			Recommendations: recommendations,
			ShowAutumnOffer: startTime == "august",
		}

		tmpl := `{{if .Recommendations}}
<div style="background: white; border-radius: 12px; padding: 1.5rem; box-shadow: 0 4px 12px rgba(0,0,0,0.1);">
    <h3 style="margin-bottom: 1.5rem; color: #333; font-size: 1.25rem;">Våre anbefalinger for deg:</h3>
    
    {{if .ShowAutumnOffer}}
    <div style="background: linear-gradient(135deg, #ff6b35, #f7931e); color: white; padding: 1rem; border-radius: 8px; margin-bottom: 1.5rem; text-align: center;">
        <strong>🍂 Spesielt Høsttilbud!</strong><br>
        Få 12-måneders pris med kun 4 måneders binding
    </div>
    {{end}}
    
    <div style="display: grid; gap: 1rem;">
        {{range .Recommendations}}
        <div style="border: 2px solid {{if .IsSpecialOffer}}#ff6b35{{else}}#e0e0e0{{end}}; border-radius: 8px; padding: 1.5rem; {{if .IsSpecialOffer}}background-color: #fff5f0;{{end}}">
            {{if .IsSpecialOffer}}
            <div style="background: #ff6b35; color: white; padding: 0.25rem 0.75rem; border-radius: 12px; font-size: 0.8rem; font-weight: 600; display: inline-block; margin-bottom: 0.5rem;">
                Spesialtilbud
            </div>
            {{end}}
            
            <h4 style="font-size: 1.1rem; margin-bottom: 0.5rem; color: #333;">{{.Name}}</h4>
            <p style="color: #666; margin-bottom: 1rem; font-size: 0.9rem;">{{.Description}}</p>
            
            <div style="display: flex; justify-content: space-between; align-items: center;">
                <div>
                    <div style="font-size: 1.5rem; font-weight: 700; color: #007cba;">{{printf "%.0f" (divf .Price 100)}} kr/mnd</div>
                    {{if gt .CommitmentMonths 0}}
                    <div style="font-size: 0.8rem; color: #666;">{{.CommitmentMonths}} måneders binding</div>
                    {{else}}
                    <div style="font-size: 0.8rem; color: #666;">Ingen binding</div>
                    {{end}}
                </div>
                <button style="background: #007cba; color: white; border: none; padding: 0.75rem 1.5rem; border-radius: 6px; cursor: pointer; font-weight: 600;">
                    Velg dette
                </button>
            </div>
        </div>
        {{end}}
    </div>
</div>
{{else}}
<div style="background: white; border-radius: 12px; padding: 2rem; box-shadow: 0 4px 12px rgba(0,0,0,0.1); text-align: center; color: #666;">
    Velg flere alternativer for å se anbefalinger
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
		}

		t, err := template.New("recommendations").Funcs(tmplFuncs).Parse(tmpl)
		if err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		if err := t.Execute(w, data); err != nil {
			http.Error(w, "Template execution error", http.StatusInternalServerError)
		}
	} else {
		// Return full page for regular form submission
		data := struct {
			Recommendations []models.Membership
			ShowAutumnOffer bool
			IsStudentSenior bool
			Commitment      string
			StartTime       string
		}{
			Recommendations: recommendations,
			ShowAutumnOffer: startTime == "august",
			IsStudentSenior: isStudentSenior,
			Commitment:      commitment,
			StartTime:       startTime,
		}

		tmpl := `<!DOCTYPE html>
<html lang="no">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Medlemskapsanbefalinger - Kjernekraft</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f5f5f5; color: #333; }
        .header { background-color: #007cba; color: white; padding: 1rem 2rem; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header h1 { font-size: 1.5rem; font-weight: 600; }
        .nav { background-color: white; border-bottom: 1px solid #e0e0e0; padding: 0; }
        .nav-list { display: flex; list-style: none; max-width: 1200px; margin: 0 auto; }
        .nav-item { border-right: 1px solid #e0e0e0; }
        .nav-item:last-child { border-right: none; }
        .nav-link { display: block; padding: 1rem 2rem; text-decoration: none; color: #333; font-weight: 500; transition: background-color 0.2s; }
        .nav-link:hover, .nav-link.active { background-color: #f0f8ff; color: #007cba; }
        .main { max-width: 800px; margin: 0 auto; padding: 2rem; }
        .page-title { font-size: 2rem; margin-bottom: 2rem; color: #333; text-align: center; border-bottom: 2px solid #007cba; padding-bottom: 0.5rem; }
        .recommendations { display: grid; gap: 1.5rem; }
        .recommendation-card { background: white; border-radius: 12px; padding: 2rem; box-shadow: 0 4px 12px rgba(0,0,0,0.1); border: 2px solid #e0e0e0; }
        .recommendation-card.special { border-color: #ff6b35; background: linear-gradient(135deg, #fff5f0, #ffffff); }
        .special-badge { background: #ff6b35; color: white; padding: 0.25rem 0.75rem; border-radius: 12px; font-size: 0.8rem; font-weight: 600; display: inline-block; margin-bottom: 1rem; }
        .card-title { font-size: 1.5rem; font-weight: 600; color: #333; margin-bottom: 0.5rem; }
        .card-description { color: #666; margin-bottom: 1.5rem; }
        .price { font-size: 2rem; font-weight: 700; color: #007cba; margin-bottom: 0.5rem; }
        .commitment { color: #666; margin-bottom: 1.5rem; }
        .back-link { display: inline-block; margin-top: 2rem; padding: 0.75rem 1.5rem; background: #6c757d; color: white; text-decoration: none; border-radius: 6px; }
        .back-link:hover { background: #5a6268; }
    </style>
</head>
<body>
    <header class="header"><h1>Kjernekraft - Medlemskapsanbefalinger</h1></header>
    <nav class="nav">
        <ul class="nav-list">
            <li class="nav-item"><a href="/elev/hjem" class="nav-link">Hjem</a></li>
            <li class="nav-item"><a href="/elev/timeplan" class="nav-link">Timeplan</a></li>
            <li class="nav-item"><a href="/elev/klippekort" class="nav-link">Klippekort</a></li>
            <li class="nav-item"><a href="/elev/medlemskap" class="nav-link active">Medlemskap</a></li>
            <li class="nav-item"><a href="/elev/min-profil" class="nav-link">Min profil</a></li>
        </ul>
    </nav>
    
    <main class="main">
        <h1 class="page-title">Våre anbefalinger for deg</h1>
        
        {{if .ShowAutumnOffer}}
        <div style="background: linear-gradient(135deg, #ff6b35, #f7931e); color: white; padding: 1.5rem; border-radius: 12px; margin-bottom: 2rem; text-align: center;">
            <strong style="font-size: 1.2rem;">🍂 Spesielt Høsttilbud!</strong><br>
            Få 12-måneders pris med kun 4 måneders binding
        </div>
        {{end}}
        
        <div class="recommendations">
            {{range .Recommendations}}
            <div class="recommendation-card {{if .IsSpecialOffer}}special{{end}}">
                {{if .IsSpecialOffer}}
                <div class="special-badge">Spesialtilbud</div>
                {{end}}
                
                <h2 class="card-title">{{.Name}}</h2>
                <p class="card-description">{{.Description}}</p>
                
                <div class="price">{{printf "%.0f" (divf .Price 100)}} kr/mnd</div>
                {{if gt .CommitmentMonths 0}}
                <div class="commitment">{{.CommitmentMonths}} måneders binding</div>
                {{else}}
                <div class="commitment">Ingen binding</div>
                {{end}}
            </div>
            {{end}}
        </div>
        
        <a href="/medlemskap" class="back-link">← Tilbake til spørreskjema</a>
    </main>
</body>
</html>`

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

		t, err := template.New("membership-results").Funcs(tmplFuncs).Parse(tmpl)
		if err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		if err := t.Execute(w, data); err != nil {
			http.Error(w, "Template execution error", http.StatusInternalServerError)
		}
	}
}

// MinProfilHandler serves the user profile page
func MinProfilHandler(w http.ResponseWriter, r *http.Request) {
	// For now, use a test user. In a real app, this would come from session/auth
	user := struct {
		ID       int64
		Name     string
		Email    string
		JoinDate string
		Phone    string
	}{
		ID:       1,
		Name:     "Test Bruker",
		Email:    "test@example.com",
		JoinDate: "1. januar 2024",
		Phone:    "+47 123 45 678",
	}

	data := map[string]interface{}{
		"Title":       "Min profil",
		"CurrentPage": "profil",
		"ID":          user.ID,
		"Name":        user.Name,
		"Email":       user.Email,
		"JoinDate":    user.JoinDate,
		"Phone":       user.Phone,
	}

	// Use the new template system
	tm := GetTemplateManager()
	if tmpl, exists := tm.GetTemplate("pages/min-profil"); exists {
		w.Header().Set("Content-Type", "text/html")
		if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
			http.Error(w, "Template execution error", http.StatusInternalServerError)
		}
		return
	}

	// If template doesn't exist, return error
	http.Error(w, "Template not found", http.StatusInternalServerError)
}

// TestDataPageHandler serves the test data generation page
func TestDataPageHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html lang="no">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Testdata - Kjernekraft</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background-color: #f5f5f5;
            color: #333;
        }
        .header {
            background-color: #007cba;
            color: white;
            padding: 1rem 2rem;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .header h1 {
            font-size: 1.5rem;
            font-weight: 600;
        }
        .nav {
            background-color: white;
            border-bottom: 1px solid #e0e0e0;
            padding: 0;
        }
        .nav-list {
            display: flex;
            list-style: none;
            max-width: 1200px;
            margin: 0 auto;
        }
        .nav-item {
            border-right: 1px solid #e0e0e0;
        }
        .nav-item:last-child {
            border-right: none;
        }
        .nav-link {
            display: block;
            padding: 1rem 2rem;
            text-decoration: none;
            color: #333;
            font-weight: 500;
            transition: background-color 0.2s;
        }
        .nav-link:hover, .nav-link.active {
            background-color: #f0f8ff;
            color: #007cba;
        }
        .main-content {
            max-width: 800px;
            margin: 0 auto;
            padding: 2rem;
        }
        .page-title {
            font-size: 2rem;
            margin-bottom: 1rem;
            color: #333;
            border-bottom: 2px solid #007cba;
            padding-bottom: 0.5rem;
        }
        .dev-warning {
            background: #fff3cd;
            border: 1px solid #ffeaa7;
            color: #856404;
            padding: 1rem;
            border-radius: 8px;
            margin-bottom: 2rem;
        }
        .test-section {
            background: white;
            border-radius: 12px;
            padding: 2rem;
            box-shadow: 0 4px 12px rgba(0,0,0,0.1);
            margin-bottom: 2rem;
        }
        .section-title {
            font-size: 1.25rem;
            margin-bottom: 1rem;
            color: #333;
        }
        .section-description {
            color: #666;
            margin-bottom: 1.5rem;
            line-height: 1.6;
        }
        .test-btn {
            background: #6c757d;
            color: white;
            border: none;
            padding: 1rem 2rem;
            border-radius: 6px;
            font-weight: 600;
            cursor: pointer;
            transition: background-color 0.2s;
            margin-right: 1rem;
            margin-bottom: 0.5rem;
        }
        .test-btn:hover {
            background: #5a6268;
        }
        .test-btn:disabled {
            background: #adb5bd;
            cursor: not-allowed;
        }
        .test-btn.danger {
            background: #dc3545;
        }
        .test-btn.danger:hover {
            background: #c82333;
        }
        .result-area {
            margin-top: 1rem;
            padding: 1rem;
            background: #f8f9fa;
            border-radius: 6px;
            display: none;
        }
        .result-area.show {
            display: block;
        }
        .result-area.success {
            background: #d4edda;
            color: #155724;
            border: 1px solid #c3e6cb;
        }
        .result-area.error {
            background: #f8d7da;
            color: #721c24;
            border: 1px solid #f1aeb5;
        }
    </style>
</head>
<body>
    <header class="header">
        <h1>Kjernekraft - Testdata</h1>
    </header>
    
    <nav class="nav">
        <ul class="nav-list">
            <li class="nav-item">
                <a href="/elev/hjem" class="nav-link">Hjem</a>
            </li>
            <li class="nav-item">
                <a href="/elev/timeplan" class="nav-link">Timeplan</a>
            </li>
            <li class="nav-item">
                <a href="/elev/klippekort" class="nav-link">Klippekort</a>
            </li>
            <li class="nav-item">
                <a href="/elev/medlemskap" class="nav-link">Medlemskap</a>
            </li>
            <li class="nav-item">
                <a href="/elev/min-profil" class="nav-link">Min profil</a>
            </li>
        </ul>
    </nav>

    <main class="main-content">
        <h1 class="page-title">🧪 Testdata generering</h1>
        
        <div class="dev-warning">
            <strong>⚠️ Utviklingsverktøy</strong><br>
            Denne siden er kun tilgjengelig i utviklingsmiljø og vil generere testdata for demonstrasjon.
        </div>
        
        <div class="test-section">
            <h2 class="section-title">Kalenderdata</h2>
            <p class="section-description">
                Generer nye tilfeldige treningsklasser for denne og neste uke. Dette vil erstatte alle eksisterende kalenderoppføringer.
            </p>
            <button class="test-btn" onclick="shuffleEvents()">
                🗓️ Generer kalenderdata
            </button>
            <div id="events-result" class="result-area"></div>
        </div>
        
        <div class="test-section">
            <h2 class="section-title">Medlemskapsdata</h2>
            <p class="section-description">
                Generer nye tilfeldige medlemskapsnavn og priser. Dette vil oppdatere eksisterende medlemskapstyper med nye verdier.
            </p>
            <button class="test-btn" onclick="shuffleMemberships()">
                💳 Generer medlemskapsdata
            </button>
            <div id="memberships-result" class="result-area"></div>
        </div>
        
        <div class="test-section">
            <h2 class="section-title">Brukerdata</h2>
            <p class="section-description">
                Oppdater den innloggede brukerens klippekort med nye tilfeldige verdier. Dette endrer antall gjenværende klipp.
            </p>
            <button class="test-btn" onclick="shuffleUserKlippekort()">
                🎫 Generer brukerklippekort
            </button>
            <div id="user-result" class="result-area"></div>
        </div>
        
        <div class="test-section">
            <h2 class="section-title">Generer alt</h2>
            <p class="section-description">
                Generer alle testdata på en gang. Dette vil oppdatere kalenderen, medlemskap og brukerdata samtidig.
            </p>
            <button class="test-btn danger" onclick="shuffleAll()">
                🎲 Generer alle testdata
            </button>
            <div id="all-result" class="result-area"></div>
        </div>
    </main>

    <script>
        async function makeRequest(endpoint, btnElement, resultElement, successMessage) {
            btnElement.disabled = true;
            btnElement.textContent = '🔄 Genererer...';
            resultElement.className = 'result-area';
            resultElement.style.display = 'none';
            
            try {
                const response = await fetch(endpoint, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                });
                
                if (response.ok) {
                    const data = await response.json();
                    resultElement.className = 'result-area success show';
                    resultElement.textContent = successMessage + (data.message ? ': ' + data.message : '');
                } else {
                    throw new Error('Request failed');
                }
            } catch (error) {
                console.error('Error:', error);
                resultElement.className = 'result-area error show';
                resultElement.textContent = 'Feil ved generering av testdata';
            } finally {
                btnElement.disabled = false;
                btnElement.textContent = btnElement.textContent.replace('🔄 Genererer...', btnElement.getAttribute('data-original-text'));
            }
        }

        function shuffleEvents() {
            const btn = event.target;
            btn.setAttribute('data-original-text', '🗓️ Generer kalenderdata');
            const result = document.getElementById('events-result');
            makeRequest('/api/shuffle-test-data', btn, result, 'Kalenderdata generert');
        }

        function shuffleMemberships() {
            const btn = event.target;
            btn.setAttribute('data-original-text', '💳 Generer medlemskapsdata');
            const result = document.getElementById('memberships-result');
            makeRequest('/api/shuffle-memberships', btn, result, 'Medlemskapsdata generert');
        }

        function shuffleUserKlippekort() {
            const btn = event.target;
            btn.setAttribute('data-original-text', '🎫 Generer brukerklippekort');
            const result = document.getElementById('user-result');
            makeRequest('/api/shuffle-user-klippekort', btn, result, 'Brukerklippekort generert');
        }

        function shuffleAll() {
            const btn = event.target;
            btn.setAttribute('data-original-text', '🎲 Generer alle testdata');
            const result = document.getElementById('all-result');
            makeRequest('/api/shuffle-all-test-data', btn, result, 'Alle testdata generert');
        }
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(tmpl))
}