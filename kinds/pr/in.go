package pr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	gh "github.com/PixelAirIO/github-resource"
)

type inRequest struct {
	Source  Source  `json:"source"`
	Version version `json:"version"`
}

type inResponse struct {
	Version  version     `json:"version"`
	Metadata gh.Metadata `json:"metadata,omitempty"`
}

func (*Pr) In(stdin []byte, dest string) {
	dc := json.NewDecoder(bytes.NewReader(stdin))
	dc.DisallowUnknownFields()

	request := inRequest{}
	err := dc.Decode(&request)
	if err != nil {
		log.Fatalf("failed to unmarshal in request: %v", err)
	}

	if request.Version.Ref == "" {
		log.Fatal("version has an empty 'ref'. Cannot proceed.")
	}

	err = validateSource(&request.Source)
	if err != nil {
		log.Fatalf("validation error: %v", err)
	}

	ghc, err := gh.NewGithubClient(request.Source.Config.Config)
	if err != nil {
		log.Fatalf("failed to create Github client: %v", err)
	}

	meta, err := in(request, dest, ghc)
	if err != nil {
		log.Fatalf("error getting PR: %v", err)
	}

	resp := inResponse{
		Version:  request.Version,
		Metadata: meta,
	}

	ver, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error marshaling version: %v", err)
	}

	fmt.Println(string(ver))
}

func in(req inRequest, dest string, ghc gh.GithubClient) (gh.Metadata, error) {
	pr, err := ghc.GetPRInfo(int(req.Source.Number))
	if err != nil {
		return nil, err
	}

	var meta gh.Metadata
	meta.Add("ref", req.Version.Ref)
	meta.Add("pr", pr.Number)
	meta.Add("url", pr.Url)
	meta.Add("target_branch", pr.TargetBranch)
	meta.Add("pr_branch", pr.Branch)
	meta.Add("author", pr.Author)

	err = os.Chdir(dest)
	if err != nil {
		return nil, fmt.Errorf("chdir: %w", err)
	}

	err = ghc.InitRepo(pr.ParentRepoUrl, pr.TargetBranch)
	if err != nil {
		return nil, err
	}

	err = ghc.FetchPr(pr.ParentRepoUrl, pr.Number, req.Source.Depth, req.Source.FetchTags, req.Source.Submodules)
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(req.Source.Config.MergeStrategy) {
	case "merge", "":
		err = ghc.PullBranch(pr.TargetBranch, req.Source.Depth, req.Source.FetchTags, req.Source.Submodules)
		if err != nil {
			return nil, err
		}

		err = ghc.CheckoutPr(pr.Branch, req.Version.Ref, req.Source.Submodules)
		if err != nil {
			return nil, err
		}

		err = ghc.MergePr(req.Version.Ref, req.Source.Submodules)
		if err != nil {
			return nil, err
		}

	case "rebase":
		err = ghc.PullBranch(pr.TargetBranch, req.Source.Depth, req.Source.FetchTags, req.Source.Submodules)
		if err != nil {
			return nil, err
		}

		err = ghc.CheckoutPr(pr.Branch, req.Version.Ref, req.Source.Submodules)
		if err != nil {
			return nil, err
		}

		err = ghc.RebasePr(pr.TargetBranch, pr.Branch, req.Source.Submodules)
		if err != nil {
			return nil, err
		}

	case "checkout":
		err = ghc.CheckoutPr(pr.Branch, req.Version.Ref, req.Source.Submodules)
		if err != nil {
			return nil, err
		}
	default:
		log.Fatalf("unknown merge strategy '%s'", req.Source.MergeStrategy)
	}

	return meta, nil
}
