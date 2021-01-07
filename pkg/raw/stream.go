package raw

import (
	log "github.com/sirupsen/logrus"

	"net"
)

type Stream struct {
	conn   net.Conn
	sendCh chan Message
	ID     uint32
}

func NewStream(conn net.Conn) (*Stream, error) {
	s := &Stream{
		conn:   conn,
		sendCh: make(chan Message),
	}
	return s, nil
}

func (s *Stream) recv() {
	b := make([]byte, 1096)
	for {
		n, err := s.conn.Read(b)
		if err != nil {
			log.Fatal(err.Error())
		}

		if n > 0 {
			header := make(Header, HeaderSize)
			header.Encode(uint32(n), s.ID)

			msg := Message{
				Header:  header,
				Payload: b[:n],
			}

			s.sendCh <- msg
			log.Info("enqueued new message")
			log.Println()
		}
	}
}
