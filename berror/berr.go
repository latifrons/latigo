package berror

import (
	"fmt"
	"github.com/pkg/errors"
)

type ErrorCategory int

const (
	CategoryBusinessFail      ErrorCategory = 1
	CategoryBusinessTemporary               = 2
	CategorySystemTemporary                 = 3
)

type StackTracer interface {
	StackTrace() errors.StackTrace
}

type BError struct {
	Code          string
	Msg           string
	CausedBy      error
	ErrorCategory ErrorCategory
}

func (b *BError) StackTrace() errors.StackTrace {
	return b.CausedBy.(StackTracer).StackTrace()
}

func (b *BError) Error() string {
	return fmt.Sprintf("code: %s, cat: %d, msg: %s, causedBy: %v", b.Code, b.ErrorCategory, b.Msg, b.CausedBy)
}

func new(code string, msg string, errorCategory ErrorCategory, causedBy error) *BError {
	if causedBy == nil {
		causedBy = errors.New(msg)
	}
	b := &BError{
		Code:          code,
		Msg:           msg,
		ErrorCategory: errorCategory,
		CausedBy:      errors.Wrap(causedBy, "caused by"),
	}
	return b
}

func NewSystemTemporary(causedBy error, code string, msg string) *BError {
	return new(code, msg, CategorySystemTemporary, causedBy) // can be resolved by retry, caused by system issue
}

func NewBusinessFail(causedBy error, code string, msg string) *BError {
	return new(code, msg, CategoryBusinessFail, causedBy) // cannot be resolved by retry
}

func NewBusinessTemporary(causedBy error, code string, msg string) *BError {
	return new(code, msg, CategoryBusinessTemporary, causedBy) // can be resolved by retry, caused by business issue.
}
