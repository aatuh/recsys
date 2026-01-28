package foosvc

import "errors"

var (
	ErrInvalid  = errors.New("invalid input")
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
	ErrInternal = errors.New("internal error")
)
