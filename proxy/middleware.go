package proxy

import (
	"errors"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	AccessControlAllowOriginHeader string = "Access-Control-Allow-Origin"
)

func RecoverMiddleware(metricPublisher MetricPublisher) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.WithFields(log.Fields{
					"error":       err,
					"stack_trace": string(debug.Stack()),
				}).Errorf("uncaught panic")
				metricPublisher.IncrementInternalErrorCount(c.Request.Method, c.Request.URL.Path)

				// Check for a broken connection
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					var se *os.SyscallError
					if errors.As(ne, &se) {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				if brokenPipe {
					// If the connection is dead, we can't write a status to it.
					c.Abort()
				} else {
					c.AbortWithStatus(http.StatusInternalServerError)
				}
			}
		}()
		c.Next()
	}
}

func LogAndMetricsMiddleware(metricPublisher MetricPublisher) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		metricPublisher.IncrementTotalRequestCount(c.Request.Method, c.Request.URL.Path)
		c.Next()

		latency := time.Since(start)
		metricPublisher.MeasureTotalDuration(c.Request.Method, c.Request.URL.Path, latency.Seconds())

		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}

		log.WithFields(log.Fields{
			"method":      c.Request.Method,
			"path":        path,
			"latency":     latency.String(),
			"status_code": c.Writer.Status(),
			"client_ip":   c.ClientIP(),
		}).Info("request summary")
	}
}

func CORSMiddleware(accessControlAllowOrigin string) gin.HandlerFunc {
	if accessControlAllowOrigin != "" {
		return func(c *gin.Context) {
			c.Writer.Header().Set(AccessControlAllowOriginHeader, accessControlAllowOrigin)
		}
	}
	return func(_ *gin.Context) {}
}
