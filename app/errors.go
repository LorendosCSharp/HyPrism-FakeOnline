package app

import "fmt"

// ErrorType represents types of errors
type ErrorType string

const (
	ErrorTypeNetwork     ErrorType = "NETWORK"
	ErrorTypeFileSystem  ErrorType = "FILESYSTEM"
	ErrorTypeValidation  ErrorType = "VALIDATION"
	ErrorTypeGame        ErrorType = "GAME"
	ErrorTypeUpdate      ErrorType = "UPDATE"
	ErrorTypeUnknown     ErrorType = "UNKNOWN"
)

// AppError represents a structured error
type AppError struct {
	Type      ErrorType `json:"type"`
	Message   string    `json:"message"`
	Technical string    `json:"technical,omitempty"`
	Timestamp string    `json:"timestamp"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Type, e.Message, e.Technical)
}

// NewAppError creates a new AppError
func NewAppError(errType ErrorType, message string, cause error) *AppError {
	technical := ""
	if cause != nil {
		technical = cause.Error()
	}
	return &AppError{
		Type:      errType,
		Message:   message,
		Technical: technical,
		Timestamp: "",
	}
}

// WrapError wraps an error with a type
func WrapError(errType ErrorType, message string, cause error) *AppError {
	return NewAppError(errType, message, cause)
}

// NetworkError creates a network error
func NetworkError(action string, cause error) *AppError {
	return NewAppError(ErrorTypeNetwork, fmt.Sprintf("Network error while %s", action), cause)
}

// FileSystemError creates a filesystem error
func FileSystemError(action string, cause error) *AppError {
	return NewAppError(ErrorTypeFileSystem, fmt.Sprintf("File system error while %s", action), cause)
}

// ValidationError creates a validation error
func ValidationError(message string) *AppError {
	return NewAppError(ErrorTypeValidation, message, nil)
}

// GameError creates a game error
func GameError(message string, cause error) *AppError {
	return NewAppError(ErrorTypeGame, message, cause)
}

// UpdateError creates an update error
func UpdateError(message string, cause error) *AppError {
	return NewAppError(ErrorTypeUpdate, message, cause)
}
