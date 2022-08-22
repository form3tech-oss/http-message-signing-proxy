package test

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/form3tech-oss/go-http-message-signatures"
	"github.com/form3tech-oss/http-message-signing-proxy/cmd"
	"github.com/stretchr/testify/suite"
)

const (
	// from ../example/config_example.yaml
	keyId     = "6f33b219-137c-467e-9a61-f61040a03363"
	proxyHost = "https://localhost:8080"

	cfgFile        = "../example/config_example.yaml"
	sslCertFile    = "../example/cert.crt"
	sslKeyFile     = "../example/private.key"
	privateKeyFile = "../example/rsa_private_key.pem"
	publicKeyFile  = "../example/rsa_public_key.pub"

	testPath = "/test/path?query=bojack"
	testBody = `{"content": "something"}`
)

type e2eTestSuite struct {
	suite.Suite
}

func (s *e2eTestSuite) msgVerifier() *httpsignatures.MessageVerifier {
	return httpsignatures.NewMessageVerifier(func(keyID string) (crypto.PublicKey, crypto.Hash, error) {
		if keyID == keyId {
			keyBytes, err := os.ReadFile(publicKeyFile)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to open public key file: %w", err)
			}
			decoded, _ := pem.Decode(keyBytes)
			pk, err := x509.ParsePKIXPublicKey(decoded.Bytes)
			return pk, crypto.SHA256, err
		}
		return nil, 0, fmt.Errorf("unknown key")
	})
}

func (s *e2eTestSuite) targetServer(verifier *httpsignatures.MessageVerifier) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := verifier.VerifyRequest(r); err != nil {
			if errors.Is(err, &httpsignatures.VerificationError{}) {
				// Signature verification error
				w.WriteHeader(http.StatusBadRequest)
				s.NoError(writeBody(w, errResp{Message: err.Error()}))
			} else {
				// Unknown error, the following statement should fail
				s.NoError(err)
			}
		} else {
			// Valid signature, echo the request back
			w.WriteHeader(http.StatusOK)
			b, err := io.ReadAll(r.Body)
			s.NoError(err)
			s.NoError(writeBody(w, successResp{
				Path:   r.URL.Path + "?" + r.URL.RawQuery,
				Method: r.Method,
				Header: r.Header,
				Body:   string(b),
			}))
		}
	}))
}

func (s *e2eTestSuite) runProxy(upstreamTarget string) {
	// Use the config file, ssl key pair and signing key from the ../example directory
	rootCmd := cmd.NewRootCmd()
	rootCmd.SetArgs(append(
		[]string{"--config", cfgFile},
		genSetFlags(map[string]string{
			"server.ssl.certFilePath":  sslCertFile,
			"server.ssl.keyFilePath":   sslKeyFile,
			"proxy.signer.keyFilePath": privateKeyFile,
			"proxy.upstreamTarget":     upstreamTarget,
		})...,
	))
	go func() {
		s.NoError(rootCmd.Execute())
	}()

	// Wait a brief time to ensure the server is running
	time.Sleep(2 * time.Second)
}

func (s *e2eTestSuite) SetupSuite() {
	// Setup test target server that verify request signature
	targetSrv := s.targetServer(s.msgVerifier())

	// Run proxy server that points to the test target above
	s.runProxy(targetSrv.URL)

	// Skip cert verification because we use a self-signed certificate
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
}

func (s *e2eTestSuite) TestHealth() {
	req, err := http.NewRequest(http.MethodGet, proxyHost+"/-/health", nil)
	s.NoError(err)
	r, err := http.DefaultClient.Do(req)
	s.NoError(err)
	b, err := io.ReadAll(r.Body)
	s.NoError(err)
	s.JSONEq(`{"status": "up"}`, string(b))
}

func (s *e2eTestSuite) TestProxy() {
	tests := []struct {
		name           string
		method         string
		header         http.Header
		body           string
		expectedStatus int
	}{
		{
			"GET with multiple valid headers",
			http.MethodGet,
			genDefaultHeader(),
			"",
			http.StatusOK,
		},
		{
			"PUT with multiple valid headers",
			http.MethodPut,
			genDefaultHeader(),
			testBody,
			http.StatusOK,
		},
		{
			"POST with multiple valid headers",
			http.MethodPost,
			genDefaultHeader(),
			testBody,
			http.StatusOK,
		},
		{
			"PATCH with multiple valid headers",
			http.MethodPatch,
			genDefaultHeader(),
			testBody,
			http.StatusOK,
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest(test.method, proxyHost+testPath, strings.NewReader(test.body))
			s.NoError(err)
			req.Header = test.header
			r, err := http.DefaultClient.Do(req)
			s.NoError(err)
			s.Equal(test.expectedStatus, r.StatusCode)

			if test.expectedStatus == http.StatusOK {
				resp, err := readHttpResp[successResp](r)
				s.NoError(err)

				// Check if the original request is preserved
				s.Equal(test.method, resp.Method)
				s.Equal(testPath, resp.Path)
				s.Equal(test.body, resp.Body)
				// ignore host header because it changes
				origHeader := test.header.Clone()
				origHeader.Del("host")
				s.Truef(headerContains(resp.Header, origHeader), "some original headers are not preserved")
			}
		})
	}
}

func TestE2ETestSuite(t *testing.T) {
	suite.Run(t, new(e2eTestSuite))
}

func genSetFlags(m map[string]string) []string {
	var flags []string
	for k, v := range m {
		flags = append(flags, "--set", k+"="+v)
	}
	return flags
}

func genHeader(headerMap map[string]string) http.Header {
	h := http.Header{}
	for k, v := range headerMap {
		h.Set(k, v)
	}
	return h
}

func genDefaultHeader() http.Header {
	return genHeader(map[string]string{
		"host":   proxyHost,
		"date":   time.Now().Format(time.RFC1123),
		"accept": "application/json",
	})
}

// check whether all elements in h2 are present in h1
func headerContains(h1, h2 http.Header) bool {
	for k := range h2 {
		if h2.Get(k) != h1.Get(k) {
			return false
		}
	}
	return true
}
