package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

type TestResponseRecorder struct {
	*httptest.ResponseRecorder
	closeChannel chan bool
}

func (r *TestResponseRecorder) CloseNotify() <-chan bool {
	return r.closeChannel
}

func NewTestResponseRecorder() *TestResponseRecorder {
	return &TestResponseRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}
}

type errResp struct {
	Message string
}

type successResp struct {
	Path   string
	Method string
	Header http.Header
	Body   string
}

func writeBody(w http.ResponseWriter, content any) error {
	b, err := json.Marshal(content)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

func readHttpResp[T any](resp *http.Response) (*T, error) {
	var eval T
	err := json.NewDecoder(resp.Body).Decode(&eval)
	if err != nil {
		return nil, err
	}
	return &eval, nil
}
