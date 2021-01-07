package raw

import (
	"bytes"
	"io"
	"log"
	"net"

	"github.com/pkg/errors"
)

type Client struct {
	downstream net.Conn
	upstream   net.Conn
	ID         uint32
}

func NewClient(upaddr string, downaddr string) (*Client, error) {
	downstream, err := net.Dial("tcp", downaddr)
	if err != nil {
		return nil, errors.Wrap(err, "could not connect to downstream")
	}
	log.Printf("🎠connected to downstream at %s", downaddr)

	upstream, err := net.Dial("tcp", upaddr)
	if err != nil {
		return nil, errors.Wrap(err, "could not connect to upstream")
	}
	log.Printf("🧵connected to upstream at %s", upaddr)
	id := uint32(4444)
	c := &Client{
		downstream: downstream,
		upstream:   upstream,
		ID:         id,
	}

	return c, nil
}
func (c *Client) Run() {
	go c.recv()
	c.send()
}

func (c *Client) recv() {
	h := Header(make([]byte, HeaderSize))
	b := make([]byte, 1096)
	for {
		n, _ := c.downstream.Read(b)
		h.Encode(uint32(n), 1)
		sent := 0
		for sent < HeaderSize {
			n, _ := c.upstream.Write(h)
			sent += n
		}
		if n > 0 {
			io.Copy(c.upstream, bytes.NewBuffer(b[:n]))
		}
	}
}

func (c *Client) send() {
	h := Header(make([]byte, HeaderSize))
	for {
		io.ReadFull(c.upstream, h)
		if h != nil {
			mb := make([]byte, int(h.Next()))
			n, _ := io.ReadFull(c.upstream, mb)
			if n > 0 {
				io.Copy(c.downstream, bytes.NewBuffer(mb))
			}
		}

	}
}
