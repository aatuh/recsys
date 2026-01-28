package validation

// Error represents a validation error with an HTTP status and machine code.
type Error struct {
	Status  int
	Code    string
	Message string
}

func (e Error) Error() string {
	return e.Message
}

func newError(status int, code, message string) error {
	return Error{Status: status, Code: code, Message: message}
}
