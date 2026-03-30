package pr

import (
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
	CommitSHA   string `json:"commit_sha"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Description string `json:"description"`
}
