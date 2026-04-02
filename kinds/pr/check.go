package pr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	gh "github.com/PixelAirIO/github-resource"
)

type checkRequest struct {
	Source  Source  `json:"source"`
	Version version `json:"version"`
}

func (*Pr) Check(stdin []byte) {
	dc := json.NewDecoder(bytes.NewReader(stdin))
	dc.DisallowUnknownFields()

	var request checkRequest
	err := dc.Decode(&request)
	if err != nil {
		log.Fatalf("failed to unmarshal check request: %v", err)
	}

	err = validateSource(&request.Source)
	if err != nil {
		log.Fatalf("validation error: %v", err)
	}

	ghc, err := gh.NewGithubClient(request.Source.Config.Config)
	if err != nil {
		log.Fatalf("failed to create Github client: %v", err)
	}

	newVersion := check(request, ghc)

	out, err := json.Marshal(newVersion)
	if err != nil {
		log.Fatalf("failed to marshal output: %v", err)
	}

	fmt.Println(string(out))
}

func check(request checkRequest, ghc gh.GithubClient) []version {
	ref, err := ghc.LatestCommitForPR(request.Source.Number)
	if err != nil {
		log.Fatalf("failed to get latest commit for PR: %v", err)
	}

	if request.Version.Ref == ref {
		return nil
	}

	return []version{
		version{
			Ref: ref,
		},
	}
}
