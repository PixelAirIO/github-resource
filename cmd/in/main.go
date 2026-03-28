package main

import (
	"io"
	"log"
	"os"

	"github.com/PixelAirIO/github-resource/factory"
)

func init() {
	log.SetFlags(0)
	log.Println("Resource maintained by Pixel Air IO (https://github.com/PixelAirIO)")
}

func main() {
	stdin, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("failed to read stdin: %v", err)
	}

	if len(os.Args) < 2 {
		log.Fatal("no output directory provided")
	}

	kind := factory.NewKind(stdin)
	kind.In(stdin, os.Args[1])
}
