package proxy

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler interface {
	Health(c *gin.Context)
	ForwardRequest(c *gin.Context)
}

type handler struct {
	proxy           *ReverseProxy
	reqSigner       RequestSigner
	metricPublisher MetricPublisher
}

func NewHandler(proxy *ReverseProxy, reqSigner RequestSigner, metricPublisher MetricPublisher) Handler {
	return &handler{
		proxy:           proxy,
		reqSigner:       reqSigner,
		metricPublisher: metricPublisher,
	}
}

func (h *handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "up"})
}

func (h *handler) ForwardRequest(c *gin.Context) {
	req := c.Request.Clone(c)
	req.Host = h.proxy.TargetHost
	req.Header.Set("Host", h.proxy.TargetHost)

	start := time.Now()
	signedReq, err := h.reqSigner.SignRequest(req)
	singingDuration := time.Since(start)

	if err != nil {
		errJson := gin.H{"error": err.Error()}
		switch err.(type) {
		case *InvalidRequestError:
			c.AbortWithStatusJSON(http.StatusBadRequest, errJson)
		default:
			h.metricPublisher.IncrementInternalErrorCount(c.Request.Method, c.Request.URL.Path)
			c.AbortWithStatusJSON(http.StatusInternalServerError, errJson)
		}
		return
	}

	h.metricPublisher.MeasureSigningDuration(c.Request.Method, c.Request.URL.Path, singingDuration.Seconds())
	h.metricPublisher.IncrementSignedRequestCount(c.Request.Method, c.Request.URL.Path)
	h.proxy.ServeHTTP(c.Writer, signedReq)
}
