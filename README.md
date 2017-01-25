# run-amqp

[![Build Status](https://travis-ci.org/mergermarket/run-amqp.svg?branch=master)](https://travis-ci.org/mergermarket/run-amqp)[![GoDoc](https://godoc.org/github.com/mergermarket/run-amqp?status.svg)](https://godoc.org/github.com/mergermarket/run-amqp)

![Run AMQP](http://i.imgur.com/ZOyxDrr.png)

**It's tricky!** to set up rabbit mq for pub/sub in a largish distributed system. This library sets up exchanges/queues in a very opinionated way described below.

Given this set up, the library provides simple interfaces to register `MessageHandler` instances that you provide to consume messages.

[The best place for documentation is of course go doc](https://godoc.org/github.com/mergermarket/run-amqp)

## running tests

You will need
- Go installed and workspace set up as described in https://golang.org/doc/code.html
- Docker & Docker Compose

Check out into `$GOPATH/src/github.com/mergermarket/run-amqp`.

`docker-compose run runamqp`

Specific test

`docker-compose run runamqp go test -run=TestRequeue_DLQ_Message_After_Retries`

There is a test harness app in `sample` so you can play around with it a bit

`docker-compose run --service-ports sampleapp`

## Exchanges & Queues autocreated

todo!