package main

import (
	"context"
	"log"

	"github.com/charconstpointer/raw/pkg/raw"
)

func main() {
	pipe, err := raw.NewPipe(true)
	if err := pipe.Run(context.Background()); err != nil {
		log.Fatal(err.Error())
	}
	log.Fatal(err.Error())
}
