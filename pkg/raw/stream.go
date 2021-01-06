package raw

import "io"

type Stream struct {
	conn io.ReadWriteCloser
}

func NewStream(conn io.ReadWriteCloser) (*Stream, error) {
	s := &Stream{
		conn: conn,
	}
	return s, nil
}
