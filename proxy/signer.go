package proxy

import "net/http"

type RequestSigner interface {
	SignRequest(req *http.Request) (*http.Request, error)
}
