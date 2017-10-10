package runamqp

import "fmt"

func startWorkers(work <-chan Message, handler MessageHandler, maxWorkers int, logger logger) {
	logger.Debug("Delegating work to", maxWorkers, "workers called", handler.Name())

	tokens := make(chan token, maxWorkers)

	go func() {
		for msg := range work {
			tokens <- token{}
			go func(newMessage Message) {

				defer func() {
					if r := recover(); r != nil {
						logger.Error(fmt.Sprintf(`handler: "%s" paniced on message "%s", panic msg: "%v"`, handler.Name(), string(newMessage.Body()), r))
						logger.Error(r)
						err := newMessage.Nack(fmt.Sprintf(`handler "%s" paniced with panic message: "%+v"`, handler.Name(), r))
						if err != nil {
							logger.Error(err)
						}
					}
				}()

				handler.Handle(newMessage)
				<-tokens
			}(msg)
		}
	}()

}

type token struct {
}
