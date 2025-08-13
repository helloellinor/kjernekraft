package handlers

import (
	"encoding/json"
	"html/template"
	"kjernekraft/database"
	"kjernekraft/models"
	"net/http"
)

var AdminDB *database.Database

type AdminData struct {
	Users []models.User
}

func AdminPageHandler(w http.ResponseWriter, r *http.Request) {
	// For now, we'll skip authentication check
	// TODO: Add proper authentication to check if user has admin role

	users, err := AdminDB.GetAllUsers()
	if err != nil {
		http.Error(w, "Kunne ikke hente brukere", http.StatusInternalServerError)
		return
	}

	data := AdminData{
		Users: users,
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
        h1 { color: #333; }
        .stats { background: #f9f9f9; padding: 15px; border-radius: 5px; margin-bottom: 20px; }
    </style>
</head>
<body>
    <h1>Admin - Brukeradministrasjon</h1>
    
    <div class="stats">
        <strong>Totalt antall brukere:</strong> {{len .Users}}
    </div>

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
</body>
</html>`

	t, err := template.New("admin").Parse(tmpl)
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
