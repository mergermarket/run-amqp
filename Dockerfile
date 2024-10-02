#syntax=docker/dockerfile:1.10.0

ARG GOLANG_VERSION=1.23.1
ARG GOLANG_LINT_VERSION=v1.61.0

FROM golang:${GOLANG_VERSION}
ENV CGO_ENABLED=0

RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin ${GOLANG_LINT_VERSION}

WORKDIR /go/src/github.com/mergermarket/run-amqp
ADD . /go/src/github.com/mergermarket/run-amqp
CMD ./build-app.sh
