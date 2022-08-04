package proxy

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler interface {
	Health(c *gin.Context)
	ForwardRequest(c *gin.Context)
}

type handler struct {
	proxy     *ReverseProxy
	reqSigner RequestSigner
}

func NewHandler(proxy *ReverseProxy, reqSigner RequestSigner) Handler {
	return &handler{
		proxy:     proxy,
		reqSigner: reqSigner,
	}
}

func (h *handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "up"})
}

func (h *handler) ForwardRequest(c *gin.Context) {
	req := c.Request.Clone(c)
	req.Host = h.proxy.TargetHost
	req.Header.Set("host", h.proxy.TargetHost)

	signedReq, err := h.reqSigner.SignRequest(req)
	if err != nil {
		errJson := gin.H{"error": err.Error()}
		switch err.(type) {
		case *InvalidRequestError:
			c.AbortWithStatusJSON(http.StatusBadRequest, errJson)
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, errJson)
		}
		return
	}

	h.proxy.ServeHTTP(c.Writer, signedReq)
}
