package pr

import (
	"errors"

	gh "github.com/PixelAirIO/github-resource"
)

type Pr struct{}

var _ gh.Kind = (*Pr)(nil)

type Source struct {
	Config
	Kind string `json:"kind"`
}

type Config struct {
	gh.Config
	Number *int `json:"number"`
}

type version struct {
	SHA          string `json:"sha"`
	TargetBranch string `json:"target_branch"`
}

func validateSource(src *Source) (err error) {
	if src.Number == nil {
		err = errors.Join(errors.New("'number' field is required and should be set to the PR's number"), err)
	}

	return err
}
