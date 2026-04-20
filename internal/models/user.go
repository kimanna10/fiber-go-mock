package models

// Сущность для базы (Entity)
type User struct {
	ID       int    `db:"id"`
	Name     string `db:"name"`
	Age      int    `db:"age"`
	Role     string `db:"role"`
	Email    string `db:"email"`
	Password string `db:"password"`
}

// Ответ клиенту (Response DTO)
type UserResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Role  string `json:"role"`
	Email string `json:"email"`
}

// Данные для входа (Request DTO)
type UserLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Данные для регистрации (Request DTO)
type UserRegisterRequest struct {
	Name     string `json:"name"`
	Age      int    `json:"age"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}
