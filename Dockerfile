#syntax=docker/dockerfile:1.10.0

ARG GOLANG_VERSION=1.23.1

FROM golang:${GOLANG_VERSION}
ENV CGO_ENABLED=0

WORKDIR /go/src/github.com/mergermarket/run-amqp
COPY *netskope-CA.pem /etc/ssl/certs
COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN go mod tidy
CMD ./build-app.sh
