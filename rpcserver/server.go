package rpcserver

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"time"
)

const ShutdownTimeoutSeconds = 5

type RouterProvider interface {
	ProvideRouter(*gin.Engine) *gin.Engine
}

type DebugFlags struct {
	GinDebug            bool
	RequestLog          bool
	ResponseLog         bool
	ReqBodyPrintLength  int
	RespBodyPrintLength int
}

type RpcServer struct {
	RouterProvider RouterProvider
	Port           string
	DebugFlags     DebugFlags
	router         *gin.Engine
	server         *http.Server
}

func (srv *RpcServer) Start() {
	router := srv.initRouter()
	srv.router = srv.RouterProvider.ProvideRouter(router)

	srv.server = &http.Server{
		Addr:    ":" + srv.Port,
		Handler: srv.router,
	}

	log.Info().Str("port", srv.Port).Msg("listening Http on " + srv.Port)
	go func() {
		// service connections
		if err := srv.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Stack().Err(err).Msg("error in Http rpcserver")
		}
	}()
}

func (srv *RpcServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeoutSeconds*time.Second)
	defer cancel()
	if err := srv.server.Shutdown(ctx); err != nil {
		log.Error().Stack().Err(err).Msg("error while shutting down the Http rpcserver")
	}
	log.Info().Msg("http rpcserver Stopped")
}

func (srv *RpcServer) Name() string {
	return fmt.Sprintf("rpcServer at port %s", srv.Port)
}

func (srv *RpcServer) initRouter() *gin.Engine {
	if srv.DebugFlags.GinDebug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	if srv.DebugFlags.GinDebug {
		logger := gin.LoggerWithConfig(gin.LoggerConfig{
			SkipPaths: []string{"/", "/health",
				"/metrics",
				"/apis",
				"/apis/swagger.json",
				"/redoc.standalone.js.map",
				"/574a25b96816f2c682b8.worker.js.map",
				"/docs/"},
		})
		router.Use(logger)
		if srv.DebugFlags.RequestLog {
			router.Use(RequestLoggerMiddleware(min(srv.DebugFlags.ReqBodyPrintLength, 4096)))
		}
		if srv.DebugFlags.ResponseLog {
			router.Use(ResponseLoggerMiddleware(min(srv.DebugFlags.RespBodyPrintLength, 4096)))
		}
	}
	return router
}

func (srv *RpcServer) InitDefault() {

}

func RequestLoggerMiddleware(reqBodyPrintLength int) gin.HandlerFunc {
	return func(c *gin.Context) {
		var buf bytes.Buffer
		tee := io.TeeReader(c.Request.Body, &buf)
		body, _ := io.ReadAll(tee)
		c.Request.Body = io.NopCloser(&buf)
		l := len(body)
		log.Trace().Str("method", c.Request.Method).Str("path", c.Request.URL.Path).Int("req", l).Any("header", c.Request.Header).Msg("REQ")
		log.Trace().Msg("REQ: " + string(body[:min(l, reqBodyPrintLength)]))
		c.Next()
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func ResponseLoggerMiddleware(respBodyPrintLength int) gin.HandlerFunc {
	return func(c *gin.Context) {
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw
		c.Next()

		s := blw.body.String()
		l := len(s)
		log.Trace().Str("method", c.Request.Method).Str("path", c.Request.URL.Path).Int("rsp", l).Msg("RSP")
		log.Trace().Msg("RSP: " + s[:min(l, respBodyPrintLength)])
	}
}
