package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/form3tech-oss/http-message-signing-proxy/test"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestHandler(t *testing.T) {
	expectedRespBody := "OK"
	mockURL := "mock"
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// Mock dependencies
	mockReqSigner := mockReqSigner(mockCtrl)
	mockMetricPublisher := mockMetricPublisher(mockCtrl, mockURL)

	// Test upstream target that returns 200 OK
	targetSrv := testTargetServer(expectedRespBody)

	// Reverse proxy pointing to test target
	rs, err := NewReverseProxy(targetSrv.URL)
	require.NoError(t, err)

	// Test handler
	h := NewHandler(rs, mockReqSigner, mockMetricPublisher)
	w := test.NewTestResponseRecorder()
	_, e := gin.CreateTestContext(w)
	e.NoRoute(
		RecoverMiddleware(mockMetricPublisher),
		LogAndMetricsMiddleware(mockMetricPublisher),
		h.ForwardRequest,
	)

	req, err := http.NewRequest(http.MethodGet, mockURL, nil)
	require.NoError(t, err)

	e.ServeHTTP(w, req)

	require.Equal(t, expectedRespBody, w.Body.String())
	require.Equal(t, http.StatusOK, w.Code)
}

func mockReqSigner(mockCtrl *gomock.Controller) *MockRequestSigner {
	mockReqSigner := NewMockRequestSigner(mockCtrl)
	mockReqSigner.EXPECT().SignRequest(gomock.Any()).DoAndReturn(func(r *http.Request) (*http.Request, error) {
		// We don't test the signer here so we return the request as-is
		return r, nil
	})
	return mockReqSigner
}

func mockMetricPublisher(mockCtrl *gomock.Controller, mockURL string) *MockMetricPublisher {
	mockMetricPublisher := NewMockMetricPublisher(mockCtrl)
	mockMetricPublisher.EXPECT().IncrementTotalRequestCount(http.MethodGet, mockURL)
	mockMetricPublisher.EXPECT().MeasureSigningDuration(http.MethodGet, mockURL, gomock.Any())
	mockMetricPublisher.EXPECT().IncrementSignedRequestCount(http.MethodGet, mockURL)
	mockMetricPublisher.EXPECT().MeasureTotalDuration(http.MethodGet, mockURL, gomock.Any())
	return mockMetricPublisher
}

func testTargetServer(expectedBody string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(expectedBody))
	}))
}
