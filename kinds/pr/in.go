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
	Version version `json:"version"`
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

	err = in(request, dest, ghc)
	if err != nil {
		log.Fatalf("error getting Prs: %v", err)
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

func in(req inRequest, dest string, ghc gh.GithubClient) error {
	pr, err := ghc.GetPRInfo(req.Source.Number)
	if err != nil {
		return err
	}

	err = os.Chdir(dest)
	if err != nil {
		return fmt.Errorf("chdir: %w", err)
	}

	err = ghc.InitRepo(pr.ParentRepoUrl, pr.TargetBranch)
	if err != nil {
		return err
	}

	err = ghc.FetchPr(pr.ParentRepoUrl, pr.Number, req.Source.Depth, req.Source.FetchTags, req.Source.Submodules)
	if err != nil {
		return err
	}

	switch strings.ToLower(req.Source.Config.MergeStrategy) {
	case "merge", "":
		err = ghc.PullBranch(pr.TargetBranch, req.Source.Depth, req.Source.FetchTags, req.Source.Submodules)
		if err != nil {
			return err
		}

		err = ghc.MergePr(req.Version.Ref, req.Source.Submodules)
		if err != nil {
			return err
		}

	case "rebase":
		err = ghc.PullBranch(pr.TargetBranch, req.Source.Depth, req.Source.FetchTags, req.Source.Submodules)
		if err != nil {
			return err
		}
		err = ghc.RebasePr(pr.TargetBranch, pr.Branch, req.Source.Submodules)
		if err != nil {
			return err
		}

	case "checkout":
		err = ghc.CheckoutPr(pr.Branch, req.Version.Ref, req.Source.Submodules)
		if err != nil {
			return err
		}
	default:
		log.Fatalf("unknown merge strategy '%s'", req.Source.MergeStrategy)
	}

	return nil
}
