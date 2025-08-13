package models

type User struct {
	ID                     int      `json:"id"`
	Name                   string   `json:"name"`
	Birthdate              string   `json:"birthdate"`
	Email                  string   `json:"email"`
	Phone                  string   `json:"phone"`
	Address                string   `json:"address"`
	PostalCode             string   `json:"postal_code"`
	City                   string   `json:"city"`
	Country                string   `json:"country"`
	Password               string   `json:"password"`
	NewsletterSubscription bool     `json:"newsletter_subscription"`
	TermsAccepted          bool     `json:"terms_accepted"`
	Roles                  []string `json:"roles"` // e.g. "admin", "user"
}
