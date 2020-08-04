package server

import (
	"fmt"
	"sync"
)

type Topic struct {
	name           string
	accs           sync.Map
	broadcastQueue chan []byte
}

func NewTopic(name string) *Topic {
	t := &Topic{
		name:           name,
		broadcastQueue: make(chan []byte, 16),
	}
	return t
}

func (t *Topic) BroadcastLoop() {
	for {
		data := <-t.broadcastQueue
		t.accs.Range(func(key, value interface{}) bool {
			fmt.Println(data)

			return true
		})
	}
}

func (t *Topic) Subscribe(acc *Account) {
	t.accs.Store(acc.conn.ConnID(), acc)
}

func (t *Topic) UnSubscribe(id interface{}) {
	t.accs.Delete(id)
}

func (t *Topic) HaveAccount() (exists bool) {
	t.accs.Range(func(key, value interface{}) bool {
		exists = true
		return false
	})
	return
}

type Topics struct {
	sync.Map
	sync.Mutex
}

func (t *Topics) RemoveTopic(name string) {
	t.Delete(name)
}

func (t *Topics) GetTopicForce(name string) *Topic {
	var (
		topic interface{}
		ok    bool
	)
	topic, ok = t.Load(name)
	if !ok {
		t.Lock()
		topic, ok = t.Load(name)
		if !ok {
			topic = NewTopic(name)
		}
		t.Unlock()
	}
	return topic.(*Topic)
}

var topics = &Topics{}
