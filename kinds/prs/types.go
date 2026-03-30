package prs

import (
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
	if len(src.Config.States) == 0 {
		src.Config.States = []gh.PullRequestState{gh.PullRequestStateOpen}
	}

	return err
}
