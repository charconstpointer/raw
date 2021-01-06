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
	conn    net.Listener
	streams map[uint32]*Stream
}

func NewServer() *Server {
	return &Server{
		streams: make(map[uint32]*Stream),
	}
}
func (s *Server) Listen(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s.conn = listener
	conn, err := s.conn.Accept()
	log.Println("new client connected")
	if err != nil {
		return err
	}
	hb := Header(make([]byte, HeaderSize))

	for {
		//read header
		n, err := io.ReadFull(conn, hb)

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
			_, err = io.ReadFull(conn, mb)
			fmt.Printf("body %s\n", mb)

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

	s.conn = listener
	for {
		conn, err := s.conn.Accept()
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
