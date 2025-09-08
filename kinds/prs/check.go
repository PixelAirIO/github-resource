package prs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	. "github.com/pixel-air/github-resource"
)

type checkRequest struct {
	Source  Source  `json:"source"`
	Version version `json:"version"`
}

func Check(stdin []byte) {
	dc := json.NewDecoder(bytes.NewReader(stdin))
	dc.DisallowUnknownFields()

	var request checkRequest
	err := dc.Decode(&request)
	if err != nil {
		log.Fatalf("failed to unmarshal check request: %v", err)
	}

	err = checkValidate(&request)
	if err != nil {
		log.Fatalf("validation error: %v", err)
	}

	gh, err := NewGithubClient(request.Source.Config.Config)
	if err != nil {
		log.Fatalf("failed to create Github client: %v", err)
	}

	prs, err := gh.ListPullRequests(
		request.Source.Config.Owner,
		request.Source.Config.Repo,
		request.Source.Config.States,
		request.Source.Config.Labels)

	if err != nil {
		log.Fatalf("failed to get pull requests: %v", err)
	}

	prsVersion := ""
	for _, p := range prs {
		prsVersion += strconv.Itoa(p.ID) + ","
	}
	prsVersion = strings.TrimSuffix(prsVersion, ",")

	if prsVersion == request.Version.Prs {
		log.Println("No new PRs found.")
		fmt.Println("[]")
		os.Exit(0)
	}

	newVersion := []version{
		{
			Prs:       prsVersion,
			Timestamp: time.Now(),
		},
	}

	out, err := json.Marshal(newVersion)
	if err != nil {
		log.Fatalf("failed to marshal output: %v", err)
	}

	fmt.Println(string(out))
}

func checkValidate(req *checkRequest) (err error) {
	if req.Source.Config.Owner == "" {
		errors.Join(errors.New("owner field is required"), err)
	}

	if req.Source.Config.Repo == "" {
		errors.Join(errors.New("repository field is required"), err)
	}

	for _, v := range req.Source.Config.States {
		if v != PullRequestStateOpen && v != PullRequestStateClosed && v != PullRequestStateMerged {
			errors.Join(fmt.Errorf("unknown state in source.config.states '%s'. Must be one of OPEN, CLOSED, MERGED", v), err)
		}
	}

	return err
}
