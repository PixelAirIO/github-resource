package prs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	gh "github.com/PixelAirIO/github-resource"
)

type inRequest struct {
	Source  Source  `json:"source"`
	Version version `json:"version"`
}

type inResponse struct {
	Version version `json:"version"`
}

func (*Prs) In(stdin []byte, dest string) {
	dc := json.NewDecoder(bytes.NewReader(stdin))
	dc.DisallowUnknownFields()

	request := inRequest{}
	err := dc.Decode(&request)
	if err != nil {
		log.Fatalf("failed to unmarshal in request: %v", err)
	}

	if request.Version.Prs == "" {
		log.Fatal("empty list of Pr's passed in")
	}

	err = validateSource(&request.Source)
	if err != nil {
		log.Fatalf("validation error: %v", err)
	}

	ghc, err := gh.NewGithubClient(request.Source.Config.Config)
	if err != nil {
		log.Fatalf("failed to create Github client: %v", err)
	}

	prs, err := in(request, ghc)
	if err != nil {
		log.Fatalf("error getting Prs: %v", err)
	}

	prsMsh, err := json.Marshal(prs)
	if err != nil {
		log.Fatalf("error marshaling PRs: %v", err)
	}

	err = os.WriteFile(filepath.Join(dest, "prs.json"), prsMsh, 0644)
	if err != nil {
		log.Fatalf("error writing to 'prs.json': %v", err)
	}

	resp := inResponse{
		Version: request.Version,
	}

	ver, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error marshaling version: %v", err)
	}

	fmt.Println(string(ver))
}

func in(req inRequest, ghc gh.GithubClient) ([]gh.PullRequest, error) {
	if req.Version.Prs == "none" {
		return []gh.PullRequest{}, nil
	}

	prs := strings.Split(req.Version.Prs, ",")
	if len(prs) == 0 {
		return nil, errors.New("got an empty list of Pr's after trying to parse the version")
	}

	remotePrs, err := ghc.ListPullRequests(
		req.Source.Owner,
		req.Source.Repo,
		req.Source.States,
		req.Source.Labels,
	)
	if err != nil {
		return nil, fmt.Errorf("listPullRequests: %v", err)
	}

	filteredPrs := []gh.PullRequest{}
	for _, p := range remotePrs {
		if slices.Contains(prs, p.Number) {
			filteredPrs = append(filteredPrs, p)
		}
	}

	if len(filteredPrs) != len(prs) {
		return nil, fmt.Errorf(
			"given PRs (%d) does not match found PRs (%d). One or more PRs likely changed.",
			len(prs),
			len(filteredPrs),
		)
	}

	return filteredPrs, nil
}
