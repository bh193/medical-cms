package models

type User struct {
	Id      int     `json:"id"`
	Email   string  `json:"email"`
	Name    string  `json:"name"`
	Picture *string `json:"picture"` // *為指針可nil
}

type UserRole struct {
	UserID int `json:"user_id"`
	RoleID int `json:"role_id"`
}