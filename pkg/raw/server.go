package raw

import (
	"bytes"
	"io"

	log "github.com/sirupsen/logrus"

	"net"

	"github.com/pkg/errors"
)

type Server struct {
	upstreams  map[uint32]*Stream
	downstream net.Conn
	sendCh     chan Message
}

func NewServer(upaddr string, downaddr string) (*Server, error) {
	l, err := net.Listen("tcp", downaddr)
	if err != nil {
		return nil, errors.Wrap(err, "cannot start downstream listener")
	}
	log.Infof("✔waiting for downstream to connect on %s", downaddr)
	downstream, err := l.Accept()
	if err != nil {
		return nil, errors.Wrap(err, "cannot accept downstream connection")
	}

	s, err := net.Listen("tcp", upaddr)
	if err != nil {
		return nil, errors.Wrap(err, "cannot start accepting upstream connections")
	}
	log.Infof("✔read to accept upstream connections on %s", upaddr)
	server := &Server{
		downstream: downstream,
		upstreams:  make(map[uint32]*Stream),
		sendCh:     make(chan Message),
	}
	go func(s net.Listener, server *Server) {
		for {
			conn, _ := s.Accept()
			upstream, err := NewStream(conn, server.sendCh)
			if err != nil {
				log.Error(err.Error())
			}

			server.upstreams[upstream.ID] = upstream
			log.Infof("🙋‍♀️%d connected", upstream.ID)
		}

	}(s, server)

	return server, err
}

func (s *Server) Run() {
	go s.recv()
	s.send()
}

func (s *Server) recv() {
	for {
		h := Header(make([]byte, HeaderSize))
		io.ReadFull(s.downstream, h)
		mb := make([]byte, int(h.Next()))
		n, _ := io.ReadFull(s.downstream, mb)
		upstream := s.upstreams[h.ID()]
		if n > 0 {
			io.Copy(upstream.conn, bytes.NewBuffer(mb))
		}
	}
}

func (s *Server) send() {
	for {
		select {
		case msg := <-s.sendCh:
			if msg.Header.Next() == 0 {
				continue
			}
			sent := 0
			for sent < HeaderSize {
				n, _ := s.downstream.Write(msg.Header)
				sent += n
			}
			_, _ = io.Copy(s.downstream, bytes.NewBuffer(msg.Payload))
		}
	}
}
