package main

import (
	"log"
	"time"

	"github.com/charconstpointer/raw/pkg/raw"
)

func main() {
	s := raw.Server{}
	go func(s raw.Server) {
		err := s.Listen(":9999")
		log.Fatal(err.Error())
	}(s)

	c, err := raw.NewClient(":9999")
	if err != nil {
		log.Fatal(err.Error())
	}
	c.Send([]byte("oemgallu"))
	time.Sleep(1 * time.Hour)
}
