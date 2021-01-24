package raw

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"

	"golang.org/x/sync/errgroup"
)

type Pipe struct {
	s         map[int]*S
	raw       net.Conn
	unwrapped chan Message
	wrapped   chan Message
	mc        bool
}

func NewPipe(raw net.Conn) (*Pipe, error) {
	return &Pipe{
		raw:       raw,
		unwrapped: make(chan Message),
		wrapped:   make(chan Message),
	}, nil
}

func (p *Pipe) Run(ctx context.Context) error {
	g, _ := errgroup.WithContext(ctx)
	g.Go(p.recvSend)
	return g.Wait()
}

func (p *Pipe) recvSend() error {
	for {
		select {
		case msg := <-p.wrapped:
			recv, e := p.s[int(msg.Header.ID())]
			if !e {
				return fmt.Errorf("cannot find s receiver for id %d", msg.Header.ID())
			}
			recv.sendCh <- msg
		}
	}
	return nil
}

func (p *Pipe) startRaw() error {
	for {

	}
	return nil
}

type S struct {
	ID        int
	raw       net.Conn
	sendCh    chan Message
	unwrapped chan Message
	wrapped   chan Message
}

func NewS(raw net.Conn, unwrapped chan Message, wrapped chan Message) (*S, error) {
	return &S{
		unwrapped: unwrapped,
		wrapped:   wrapped,
	}, nil
}

func (s *S) Run(ctx context.Context, p *Pipe, mc bool) error {
	if mc {
		mc, err := net.Dial("tcp", ":25565")
		if err != nil {
			return err
		}
		s, err := NewS(mc, p.unwrapped, p.wrapped)
		if err != nil {
			log.Fatal(err)
		}
		p.s[1] = s
		log.Println("connected to mc")
	}

	g, _ := errgroup.WithContext(ctx)
	g.Go(s.send)
	g.Go(s.recv)
	return g.Wait()
}

func (s *S) recv() error {
	for {
		h := Header(make([]byte, HeaderSize))
		b := make([]byte, 1096)
		n, _ := s.raw.Read(b)
		if n > 0 {
			h.Encode(TICK, uint32(n), uint32(s.ID))
			msg := Message{
				Header:  h,
				Payload: b[:n],
			}
			s.wrapped <- msg
		}
	}
	return nil
}

func (s *S) send() error {
	for {
		select {
		case msg := <-s.sendCh:
			if msg.Header.MessageType() == TICK {
				n, err := io.Copy(s.raw, bytes.NewBuffer(msg.Payload))
				if err != nil {
					log.Fatal(err.Error())
				}

				if n != int64(len(msg.Payload)) {
					return fmt.Errorf("n different than payload length of %d", len(msg.Payload))
				}
			}
		}
	}
	return nil
}
