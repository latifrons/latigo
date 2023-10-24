package grpcserver

import (
	"errors"
	"github.com/latifrons/latigo/berror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Gerror(c codes.Code, berror, format string, a ...any) error {
	return status.Errorf(c, berror+": "+format, a...)
}

func WrapGError(err error) error {
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
		return Gerror(codes.Internal, berr.Code, berr.Msg)
	default:
		return Gerror(codes.Internal, berror.ErrInternal, err.Error())
	}
}
