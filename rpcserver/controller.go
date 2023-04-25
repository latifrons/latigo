package rpcserver

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/latifrons/latigo/berror"
	"github.com/sirupsen/logrus"
	"net/http"
)

const CodeOK = "OK"

type RpcWrapperFlags struct {
	ReturnDetailError bool
	ResponseLog       bool
}

type RpcWrapper struct {
	Flags RpcWrapperFlags
}

func (rpc *RpcWrapper) Response(c *gin.Context, status int, code string, msg string, data interface{}) {
	if rpc.Flags.ResponseLog {
		logrus.WithField("data", data).WithField("msg", msg).WithField("code", code).WithField("status", status).Info("resp")
	}
	c.JSON(status, GeneralResponse{
		Code: code,
		Msg:  msg,
		Data: data,
	})
}
func (rpc *RpcWrapper) ResponseOK(c *gin.Context, data interface{}) {
	if rpc.Flags.ResponseLog {
		logrus.WithField("data", data).Info("resp ok")
	}
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
	logrus.WithError(err).Debug("bad request")
	if userMessage != "" {
		rpc.Response(c, http.StatusBadRequest, ErrBadRequest, userMessage, nil)
	} else if rpc.Flags.ReturnDetailError {
		rpc.Response(c, http.StatusBadRequest, ErrBadRequest, err.Error(), nil)
	} else {
		rpc.Response(c, http.StatusBadRequest, ErrBadRequest, "Bad request. Check your input.", nil)
	}
	return true
}

func (rpc *RpcWrapper) ResponseInternalServerError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	logrus.WithError(err).Error("internal error")
	if rpc.Flags.ReturnDetailError {
		rpc.Response(c, http.StatusInternalServerError, ErrInternal, err.Error(), nil)
	} else {
		rpc.Response(c, http.StatusInternalServerError, ErrInternal, "Internal server error", nil)
	}
	return true
}

func (rpc *RpcWrapper) ResponseError(c *gin.Context, err error) bool {
	switch err.(type) {
	case *berror.BError:
		berr := err.(*berror.BError)
		if berr == nil {
			return false
		}

		var msg = "fail"
		if rpc.Flags.ReturnDetailError {
			msg = "DEBUG ONLY >>>" + berr.Msg
		}

		// response http code according to DTM
		switch berr.ErrorCategory {
		case berror.CategoryBusinessFail:
			rpc.Response(c, http.StatusOK, berr.Code, msg, nil)
		case berror.CategoryBusinessTemporary:
			rpc.Response(c, http.StatusOK, berr.Code, msg, nil)
		case berror.CategorySystemTemporary:
			rpc.Response(c, http.StatusOK, berr.Code, msg, nil)
		}
	default:
		if err == nil {
			return false
		}
		fmt.Printf("%s", err.Error())
		if v, ok := err.(berror.StackTracer); ok {
			fmt.Println(v.StackTrace())
		} else {
			fmt.Printf("%s", err.Error())
		}

		rpc.Response(c, http.StatusInternalServerError, ErrInternal, err.Error(), nil)
	}
	return true
}

func (rpc *RpcWrapper) ResponseEmptyParam(c *gin.Context, name string, value string) bool {
	if value == "" {
		logrus.Debug("param missing")
		rpc.Response(c, http.StatusBadRequest, ErrBadRequest, "param missing: "+name, nil)
		return true
	}
	return false
}

func (rpc *RpcWrapper) ResponseEmptyField(c *gin.Context, name string, value string) bool {
	if value == "" {
		logrus.Debug("body field missing")
		rpc.Response(c, http.StatusBadRequest, ErrBadRequest, "body field missing: "+name, nil)
		return true
	}
	return false
}
