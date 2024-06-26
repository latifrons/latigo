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

const (
	ErrInternal   = "ErrInternal"
	ErrBusiness   = "ErrBusiness"
	ErrBadRequest = "ErrBadRequest"
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
	if b.CausedBy == nil {
		return []errors.Frame{}
	}
	return b.CausedBy.(StackTracer).StackTrace()
}

func (b *BError) Error() string {
	if b.CausedBy == nil {
		return fmt.Sprintf("code: %s, cat: %d, msg: %s", b.Code, b.ErrorCategory, b.Msg)
	} else {
		return fmt.Sprintf("code: %s, cat: %d, msg: %s, causedBy: %v", b.Code, b.ErrorCategory, b.Msg, b.CausedBy)
	}

}

func new(code string, msg string, errorCategory ErrorCategory, causedBy error) (b *BError) {

	if causedBy != nil {
		if _, ok := causedBy.(StackTracer); !ok {
			causedBy = errors.Wrap(causedBy, "")
		}
	}
	b = &BError{
		Code:          code,
		Msg:           msg,
		ErrorCategory: errorCategory,
		CausedBy:      causedBy,
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
