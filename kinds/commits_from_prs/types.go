package commits_from_prs

import (
	gh "github.com/PixelAirIO/github-resource"
)

type CommitsFromPrs struct{}

var _ gh.Kind = (*CommitsFromPrs)(nil)

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
	MergeStrategy string                `json:"merge_strategy,omitempty"`
	Depth         int                   `json:"depth,omitempty"`
	Submodules    bool                  `json:"submodules,omitempty"`
	FetchTags     bool                  `json:"fetch_tags,omitempty"`
}

type version struct {
	Ref        string `json:"ref"`
	Pr         string `json:"pr"`
	CommitDate string `json:"commit_date"`
}

func validateSource(src *Source) error {
	if len(src.Config.States) == 0 {
		src.Config.States = []gh.PullRequestState{gh.PullRequestStateOpen}
	}

	return nil
}
