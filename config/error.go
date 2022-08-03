package config

type ValidationError struct {
	message string
}

func NewValidationError(message string) error {
	return &ValidationError{
		message: message,
	}
}

func (e *ValidationError) Error() string {
	return e.message
}
