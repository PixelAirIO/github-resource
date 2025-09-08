package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"

	githubresource "github.com/pixel-air/github-resource"
	"github.com/pixel-air/github-resource/kinds/prs"
)

func init() {
	log.SetFlags(0)
	log.Println("Resource maintained by Pixel Air IO (https://github.com/PixelAirIO)")
}

func main() {
	stdin, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("failed to stdin: %v", err)
	}

	var req githubresource.BaseRequest
	err = json.Unmarshal(stdin, &req)
	if err != nil {
		log.Fatalf("failed to unmarshal to base request: %v", err)
	}

	switch strings.ToLower(req.Source.Kind) {
	case "prs":
		prs.Check(stdin)
	default:
		log.Fatalf("unknown kind: %s", req.Source.Kind)
	}
}
