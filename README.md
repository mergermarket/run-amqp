# RunAMQP

[![Build Status](https://travis-ci.org/mergermarket/run-amqp.svg?branch=master)](https://travis-ci.org/mergermarket/run-amqp)[![GoDoc](https://godoc.org/github.com/mergermarket/run-amqp?status.svg)](https://godoc.org/github.com/mergermarket/run-amqp)

![Run AMQP](http://i.imgur.com/ZOyxDrr.png)

**It's tricky!** to set up rabbit mq for pub/sub in a largish distributed system. This library sets up exchanges/queues in a very opinionated way described below.

Given this set up, the library provides simple interfaces to register `MessageHandler` instances that you provide to consume messages.

[The best place for documentation is of course go doc](https://godoc.org/github.com/mergermarket/run-amqp)

## Running Tests

Prerequisites:
- Install Golang and workspace as described in https://golang.org/doc/code.html
- Install Docker and Docker Compose
- Checkout RunAMQP into: `$GOPATH/src/github.com/mergermarket/run-amqp`.

Run all tests:

    ```
    ./scripts/docker-run-all-tests.sh
    ```

## Test Harness Application

A test harness app exists in `/sample/` so you can play around with it a bit:

    ```
    docker compose run --service-ports sampleapp
    ```

View the test harness app at `<URL>/entry`

    ```
    http://localhost:8080/entry
    ```

## Usage

Read the [Godocs](https://godoc.org/github.com/mergermarket/run-amqp) for a comprehensive guide on how to implement RunAMQP.

A good place to see RunAMQP used is in the test harness app mentioned above.

## Contributing

How to create a new version of this library:

- Branch this git repository
- Make your changes in the branch and run the tests.
- Once the libraries tests pass, test the changes one of the services that use this library.
  - Copy this library's branched directory inside that service's directory
  - Use golang's module replace to use your local copy.`go mod edit -replace github.com/mergermarket/gotools=gotools`
- Once you are satisfied that it works, commit, push, and tag main with the new version number
- Update the services with the new version by running `go get -u; go mod tidy` in each service.
