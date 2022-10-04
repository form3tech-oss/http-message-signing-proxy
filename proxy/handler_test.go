package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/form3tech-oss/http-message-signing-proxy/test"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestHandler(t *testing.T) {
	tests := []struct {
		name         string
		inputHeaders map[string]string
		headerTestFn func(responseHeader http.Header)
	}{
		{
			"automatic date header injection",
			nil,
			func(h http.Header) {
				_, err := time.Parse(http.TimeFormat, h.Get("Date"))
				require.NoError(t, err)
			},
		},
		{
			"date header present",
			map[string]string{
				"Date": time.Date(1998, time.May, 1, 1, 2, 3, 4, time.UTC).Format(http.TimeFormat),
			},
			func(h http.Header) {
				require.Equal(t, "Fri, 01 May 1998 01:02:03 GMT", h.Get("Date"))
			},
		},
	}

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
	var w *test.TestResponseRecorder
	h := NewHandler(rs, mockReqSigner, mockMetricPublisher)
	_, e := gin.CreateTestContext(w)
	e.NoRoute(
		RecoverMiddleware(mockMetricPublisher),
		LogAndMetricsMiddleware(mockMetricPublisher),
		h.ForwardRequest,
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w = test.NewTestResponseRecorder()

			// Test request
			req, err := http.NewRequest(http.MethodGet, mockURL, nil)
			require.NoError(t, err)

			for k, v := range tt.inputHeaders {
				req.Header.Set(k, v)
			}

			e.ServeHTTP(w, req)

			require.Equal(t, expectedRespBody, w.Body.String())
			require.Equal(t, http.StatusOK, w.Code)
			tt.headerTestFn(w.Header())
		})
	}
}

func TestHandlerCORS(t *testing.T) {
	tests := []struct {
		name                     string
		accessControlAllowOrigin string
	}{
		{
			"no value",
			"",
		},
		{
			"*",
			"*",
		},
		{
			"single domain",
			"https://test",
		},
	}

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
	var w *test.TestResponseRecorder

	for _, tt := range tests {
		h := NewHandler(rs, mockReqSigner, mockMetricPublisher)
		_, e := gin.CreateTestContext(w)
		e.NoRoute(
			RecoverMiddleware(mockMetricPublisher),
			LogAndMetricsMiddleware(mockMetricPublisher),
			CORSMiddleware(tt.accessControlAllowOrigin),
			h.ForwardRequest,
		)

		t.Run(tt.name, func(t *testing.T) {
			w = test.NewTestResponseRecorder()

			// Test request
			req, err := http.NewRequest(http.MethodGet, mockURL, nil)
			require.NoError(t, err)

			e.ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Code)
			require.Equal(t, w.Header().Get(AccessControlAllowOriginHeader), tt.accessControlAllowOrigin)
		})
	}
}

func mockReqSigner(mockCtrl *gomock.Controller) *MockRequestSigner {
	mockReqSigner := NewMockRequestSigner(mockCtrl)
	mockReqSigner.EXPECT().SignRequest(gomock.Any()).DoAndReturn(func(r *http.Request) (*http.Request, error) {
		// We don't test the signer here, so we return the request as-is
		return r, nil
	}).AnyTimes()
	return mockReqSigner
}

func mockMetricPublisher(mockCtrl *gomock.Controller, mockURL string) *MockMetricPublisher {
	mockMetricPublisher := NewMockMetricPublisher(mockCtrl)
	mockMetricPublisher.EXPECT().IncrementTotalRequestCount(http.MethodGet, mockURL).AnyTimes()
	mockMetricPublisher.EXPECT().MeasureSigningDuration(http.MethodGet, mockURL, gomock.Any()).AnyTimes()
	mockMetricPublisher.EXPECT().IncrementSignedRequestCount(http.MethodGet, mockURL).AnyTimes()
	mockMetricPublisher.EXPECT().MeasureTotalDuration(http.MethodGet, mockURL, gomock.Any()).AnyTimes()
	return mockMetricPublisher
}

func testTargetServer(expectedBody string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Date", r.Header.Get("Date"))
		w.Header().Set("haha", "haha")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(expectedBody))
	}))
}
