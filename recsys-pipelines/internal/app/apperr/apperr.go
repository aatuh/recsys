package apperr

import (
	"errors"
	"fmt"
)

type Kind string

const (
	KindInvalidArgument  Kind = "invalid_argument"
	KindLimitExceeded    Kind = "limit_exceeded"
	KindValidationFailed Kind = "validation_failed"
	KindDependency       Kind = "dependency"
)

// Error is a small structured error for user-facing commands. It supports
// errors.Is matching by Kind sentinel values.
type Error struct {
	Kind Kind
	Msg  string
	Err  error
}

func (e Error) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("%s: %s", e.Kind, e.Msg)
	}
	return fmt.Sprintf("%s: %s: %v", e.Kind, e.Msg, e.Err)
}

func (e Error) Unwrap() error { return e.Err }

// Sentinel errors for errors.Is matching.
var (
	ErrInvalidArgument  = errors.New(string(KindInvalidArgument))
	ErrLimitExceeded    = errors.New(string(KindLimitExceeded))
	ErrValidationFailed = errors.New(string(KindValidationFailed))
	ErrDependency       = errors.New(string(KindDependency))
)

func New(kind Kind, msg string, err error) error {
	switch kind {
	case KindInvalidArgument:
		return Error{Kind: kind, Msg: msg, Err: errors.Join(ErrInvalidArgument, err)}
	case KindLimitExceeded:
		return Error{Kind: kind, Msg: msg, Err: errors.Join(ErrLimitExceeded, err)}
	case KindValidationFailed:
		return Error{Kind: kind, Msg: msg, Err: errors.Join(ErrValidationFailed, err)}
	case KindDependency:
		return Error{Kind: kind, Msg: msg, Err: errors.Join(ErrDependency, err)}
	default:
		return Error{Kind: kind, Msg: msg, Err: err}
	}
}
