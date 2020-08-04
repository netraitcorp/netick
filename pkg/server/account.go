package server

import (
	"errors"
	"sync"
)

var (
	ErrAccountNotExists    = errors.New("account: not exists")
	ErrAccountNotSubscribe = errors.New("account: not subscribe topic")
)

type Account struct {
	conn   Conn
	topics sync.Map
}

type Accounts struct {
	sync.Map
	sync.Mutex
}

func (accs *Accounts) CreateAccount(conn Conn) {
	accs.Store(conn.ConnID(), &Account{
		conn: conn,
	})
}

func (accs *Accounts) Subscribe(id int64, topicName string) error {
	account, ok := accs.Load(id)
	if !ok {
		return ErrAccountNotExists
	}
	ac := account.(*Account)
	topic := topics.GetTopicForce(topicName)

	topic.Subscribe(ac)
	ac.topics.Store(topicName, topic)
	return nil
}

func (accs *Accounts) UnSubscribe(id int64, topicName string) error {
	account, ok := accs.Load(id)
	if !ok {
		return ErrAccountNotExists
	}
	ac := account.(*Account)
	topic, ok := ac.topics.Load(topicName)
	if !ok {
		return ErrAccountNotSubscribe
	}

	topic.(*Topic).UnSubscribe(id)
	ac.topics.Delete(topicName)

	return nil
}

func (accs *Accounts) RemoveAccount(id int64) {
	acc, ok := accs.Load(id)
	if !ok {
		return
	}
	acc.(*Account).topics.Range(func(_, value interface{}) bool {
		value.(*Topic).UnSubscribe(id)
		return true
	})
	accs.Delete(id)
}

var accounts = &Accounts{}
