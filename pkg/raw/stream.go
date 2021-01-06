package raw

import (
	"net"
)

type Stream struct {
	conn net.Conn
}

func NewStream(conn net.Conn) (*Stream, error) {
	s := &Stream{
		conn: conn,
	}
	return s, nil
}
