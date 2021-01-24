package main

import (
	"context"
	"log"
	"net"

	"github.com/charconstpointer/raw/pkg/raw"
)

func main() {
	up, err := net.Dial("tcp", ":5555")
	if err != nil {
		log.Fatal(err.Error())
	}
	pipe, err := raw.NewPipe(up)
	err = pipe.Run(context.Background())
	if err != nil {
		log.Fatal(err.Error())
	}
}
