package prs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	gh "github.com/pixel-air/github-resource"
	ghf "github.com/pixel-air/github-resource/github-resourcefakes"
)

func TestCheckValidateReturnsErrors(t *testing.T) {
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

func TestCheckValidateReturnsNoErrors(t *testing.T) {
	assert := require.New(t)
	req := checkRequest{
		Source: Source{
			Config: Config{
				Owner:  "some-owner",
				Repo:   "some-repo",
				States: nil,
				Labels: nil,
			},
		},
	}

	actualErr := checkValidate(&req)

	assert.NoError(actualErr)
	assert.Contains(req.Source.Config.States, gh.PullRequestStateOpen, "sets an empty 'States' to OPEN")
	assert.Nil(req.Source.Config.Labels, "labels is unmodified")
}

func TestInternalCheckNoPriorVersion(t *testing.T) {
	assert := require.New(t)
	req := checkRequest{
		Source: Source{
			Config: Config{
				Owner:  "some-owner",
				Repo:   "some-repo",
				States: nil,
				Labels: nil,
			},
		},
	}

	client := &ghf.FakeGithubClient{}
	client.ListPullRequestsReturns([]gh.PullRequest{
		{ID: 1},
		{ID: 3},
		{ID: 6},
		{ID: 88},
	}, nil)

	versions := check(req, client)
	assert.Len(versions, 1)
	assert.Equal(versions[0].Prs, "1,3,6,88")
}

func TestInternalCheckMatchesPriorVersion(t *testing.T) {
	assert := require.New(t)
	req := checkRequest{
		Source: Source{
			Config: Config{
				Owner:  "some-owner",
				Repo:   "some-repo",
				States: nil,
				Labels: nil,
			},
		},
		Version: version{
			Prs:       "1,3,6,88",
			Timestamp: time.Now(),
		},
	}

	client := &ghf.FakeGithubClient{}
	client.ListPullRequestsReturns([]gh.PullRequest{
		{ID: 1},
		{ID: 3},
		{ID: 6},
		{ID: 88},
	}, nil)

	versions := check(req, client)
	assert.Len(versions, 0, "does not return a version because it matches the prior version")
}

func TestInternalCheckDoesNotMatchesPriorVersion(t *testing.T) {
	assert := require.New(t)
	req := checkRequest{
		Source: Source{
			Config: Config{
				Owner:  "some-owner",
				Repo:   "some-repo",
				States: nil,
				Labels: nil,
			},
		},
		Version: version{
			Prs:       "1,3,6,88",
			Timestamp: time.Now(),
		},
	}

	client := &ghf.FakeGithubClient{}
	client.ListPullRequestsReturns([]gh.PullRequest{
		{ID: 1},
		{ID: 3},
		{ID: 6},
		{ID: 88},
		{ID: 102},
	}, nil)

	versions := check(req, client)
	assert.Len(versions, 1, "returns a new version because it doesn't match the prior version")
	assert.Equal(versions[0].Prs, "1,3,6,88,102")
}
