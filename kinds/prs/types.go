package prs

import (
	"time"

	gh "github.com/pixel-air/github-resource"
)

type Source struct {
	Kind   string `json:"kind"`
	Config Config `json:"config"`
}
type Config struct {
	gh.Config
	Owner  string                `json:"owner"`
	Repo   string                `json:"repository"`
	States []gh.PullRequestState `json:"states,omitempty"`
	Labels []string              `json:"labels,omitempty"`
}

type version struct {
	Prs       string    `json:"prs"`
	Timestamp time.Time `json:"timestamp"`
}
