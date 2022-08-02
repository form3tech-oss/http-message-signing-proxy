package proxy

import (
	"github.com/gin-gonic/gin"
)

type Handler interface {
	Health(c *gin.Context)
	ForwardRequest(c *gin.Context)
}

type handler struct {
}

func NewHandler() Handler {
	return &handler{}
}

func (h *handler) Health(c *gin.Context) {

}

func (h *handler) ForwardRequest(c *gin.Context) {
	//proxy := &httputil.ReverseProxy{}
}
