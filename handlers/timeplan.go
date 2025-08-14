package handlers

import (
	"kjernekraft/handlers/config"
	"kjernekraft/models"
	"net/http"
	"time"
)

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
	settings := config.GetInstance()
	now := settings.GetCurrentTime()

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
        
        
        .no-events {
            text-align: center;
            color: #999;
            font-style: italic;
            padding: 1rem;
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
                                {{template "event_card" .}}
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

    <script>
        function toggleEventCard(element) {
            const isExpanded = element.classList.contains('expanded');
            
            // Close all other expanded cards first
            document.querySelectorAll('.event-card.expanded').forEach(card => {
                if (card !== element) {
                    card.classList.remove('expanded');
                }
            });
            
            // Toggle the clicked card
            if (isExpanded) {
                element.classList.remove('expanded');
            } else {
                element.classList.add('expanded');
            }
        }
        
        function signupForClass(classId) {
            // TODO: Implement class signup functionality
            alert('Påmelding for klasse ' + classId + ' - kommer snart!');
        }
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
