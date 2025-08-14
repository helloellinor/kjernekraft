package handlers

import (
	"html/template"
	"kjernekraft/models"
	"net/http"
	"time"
)

// ElevDashboardHandler serves the Elev dashboard home page
func ElevDashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Get today's events
	allTodaysEvents, err := DB.GetTodaysEvents()
	if err != nil {
		http.Error(w, "Could not fetch today's events", http.StatusInternalServerError)
		return
	}

	// Filter out events that have already started
	now := time.Now()
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
        .class-stack {
            display: flex;
            flex-direction: column;
            gap: 0.5rem;
        }
        .class-item {
            background: white;
            border: 1px solid #e0e0e0;
            border-radius: 6px;
            overflow: hidden;
            transition: all 0.3s ease;
            cursor: pointer;
        }
        .class-item.expanded {
            box-shadow: 0 4px 12px rgba(0,0,0,0.15);
            transform: translateY(-2px);
        }
        .class-header {
            padding: 1rem;
            display: flex;
            justify-content: space-between;
            align-items: center;
            border-left: 4px solid #007cba;
        }
        .class-title-line {
            display: flex;
            align-items: center;
            gap: 1rem;
        }
        .class-time {
            font-size: 0.875rem;
            color: #666;
            font-weight: 500;
        }
        .class-title {
            font-size: 1rem;
            font-weight: 600;
            color: #333;
        }
        .class-expand-icon {
            font-size: 1.2rem;
            color: #007cba;
            transition: transform 0.3s ease;
        }
        .class-item.expanded .class-expand-icon {
            transform: rotate(180deg);
        }
        .class-details {
            padding: 0 1rem 1rem 1rem;
            display: none;
            border-top: 1px solid #f0f0f0;
            background: #fafafa;
        }
        .class-item.expanded .class-details {
            display: block;
            animation: slideDown 0.3s ease;
        }
        .class-teacher {
            font-size: 0.875rem;
            color: #666;
            margin-bottom: 0.5rem;
        }
        .class-spaces {
            font-size: 0.875rem;
            margin-bottom: 1rem;
            color: #333;
        }
        .signup-btn {
            width: 100%;
            padding: 0.75rem;
            background-color: #007cba;
            color: white;
            border: none;
            border-radius: 4px;
            font-size: 0.875rem;
            cursor: pointer;
            transition: background-color 0.2s;
        }
        .signup-btn:hover {
            background-color: #005a87;
        }
        .signup-btn.waitlist {
            background-color: #ff6b35;
        }
        .signup-btn.waitlist:hover {
            background-color: #e55a2b;
        }
        @keyframes slideDown {
            from {
                opacity: 0;
                max-height: 0;
            }
            to {
                opacity: 1;
                max-height: 200px;
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
        /* Color scheme for different class types */
        .yoga { background-color: #8e44ad; }
        .pilates { background-color: #27ae60; }
        .strength { background-color: #e74c3c; }
        .cardio { background-color: #f39c12; }
        .flexibility { background-color: #3498db; }
        .loading {
            text-align: center;
            padding: 2rem;
            color: #666;
            font-style: italic;
        }
        .error {
            text-align: center;
            padding: 2rem;
            color: #dc3545;
            background-color: #f8d7da;
            border-radius: 6px;
            border: 1px solid #f1aeb5;
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
                        <div class="class-item" onclick="toggleClass(this)">
                            <div class="class-header">
                                <div class="class-title-line">
                                    <span class="class-time">{{.StartTime.Format "15:04"}}</span>
                                    <span class="class-title">{{.Title}}</span>
                                </div>
                                <span class="class-expand-icon">▼</span>
                            </div>
                            <div class="class-details">
                                <div class="class-teacher">{{.TeacherName}}</div>
                                <div class="class-spaces">
                                    {{if lt .CurrentEnrolment .Capacity}}
                                        {{sub .Capacity .CurrentEnrolment}} plasser igjen
                                    {{else}}
                                        Venteliste
                                    {{end}}
                                </div>
                                <button class="signup-btn {{if ge .CurrentEnrolment .Capacity}}waitlist{{end}}" 
                                        onclick="signupForClass({{.ID}}); event.stopPropagation();">
                                    Meld på
                                </button>
                            </div>
                        </div>
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
        function toggleClass(element) {
            // Close other expanded classes
            const allItems = document.querySelectorAll('.class-item');
            allItems.forEach(item => {
                if (item !== element) {
                    item.classList.remove('expanded');
                }
            });
            
            // Toggle the clicked class
            element.classList.toggle('expanded');
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

	t, err := template.New("elev-dashboard").Funcs(template.FuncMap{
		"sub": func(a, b int) int {
			return a - b
		},
	}).Parse(tmpl)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, data)
}

// ElevTimeplanHandler serves the Elev timeplan (schedule) page  
func ElevTimeplanHandler(w http.ResponseWriter, r *http.Request) {
	// Get this week's events
	weekEvents, err := DB.GetThisWeeksEvents()
	if err != nil {
		http.Error(w, "Could not fetch week's events", http.StatusInternalServerError)
		return
	}

	// Group events by day
	eventsByDay := make(map[string][]models.Event)
	now := time.Now()
	
	// Calculate this week's dates (Monday to Sunday)
	weekdays := []string{"Mandag", "Tirsdag", "Onsdag", "Torsdag", "Fredag", "Lørdag", "Søndag"}
	weekDates := make([]time.Time, 7)
	
	// Find Monday of this week
	monday := now.AddDate(0, 0, -int(now.Weekday())+1)
	if now.Weekday() == time.Sunday {
		monday = monday.AddDate(0, 0, -7)
	}
	
	for i := 0; i < 7; i++ {
		weekDates[i] = monday.AddDate(0, 0, i)
		dateKey := weekDates[i].Format("2006-01-02")
		eventsByDay[dateKey] = []models.Event{}
	}

	// Group events by date
	for _, event := range weekEvents {
		dateKey := event.StartTime.Format("2006-01-02")
		if _, exists := eventsByDay[dateKey]; exists {
			eventsByDay[dateKey] = append(eventsByDay[dateKey], event)
		}
	}

	data := struct {
		WeekDays    []string
		WeekDates   []time.Time
		EventsByDay map[string][]models.Event
		Today       string
	}{
		WeekDays:    weekdays,
		WeekDates:   weekDates,
		EventsByDay: eventsByDay,
		Today:       now.Format("2006-01-02"),
	}

	tmpl := `<!DOCTYPE html>
<html lang="no">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Kjernekraft - Timeplan</title>
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
        }
        .page-title {
            font-size: 2rem;
            margin-bottom: 2rem;
            color: #333;
        }
        .module {
            background: white;
            border-radius: 8px;
            padding: 1.5rem;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        .module-title {
            font-size: 1.25rem;
            margin-bottom: 1.5rem;
            color: #333;
            font-weight: 600;
        }
        .week-grid {
            display: grid;
            grid-template-columns: repeat(7, 1fr);
            gap: 1rem;
        }
        .day-column {
            min-height: 300px;
        }
        .day-header {
            text-align: center;
            margin-bottom: 1rem;
            padding-bottom: 0.5rem;
            border-bottom: 2px solid #e0e0e0;
        }
        .day-name {
            font-weight: 600;
            color: #333;
            margin-bottom: 0.25rem;
        }
        .day-date {
            font-size: 0.875rem;
            color: #666;
        }
        .day-column.past {
            opacity: 0.5;
        }
        .day-column.past .day-content {
            background-color: #f0f0f0;
            border-radius: 4px;
            padding: 1rem;
            text-align: center;
            color: #999;
            font-style: italic;
        }
        .day-events {
            display: flex;
            flex-direction: column;
            gap: 0.5rem;
        }
        .event-card {
            background: white;
            border: 1px solid #e0e0e0;
            border-radius: 6px;
            padding: 0.75rem;
            transition: transform 0.2s, box-shadow 0.2s;
            border-left: 4px solid #007cba;
        }
        .event-card:hover {
            transform: translateY(-1px);
            box-shadow: 0 2px 8px rgba(0,0,0,0.15);
        }
        .event-time {
            font-size: 0.75rem;
            color: #666;
            font-weight: 500;
        }
        .event-title {
            font-size: 0.875rem;
            font-weight: 600;
            margin: 0.25rem 0;
            color: #333;
        }
        .event-teacher {
            font-size: 0.75rem;
            color: #666;
        }
        .event-spaces {
            font-size: 0.75rem;
            color: #333;
            margin-top: 0.25rem;
        }
        /* Color scheme for different class types */
        .event-card.yoga { border-left-color: #8e44ad; }
        .event-card.pilates { border-left-color: #27ae60; }
        .event-card.strength { border-left-color: #e74c3c; }
        .event-card.cardio { border-left-color: #f39c12; }
        .event-card.flexibility { border-left-color: #3498db; }
        
        .no-events {
            text-align: center;
            color: #999;
            font-style: italic;
            padding: 1rem;
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
            .week-grid {
                grid-template-columns: 1fr;
                gap: 1.5rem;
            }
            .day-column {
                min-height: auto;
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
                <a href="/elev/hjem" class="nav-link">Hjem</a>
            </li>
            <li class="nav-item">
                <a href="/elev/timeplan" class="nav-link active">Timeplan</a>
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
        <h1 class="page-title">Timeplan</h1>
        
        <div class="module">
            <h2 class="module-title">Denne uka</h2>
            <div class="week-grid">
                {{range $index, $day := .WeekDays}}
                {{$date := index $.WeekDates $index}}
                {{$dateKey := $date.Format "2006-01-02"}}
                {{$events := index $.EventsByDay $dateKey}}
                <div class="day-column {{if lt $dateKey $.Today}}past{{end}}">
                    <div class="day-header">
                        <div class="day-name">{{$day}}</div>
                        <div class="day-date">{{$date.Format "02.01"}}</div>
                    </div>
                    {{if lt $dateKey $.Today}}
                        <div class="day-content">Avsluttet</div>
                    {{else}}
                        <div class="day-events">
                            {{if $events}}
                                {{range $events}}
                                <div class="event-card {{.ClassType}}">
                                    <div class="event-time">{{.StartTime.Format "15:04"}}-{{.EndTime.Format "15:04"}}</div>
                                    <div class="event-title">{{.Title}}</div>
                                    <div class="event-teacher">{{.TeacherName}}</div>
                                    <div class="event-spaces">
                                        {{if lt .CurrentEnrolment .Capacity}}
                                            {{sub .Capacity .CurrentEnrolment}} plasser
                                        {{else}}
                                            Venteliste
                                        {{end}}
                                    </div>
                                </div>
                                {{end}}
                            {{else}}
                                <div class="no-events">Ingen klasser</div>
                            {{end}}
                        </div>
                    {{end}}
                </div>
                {{end}}
            </div>
        </div>
    </main>
</body>
</html>`

	t, err := template.New("elev-timeplan").Funcs(template.FuncMap{
		"sub": func(a, b int) int {
			return a - b
		},
	}).Parse(tmpl)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, data)
}