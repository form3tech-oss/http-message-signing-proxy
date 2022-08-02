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
# URL where the proxy should forward the request to. It can be a server or another proxy.
upstreamTarget: "https://api.form3.tech/v1"

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

# Request signing config
signer:
  # The key id stored on remote server that mapped to the public key
  keyId: ""
  # Location of the private key which will be used to sign requests
  keyFilePath: ""
  # The algorithm used to create a digest for body content, can be either SHA-256 or SHA-512
  bodyDigestAlgo: ""
  # The algorithm used to hash the signature, can be either SHA-256 or SHA-512
  signatureHashAlgo: ""
  # List of headers to create signature from 
  signatureHeaders: 
    - (request-target)
    - host
    - date

# Log config
log:
  level: info
```
