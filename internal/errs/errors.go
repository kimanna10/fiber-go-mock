package errs

type AppError struct {
	Message    string
	StatusCode int
}

func (e *AppError) Error() string {
	return e.Message
}

func New(message string, status int) *AppError {
	return &AppError{
		Message:    message,
		StatusCode: status,
	}
}

var (
	ErrBadRequest = New("Bad request", 400)
	ErrNotFound   = New("Resource not found", 404)
	ErrInternal   = New("Internal server error", 500)

	ErrUserNotFound = New("User not found", 404)
	ErrTimeout      = New("Request timeout", 408)

	ErrUnauthorized       = New("Unauthorized", 401)
	ErrInvalidCredentials = New("Invalid credentials", 401)
	ErrInvalidToken       = New("Invalid token", 401)
	ErrForbidden          = New("Forbidden", 403)

	// Для валидации
	ErrFieldEmpty  = New("Field cannot be empty", 400)
	ErrInvalidData = New("Invalid data", 400)
)
