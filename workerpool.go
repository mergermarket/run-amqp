package runamqp

func startWorkers(work <-chan Message, handler MessageHandler, maxWorkers int, logger logger) {
	logger.Debug("Delegating work to", maxWorkers, "workers called", handler.Name())

	tokens := make(chan token, maxWorkers)

	go func() {
		for msg := range work {
			tokens <- token{}

			go func(newMessage Message) {
				handler.Handle(newMessage)
				<-tokens
			}(msg)
		}
	}()

}

type token struct {
}
