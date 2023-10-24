package grpcserver

import (
	"errors"
	"github.com/latifrons/latigo/berror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Gerror(from string, c codes.Code, berror, format string, a ...any) error {
	from = "[" + from + "]"
	return status.Errorf(c, berror+":"+from+":"+format, a...)
}

func WrapGError(from string, err error) error {
	if err == nil {
		return nil
	}
	var BError *berror.BError
	switch {
	case errors.As(err, &BError):
		var berr *berror.BError
		errors.As(err, &berr)
		if berr == nil {
			return nil
		}
		return Gerror(from, codes.Internal, berr.Code, berr.Msg)
	default:
		return Gerror(from, codes.Internal, berror.ErrInternal, err.Error())
	}
}
