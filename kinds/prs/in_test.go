package prs

import (
	"testing"
	"time"

	gh "github.com/PixelAirIO/github-resource"
	ghf "github.com/PixelAirIO/github-resource/github-resourcefakes"
	"github.com/stretchr/testify/require"
)

func TestInReturnsThePRs(t *testing.T) {
	assert := require.New(t)
	req := inRequest{
		Source: Source{
			Config: Config{
				Owner: "some-owner",
				Repo:  "some-repo",
			},
		},
		Version: version{
			Prs:       "1,3,6,88",
			Timestamp: time.Now(),
		},
	}

	client := &ghf.FakeGithubClient{}
	expectedPrs := []gh.PullRequest{
		{Number: "1"},
		{Number: "3"},
		{Number: "6"},
		{Number: "88"},
	}
	client.ListPullRequestsReturns(expectedPrs, nil)

	versions, err := in(req, client)
	assert.NoError(err)
	assert.Len(versions, 4)
	assert.ElementsMatch(versions, expectedPrs)
}

func TestInErrorsWhenPRsDontMatchTheVersion(t *testing.T) {
	assert := require.New(t)
	req := inRequest{
		Source: Source{
			Config: Config{
				Owner: "some-owner",
				Repo:  "some-repo",
			},
		},
		Version: version{
			Prs:       "1,3,6,88",
			Timestamp: time.Now(),
		},
	}

	client := &ghf.FakeGithubClient{}
	expectedPrs := []gh.PullRequest{
		{Number: "1"},
		{Number: "3"},
		{Number: "6"},
		// PR 88 is not returned from GitHub
	}
	client.ListPullRequestsReturns(expectedPrs, nil)

	versions, err := in(req, client)
	assert.ErrorContains(err, "One or more PRs likely changed.")
	assert.Len(versions, 0)
}
