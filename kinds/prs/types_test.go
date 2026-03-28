package prs

import (
	"testing"

	gh "github.com/PixelAirIO/github-resource"
	"github.com/stretchr/testify/require"
)

func TestValidateSourceReturnsErrors(t *testing.T) {
	assert := require.New(t)
	src := Source{
		Config: Config{
			Owner: "",
			Repo:  "",
		},
	}

	actualErr := validateSource(&src)

	assert.ErrorContains(actualErr, "owner field is required")
	assert.ErrorContains(actualErr, "repository field is required")
	assert.Contains(src.Config.States, gh.PullRequestStateOpen, "sets an empty 'States' to OPEN")
	assert.Nil(src.Config.Labels, "labels is unmodified")
}

func TestValidateSourceReturnsNoErrors(t *testing.T) {
	assert := require.New(t)
	src := Source{
		Config: Config{
			Owner: "some-owner",
			Repo:  "some-repo",
		},
	}

	actualErr := validateSource(&src)

	assert.NoError(actualErr)
	assert.Contains(src.Config.States, gh.PullRequestStateOpen, "sets an empty 'States' to OPEN")
	assert.Nil(src.Config.Labels, "labels is unmodified")
}
