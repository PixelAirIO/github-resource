package commits_from_prs

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

func (*CommitsFromPrs) Check(stdin []byte) {
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
	commits, err := ghc.LatestCommitsFromPrs(
		request.Source.States,
		request.Source.Labels,
	)
	if err != nil {
		log.Fatalf("failed to get commits from PRs: %v", err)
	}

	if request.Source.Config.ExcludeDrafts {
		filtered := []gh.PRCommit{}
		for _, c := range commits {
			if !c.IsDraft {
				filtered = append(filtered, c)
			}
		}
		commits = filtered
	}

	if request.Source.Config.TargetBranch != "" {
		filtered := []gh.PRCommit{}
		for _, c := range commits {
			if c.TargetBranch == request.Source.Config.TargetBranch {
				filtered = append(filtered, c)
			}
		}
		commits = filtered
	}

	if len(commits) == 0 {
		log.Println("No matching commits found.")
		return []version{}
	}

	versions := make([]version, 0, len(commits))
	for _, c := range commits {
		versions = append(versions, version{
			Ref:        c.Ref,
			Pr:         c.Number,
			CommitDate: c.Date,
		})
	}

	return versions
}
