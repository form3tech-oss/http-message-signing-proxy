package signer

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"strings"

	msgsigner "github.com/form3tech-oss/go-http-message-signatures"
	"github.com/form3tech-oss/http-message-signing-proxy/config"
	"github.com/form3tech-oss/http-message-signing-proxy/proxy"
)

const (
	requestTargetHeaderKey = "(request-target)"
	digestHeaderKey        = "digest"
)

type DataError = msgsigner.DataError

type requestSigner struct {
	*msgsigner.MessageSigner
	headerConfig config.HeadersConfig
}

func NewRequestSigner(cfg config.SignerConfig) (proxy.RequestSigner, error) {
	key, err := loadKey(cfg.KeyFilePath)
	if err != nil {
		return nil, err
	}

	signatureHashAlgo, err := getHashAlgo(cfg.SignatureHashAlgo)
	if err != nil {
		return nil, err
	}

	digestHashAlgo, err := getHashAlgo(cfg.SignatureHashAlgo)
	if err != nil {
		return nil, err
	}

	signer, err := msgsigner.NewRSASigner(key, signatureHashAlgo)
	if err != nil {
		return nil, err
	}

	msgSigner, err := msgsigner.NewMessageSigner(digestHashAlgo, signer, cfg.KeyId, msgsigner.Signature)
	if err != nil {
		return nil, err
	}

	return &requestSigner{
		MessageSigner: msgSigner,
		headerConfig:  cfg.Headers,
	}, err
}

func (rs *requestSigner) SignRequest(req *http.Request) (*http.Request, error) {
	headers, err := rs.getSignatureHeaders(req)
	if err != nil {
		return nil, err
	}

	signedReq, err := rs.MessageSigner.SignRequest(req, headers)
	switch err.(type) {
	case *msgsigner.DataError:
		return nil, proxy.NewInvalidRequestError(err)
	default:
		return signedReq, err
	}
}

func (rs *requestSigner) getSignatureHeaders(req *http.Request) ([]string, error) {
	// Get the intersection of request's headers and config's signature headers
	var headers []string
	for _, header := range rs.headerConfig.SignatureHeaders {
		if req.Header.Get(header) != "" {
			headers = append(headers, header)
		}
	}
	if len(headers) == 0 {
		return nil, proxy.NewInvalidRequestError(fmt.Errorf("none of the signature headers found in the request, expected at least one"))
	}

	// Include 'digest' header for PUT, POST and PATCH requests only
	if rs.headerConfig.IncludeDigest && shouldHaveBody(req) {
		headers = append(headers, digestHeaderKey)
	}

	// Include '(request-target)' header
	if rs.headerConfig.IncludeRequestTarget {
		headers = append(headers, requestTargetHeaderKey)
	}

	return headers, nil
}

func getHashAlgo(algo string) (crypto.Hash, error) {
	upper := strings.ToUpper(algo)
	switch upper {
	case crypto.SHA256.String():
		return crypto.SHA256, nil
	case crypto.SHA512.String():
		return crypto.SHA512, nil
	default:
		return 0, fmt.Errorf("unknown algo hash '%s'", algo)
	}
}

func loadKey(keyFile string) (*rsa.PrivateKey, error) {
	rawKey, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read signer key file: %w", err)
	}
	decodedKey, _ := pem.Decode(rawKey)
	key, err := x509.ParsePKCS1PrivateKey(decodedKey.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse signer key file: %w", err)
	}
	return key, nil
}

func shouldHaveBody(req *http.Request) bool {
	switch req.Method {
	case http.MethodPut, http.MethodPost, http.MethodPatch:
		return req.Body != http.NoBody
	default:
		return false
	}
}
