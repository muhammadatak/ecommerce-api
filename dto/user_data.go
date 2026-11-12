package dto

type UserData struct {
	Id       string `json:"id"`
	Email    string `json:"email"`
	Password string `json:"-"`
}
