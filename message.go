package pipkin

import "context"

type Message struct {
	ctx context.Context

	Value interface{}
	Err   error

	chEcho chan *Message

	Records []record
}

func (m *Message) GetContext() context.Context {
	return m.ctx
}

func (m *Message) Echo() {
	m.chEcho <- m
}

func NewMessage(value interface{}) *Message {
	return NewMessageWithContext(context.Background(), value)
}

func NewMessageWithContext(ctx context.Context, value interface{}) *Message {
	return &Message{
		ctx:    ctx,
		Value:  value,
		chEcho: make(chan *Message, 1),
	}
}

func (m *Message) Wait() *Message {
	return <-m.chEcho
}

type record struct {
	UnitID uint8
	Err    error
}
