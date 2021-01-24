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
	pipe      net.Conn
	unwrapped chan Message
	wrapped   chan Message
	init      bool
}

func NewPipe(init bool) (*Pipe, error) {
	return &Pipe{
		unwrapped: make(chan Message),
		wrapped:   make(chan Message),
		s:         make(map[int]*S),
		init:      init,
	}, nil
}

func (p *Pipe) Run(ctx context.Context) error {
	if p.init {
		p.dialPipe(":5555")
	} else {
		p.recvPipe(":5555")
	}

	g, _ := errgroup.WithContext(ctx)
	//if pipe should handle external connections
	if !p.init {
		g.Go(p.recvConn)
	}
	g.Go(p.recvSend)
	g.Go(p.startpipe)
	return g.Wait()
}
func (p *Pipe) dialPipe(addr string) error {
	log.Printf("connecting to a pipe at %s\n", addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatal(err.Error())
	}

	p.pipe = conn
	return err
}

func (p *Pipe) recvPipe(addr string) error {
	log.Printf("waiting for pipes to connect on %s\n", addr)
	up, err := net.Listen("tcp", ":5555")
	if err != nil {
		log.Fatal(err.Error())
	}

	conn, err := up.Accept()
	p.pipe = conn
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Printf("successfully received new pipe connection at %s", addr)
	return err
}

func (p *Pipe) recvConn() error {
	s, err := net.Listen("tcp", ":6000")
	if err != nil {
		log.Fatal(err.Error())
	}
	for {
		log.Println("listening for new conns")
		client, err := s.Accept()
		err = p.AddS(client)
		return err
	}
}

func (p *Pipe) AddS(pipe net.Conn) error {
	s, err := NewS(pipe, p.unwrapped, p.wrapped)
	id := 1
	p.s[id] = s
	go s.Run(context.Background())
	return err
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
		case msg := <-p.unwrapped:
			sent := 0
			for sent < HeaderSize {
				n, _ := p.pipe.Write(msg.Header)
				sent += n
			}
			if msg.Header.MessageType() == TERM {
				continue
			}
			n, err := io.Copy(p.pipe, bytes.NewBuffer(msg.Payload))
			if err != nil {
				log.Fatal(err.Error())
			}

			if n != int64(len(msg.Payload)) {
				return fmt.Errorf("n different than payload length of %d", len(msg.Payload))
			}
		}
	}
}

func (p *Pipe) startpipe() error {
	for {
		h := Header(make([]byte, HeaderSize))
		n, _ := io.ReadFull(p.pipe, h)
		mb := make([]byte, int(h.Next()))
		n, _ = io.ReadFull(p.pipe, mb)
		s, e := p.s[int(h.ID())]
		if !e {
			log.Println("creating new s")
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
}

type S struct {
	ID        int
	pipe      net.Conn
	sendCh    chan Message
	unwrapped chan Message
	wrapped   chan Message
}

func NewS(pipe net.Conn, unwrapped chan Message, wrapped chan Message) (*S, error) {
	return &S{
		pipe:      pipe,
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
		n, _ := s.pipe.Read(b)
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
}

func (s *S) send() error {
	for {
		select {
		case msg := <-s.sendCh:
			log.Println("new msg to send")
			if msg.Header.MessageType() == TICK {
				n, err := io.Copy(s.pipe, bytes.NewBuffer(msg.Payload))
				if err != nil {
					log.Fatal(err.Error())
				}

				if n != int64(len(msg.Payload)) {
					return fmt.Errorf("n different than payload length of %d", len(msg.Payload))
				}
			}
		}
	}
}
