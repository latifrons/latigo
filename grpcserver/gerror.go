package grpcserver

import (
	"encoding/json"
	"errors"
	"github.com/latifrons/latigo/berror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (e *GError) Error() string {
	b, _ := json.Marshal(e)
	return string(b)
}

func Parse(err string) (*GError, bool) {
	e := new(GError)
	errx := json.Unmarshal([]byte(err), e)
	if errx != nil {
		return nil, false
	}
	return e, true
}

func NewGError(module string, berrorCode, userMessage, debugMessage, stackTrace string, causedBy []*GError, category Category) (gerr *GError) {
	gerr = &GError{
		Code:         berrorCode,
		ModuleName:   module,
		UserMessage:  userMessage,
		DebugMessage: debugMessage,
		StackTrace:   stackTrace,
		CausedBy:     causedBy,
		Category:     category,
	}
	return
}

// convert to grpc error
func WrapGRpcError(module string, code codes.Code, err error) (errOut error) {
	if err == nil {
		return nil
	}

	// already a grpc wrapped error
	if _, ok := status.FromError(err); ok {
		return err
	}

	var gerror *GError

	switch {
	case errors.As(err, &gerror):
		var berr *GError
		errors.As(err, &berr)
		if berr == nil {
			return nil
		}
		return status.New(code, gerror.Error()).Err()
	default:
		gerror = NewGError(module, berror.ErrInternal, berror.ErrInternal, err.Error(), "", nil, Category_System)
		return status.New(code, gerror.Error()).Err()
	}

}

func Gerror(from string, c codes.Code, berror, format string, a ...any) error {
	from = "[" + from + "]"
	return status.Errorf(c, berror+":"+from+":"+format, a...)
}

func WrapGError(from string, err error) error {
	if err == nil {
		return nil
	}

	if _, ok := status.FromError(err); ok {
		return err
	}

	var BError *berror.BError
	switch {
	case errors.As(err, &BError):
		var berr *berror.BError
		errors.As(err, &berr)
		if berr == nil {
			return nil
		}
		return Gerror(from, codes.FailedPrecondition, berr.Code, berr.Msg)
	default:
		return Gerror(from, codes.Internal, berror.ErrInternal, err.Error())
	}
}

// FromError try to convert go error to *Error.
func FromError(err error) (gerr *GError, ok bool) {
	if err == nil {
		return nil, true
	}
	if verr, ok := status.FromError(err); ok && verr != nil {
		return Parse(verr.Message())
	}

	var verr *GError
	if errors.As(err, &verr) && verr != nil {
		return verr, true
	}

	// other unknown error
	return nil, false
}
