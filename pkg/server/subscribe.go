package server

import "sync"

type Subscribe struct {
	topics sync.Map
	mu     sync.Mutex
}

var subscribe = NewSubscribe()

func NewSubscribe() *Subscribe {
	return &Subscribe{}
}

func (s *Subscribe) Subscribe(subsName string) *Topic {
	s.mu.Lock()
	defer s.mu.Unlock()

	var topic *Topic
	t, ok := s.topics.Load(subsName)
	if !ok {
		topic = NewTopic(subsName)
		go topic.BroadcastLoop()

		s.topics.Store(subsName, topic)
	} else {
		topic = t.(*Topic)
	}
	//topic.Subscribe(c)

	return topic
}
