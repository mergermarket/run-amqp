# run-amqp

[![GoDoc](https://godoc.org/github.com/mergermarket/run-amqp?status.svg)](https://godoc.org/github.com/mergermarket/run-amqp)

![Run AMQP](http://i.imgur.com/ZOyxDrr.png)

**It's tricky!** to set up rabbit mq for pub/sub in a largish distributed system. This library sets up exchanges/queues in a very opinionated way described below.

Given this set up, the library provides simple interfaces to register `MessageHandler` instances that you provide to consume messages.

[The best place for documentation is of course go doc](https://godoc.org/github.com/mergermarket/run-amqp)

## Exchanges & Queues autocreated

todo!