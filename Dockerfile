FROM golang:1.12

RUN go get -u github.com/alecthomas/gometalinter
RUN go get -u github.com/kardianos/govendor
RUN gometalinter --install
WORKDIR /go/src/github.com/mergermarket/run-amqp
ADD . /go/src/github.com/mergermarket/run-amqp
CMD ./build-app.sh
