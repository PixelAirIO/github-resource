package pr

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

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
	Number        StrInt `json:"number"`
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

// StrInt is an int that can be unmarshaled from a JSON int or string.
type StrInt int

func (fi *StrInt) UnmarshalJSON(b []byte) error {
	var n int
	if err := json.Unmarshal(b, &n); err == nil {
		*fi = StrInt(n)
		return nil
	}

	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		n, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("cannot convert %q to int: %w", s, err)
		}
		*fi = StrInt(n)
		return nil
	}

	return fmt.Errorf("cannot unmarshal %s into int", string(b))
}
