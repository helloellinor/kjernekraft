package handlers

import (
	"io/ioutil"
	"net/http"

	"github.com/russross/blackfriday/v2"
)

func TermsHandler(w http.ResponseWriter, r *http.Request) {
	// Read the terms markdown file
	content, err := ioutil.ReadFile("static/vilkår.md")
	if err != nil {
		http.Error(w, "Could not load terms", http.StatusInternalServerError)
		return
	}

	// Convert markdown to HTML
	html := blackfriday.Run(content)
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(html)
}

func SignUpPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "signup.html")
}