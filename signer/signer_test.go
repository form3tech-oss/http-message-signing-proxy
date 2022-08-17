package signer

import (
	"crypto"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/form3tech-oss/http-message-signing-proxy/config"
	"github.com/stretchr/testify/require"
)

type errCheckFn func(require.TestingT, error, ...interface{})

func TestGetHashAlgo(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		expected   crypto.Hash
		errCheckFn errCheckFn
	}{
		{
			"sha-256",
			"SHA-256",
			crypto.SHA256,
			require.NoError,
		},
		{
			"sha-256 mixed case",
			"sHa-256",
			crypto.SHA256,
			require.NoError,
		},
		{
			"sha-512",
			"SHA-512",
			crypto.SHA512,
			require.NoError,
		},
		{
			"sha-512 mixed case",
			"sHa-512",
			crypto.SHA512,
			require.NoError,
		},
		{
			"invalid algo",
			"lockpickinglawyer",
			0,
			require.Error,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := getHashAlgo(test.input)
			test.errCheckFn(t, err)
			require.Equal(t, test.expected, actual)
		})
	}
}

func TestLoadKey(t *testing.T) {
	key, err := loadKey("rsa_test.pem")
	require.NoError(t, err)
	require.NotNil(t, key)
}

func TestGetSignatureHeaders(t *testing.T) {
	dummyUrl := "https://oompaloompa.localhost:1234"
	dummyBody := "{\"name\":\"travis\"}"

	tests := []struct {
		name            string
		headerCfg       config.HeadersConfig
		expectedHeaders []string
		errCheckFn      errCheckFn

		// incoming request config
		method  string
		body    io.Reader
		headers map[string]string
	}{
		{
			"GET with only host header",
			config.HeadersConfig{
				SignatureHeaders: []string{"host"},
			},
			[]string{"host"},
			require.NoError,
			http.MethodGet,
			nil,
			map[string]string{
				"host": "foo",
			},
		},
		{
			"GET with special headers",
			config.HeadersConfig{
				IncludeDigest:        true,
				IncludeRequestTarget: true,
				SignatureHeaders:     []string{"host"},
			},
			[]string{"host", requestTargetHeaderKey},
			require.NoError,
			http.MethodGet,
			nil,
			map[string]string{
				"host": "foo",
			},
		},
		{
			"GET with special headers and partially matched signature headers",
			config.HeadersConfig{
				IncludeDigest:        true,
				IncludeRequestTarget: true,
				SignatureHeaders:     []string{"host", "content-type", "content-length"},
			},
			[]string{"host", requestTargetHeaderKey},
			require.NoError,
			http.MethodGet,
			nil,
			map[string]string{
				"host": "foo",
				"tip":  "top",
			},
		},
		{
			"POST with special headers",
			config.HeadersConfig{
				IncludeDigest:        true,
				IncludeRequestTarget: true,
				SignatureHeaders:     []string{"host"},
			},
			[]string{"host", digestHeaderKey, requestTargetHeaderKey},
			require.NoError,
			http.MethodPost,
			strings.NewReader(dummyBody),
			map[string]string{
				"host": "foo",
			},
		},
		{
			"PUT with special headers but no body",
			config.HeadersConfig{
				IncludeDigest:        true,
				IncludeRequestTarget: true,
				SignatureHeaders:     []string{"host"},
			},
			[]string{"host", requestTargetHeaderKey},
			require.NoError,
			http.MethodPut,
			nil,
			map[string]string{
				"host": "foo",
			},
		},
		{
			"POST with special headers but no declared signature headers",
			config.HeadersConfig{
				IncludeDigest:        true,
				IncludeRequestTarget: true,
				SignatureHeaders:     []string{},
			},
			nil,
			require.Error,
			http.MethodPost,
			strings.NewReader(dummyBody),
			nil,
		},
		{
			"POST with special headers but unmatched declared signature headers",
			config.HeadersConfig{
				IncludeDigest:        true,
				IncludeRequestTarget: true,
				SignatureHeaders:     []string{"foo", "bar"},
			},
			nil,
			require.Error,
			http.MethodPost,
			strings.NewReader(dummyBody),
			map[string]string{
				"tip": "top",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest(test.method, dummyUrl, test.body)
			require.NoError(t, err)
			for k, v := range test.headers {
				req.Header.Set(k, v)
			}

			reqSigner, err := NewRequestSigner(config.SignerConfig{
				KeyId:             "dfb4c78a-e141-4144-aa68-8ec605484d63",
				KeyFilePath:       "rsa_test.pem",
				BodyDigestAlgo:    "SHA-256",
				SignatureHashAlgo: "SHA-256",
				Headers:           test.headerCfg,
			})
			require.NoError(t, err)

			actualHeaders, err := reqSigner.(*requestSigner).getSignatureHeaders(req)
			test.errCheckFn(t, err)

			require.Equal(t, test.expectedHeaders, actualHeaders)
		})
	}
}
