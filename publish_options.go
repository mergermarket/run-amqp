package runamqp

import "fmt"

// PublishOptions will enable options being sent with the message
type PublishOptions struct {
	// Priority will dictate which messages are processed by the consumers first.  The higher the number, the higher the priority
	Priority uint8
	// PublishToQueue will send the message directly to a specific existing queue and the message will not be routed to any other queue attached to the exchange
	PublishToQueue string
	// Pattern is the routing key between the exchange and queues
	Pattern string
}

func (p PublishOptions) String() string {
	return fmt.Sprintf(`Priority: "%d" Publish to queue: "%s" Pattern "%s"`, p.Priority, p.PublishToQueue, p.Pattern)
}

