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
	Number        int    `json:"number"`
	Depth         int    `json:"depth"`
	Submodules    bool   `json:"submodules"`
	FetchTags     bool   `json:"fetch_tags"`
	MergeStrategy string `json:"merge_strategy"`
}

type version struct {
	Ref string `json:"ref"`
}

func validateSource(src *Source) (err error) {
	if src.Number == 0 {
		err = errors.Join(errors.New("'number' field is required and should be set to the PR's number"), err)
	}

	return err
}
