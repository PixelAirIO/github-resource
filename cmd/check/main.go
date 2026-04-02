package main

import (
	"io"
	"log"
	"os"

	"github.com/PixelAirIO/github-resource/factory"
)

func init() {
	log.SetFlags(0)
}

func main() {
	stdin, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("failed to read stdin: %v", err)
	}

	kind := factory.NewKind(stdin)
	kind.Check(stdin)
}
