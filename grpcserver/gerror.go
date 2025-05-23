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

func NewGRpcError(module string, code codes.Code, berrorCode string, userMessage string, debugMessage string, category Category) (errOut error) {
	gerror := NewGError(module, berrorCode, userMessage, debugMessage, "", nil, category)
	return status.New(code, gerror.Error()).Err()

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
	var berrorx *berror.BError

	switch {
	case errors.As(err, &gerror):
		var berr *GError
		errors.As(err, &berr)
		if berr == nil {
			return nil
		}
		return status.New(code, gerror.Error()).Err()
	case errors.As(err, &berrorx):
		e := berrorx.Msg
		if berrorx.CausedBy != nil {
			e += " caused by: " + berrorx.CausedBy.Error()
		}
		return NewGRpcError(module, code, berrorx.Code, berrorx.Msg, e, Category_Business)
	default:
		return NewGRpcError(module, code, berror.ErrInternal, "", err.Error(), Category_System)
	}

}

func WrapGRpcErrorLogic(from string, err error) error {
	return WrapGRpcError(from, codes.FailedPrecondition, err)
}

func WrapGRpcErrorBadRequest(from string, err error) error {
	return WrapGRpcError(from, codes.InvalidArgument, berror.NewBusinessFail(err, berror.ErrBadRequest, err.Error()))
}

// FromError try to convert go error to *Error.
func FromError(err error) (gerr *GError, code codes.Code, ok bool) {
	if err == nil {
		return nil, codes.OK, true
	}
	var verr *GError
	if errors.As(err, &verr) && verr != nil {
		return verr, codes.Internal, true
	}

	var serr *status.Status
	if serr, ok = status.FromError(err); ok && serr != nil {
		code = serr.Code()
		gerr, ok = Parse(serr.Message())
		return
	}

	// other unknown error
	return nil, codes.Unknown, false
}
