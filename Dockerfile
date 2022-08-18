FROM golang:1.19.0-alpine3.16 AS build-env

WORKDIR /app

COPY ./go.mod ./
COPY ./go.sum ./

RUN go mod download

COPY . ./

RUN go build -o /go/bin/proxy

FROM alpine:3.16

RUN apk add --no-cache ca-certificates && update-ca-certificates

COPY --from=build-env /go/bin/proxy /

ENTRYPOINT ["./proxy"]
