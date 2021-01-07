package raw

import (
	"bytes"
	"io"
	"log"
	"net"

	"github.com/pkg/errors"
)

type Server struct {
	upstreams  map[uint32]net.Conn
	upstream   net.Conn
	downstream net.Conn
}

func NewServer(upaddr string, downaddr string) (*Server, error) {
	l, err := net.Listen("tcp", downaddr)
	if err != nil {
		return nil, errors.Wrap(err, "cannot start downstream listener")
	}
	log.Printf("✔waiting for downstream to connect on %s", downaddr)
	downstream, err := l.Accept()
	if err != nil {
		return nil, errors.Wrap(err, "cannot accept downstream connection")
	}

	s, err := net.Listen("tcp", upaddr)
	if err != nil {
		return nil, errors.Wrap(err, "cannot start accepting upstream connections")
	}
	log.Printf("✔read to accept upstream connections on %s", upaddr)
	conn, err := s.Accept()
	if err != nil {
		return nil, errors.Wrap(err, "cannot accept upstream connection")
	}
	log.Printf("🙋‍♀️%s connected", conn.RemoteAddr().String())
	server := &Server{
		downstream: downstream,
		upstream:   conn,
	}
	return server, nil
}

func (s *Server) Run() {
	go s.recv()
	s.send()
}

func (s *Server) recv() {
	h := Header(make([]byte, HeaderSize))
	for {
		io.ReadFull(s.downstream, h)
		mb := make([]byte, int(h.Next()))
		n, _ := io.ReadFull(s.downstream, mb)

		if n > 0 {

			io.Copy(s.upstream, bytes.NewBuffer(mb))
		}
	}
}

func (s *Server) send() {
	b := make([]byte, 1096)
	h := Header(make([]byte, HeaderSize))
	for {
		n, _ := s.upstream.Read(b)
		h.Encode(uint32(n), 1)
		sent := 0
		for sent < HeaderSize {
			n, _ := s.downstream.Write(h)
			sent += n
		}
		if n > 0 {
			io.Copy(s.downstream, bytes.NewBuffer(b[:n]))
		}
	}
}
