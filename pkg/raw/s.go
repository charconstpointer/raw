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
		s:         make(map[int]*S),
	}, nil
}

func (p *Pipe) Run(ctx context.Context) error {
	g, _ := errgroup.WithContext(ctx)
	g.Go(p.recvSend)
	g.Go(p.startRaw)
	return g.Wait()
}

func (p *Pipe) AddS(raw net.Conn) error {
	log.Println("creating new s")
	s, err := NewS(raw, p.unwrapped, p.wrapped)
	id := 1
	p.s[id] = s
	go s.Run(context.Background())
	return err
}

func (p *Pipe) recvSend() error {
	for {
		select {
		case msg := <-p.wrapped:
			log.Println("received new wrapped msg", msg.Header.ID())
			recv, e := p.s[int(msg.Header.ID())]
			if !e {
				return fmt.Errorf("cannot find s receiver for id %d", msg.Header.ID())
			}
			recv.sendCh <- msg
		case msg := <-p.unwrapped:
			log.Println("received new unwrapped msg", msg.Header.ID())
			sent := 0
			for sent < HeaderSize {
				n, _ := p.raw.Write(msg.Header)
				sent += n
			}
			if msg.Header.MessageType() == TERM {
				continue
			}
			n, err := io.Copy(p.raw, bytes.NewBuffer(msg.Payload))
			if err != nil {
				log.Fatal(err.Error())
			}

			if n != int64(len(msg.Payload)) {
				return fmt.Errorf("n different than payload length of %d", len(msg.Payload))
			}
		}
	}
	return nil
}

func (p *Pipe) startRaw() error {
	for {
		h := Header(make([]byte, HeaderSize))
		n, _ := io.ReadFull(p.raw, h)
		log.Println("read", n)
		mb := make([]byte, int(h.Next()))
		n, _ = io.ReadFull(p.raw, mb)
		s, e := p.s[int(h.ID())]
		log.Println("ID", h.ID())
		if !e {
			log.Println("not e")
			mc, err := net.Dial("tcp", ":25565")
			if err != nil {
				log.Fatal(err.Error())
			}
			p.AddS(mc)
			ss, _ := p.s[int(h.ID())]
			s = ss
		}
		log.Println(s.ID)
		h.Encode(TICK, uint32(n), uint32(s.ID))
		msg := Message{
			Header:  h,
			Payload: mb[:n],
		}
		s.sendCh <- msg
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
		raw:       raw,
		unwrapped: unwrapped,
		wrapped:   wrapped,
		sendCh:    make(chan Message),
		ID:        1,
	}, nil
}

func (s *S) Run(ctx context.Context) error {
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
		log.Println("recv", n)
		if n > 0 {
			h.Encode(TICK, uint32(n), uint32(s.ID))
			msg := Message{
				Header:  h,
				Payload: b[:n],
			}
			s.unwrapped <- msg

		}
	}
	return nil
}

func (s *S) send() error {
	for {
		select {
		case msg := <-s.sendCh:
			log.Println("new msg to send")
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
