package pipkin

import (
	"testing"
	"context"
	"time"
	"log"
)

func TestNewProcess(t *testing.T) {
	type test struct {
		id int
		s  string
	}

	p := NewProcess(context.Background())

	u1, ch := NewEnterUnit(func(msg *Message) (uint8, error) {
		t := msg.Value.(test)
		log.Print("unit1", t)
		t.s += "11"
		msg.Value = t
		time.Sleep(time.Second)
		return 2, nil
	}, 1, 0, 2, 0)

	u2 := NewUnit(func(msg *Message) (uint8, error) {
		t := msg.Value.(test)
		log.Print("unit2", t)
		t.s += "22"
		msg.Value = t
		msg.Echo()
		time.Sleep(time.Second)
		return 0, nil
	}, 2, 0, 1, 0)

	p.Use(u1, u2)

	p.Run()

	for i := 0; i < 10; i++ {
		val := test{
			id: i,
		}
		m := NewMessage(val)
		ch <- m
		go func() {
			m2 := m.Wait()

			log.Print("echo ", m2.Value)
		}()
	}

	time.Sleep(time.Minute)
}
