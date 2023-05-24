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
	ErrorCategory ErrorCategory
	StackTrace    errors.StackTrace
	//InnerError    error
}

func (b *BError) Error() string {
	//if b.InnerError != nil {
	//	return fmt.Sprintf("msg: %s, inner: %s", b.Msg, b.InnerError.Error())
	//} else {
	return fmt.Sprintf("code: %s, cat: %d, msg: %s", b.Code, b.ErrorCategory, b.Msg)
	//}
}

func New(code string, msg string, errorCategory ErrorCategory) *BError {
	b := &BError{
		Code:          code,
		Msg:           msg,
		ErrorCategory: errorCategory,
		StackTrace:    errors.New("").(StackTracer).StackTrace(),
	}
	return b
}

func NewFail(code string, msg string) *BError {
	return New(code, msg, CategoryBusinessFail)
}
