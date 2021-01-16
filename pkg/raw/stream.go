package raw

import (
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"net"
)

type Stream struct {
	conn   net.Conn
	sendCh chan Message
	ID     uint32
}

func NewStream(conn net.Conn, sendCh chan Message) (*Stream, error) {
	s := &Stream{
		conn:   conn,
		sendCh: sendCh,
		ID:     getID(conn.RemoteAddr().String()),
	}
	go s.recv()
	return s, nil
}

func getID(addr string) uint32 {
	i := strings.LastIndex(addr, ":")
	id := addr[i+1:]
	parsed, err := strconv.Atoi(id)
	if err != nil {
		log.Fatal(err.Error())
	}

	return uint32(parsed)
}

func (s *Stream) recv() {
	b := make([]byte, 1096)
	for {
		n, err := s.conn.Read(b)

		if n == 0 || err != nil {
			//upstream DC, propagate down
			h := Header(make([]byte, HeaderSize))
			h.Encode(TERM, 0, s.ID)
			msg := Message{
				Header:  h,
				Payload: nil,
			}

			s.sendCh <- msg
			s.conn.Close()

			break
		}

		if n > 0 {
			h := Header(make([]byte, HeaderSize))
			h.Encode(TICK, uint32(n), s.ID)
			payload := make([]byte, n)
			copy(payload, b[:n])
			msg := Message{
				Header:  h,
				Payload: payload,
			}
			s.sendCh <- msg
		}

	}
}
