package handlers

import (
    "html/template"
    "kjernekraft/models"
    "net/http"
    "os"
    "path/filepath"
    "time"
)

var OsloLoc *time.Location

import (
	"html/template"
	"kjernekraft/models"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// parseTemplateWithEventCard creates a template with the event card partial
func parseTemplateWithEventCard(mainTemplate string) (*template.Template, error) {
	// Get the current working directory
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Parse the main template
	t := template.New("main").Funcs(template.FuncMap{
		"sub": func(a, b int) int {
			return a - b
		},
	})

	// Parse the event card template
	eventCardPath := filepath.Join(wd, "handlers", "templates", "event_card.html")
	t, err = t.ParseFiles(eventCardPath)
	if err != nil {
		return nil, err
	}

	// Parse the main template
	t, err = t.Parse(mainTemplate)
	if err != nil {
		return nil, err
	}

	return t, nil
}

// ElevDashboardHandler serves the Elev dashboard home page
func ElevDashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Get today's events
	allTodaysEvents, err := DB.GetTodaysEvents()
	if err != nil {
		http.Error(w, "Could not fetch today's events", http.StatusInternalServerError)
		return
	}

	// Filter out events that have already started
    now := time.Now().In(OsloLoc)
	var upcomingEvents []models.Event
	for _, event := range allTodaysEvents {
		if event.StartTime.After(now) {
			upcomingEvents = append(upcomingEvents, event)
		}
	}

	data := struct {
		TodaysEvents []models.Event
	}{
		TodaysEvents: upcomingEvents,
	}

	tmpl := `<!DOCTYPE html>
<html lang="no">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Kjernekraft - Elev Dashboard</title>
    <link rel="stylesheet" href="/static/css/event-card.css">
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
            max-width: 1200px;
            margin: 0 auto;
            padding: 2rem;
            overflow-x: hidden;
        }
        .page-title {
            font-size: 2rem;
            margin-bottom: 2rem;
            color: #333;
        }
        .modules-grid {
            display: grid;
            grid-template-columns: 1fr;
            gap: 2rem;
        }
        @media (min-width: 768px) {
            .modules-grid {
                grid-template-columns: 2fr 1fr;
            }
        }
        @media (min-width: 992px) {
            .modules-grid {
                grid-template-columns: 2fr 1fr 1fr 1fr;
            }
        }
        .module {
            background: white;
            border-radius: 8px;
            padding: 1.5rem;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            min-width: 0; /* Allow grid items to shrink */
        }
        .module-title {
            font-size: 1.25rem;
            margin-bottom: 1rem;
            color: #333;
            font-weight: 600;
        }
        
        @keyframes fadeIn {
            from {
                opacity: 0;
                transform: translateY(-10px);
            }
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }
        .enrolled-classes {
            display: flex;
            flex-direction: column;
            gap: 1.5rem;
        }
        .enrolled-section {
            background: #f8f9fa;
            border-radius: 6px;
            padding: 1rem;
        }
        .enrolled-subtitle {
            font-size: 1rem;
            font-weight: 600;
            color: #333;
            margin-bottom: 1rem;
            padding-bottom: 0.5rem;
            border-bottom: 1px solid #e0e0e0;
        }
        .no-classes {
            text-align: center;
            color: #666;
            font-style: italic;
            padding: 2rem;
        }
        .activity-placeholder {
            text-align: center;
            color: #666;
            font-style: italic;
            padding: 2rem;
        }
        
        /* Responsive styles */
        @media (max-width: 767px) {
            .nav-list {
                flex-direction: column;
            }
            .nav-item {
                border-right: none;
                border-bottom: 1px solid #e0e0e0;
            }
            .nav-item:last-child {
                border-bottom: none;
            }
            .main-content {
                padding: 1rem;
            }
        }
    </style>
</head>
<body>
    <header class="header">
        <h1>Kjernekraft - Elev Dashboard</h1>
    </header>
    
    <nav class="nav">
        <ul class="nav-list">
            <li class="nav-item">
                <a href="/elev/hjem" class="nav-link active">Hjem</a>
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
        <h1 class="page-title">Hjem</h1>
        
        <div class="modules-grid">
            <div class="module">
                <h2 class="module-title">I dag</h2>
                {{if .TodaysEvents}}
                    <div class="class-stack">
                        {{range .TodaysEvents}}
                        {{template "event_card" .}}
                        {{end}}
                    </div>
                {{else}}
                    <div class="no-classes">Ingen klasser i dag</div>
                {{end}}
            </div>
            
            <div class="module">
                <h2 class="module-title">Påmeldte timer</h2>
                <div class="enrolled-classes">
                    <div class="enrolled-section">
                        <h3 class="enrolled-subtitle">Neste</h3>
                        <div id="next-class-container">
                            <div class="loading">Laster neste time...</div>
                        </div>
                    </div>
                    <div class="enrolled-section">
                        <h3 class="enrolled-subtitle">Alle</h3>
                        <div id="all-classes-container">
                            <div class="loading">Laster påmeldte timer...</div>
                        </div>
                    </div>
                </div>
            </div>
            
            <div class="module">
                <h2 class="module-title">Ditt medlemskap</h2>
                <div id="membership-container">
                    <div class="loading">Laster medlemskap...</div>
                </div>
            </div>
            
            <div class="module">
                <h2 class="module-title">Dine klippekort</h2>
                <div id="klippekort-container">
                    <div class="loading">Laster klippekort...</div>
                </div>
            </div>
            
            <div class="module">
                <h2 class="module-title">Aktivitet</h2>
                <div class="activity-placeholder">
                    Aktivitetsporing kommer snart...
                </div>
            </div>
        </div>
    </main>

    <script>
        function toggleEventCard(element) {
            const stack = element.closest('.class-stack');
            const allCards = stack.querySelectorAll('.event-card');
            
            // Check if this card is already expanded
            const isExpanded = element.classList.contains('expanded');
            
            // Close all other expanded cards first
            allCards.forEach(card => {
                if (card !== element) {
                    card.classList.remove('expanded');
                }
            });
            
            // Toggle the clicked card and stack state
            if (isExpanded) {
                element.classList.remove('expanded');
                stack.classList.remove('has-expanded');
            } else {
                element.classList.add('expanded');
                stack.classList.add('has-expanded');
            }
        }
        
        function signupForClass(classId) {
            // TODO: Implement class signup functionality
            alert('Påmelding for klasse ' + classId + ' - kommer snart!');
        }

        // Load dashboard components
        async function loadMembership() {
            try {
                const response = await fetch('/api/user/membership?user_id=1');
                if (response.ok) {
                    const html = await response.text();
                    document.getElementById('membership-container').innerHTML = html;
                }
            } catch (error) {
                console.error('Error loading membership:', error);
                document.getElementById('membership-container').innerHTML = '<div class="error">Kunne ikke laste medlemskap</div>';
            }
        }

        async function loadKlippekort() {
            try {
                const response = await fetch('/api/user/klippekort?user_id=1');
                if (response.ok) {
                    const html = await response.text();
                    document.getElementById('klippekort-container').innerHTML = html;
                }
            } catch (error) {
                console.error('Error loading klippekort:', error);
                document.getElementById('klippekort-container').innerHTML = '<div class="error">Kunne ikke laste klippekort</div>';
            }
        }

        // Load components when page loads
        document.addEventListener('DOMContentLoaded', function() {
            loadMembership();
            loadKlippekort();
        });
    </script>
</body>
</html>`

	t, err := parseTemplateWithEventCard(tmpl)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, data)
}
