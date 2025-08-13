package models

type User struct {
	ID        int      `json:"id"`
	Name      string   `json:"name"`
	Birthdate string   `json:"birthdate"`
	Email     string   `json:"email"`
	Phone     string   `json:"phone"`
	Roles     []string `json:"roles"` // e.g. "admin", "user"
}
