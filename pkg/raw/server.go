package raw

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
)

type Server struct {
	downstream net.Conn
	streams    map[uint32]*Stream
	sendCh     chan Msg
}
type Msg struct {
	H       Header
	Payload io.Reader
}

func NewServer() *Server {
	s := &Server{
		streams: make(map[uint32]*Stream),
		sendCh:  make(chan Msg),
	}

	return s
}

func (s *Server) Listen(addr string) error {
	conn, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err.Error())
	}

	downstream, err := conn.Accept()
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println("downstream connected🥳")

	s.downstream = downstream

	go s.send()
	go s.recv()
	return nil
}
func (s *Server) send() error {
	for {
		select {
		case msg := <-s.sendCh:
			sent := 0
			for sent < len(msg.H) {
				n, err := s.downstream.Write(msg.H[sent:])
				if err != nil {
					log.Fatal(err.Error())
				}
				sent += n
			}
			// Send data from a body if given
			if msg.Payload != nil {
				_, err := io.Copy(s.downstream, msg.Payload)
				if err != nil {
					log.Fatal(err.Error())
				}
			}

		}
	}
}

func (s *Server) recv() error {
	hb := Header(make([]byte, HeaderSize))
	for {
		//read header
		n, err := io.ReadFull(s.downstream, hb)

		if n > 0 {
			if err != nil {
				fmt.Println("failed to read header")
				return err
			}

			//read body
			mb := make([]byte, int(hb.Next()))
			n, err = io.ReadFull(s.downstream, mb)
			log.Printf("expected to read %d, got %d", hb.Next(), n)
			upstream, exists := s.streams[4]
			if !exists {
				log.Fatal("upstream does not exists", hb.ID())
			}

			n, err := io.Copy(upstream.conn, bytes.NewReader(mb))
			if err != nil {
				log.Fatal(err.Error())
			}

			log.Printf("sent %d bytes to client, expected to send %d \r\n", n, hb.Next())
			if err != nil {
				log.Fatal(err.Error())
			}
		}
	}
}
func (s *Server) ListenUp(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Println("new client connected")

		stream, err := NewStream(conn)
		if err != nil {
			return errors.New("cannot create new stream")
		}
		id := ID(conn)
		log.Println("added new upstream", id)
		s.streams[4] = stream
		go s.handleUpstream(stream)
	}

}

func (s *Server) handleUpstream(stream *Stream) {
	b := make([]byte, 32*1024)
	for {
		n, _ := stream.conn.Read(b)
		if n > 0 {
			log.Printf("received message from upstream of len %d", n)
			h := Header(make([]byte, HeaderSize))
			h.Encode(uint32(n), ID(stream.conn))
			payload := bytes.NewBuffer(b[:n])
			log.Printf("payload len %d", h.Next())
			s.sendCh <- Msg{
				H:       h,
				Payload: payload,
			}
		}
	}
}

func ID(conn net.Conn) uint32 {
	index := strings.LastIndex(conn.RemoteAddr().String(), ":") + 1
	port := conn.RemoteAddr().String()[index:]

	parsed, err := strconv.Atoi(port)
	if err != nil {
		log.Fatal(err.Error())
	}
	return uint32(parsed)
}
