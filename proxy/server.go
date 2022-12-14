package proxy

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/form3tech-oss/http-message-signing-proxy/config"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	http.Server
	sslConfig config.SSLConfig
}

func NewServer(cfg config.ServerConfig, handler Handler, metric MetricPublisher) *Server {
	router := gin.New()

	router.GET("/-/health", handler.Health)
	router.GET("/-/prometheus", func(c *gin.Context) {
		promhttp.Handler().ServeHTTP(c.Writer, c.Request)
	})

	// NoRoute means all other routes.
	// We cannot use wildcard here because it will conflict with /-/health and /-/prometheus above.
	router.NoRoute(
		RecoverMiddleware(metric),
		LogAndMetricsMiddleware(metric),
		CORSMiddleware(cfg.AccessControlAllowOrigin),
		handler.ForwardRequest,
	)

	return &Server{
		Server: http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.Port),
			Handler: router,
		},
		sslConfig: cfg.SSL,
	}
}

func (s *Server) Start() {
	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		var err error
		if s.sslConfig.Enable {
			log.Info("starting server in TLS mode")
			err = s.ListenAndServeTLS(s.sslConfig.CertFilePath, s.sslConfig.KeyFilePath)
		} else {
			log.Info("starting server without TLS")
			err = s.ListenAndServe()
		}
		if err != nil {
			log.Fatalf("failed to start server: %s", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down server")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		log.Fatalf("failed to shutdown server: %s", err)
	}

	log.Info("server stopped")
}
