package rpcserver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const ShutdownTimeoutSeconds = 5

type RouterProvider interface {
	ProvideRouter(*gin.Engine) *gin.Engine
}

type DebugFlags struct {
	GinDebug    bool
	RequestLog  bool
	ResponseLog bool
}

type RpcServer struct {
	RouterProvider RouterProvider
	Port           string
	DebugFlags     DebugFlags
	Logger         *zap.SugaredLogger

	router *gin.Engine
	server *http.Server
}

func (srv *RpcServer) Start() {
	router := srv.initRouter()
	srv.router = srv.RouterProvider.ProvideRouter(router)
	srv.router.Use(BreakerWrapper)

	srv.server = &http.Server{
		Addr:    ":" + srv.Port,
		Handler: srv.router,
	}

	srv.Logger.Infow("listening Http on " + srv.Port)
	go func() {
		// service connections
		if err := srv.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			srv.Logger.Fatalw("error in Http rpcserver", "err", err)
		}
	}()
}

func (srv *RpcServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeoutSeconds*time.Second)
	defer cancel()
	if err := srv.server.Shutdown(ctx); err != nil {
		srv.Logger.Errorw("error while shutting down the Http rpcserver", "err", err)
	}
	srv.Logger.Infow("http rpcserver Stopped")
}

func (srv *RpcServer) Name() string {
	return fmt.Sprintf("rpcServer at port %s", srv.Port)
}

func (srv *RpcServer) InitDefault() {
	if srv.Logger == nil {
		srv.Logger = zap.NewExample().Sugar()
	}
}

func (srv *RpcServer) initRouter() *gin.Engine {
	if srv.DebugFlags.GinDebug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	if srv.DebugFlags.RequestLog || srv.DebugFlags.ResponseLog {
		logger := gin.LoggerWithConfig(gin.LoggerConfig{
			SkipPaths: []string{"/", "/health"},
		})
		router.Use(logger)
		if srv.DebugFlags.RequestLog {
			router.Use(RequestLoggerMiddleware())
		}

		if srv.DebugFlags.ResponseLog {
			router.Use(ResponseLoggerMiddleware())
		}
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	router.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	router.Use(ginzap.RecoveryWithZap(logger, true))
	return router
}

var ginLogFormatter = func(param gin.LogFormatterParams) string {
	var statusColor, methodColor, resetColor string
	if param.IsOutputColor() {
		statusColor = param.StatusCodeColor()
		methodColor = param.MethodColor()
		resetColor = param.ResetColor()
	}

	if param.Latency > time.Minute {
		// Truncate in a golang < 1.8 safe way
		param.Latency = param.Latency - param.Latency%time.Second
	}

	logEntry := fmt.Sprintf("GIN %v %s %3d %s %13v  %15s %s %-7s %s %s %s",
		param.TimeStamp.Format("2006/01/02 - 15:04:05"),
		statusColor, param.StatusCode, resetColor,
		param.Latency,
		param.ClientIP,
		methodColor, param.Method, resetColor,
		param.Path,
		param.ErrorMessage,
	)
	//logrus.Tracef("gin log %v ", logEntry)
	return logEntry
	//return ""
}

var FullDateFormatPattern = "2 Jan 2006 15:04:05"
var ShortDateFormatPattern = "2 Jan 2006"

func BreakerWrapper(c *gin.Context) {
	name := c.Request.Method + "-" + c.Request.RequestURI
	hystrix.Do(name, func() error {
		c.Next()

		statusCode := c.Writer.Status()

		if statusCode >= http.StatusInternalServerError {
			str := fmt.Sprintf("status code %d", statusCode)
			return errors.New(str)
		}

		return nil
	}, func(e error) error {
		if e == hystrix.ErrCircuitOpen {
			c.String(http.StatusAccepted, "Please try again later")
		}

		return e
	})
}

func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var buf bytes.Buffer
		tee := io.TeeReader(c.Request.Body, &buf)
		body, _ := io.ReadAll(tee)
		c.Request.Body = io.NopCloser(&buf)
		zap.S().Debugw("Received request", "uri", c.Request.RequestURI, "headers", c.Request.Header, "body", string(body))
		c.Next()
	}
}

func tryParseIntDefault(v string, d int) int {
	c, err := strconv.Atoi(v)
	if err != nil {
		return d
	}
	return c
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func ResponseLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw
		c.Next()
		if strings.HasPrefix(c.Request.RequestURI, "/swagger/") {
			return
		}
		zap.S().Debugw("Response", "body", blw.body.String())
	}
}
