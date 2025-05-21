package rpcserver

import (
	"github.com/gin-gonic/gin"
	"github.com/latifrons/latigo/berror"
	"github.com/latifrons/latigo/grpcserver"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"strconv"
)

const CodeOK = "OK"

type RpcWrapperFlags struct {
	ReturnDetailError bool
}

type RpcWrapper struct {
	Flags RpcWrapperFlags
}

func (rpc *RpcWrapper) ResponseDebug(c *gin.Context, status int, code string, msg string, debugMessage string, data interface{}) {
	c.Error(errors.Errorf("%s: %s", code, debugMessage))
	if !rpc.Flags.ReturnDetailError {
		debugMessage = ""
	}
	c.JSON(status, GeneralResponse{
		Code:     code,
		Msg:      msg,
		DebugMsg: debugMessage,
		Data:     data,
	})
}

func (rpc *RpcWrapper) Response(c *gin.Context, status int, code string, msg string, data interface{}) {

	c.JSON(status, GeneralResponse{
		Code: code,
		Msg:  msg,
		Data: data,
	})
}
func (rpc *RpcWrapper) ResponseOK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, GeneralResponse{
		Code: CodeOK,
		Msg:  "",
		Data: data,
	})
}

func (rpc *RpcWrapper) ResponsePaging(c *gin.Context, pagingResult PagingResult, data interface{}, list interface{}) {
	if pagingResult.Limit != 0 {
		c.JSON(http.StatusOK, PagingResponse{
			GeneralResponse: GeneralResponse{
				Code: CodeOK,
				Msg:  "",
				Data: data,
			},
			List:  list,
			Size:  pagingResult.Limit,
			Total: pagingResult.Total,
			Page:  pagingResult.Offset/pagingResult.Limit + 1,
		})
	} else {
		c.JSON(http.StatusOK, PagingResponse{
			GeneralResponse: GeneralResponse{
				Code: CodeOK,
				Msg:  "",
				Data: data,
			},
			List:  list,
			Size:  pagingResult.Limit,
			Total: pagingResult.Total,
			Page:  1,
		})
	}

}

func (rpc *RpcWrapper) ResponseBadRequest(c *gin.Context, err error, userMessage string) bool {
	if err == nil {
		return false
	}
	if userMessage != "" {
		rpc.ResponseDebug(c, http.StatusBadRequest, ErrBadRequest, userMessage, err.Error(), nil)
	} else if rpc.Flags.ReturnDetailError {
		rpc.ResponseDebug(c, http.StatusBadRequest, ErrBadRequest, err.Error(), err.Error(), nil)
	} else {
		rpc.ResponseDebug(c, http.StatusBadRequest, ErrBadRequest, "Bad request. Check your input.", err.Error(), nil)
	}
	return true
}

func (rpc *RpcWrapper) ResponseInternalServerError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	log.Error().Stack().Err(err).Msg("internal server error")

	if rpc.Flags.ReturnDetailError {
		rpc.ResponseDebug(c, http.StatusInternalServerError, ErrInternal, err.Error(), err.Error(), nil)
	} else {
		rpc.ResponseDebug(c, http.StatusInternalServerError, ErrInternal, "Internal server error", err.Error(), nil)
	}
	return true
}

func (rpc *RpcWrapper) ResponseError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	// grpc error check
	grpcError, ok := status.FromError(err)
	if ok {
		switch grpcError.Code() {
		case codes.Unavailable:
			rpc.ResponseDebug(c, http.StatusServiceUnavailable, ErrInternal, "service unreachable", grpcError.Message(), nil)
			return true
		case codes.NotFound:
			rpc.ResponseDebug(c, http.StatusNotFound, ErrNotFound, "service not found", grpcError.Message(), nil)
			return true
		case codes.InvalidArgument:
			rpc.ResponseDebug(c, http.StatusBadRequest, ErrBadRequest, "bad request", grpcError.Message(), nil)
			return true
		default:
			// no action

		}

		gerr, ok := grpcserver.Parse(grpcError.Message())
		if ok {
			switch gerr.Category {
			case grpcserver.Category_System:
				rpc.ResponseDebug(c, http.StatusInternalServerError, ErrInternal, "internal error", grpcError.Message(), nil)
			case grpcserver.Category_Business:
				switch gerr.Code {
				case ErrBadRequest:
					rpc.ResponseDebug(c, http.StatusBadRequest, gerr.Code, gerr.UserMessage, gerr.ModuleName+":"+gerr.DebugMessage, nil)
				default:
					rpc.ResponseDebug(c, http.StatusOK, gerr.Code, gerr.UserMessage, gerr.ModuleName+":"+gerr.DebugMessage, nil)
				}
			default:
				rpc.ResponseDebug(c, http.StatusOK, gerr.Code, gerr.UserMessage, gerr.ModuleName+":"+gerr.DebugMessage, nil)
			}
		} else {
			rpc.ResponseDebug(c, http.StatusInternalServerError, ErrInternal, "internal error", grpcError.Message(), nil)
		}

		return true
	} else {
		gerr, _, ok := grpcserver.FromError(err)
		if ok {
			switch gerr.Category {
			case grpcserver.Category_System:
				rpc.ResponseDebug(c, http.StatusInternalServerError, ErrInternal, "internal error", grpcError.Message(), nil)
			case grpcserver.Category_Business:
				switch gerr.Code {
				case ErrBadRequest:
					rpc.ResponseDebug(c, http.StatusBadRequest, gerr.Code, gerr.UserMessage, gerr.ModuleName+":"+gerr.DebugMessage, nil)
				default:
					rpc.ResponseDebug(c, http.StatusOK, gerr.Code, gerr.UserMessage, gerr.ModuleName+":"+gerr.DebugMessage, nil)
				}
			default:
				rpc.ResponseDebug(c, http.StatusOK, gerr.Code, gerr.UserMessage, gerr.ModuleName+":"+gerr.DebugMessage, nil)
			}
		} else {
			berr, ok := err.(*berror.BError)
			if ok {
				switch berr.Code {
				case berror.ErrInternal:
					rpc.ResponseDebug(c, http.StatusInternalServerError, ErrInternal, "internal server error", berr.Msg, nil)
					return true
				case berror.ErrBadRequest:
					rpc.ResponseDebug(c, http.StatusBadRequest, ErrBadRequest, berr.Msg, berr.Msg, nil)
					return true
				default:
					rpc.ResponseDebug(c, http.StatusOK, berr.Code, berr.Msg, berr.Error(), nil)
				}
			} else {
				rpc.ResponseDebug(c, http.StatusInternalServerError, ErrInternal, "internal error", err.Error(), nil)
			}
		}
	}

	return true
}

func (rpc *RpcWrapper) ResponseEmptyParam(c *gin.Context, name string, value string) bool {
	if value == "" {
		rpc.ResponseDebug(c, http.StatusBadRequest, ErrBadRequest, "param missing", "param missing: "+name, nil)
		return true
	}
	return false
}

func (rpc *RpcWrapper) ResponseEmptyField(c *gin.Context, name string, value string) bool {
	if value == "" {
		rpc.ResponseDebug(c, http.StatusBadRequest, ErrBadRequest, "body field missing", "body field missing: "+name, nil)
		return true
	}
	return false
}

func (rpc *RpcWrapper) ParsePagingGet(c *gin.Context) (paging PagingParams, err error) {
	paging.Limit, err = strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil {
		return
	}
	paging.Offset, err = strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil {
		return
	}
	return

}
