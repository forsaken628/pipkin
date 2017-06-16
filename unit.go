package pipkin

import (
	"context"
	"fmt"
	"time"
	"errors"
	golog "log"
)

var ErrEnd = errors.New("end")

type Unit struct {
	ctx context.Context
	p   *Process

	id     uint8
	buffer uint8
	ch     chan *Message
	fn     HandlerFunc

	concurrent uint8
	timeout    time.Duration //发送超时
}

type HandlerFunc func(message *Message) (uint8, error)

func (u *Unit) runLoop() {
	for {
		select {
		case <-u.ctx.Done():
			return
		case m := <-u.ch:
			func() {
				var next uint8
				defer func() {
					if m.Err == ErrEnd {
						return
					}
					if e := recover(); e != nil {
						if err, ok := e.(error); ok {
							m.Err = err
						} else {
							m.Err = fmt.Errorf("%s", e)
						}
					}
					m.Records = append(m.Records, record{
						UnitID: u.id,
						Err:    m.Err,
					})
					if u.timeout == 0 {
						u.p.units[next].ch <- m
					} else {
						select {
						case <-time.After(u.timeout):
							m.Records = append(m.Records, record{
								UnitID: next,
								Err:    errors.New("timeout"),
							})
							u.p.units[0].ch <- m
						case u.p.units[next].ch <- m:
						}
					}
				}()
				select {
				case <-m.ctx.Done():
					m.Err = m.ctx.Err()
				default:
					next, m.Err = u.fn(m)
					if u.p.units[next]==nil{
						next=0
						m.Err=errors.New("invalid units")
					}
				}
			}()
		}
	}
}

func NewUnit(fn HandlerFunc, id, buffer, concurrent uint8, timeout time.Duration) *Unit {
	u, _ := NewEnterUnit(fn, id, buffer, concurrent, timeout)
	return u
}

func NewEnterUnit(fn HandlerFunc, id, buffer, concurrent uint8, timeout time.Duration) (*Unit, chan<- *Message) {
	u := &Unit{
		id:         id,
		fn:         fn,
		buffer:     buffer,
		concurrent: concurrent,
		timeout:    timeout,
	}
	u.ch = make(chan *Message, buffer)
	return u, u.ch
}

var DefaultErrUnit = NewUnit(func(msg *Message) (uint8, error) {
	golog.Print(msg.Err, msg.Records)
	return 0, ErrEnd
}, 0, 0, 1, 0)
