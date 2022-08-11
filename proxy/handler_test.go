package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestHandler(t *testing.T) {
	expectedRespBody := "OK"
	mockURL := "mock"

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockReqSigner := NewMockRequestSigner(mockCtrl)
	mockReqSigner.EXPECT().SignRequest(gomock.Any()).DoAndReturn(func(r *http.Request) (*http.Request, error) {
		// We don't test the signer here so we return the request as-is
		return r, nil
	})

	mockMetricPublisher := NewMockMetricPublisher(mockCtrl)
	mockMetricPublisher.EXPECT().MeasureSigningDuration(http.MethodGet, mockURL, gomock.Any())
	mockMetricPublisher.EXPECT().IncrementSignedRequestCount(http.MethodGet, mockURL)

	targetSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(expectedRespBody))
		require.NoError(t, err)
	}))

	rs, err := NewReverseProxy(targetSrv.URL)
	require.NoError(t, err)

	h := NewHandler(rs, mockReqSigner, mockMetricPublisher)

	w := CreateTestResponseRecorder()
	_, e := gin.CreateTestContext(w)
	e.NoRoute(h.ForwardRequest)

	req, err := http.NewRequest(http.MethodGet, mockURL, nil)
	require.NoError(t, err)

	e.ServeHTTP(w, req)

	require.Equal(t, expectedRespBody, w.Body.String())
	require.Equal(t, http.StatusOK, w.Code)
}

type TestResponseRecorder struct {
	*httptest.ResponseRecorder
	closeChannel chan bool
}

func (r *TestResponseRecorder) CloseNotify() <-chan bool {
	return r.closeChannel
}

func CreateTestResponseRecorder() *TestResponseRecorder {
	return &TestResponseRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}
}
