package proxy

import (
	"fmt"
	"net/http/httputil"
	"net/url"
)

type ReverseProxy struct {
	*httputil.ReverseProxy
	TargetHost string
}

func NewReverseProxy(target string) (*ReverseProxy, error) {
	upstreamURL, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("failed to parse upstream target: %w", err)
	}
	rp := httputil.NewSingleHostReverseProxy(upstreamURL)
	return &ReverseProxy{
		ReverseProxy: rp,
		TargetHost:   upstreamURL.Host,
	}, nil
}
