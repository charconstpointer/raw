package main

import (
	"bytes"
	"io"
	"log"
	"net"

	"github.com/charconstpointer/raw/pkg/raw"
)

func main() {
	l, _ := net.Listen("tcp", ":7000")
	downstream, _ := l.Accept()
	s, _ := net.Listen("tcp", ":6444")
	conn, _ := s.Accept()
	log.Println("newone")

	go recv(conn, downstream)
	send(conn, downstream)
}

func recv(conn net.Conn, mc net.Conn) {
	h := raw.Header(make([]byte, raw.HeaderSize))
	for {
		io.ReadFull(mc, h)
		log.Println("packet to", h.ID())
		mb := make([]byte, int(h.Next()))
		n, _ := io.ReadFull(mc, mb)

		if n > 0 {
			// log.Println(n)
			io.Copy(conn, bytes.NewBuffer(mb))
		}
	}
}

func send(conn net.Conn, mc net.Conn) {
	b := make([]byte, 1096)
	h := raw.Header(make([]byte, raw.HeaderSize))
	for {
		n, _ := conn.Read(b)
		h.Encode(uint32(n), 1)
		sent := 0
		for sent < raw.HeaderSize {
			n, _ := mc.Write(h)
			sent += n
		}
		if n > 0 {
			// log.Println(n)
			io.Copy(mc, bytes.NewBuffer(b[:n]))
		}
	}
}
