package prs

import (
	"errors"
	"time"

	gh "github.com/PixelAirIO/github-resource"
)

type Prs struct{}

var _ gh.Kind = (*Prs)(nil)

type Source struct {
	Config
	Kind string `json:"kind"`
}

type Config struct {
	gh.Config
	Owner         string                `json:"owner"`
	Repo          string                `json:"repository"`
	States        []gh.PullRequestState `json:"states,omitempty"`
	Labels        []string              `json:"labels,omitempty"`
	TargetBranch  string                `json:"target_branch,omitempty"`
	ExcludeDrafts bool                  `json:"exclude_drafts,omitempty"`
}

type version struct {
	Prs       string    `json:"prs"`
	Timestamp time.Time `json:"timestamp"`
}

func validateSource(src *Source) (err error) {
	if src.Config.Owner == "" {
		err = errors.Join(errors.New("owner field is required"), err)
	}

	if src.Config.Repo == "" {
		err = errors.Join(errors.New("repository field is required"), err)
	}

	if len(src.Config.States) == 0 {
		src.Config.States = []gh.PullRequestState{gh.PullRequestStateOpen}
	}

	return err
}
