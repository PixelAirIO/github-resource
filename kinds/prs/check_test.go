package prs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	gh "github.com/PixelAirIO/github-resource"
	ghf "github.com/PixelAirIO/github-resource/github-resourcefakes"
)

func TestInternalCheckNoPriorVersion(t *testing.T) {
	assert := require.New(t)
	req := checkRequest{
		Source: Source{
			Config: Config{
				Owner: "some-owner",
				Repo:  "some-repo",
			},
		},
	}

	client := &ghf.FakeGithubClient{}
	client.ListPullRequestsReturns([]gh.PullRequest{
		{Number: "1"},
		{Number: "3"},
		{Number: "6"},
		{Number: "88"},
	}, nil)

	versions := check(req, client)
	assert.Len(versions, 1)
	assert.Equal("1,3,6,88", versions[0].Prs)
}

func TestInternalCheckMatchesPriorVersion(t *testing.T) {
	assert := require.New(t)
	req := checkRequest{
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
	client.ListPullRequestsReturns([]gh.PullRequest{
		{Number: "1"},
		{Number: "3"},
		{Number: "6"},
		{Number: "88"},
	}, nil)

	versions := check(req, client)
	assert.Len(versions, 0, "does not return a version because it matches the prior version")
}

func TestInternalCheckDoesNotMatchesPriorVersion(t *testing.T) {
	assert := require.New(t)
	req := checkRequest{
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
	client.ListPullRequestsReturns([]gh.PullRequest{
		{Number: "1"},
		{Number: "3"},
		{Number: "6"},
		{Number: "88"},
		{Number: "102"},
	}, nil)

	versions := check(req, client)
	assert.Len(versions, 1, "returns a new version because it doesn't match the prior version")
	assert.Equal("1,3,6,88,102", versions[0].Prs)
}

func TestExcludingDrafts(t *testing.T) {
	assert := require.New(t)
	req := checkRequest{
		Source: Source{
			Config: Config{
				Owner:         "some-owner",
				Repo:          "some-repo",
				ExcludeDrafts: true,
			},
		},
	}

	client := &ghf.FakeGithubClient{}
	client.ListPullRequestsReturns([]gh.PullRequest{
		{Number: "1", IsDraft: false},
		{Number: "3", IsDraft: true},
		{Number: "6", IsDraft: false},
		{Number: "88", IsDraft: false},
	}, nil)

	versions := check(req, client)
	assert.Len(versions, 1)
	assert.Equal("1,6,88", versions[0].Prs)
}

func TestMatchingTargetBranch(t *testing.T) {
	assert := require.New(t)
	req := checkRequest{
		Source: Source{
			Config: Config{
				Owner:        "some-owner",
				Repo:         "some-repo",
				TargetBranch: "other",
			},
		},
	}

	client := &ghf.FakeGithubClient{}
	client.ListPullRequestsReturns([]gh.PullRequest{
		{Number: "1", TargetBranch: "main"},
		{Number: "3", TargetBranch: "other"},
		{Number: "6", TargetBranch: "other"},
		{Number: "88", TargetBranch: "main"},
	}, nil)

	versions := check(req, client)
	assert.Len(versions, 1)
	assert.Equal("3,6", versions[0].Prs)
}

func TestNoMatchingPRsReturnsNoneAsTheVersion(t *testing.T) {
	assert := require.New(t)
	req := checkRequest{
		Source: Source{
			Config: Config{
				Owner: "some-owner",
				Repo:  "some-repo",
			},
		},
	}

	client := &ghf.FakeGithubClient{}
	client.ListPullRequestsReturns([]gh.PullRequest{}, nil)

	versions := check(req, client)
	assert.Len(versions, 1)
	assert.Equal("none", versions[0].Prs, "version should be set to the string 'none'")
}

func TestNoMatchingPRsReturnsEmptyListWhenPriorVersionIsNone(t *testing.T) {
	assert := require.New(t)
	req := checkRequest{
		Source: Source{
			Config: Config{
				Owner: "some-owner",
				Repo:  "some-repo",
			},
		},
		Version: version{
			Prs:       "none",
			Timestamp: time.Now(),
		},
	}

	client := &ghf.FakeGithubClient{}
	client.ListPullRequestsReturns([]gh.PullRequest{}, nil)

	versions := check(req, client)
	assert.Len(versions, 0, "prior version is 'none' so duplicate version should not be returned")
}
