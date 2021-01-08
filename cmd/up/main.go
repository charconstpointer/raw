package main

import (
	"log"

	"github.com/charconstpointer/raw/pkg/raw"
	"github.com/common-nighthawk/go-figure"
)

func main() {
	figure.NewColorFigure("passt", "slant", "cyan", true).Print()

	s, err := raw.NewServer(":6444", ":7000")
	if err != nil {
		log.Fatal(err.Error())
	}

	s.Run()

}
