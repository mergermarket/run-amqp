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

func (s *stubPublisher) PublishWithOptions(message []byte, pattern string, options PublishOptions) error {
	s.publishCalled = true
	s.publishCalledWithMessage = string(message)
	s.publishCalleWithPattern = pattern
	s.publishCalledWithOptions = options
	return s.err
}

const testExchangeName = "experts exchange"
