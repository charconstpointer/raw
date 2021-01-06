package raw

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"

	"golang.org/x/sync/errgroup"
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
	return &Server{
		streams: make(map[uint32]*Stream),
	}
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

	ctx, _ := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	g.Go(s.send)
	g.Go(s.recv)
	return g.Wait()

}
func (s *Server) send() error {
	for {
		select {
		case msg := <-s.sendCh:
			sent := 0
			for sent < len(msg.H) {
				s.downstream.Write(msg.H)
			}

			n, err := io.Copy(s.downstream, msg.Payload)
			if err != nil {
				log.Fatalf("could not propagate msg downstream, %s", err.Error())
			}

			log.Printf("wrote %d bytes downstream", n)
		}
	}
}

func (s *Server) recv() error {
	hb := Header(make([]byte, HeaderSize))
	for {
		//read header
		n, err := io.ReadFull(s.downstream, hb)

		if n > 0 {
			fmt.Println("request from", hb.ID())
			fmt.Println("header :", hb.Next())
			if err != nil {
				fmt.Println("failed to read header")
				return err
			}

			//read body
			mb := make([]byte, int(hb.Next()))
			log.Println(len(mb))
			_, err = io.ReadFull(s.downstream, mb)

			upstream, exists := s.streams[hb.ID()]
			if !exists {
				log.Fatal("upstream does not exists", hb.ID())
			}
			_, err := io.Copy(upstream.conn, bytes.NewReader(mb))
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
		s.streams[4444] = stream
		go s.handleUpstream(stream)
	}

}

func (s *Server) handleUpstream(stream *Stream) {
	b := make([]byte, 1024)
	for {
		n, _ := stream.conn.Read(b)
		if n > 0 {
			log.Printf("received message from upstream of len %d", n)
			h := Header{}
			h.Encode(uint32(n), ID(stream.conn))
			payload := bytes.NewBuffer(b[:n])
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
