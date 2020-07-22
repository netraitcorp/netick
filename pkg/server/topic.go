package server

import (
	"fmt"
	"sync"
)

type Topic struct {
	name           string
	conns          sync.Map
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
		t.conns.Range(func(key, value interface{}) bool {
			//value.(*Conn).Write(data)
			fmt.Println(data)
			return true
		})
	}
}

/*
func (t *Topic) Subscribe(c *Conn) {
	t.conns.Store(c.ID(), c)
}

*/

func (t *Topic) UnSubscribe(id interface{}) {
	t.conns.Delete(id)
}
