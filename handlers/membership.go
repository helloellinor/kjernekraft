package handlers

import (
	"html/template"
	"kjernekraft/models"
	"log"
	"net/http"
	"strconv"
)

// KlippekortPageHandler serves the klippekort pricing page
func KlippekortPageHandler(w http.ResponseWriter, r *http.Request) {
	packages, err := DB.GetAllKlippekortPackages()
	if err != nil {
		http.Error(w, "Could not fetch klippekort packages", http.StatusInternalServerError)
		log.Printf("Error fetching klippekort packages: %v", err)
		return
	}

	// Group packages by category
	packagesByCategory := make(map[string][]models.KlippekortPackage)
	for _, pkg := range packages {
		packagesByCategory[pkg.Category] = append(packagesByCategory[pkg.Category], pkg)
	}

	data := struct {
		PackagesByCategory map[string][]models.KlippekortPackage
	}{
		PackagesByCategory: packagesByCategory,
	}

	tmpl := `<!DOCTYPE html>
<html lang="no">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Klippekort - Kjernekraft</title>
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
            background-color: #005a87;
            padding: 0.5rem 0;
        }
        .nav-list {
            list-style: none;
            display: flex;
            gap: 2rem;
            max-width: 1200px;
            margin: 0 auto;
            padding: 0 2rem;
        }
        .nav-item a {
            color: white;
            text-decoration: none;
            padding: 0.5rem 1rem;
            border-radius: 4px;
            transition: background-color 0.2s;
        }
        .nav-item a:hover {
            background-color: rgba(255,255,255,0.1);
        }
        .nav-item a.active {
            background-color: rgba(255,255,255,0.2);
        }
        .main {
            max-width: 1200px;
            margin: 0 auto;
            padding: 2rem;
        }
        .page-title {
            font-size: 2rem;
            margin-bottom: 1rem;
            color: #333;
        }
        .page-description {
            font-size: 1.1rem;
            color: #666;
            margin-bottom: 3rem;
            text-align: center;
            max-width: 600px;
            margin-left: auto;
            margin-right: auto;
            margin-bottom: 3rem;
        }
        .category-section {
            margin-bottom: 4rem;
        }
        .category-title {
            font-size: 1.5rem;
            margin-bottom: 1.5rem;
            color: #333;
            text-align: center;
        }
        .packages-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
            gap: 1.5rem;
            margin-bottom: 2rem;
        }
        .package-card {
            background: white;
            border-radius: 12px;
            padding: 1.5rem;
            box-shadow: 0 4px 12px rgba(0,0,0,0.1);
            transition: transform 0.2s, box-shadow 0.2s;
            position: relative;
            border: 2px solid transparent;
        }
        .package-card:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 25px rgba(0,0,0,0.15);
        }
        .package-card.popular {
            border-color: #ff6b35;
        }
        .popular-badge {
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
        .package-name {
            font-size: 1.25rem;
            font-weight: 600;
            margin-bottom: 0.5rem;
            color: #333;
        }
        .package-description {
            color: #666;
            margin-bottom: 1rem;
            font-size: 0.9rem;
        }
        .package-details {
            margin-bottom: 1.5rem;
        }
        .package-price {
            font-size: 1.8rem;
            font-weight: 700;
            color: #007cba;
            margin-bottom: 0.5rem;
        }
        .package-count {
            font-size: 1rem;
            color: #666;
            margin-bottom: 0.5rem;
        }
        .price-per-session {
            font-size: 0.9rem;
            color: #333;
            font-weight: 500;
        }
        .value-graph {
            margin-top: 2rem;
            padding: 1rem;
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        .graph-title {
            text-align: center;
            margin-bottom: 1rem;
            font-weight: 600;
            color: #333;
        }
        .graph-bars {
            display: flex;
            align-items: end;
            justify-content: space-around;
            height: 120px;
            gap: 0.5rem;
        }
        .graph-bar {
            background: linear-gradient(to top, #007cba, #4a9fd1);
            border-radius: 4px 4px 0 0;
            min-width: 40px;
            display: flex;
            flex-direction: column;
            align-items: center;
            position: relative;
        }
        .bar-label {
            font-size: 0.7rem;
            color: #666;
            margin-top: 0.5rem;
            text-align: center;
        }
        .bar-value {
            position: absolute;
            top: -25px;
            font-size: 0.7rem;
            color: #333;
            font-weight: 600;
        }
        @media (max-width: 768px) {
            .packages-grid {
                grid-template-columns: 1fr;
            }
            .nav-list {
                flex-direction: column;
                gap: 0.5rem;
            }
        }
    </style>
</head>
<body>
    <header class="header">
        <h1>Kjernekraft - Klippekort</h1>
    </header>
    
    <nav class="nav">
        <ul class="nav-list">
            <li class="nav-item">
                <a href="/elev/hjem">Hjem</a>
            </li>
            <li class="nav-item">
                <a href="/elev/timeplan">Timeplan</a>
            </li>
            <li class="nav-item">
                <a href="/klippekort" class="active">Klippekort</a>
            </li>
            <li class="nav-item">
                <a href="/medlemskap">Medlemskap</a>
            </li>
        </ul>
    </nav>
    
    <main class="main">
        <h1 class="page-title">Klippekort</h1>
        <p class="page-description">
            Kjøp klippekort for personlig trening og reformer. Jo flere klipp du kjøper, desto mindre blir prisen per økt.
        </p>
        
        {{range $category, $packages := .PackagesByCategory}}
        <section class="category-section">
            <h2 class="category-title">{{$category}}</h2>
            
            <div class="packages-grid">
                {{range $packages}}
                <div class="package-card {{if .IsPopular}}popular{{end}}">
                    {{if .IsPopular}}
                    <div class="popular-badge">Mest populær</div>
                    {{end}}
                    
                    <h3 class="package-name">{{.Name}}</h3>
                    <p class="package-description">{{.Description}}</p>
                    
                    <div class="package-details">
                        <div class="package-price">{{printf "%.0f" (divf .Price 100)}} kr</div>
                        <div class="package-count">{{.KlippCount}} klipp</div>
                        <div class="price-per-session">{{printf "%.0f" (divf .PricePerSession 100)}} kr per økt</div>
                    </div>
                </div>
                {{end}}
            </div>
            
            <div class="value-graph">
                <div class="graph-title">Pris per økt - {{$category}}</div>
                <div class="graph-bars">
                    {{range $packages}}
                    <div class="graph-bar" style="height: {{printf "%.0f" (multf (divf (subf 80000 .PricePerSession) 80000) 100)}}%;">
                        <div class="bar-value">{{printf "%.0f" (divf .PricePerSession 100)}} kr</div>
                        <div class="bar-label">{{.KlippCount}} klipp</div>
                    </div>
                    {{end}}
                </div>
            </div>
        </section>
        {{end}}
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
		"subf": func(a, b interface{}) float64 {
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
			
			return aFloat - bFloat
		},
		"multf": func(a, b float64) float64 {
			return a * b
		},
	}

	t, err := template.New("klippekort").Funcs(tmplFuncs).Parse(tmpl)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Template parse error: %v", err)
		return
	}

	if err := t.Execute(w, data); err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		log.Printf("Template execution error: %v", err)
	}
}

// MembershipSelectorHandler serves the interactive membership selector page
func MembershipSelectorHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html lang="no">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Finn ditt medlemskap - Kjernekraft</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
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
            background-color: #005a87;
            padding: 0.5rem 0;
        }
        .nav-list {
            list-style: none;
            display: flex;
            gap: 2rem;
            max-width: 1200px;
            margin: 0 auto;
            padding: 0 2rem;
        }
        .nav-item a {
            color: white;
            text-decoration: none;
            padding: 0.5rem 1rem;
            border-radius: 4px;
            transition: background-color 0.2s;
        }
        .nav-item a:hover {
            background-color: rgba(255,255,255,0.1);
        }
        .nav-item a.active {
            background-color: rgba(255,255,255,0.2);
        }
        .main {
            max-width: 1200px;
            margin: 0 auto;
            padding: 2rem;
        }
        .page-title {
            font-size: 2rem;
            margin-bottom: 1rem;
            color: #333;
            text-align: center;
        }
        .page-description {
            font-size: 1.1rem;
            color: #666;
            margin-bottom: 3rem;
            text-align: center;
            max-width: 600px;
            margin-left: auto;
            margin-right: auto;
            margin-bottom: 3rem;
        }
        .selector-container {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 3rem;
            align-items: start;
        }
        .question-form {
            background: white;
            border-radius: 12px;
            padding: 2rem;
            box-shadow: 0 4px 12px rgba(0,0,0,0.1);
        }
        .form-title {
            font-size: 1.25rem;
            margin-bottom: 1.5rem;
            color: #333;
        }
        .question-group {
            margin-bottom: 1.5rem;
        }
        .question-label {
            display: block;
            margin-bottom: 0.5rem;
            font-weight: 600;
            color: #333;
        }
        .question-options {
            display: grid;
            gap: 0.5rem;
        }
        .option-label {
            display: flex;
            align-items: center;
            padding: 0.75rem;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            cursor: pointer;
            transition: all 0.2s;
        }
        .option-label:hover {
            border-color: #007cba;
            background-color: #f8f9fa;
        }
        .option-label input[type="radio"] {
            margin-right: 0.75rem;
        }
        .option-label input[type="radio"]:checked + span {
            font-weight: 600;
        }
        .option-label:has(input[type="radio"]:checked) {
            border-color: #007cba;
            background-color: #e8f4fd;
        }
        .results-container {
            min-height: 400px;
        }
        .results-placeholder {
            background: white;
            border-radius: 12px;
            padding: 3rem 2rem;
            box-shadow: 0 4px 12px rgba(0,0,0,0.1);
            text-align: center;
            color: #666;
        }
        .loading {
            color: #007cba;
        }
        @media (max-width: 768px) {
            .selector-container {
                grid-template-columns: 1fr;
                gap: 2rem;
            }
            .nav-list {
                flex-direction: column;
                gap: 0.5rem;
            }
        }
    </style>
</head>
<body>
    <header class="header">
        <h1>Kjernekraft - Finn ditt medlemskap</h1>
    </header>
    
    <nav class="nav">
        <ul class="nav-list">
            <li class="nav-item">
                <a href="/elev/hjem">Hjem</a>
            </li>
            <li class="nav-item">
                <a href="/elev/timeplan">Timeplan</a>
            </li>
            <li class="nav-item">
                <a href="/klippekort">Klippekort</a>
            </li>
            <li class="nav-item">
                <a href="/medlemskap" class="active">Medlemskap</a>
            </li>
        </ul>
    </nav>
    
    <main class="main">
        <h1 class="page-title">Finn ditt perfekte medlemskap</h1>
        <p class="page-description">
            Svar på noen enkle spørsmål så viser vi deg medlemskapet som passer best for deg.
        </p>
        
        <div class="selector-container">
            <form class="question-form" 
                  hx-post="/api/membership-recommendations" 
                  hx-trigger="change" 
                  hx-target="#membership-results"
                  hx-indicator="#loading">
                
                <h2 class="form-title">Fortell oss om deg</h2>
                
                <div class="question-group">
                    <label class="question-label">Er du student eller senior?</label>
                    <div class="question-options">
                        <label class="option-label">
                            <input type="radio" name="is_student_senior" value="true">
                            <span>Ja, jeg er student eller 67+ år</span>
                        </label>
                        <label class="option-label">
                            <input type="radio" name="is_student_senior" value="false">
                            <span>Nei</span>
                        </label>
                    </div>
                </div>
                
                <div class="question-group">
                    <label class="question-label">Hvor lenge ønsker du å binde deg?</label>
                    <div class="question-options">
                        <label class="option-label">
                            <input type="radio" name="commitment" value="12">
                            <span>12 måneder (best pris)</span>
                        </label>
                        <label class="option-label">
                            <input type="radio" name="commitment" value="6">
                            <span>6 måneder</span>
                        </label>
                        <label class="option-label">
                            <input type="radio" name="commitment" value="0">
                            <span>Ingen binding (mest fleksibelt)</span>
                        </label>
                        <label class="option-label">
                            <input type="radio" name="commitment" value="trial">
                            <span>Jeg vil bare prøve</span>
                        </label>
                    </div>
                </div>
                
                <div class="question-group">
                    <label class="question-label">Når vil du starte?</label>
                    <div class="question-options">
                        <label class="option-label">
                            <input type="radio" name="start_time" value="august">
                            <span>I august (Høsttilbud!)</span>
                        </label>
                        <label class="option-label">
                            <input type="radio" name="start_time" value="now">
                            <span>Så snart som mulig</span>
                        </label>
                    </div>
                </div>
            </form>
            
            <div class="results-container">
                <div id="membership-results" class="results-placeholder">
                    <p>Velg alternativene til venstre for å se våre anbefalinger</p>
                    <div id="loading" class="loading" style="display:none;">Laster anbefalinger...</div>
                </div>
            </div>
        </div>
    </main>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(tmpl))
}

// MembershipRecommendationsHandler provides HTMX endpoint for membership filtering
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

	// Generate HTML response
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
}