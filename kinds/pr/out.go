package pr

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"

	gh "github.com/PixelAirIO/github-resource"
)

type outRequest struct {
	Source Source    `json:"source"`
	Params outParams `json:"params"`
}

type outParams struct {
	Ref         string `json:"ref"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

type outResponse struct {
	Version version `json:"version"`
}

func (*Pr) Out(stdin []byte, src string) {
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

	if request.Params.Name == "" {
		err = errors.Join(errors.New("params.name cannot be blank"))
	}

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

	err = ghc.UpdatePRStatus(request.Params.Ref, request.Params.Name, request.Params.Status, request.Params.Description)
	if err != nil {
		log.Fatal(err.Error())
	}

	resp := outResponse{
		Version: version{
			Ref: request.Params.Ref,
		},
	}

	ver, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error marshaling version: %v", err)
	}

	fmt.Println(string(ver))
}
