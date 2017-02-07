package runamqp

type stubPublisher struct {
	ready                    bool
	publishCalled            bool
	publishCalledWithMessage string
	publishCalleWithPattern  string
	publishCalledWithOptions PublishOptions
	err                      error
}

func (s *stubPublisher) IsReady() bool {
	return s.ready
}

func (s *stubPublisher) Publish(message []byte, pattern string) error {
	s.publishCalled = true
	s.publishCalledWithMessage = string(message)
	s.publishCalleWithPattern = pattern
	return s.err
}

func (s *stubPublisher) PublishWithOptions(message []byte, options PublishOptions) error {
	s.publishCalled = true
	s.publishCalledWithMessage = string(message)
	s.publishCalledWithOptions = options
	s.publishCalleWithPattern = options.Pattern
	return s.err
}

const testExchangeName = "experts exchange"
