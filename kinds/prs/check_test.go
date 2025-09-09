package prs

import (
	"testing"

	"github.com/stretchr/testify/require"

	gh "github.com/pixel-air/github-resource"
)

func TestCheckValidateWorks(t *testing.T) {
	assert := require.New(t)
	req := checkRequest{
		Source: Source{
			Config: Config{
				Owner:  "",
				Repo:   "",
				States: nil,
				Labels: nil,
			},
		},
	}

	actualErr := checkValidate(&req)

	assert.ErrorContains(actualErr, "owner field is required")
	assert.ErrorContains(actualErr, "repository field is required")
	assert.Contains(req.Source.Config.States, gh.PullRequestStateOpen, "sets an empty 'States' to OPEN")
	assert.Nil(req.Source.Config.Labels, "labels is unmodified")
}
