package main

import (
	"io"
	"log"
	"os"

	"github.com/pixel-air/github-resource/factory"
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

	kind := factory.NewKind(stdin)
	kind.Check(stdin)
}
