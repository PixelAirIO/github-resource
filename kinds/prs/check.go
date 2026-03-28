package prs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	gh "github.com/PixelAirIO/github-resource"
)

type checkRequest struct {
	Source  Source  `json:"source"`
	Version version `json:"version"`
}

func (*Prs) Check(stdin []byte) {
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
	prs, err := ghc.ListPullRequests(
		request.Source.Owner,
		request.Source.Repo,
		request.Source.States,
		request.Source.Labels)

	if err != nil {
		log.Fatalf("failed to get pull requests: %v", err)
	}

	if len(prs) == 0 {
		log.Println("No matching PRs found.")
		if request.Version.Prs == "none" {
			return []version{}
		}

		return []version{
			{
				Prs:       "none",
				Timestamp: time.Now(),
			},
		}
	}

	if request.Source.Config.ExcludeDrafts {
		nonDraftPrs := []gh.PullRequest{}
		for _, p := range prs {
			if !p.IsDraft {
				nonDraftPrs = append(nonDraftPrs, p)
			}
		}
		prs = nonDraftPrs
	}

	if request.Source.Config.TargetBranch != "" {
		matchingPrs := []gh.PullRequest{}
		for _, p := range prs {
			if p.TargetBranch == request.Source.Config.TargetBranch {
				matchingPrs = append(matchingPrs, p)
			}
		}
		prs = matchingPrs
	}

	prsVersion := ""
	for _, p := range prs {
		prsVersion += p.Number + ","
	}
	prsVersion = strings.TrimSuffix(prsVersion, ",")

	if prsVersion == request.Version.Prs {
		log.Println("No new PRs found.")
		return []version{}
	}

	newVersion := []version{
		{
			Prs:       prsVersion,
			Timestamp: time.Now(),
		},
	}
	return newVersion
}
