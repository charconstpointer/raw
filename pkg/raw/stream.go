package raw

import (
	"net"
	"sync"
)

type Stream struct {
	conn     net.Conn
	sendLock sync.Mutex
}

func NewStream(conn net.Conn) (*Stream, error) {
	s := &Stream{
		conn: conn,
	}
	return s, nil
}
