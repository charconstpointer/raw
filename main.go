package main

import (
	"log"
	"net"
	"time"

	"github.com/charconstpointer/raw/pkg/raw"
)

func main() {
	s := raw.NewServer()
	go func(s *raw.Server) {
		err := s.Listen(":9999")
		log.Fatal(err.Error())
	}(s)

	go func(s *raw.Server) {
		err := s.ListenUp(":9998")
		log.Fatal(err.Error())
	}(s)
	time.Sleep(time.Second)
	up, _ := net.Dial("tcp", ":9998")
	go func(conn net.Conn) {

		b := make([]byte, 1024)
		for {
			n, _ := conn.Read(b)
			if n > 0 {
				log.Println("read from downstream", string(b[:n]))
			}
		}
	}(up)
	c, err := raw.NewClient(":9999")
	if err != nil {
		log.Fatal(err.Error())
	}
	c.Send([]byte("oemgallu"))
	time.Sleep(1 * time.Hour)
}
