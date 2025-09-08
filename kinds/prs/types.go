package prs

import (
	"time"

	. "github.com/pixel-air/github-resource"
)

type Source struct {
	Kind   string `json:"kind"`
	Config struct {
		Config
		Owner  string             `json:"owner"`
		Repo   string             `json:"repository"`
		States []PullRequestState `json:"states"`
		Labels []string           `json:"labels,omitempty"`
	} `json:"config"`
}

type version struct {
	Prs       string    `json:"prs"`
	Timestamp time.Time `json:"timestamp"`
}
