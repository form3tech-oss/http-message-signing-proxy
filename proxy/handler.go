package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/form3tech-oss/https-signing-proxy/config"
	"github.com/gin-gonic/gin"
)

type Handler interface {
	Health(c *gin.Context)
	ForwardRequest(c *gin.Context)
}

type handler struct {
	proxy *httputil.ReverseProxy
}

func NewHandler(proxy *httputil.ReverseProxy) Handler {
	return &handler{
		proxy: proxy,
	}
}

func (h *handler) Health(c *gin.Context) {

}

func (h *handler) ForwardRequest(c *gin.Context) {
	h.proxy.ServeHTTP(c.Writer, c.Request)
}

func NewProxy(cfg config.ProxyConfig) (*httputil.ReverseProxy, error) {
	upstreamURL, err := url.Parse(cfg.UpstreamTarget)
	if err != nil {
		return nil, fmt.Errorf("failed to parse upstream target: %w", err)
	}

	rp := httputil.NewSingleHostReverseProxy(upstreamURL)
	rp.Director = func(req *http.Request) {
		// TODO: Request signer here
	}
	return rp, nil
}
