package commits_from_prs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"

	gh "github.com/PixelAirIO/github-resource"
)

type outRequest struct {
	Source Source    `json:"source"`
	Params outParams `json:"params"`
}

type outParams struct {
	Ref         string `json:"ref"`
	Pr          string `json:"pr"`
	CommitDate  string `json:"commit_date"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

type outResponse struct {
	Version  version     `json:"version"`
	Metadata gh.Metadata `json:"metadata,omitempty"`
}

func (*CommitsFromPrs) Out(stdin []byte, src string) {
	dc := json.NewDecoder(bytes.NewReader(stdin))
	dc.DisallowUnknownFields()

	request := outRequest{}
	err := dc.Decode(&request)
	if err != nil {
		log.Fatalf("failed to unmarshal in request: %v", err)
	}

	if request.Params.Ref == "" {
		err = errors.Join(errors.New("params.ref cannot be blank"))
	}

	if request.Params.Pr == "" {
		err = errors.Join(errors.New("params.pr cannot be blank"))
	}

	if request.Params.Name == "" {
		err = errors.Join(errors.New("params.name cannot be blank"))
	}

	name := gh.InterpolateBuildMetadata(request.Params.Name)

	if request.Params.Status == "" {
		err = errors.Join(errors.New("params.status cannot be blank. Must be one of: pending, success, error, failure"))
	} else {
		if !slices.Contains([]string{"pending", "success", "error", "failure"}, strings.ToLower(request.Params.Status)) {
			err = errors.Join(errors.New("unknown value in params.status. Must be one of: pending, success, error, failure"))
		}
	}

	if err != nil {
		log.Fatal(err.Error())
	}

	err = validateSource(&request.Source)
	if err != nil {
		log.Fatalf("validation error: %v", err)
	}

	ghc, err := gh.NewGithubClient(request.Source.Config.Config)
	if err != nil {
		log.Fatalf("failed to create Github client: %v", err)
	}

	err = ghc.UpdatePRStatus(request.Params.Ref, name, request.Params.Status, request.Params.Description)
	if err != nil {
		log.Fatalf("error updating PR status: %v", err)
	}

	num, err := strconv.Atoi(request.Params.Pr)
	if err != nil {
		log.Fatalf("error parsing params.pr: %v", err)
	}

	pr, err := ghc.GetPRInfo(num)
	if err != nil {
		log.Fatalf("error getting PR info: %v", err)
	}

	// We have to populate metadata otherwise we overwrite what the get step
	// fetched with nothing, clearing out the metadata for the version
	var meta gh.Metadata
	meta.Add("ref", request.Params.Ref)
	meta.Add("pr", pr.Number)
	meta.Add("url", pr.Url)
	meta.Add("target_branch", pr.TargetBranch)
	meta.Add("pr_branch", pr.Branch)
	meta.Add("author", pr.Author)

	resp := outResponse{
		Version: version{
			Ref:        request.Params.Ref,
			Pr:         request.Params.Pr,
			CommitDate: request.Params.CommitDate,
		},
		Metadata: meta,
	}

	ver, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error marshaling version: %v", err)
	}

	fmt.Println(string(ver))
}
