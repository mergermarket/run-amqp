// Package runamqp provides two general interfaces for working with AMQP, consuming and publishing.
package runamqp

/*
When you create a consumer you will recieve a read-only channel of *Messages which you can then process and then run the standard AMQP commands like Ack, nackCalls, etc.

Producers allow you to publish content to an exchange with a routing key

In addition the library will create dead-letter-exchanges (DLE) and dead-letter-queues according to your configuration.

To get around buffer limits there is also an exchange made to put content into at high load, this is handled for you automatically.

When you retry you can specify a delay and exchanges are made to facilitate this.

In theory though you shouldn't have to "care" about these details, just use the API provided.
*/
