package runamqp

//todo: Once we have some tests around this lets fire off goroutines for every message, rather than long-living workers that we have to manage like now
func startWorkers(work <-chan Message, handler MessageHandler, numberOfWorkers int, logger logger) {
	logger.Debug("Registering", numberOfWorkers, "workers for", handler.Name())
	for w := 0; w < numberOfWorkers; w++ {
		go func() {
			for msg := range work {
				handler.Handle(msg)
			}
			logger.Error("Messages channel was closed, worker for", handler.Name(), "has stopped")
		}()
	}

}
