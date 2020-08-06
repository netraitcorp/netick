package server

import (
	"sync"
)

type Account struct {
	conn Conn
}

func (acc *Account) ID() string {
	return acc.conn.ConnID()
}

func NewAccount(conn Conn) *Account {
	return &Account{
		conn: conn,
	}
}

type Accounts struct {
	accs sync.Map
	mu   sync.Mutex
}

func (as *Accounts) AddAccount(acc *Account) {
	as.accs.Store(acc.ID(), acc)
}

func (as *Accounts) RemoveAccount(id string) {
	as.accs.Delete(id)
}

func NewAccounts() *Accounts {
	as := &Accounts{}
	return as
}

var accounts = NewAccounts()

/*

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

*/
