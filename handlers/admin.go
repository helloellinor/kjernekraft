package handlers

import (
	"encoding/json"
	"kjernekraft/database"
	"kjernekraft/models"
	"net/http"
)

var AdminDB *database.Database

type AdminData struct {
	Users          []models.User
	Events         []models.Event
	FreezeRequests []models.FreezeRequest
}

func AdminPageHandler(w http.ResponseWriter, r *http.Request) {
	// For now, we'll skip authentication check
	// TODO: Add proper authentication to check if user has admin role

	users, err := AdminDB.GetAllUsers()
	if err != nil {
		http.Error(w, "Kunne ikke hente brukere", http.StatusInternalServerError)
		return
	}

	events, err := AdminDB.GetAllEvents()
	if err != nil {
		http.Error(w, "Kunne ikke hente events", http.StatusInternalServerError)
		return
	}

	freezeRequests, err := AdminDB.GetPendingFreezeRequests()
	if err != nil {
		http.Error(w, "Kunne ikke hente frysingsforespørsler", http.StatusInternalServerError)
		return
	}

	data := AdminData{
		Users:          users,
		Events:         events,
		FreezeRequests: freezeRequests,
	}

	tmpl := `
<!DOCTYPE html>
<html lang="no">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Admin - Brukeradministrasjon</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        table { border-collapse: collapse; width: 100%; margin-top: 20px; }
        th, td { border: 1px solid #ddd; padding: 12px; text-align: left; }
        th { background-color: #f2f2f2; font-weight: bold; }
        tr:hover { background-color: #f5f5f5; }
        .roles { color: #666; font-style: italic; }
        h1, h2 { color: #333; }
        .stats { background: #f9f9f9; padding: 15px; border-radius: 5px; margin-bottom: 20px; }
        .section { margin-bottom: 40px; }
        .form-group { margin: 10px 0; }
        .form-group label { display: inline-block; width: 120px; }
        .form-group input { padding: 5px; margin-left: 10px; }
        button { background: #007cba; color: white; padding: 8px 16px; border: none; border-radius: 4px; cursor: pointer; }
        button:hover { background: #005a87; }
        .update-form { background: #f9f9f9; padding: 10px; margin: 5px 0; border-radius: 4px; }
    </style>
</head>
<body>
    <h1>Admin - Brukeradministrasjon</h1>
    
    {{template "admin_settings" .}}
    
    <div class="stats">
        <strong>Totalt antall brukere:</strong> {{len .Users}} | 
        <strong>Totalt antall events:</strong> {{len .Events}} |
        <strong>Ventende frysingsforespørsler:</strong> {{len .FreezeRequests}}
    </div>

    <div class="section">
        <h2>Brukere</h2>
        <table>
            <thead>
                <tr>
                    <th>ID</th>
                    <th>Navn</th>
                    <th>Fødselsdato</th>
                    <th>E-post</th>
                    <th>Telefon</th>
                    <th>Roller</th>
                    <th>Opprettet</th>
                </tr>
            </thead>
            <tbody>
                {{range .Users}}
                <tr>
                    <td>{{.ID}}</td>
                    <td>{{.Name}}</td>
                    <td>{{.Birthdate}}</td>
                    <td>{{.Email}}</td>
                    <td>{{.Phone}}</td>
                    <td class="roles">{{range $i, $role := .Roles}}{{if $i}}, {{end}}{{$role}}{{end}}</td>
                    <td>-</td>
                </tr>
                {{end}}
            </tbody>
        </table>
    </div>

    <div class="section">
        <h2>Frysingsforespørsler</h2>
        {{if .FreezeRequests}}
        <table>
            <thead>
                <tr>
                    <th>Bruker</th>
                    <th>E-post</th>
                    <th>Telefon</th>
                    <th>Medlemskap</th>
                    <th>Pris</th>
                    <th>Binding (mnd)</th>
                    <th>Opprettet</th>
                    <th>Handlinger</th>
                </tr>
            </thead>
            <tbody>
                {{range .FreezeRequests}}
                <tr>
                    <td>{{.UserName}}</td>
                    <td>{{.UserEmail}}</td>
                    <td>{{.UserPhone}}</td>
                    <td>{{.MembershipName}}</td>
                    <td>{{.MembershipPrice}} kr/mnd</td>
                    <td>{{.CommitmentMonths}}</td>
                    <td>{{.CreatedAt.Format "02.01.2006 15:04"}}</td>
                    <td>
                        <button onclick="approveFreezeRequest({{.UserID}})" style="background: #28a745; margin-right: 5px;">Godkjenn</button>
                        <button onclick="rejectFreezeRequest({{.UserID}})" style="background: #dc3545;">Avvis</button>
                    </td>
                </tr>
                {{end}}
            </tbody>
        </table>
        {{else}}
        <p style="font-style: italic; color: #666;">Ingen ventende frysingsforespørsler</p>
        {{end}}
    </div>

    <div class="section">
        <h2>Event Tidsadministrasjon</h2>
        <table>
            <thead>
                <tr>
                    <th>ID</th>
                    <th>Tittel</th>
                    <th>Starttid</th>
                    <th>Sluttid</th>
                    <th>Lokasjon</th>
                    <th>Handlinger</th>
                </tr>
            </thead>
            <tbody>
                {{range .Events}}
                <tr>
                    <td>{{.ID}}</td>
                    <td>{{.Title}}</td>
                    <td>{{.StartTime.Format "2006-01-02 15:04"}}</td>
                    <td>{{.EndTime.Format "2006-01-02 15:04"}}</td>
                    <td>{{.Location}}</td>
                    <td>
                        <div class="update-form">
                            <form onsubmit="updateEventTime(event, {{.ID}})">
                                <div class="form-group">
                                    <label>Start:</label>
                                    <input type="datetime-local" id="start_{{.ID}}" value="{{.StartTime.Format "2006-01-02T15:04"}}" required>
                                </div>
                                <div class="form-group">
                                    <label>Slutt:</label>
                                    <input type="datetime-local" id="end_{{.ID}}" value="{{.EndTime.Format "2006-01-02T15:04"}}" required>
                                </div>
                                <button type="submit">Oppdater tid</button>
                            </form>
                        </div>
                    </td>
                </tr>
                {{end}}
            </tbody>
        </table>
    </div>

    <script>
        function updateEventTime(event, eventId) {
            event.preventDefault();
            
            const startTime = document.getElementById('start_' + eventId).value;
            const endTime = document.getElementById('end_' + eventId).value;
            
            if (!startTime || !endTime) {
                alert('Begge tidsfelt må fylles ut');
                return;
            }
            
            const url = '/api/admin/events/update-time?event_id=' + eventId + 
                       '&start_time=' + encodeURIComponent(startTime) + 
                       '&end_time=' + encodeURIComponent(endTime);
            
            fetch(url, { method: 'POST' })
                .then(response => {
                    if (response.ok) {
                        alert('Event tid oppdatert!');
                        location.reload();
                    } else {
                        response.text().then(text => alert('Feil: ' + text));
                    }
                })
                .catch(error => alert('Feil: ' + error));
        }

        function approveFreezeRequest(userId) {
            if (!confirm('Er du sikker på at du vil godkjenne frysingsforespørselen?')) {
                return;
            }
            
            fetch('/api/admin/freeze-requests/approve?user_id=' + userId, { method: 'POST' })
                .then(response => {
                    if (response.ok) {
                        alert('Frysingsforespørsel godkjent!');
                        location.reload();
                    } else {
                        response.text().then(text => alert('Feil: ' + text));
                    }
                })
                .catch(error => alert('Feil: ' + error));
        }

        function rejectFreezeRequest(userId) {
            if (!confirm('Er du sikker på at du vil avvise frysingsforespørselen?')) {
                return;
            }
            
            fetch('/api/admin/freeze-requests/reject?user_id=' + userId, { method: 'POST' })
                .then(response => {
                    if (response.ok) {
                        alert('Frysingsforespørsel avvist!');
                        location.reload();
                    } else {
                        response.text().then(text => alert('Feil: ' + text));
                    }
                })
                .catch(error => alert('Feil: ' + error));
        }
    </script>
</body>
</html>`

	// Try to use the new template system with components
	tm := GetTemplateManager()
	t, err := tm.ParseTemplate(tmpl, "admin")
	if err != nil {
		http.Error(w, "Template-feil", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, data)
}

func GetUsersAPIHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Add authentication check for admin role

	users, err := AdminDB.GetAllUsers()
	if err != nil {
		http.Error(w, "Kunne ikke hente brukere", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// ApproveFreezeRequestHandler handles approval of freeze requests
func ApproveFreezeRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Add authentication check for admin role

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "Missing user_id parameter", http.StatusBadRequest)
		return
	}

	userID := parseInt64(userIDStr)
	if userID == 0 {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	err := AdminDB.ApproveFreezeRequest(userID)
	if err != nil {
		http.Error(w, "Could not approve freeze request", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Freeze request approved"))
}

// RejectFreezeRequestHandler handles rejection of freeze requests
func RejectFreezeRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Add authentication check for admin role

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "Missing user_id parameter", http.StatusBadRequest)
		return
	}

	userID := parseInt64(userIDStr)
	if userID == 0 {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	err := AdminDB.RejectFreezeRequest(userID)
	if err != nil {
		http.Error(w, "Could not reject freeze request", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Freeze request rejected"))
}

// Helper function to parse int64 from string
func parseInt64(s string) int64 {
	if s == "" {
		return 0
	}
	// Simple conversion - in production, use strconv.ParseInt with error handling
	var result int64
	for _, r := range s {
		if r >= '0' && r <= '9' {
			result = result*10 + int64(r-'0')
		} else {
			return 0
		}
	}
	return result
}
