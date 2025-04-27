package models

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type User struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}