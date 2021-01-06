package raw

import (
	"fmt"
	"io"
	"log"
	"net"
)

type Server struct {
	conn net.Listener
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
	hb := header(make([]byte, HeaderSize))

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
		}
	}

}
