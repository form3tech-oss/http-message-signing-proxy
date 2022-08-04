# HTTPS signing proxy

## Introduction

The request signing proxy would sit in front of the client and intercept outbound requests, sign them with client's 
private key and transfer signed requests to the server.

![design.png](doc/images/design.png)

More about request signing [here](https://api-docs.form3.tech/tutorial-request-signing.html).

## Run the proxy

The proxy requires `--config` flag, which point to the config file. See [configuration](#configuration) section below.

```shell
./signing-proxy --config <config_file_path>
```

## Configuration

```yaml
# HTTP server config
server:
  # Listening port
  port: 8080
  # If SSL is used, the proxy will receive https request from downstream, terminate it, sign it then establish a new
  # https connection to upstream
  ssl:
    enable: true
    # Location of the proxy's certificate, if SSL is enabled  
    certFilePath: "/etc/ssl/certs/cert.crt"
    # Location of the proxy's private key, if SSL is enabled
    keyFilePath: "/etc/ssl/private/private.key"

# Request forward proxy config
proxy:
  # URL where the proxy should forward the request to. It can be a server or another proxy.
  upstreamTarget: "https://api.form3.tech/v1"
  # Request signing config
  signer:
    # The key id stored on remote server that mapped to the public key
    keyId: "27781930-f2d0-463e-b6cf-0ed0ec0d8dd9"
    # Location of the private key which will be used to sign requests
    keyFilePath: "/etc/form3/private/private.key"
    # The algorithm used to create a digest for body content, can be either SHA-256 or SHA-512
    bodyDigestAlgo: "SHA-256"
    # The algorithm used to hash the signature, can be either SHA-256 or SHA-512
    signatureHashAlgo: "SHA-256"
    # Signature headers config
    headers:
      # For PUT and POST request, whether a digest header should be included.
      includeDigest: true
      # Whether a special (request-target) header should be included.
      includeRequestTarget: true
      # List of headers to create signature from, only those which present in the request will be added. 
      # However, at least one must be specified in the request. 
      # For example, a GET request does not have content-length header, 
      # so the proxy will not include content-length to the signature.
      signatureHeaders: 
        - host
        - date
        - content-length

# Log config
log:
  # Log level
  level: info
  # Log format, can be either 'text' or 'json'
  format: json
```

### Configuration override

One can override any `string` field (list field override is not supported) with `--set` flag or environment variable.

#### Override config using `--set` flag

To override specific fields, `--set key=value` flag can be set multiple times.
For example, `proxy.signer.keyId` and `log.level` in the yaml file above can be overridden by: 

```shell
./signing-proxy --config <config_file_path> \
  --set proxy.signer.keyid=5099392e-3040-40f9-ac70-ce66a9ee0ed6 \
  --set log.level=debug
```

#### Override config using env var

A `a.b.c` field can be automatically overridden by setting a `A_B_C` env var 
(all capitalised and dots replaced by underscore).
For example, `proxy.signer.keyId` and `proxy.signer.bodyDigestAlgo` in the yaml file above can be overridden by: 

```shell
export PROXY_SIGNER_KEYID=5099392e-3040-40f9-ac70-ce66a9ee0ed6
export PROXY_SIGNER_BODYDIGESTALGO=SHA-512
```

## Proxying

Any request coming in the proxy will be signed and forwarded to the upstream target, meaning the host will be replaced 
by the target host and a signature header will be added, the rest of the request is kept as-is.

Additionally, there are two endpoints explicitly exposed by the proxy:

- `GET /-/health` for health check.
- `GET /-/prometheus` for metrics.

Incoming requests like above will not be signed nor forwarded to the upstream target.
