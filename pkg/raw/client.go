package raw

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
)

type Client struct {
	conn net.Conn
	ID   uint32
}

func NewClient(addr string) (*Client, error) {
	c, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	id := uint32(rand.Intn(9999))
	log.Println("id", id)
	return &Client{
		conn: c,
		ID:   id,
	}, nil
}

func (c *Client) Send(payload []byte) error {
	h := header(make([]byte, HeaderSize))
	h.encode(uint32(len(payload)), uint16(c.ID))

	sent := 0
	for sent < len(h) {
		n, err := c.conn.Write(h)
		if err != nil {
			log.Fatal(err.Error())
		}

		sent += n
	}

	n, err := io.Copy(c.conn, bytes.NewReader(payload))
	if n != int64(len(payload)) {
		return fmt.Errorf("something went wrong could not send full payload")
	}
	return err
}
