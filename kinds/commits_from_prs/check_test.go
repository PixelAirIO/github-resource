package commits_from_prs

import (
	"testing"

	"github.com/stretchr/testify/require"

	gh "github.com/PixelAirIO/github-resource"
	ghf "github.com/PixelAirIO/github-resource/github-resourcefakes"
)

func TestCheckNoPriorVersion(t *testing.T) {
	assert := require.New(t)
	req := checkRequest{
		Source: Source{
			Config: Config{
				Config: gh.Config{
					Repository: "owner/repo",
				},
			},
		},
	}

	client := &ghf.FakeGithubClient{}
	client.LatestCommitsFromPrsReturns([]gh.PRCommit{
		{PullRequest: gh.PullRequest{Number: "1"}, Ref: "abc123"},
		{PullRequest: gh.PullRequest{Number: "3"}, Ref: "def456"},
	}, nil)

	versions := check(req, client)
	assert.Len(versions, 2)
	assert.Equal("abc123", versions[0].Ref)
	assert.Equal("1", versions[0].Pr)
	assert.Equal("def456", versions[1].Ref)
	assert.Equal("3", versions[1].Pr)
}

func TestCheckExcludeDrafts(t *testing.T) {
	assert := require.New(t)
	req := checkRequest{
		Source: Source{
			Config: Config{
				Config: gh.Config{
					Repository: "owner/repo",
				},
				ExcludeDrafts: true,
			},
		},
	}

	client := &ghf.FakeGithubClient{}
	client.LatestCommitsFromPrsReturns([]gh.PRCommit{
		{PullRequest: gh.PullRequest{Number: "1", IsDraft: false}, Ref: "abc123"},
		{PullRequest: gh.PullRequest{Number: "3", IsDraft: true}, Ref: "def456"},
		{PullRequest: gh.PullRequest{Number: "5", IsDraft: false}, Ref: "ghi789"},
	}, nil)

	versions := check(req, client)
	assert.Len(versions, 2)
	assert.Equal("1", versions[0].Pr)
	assert.Equal("5", versions[1].Pr)
}

func TestCheckTargetBranch(t *testing.T) {
	assert := require.New(t)
	req := checkRequest{
		Source: Source{
			Config: Config{
				Config: gh.Config{
					Repository: "owner/repo",
				},
				TargetBranch: "develop",
			},
		},
	}

	client := &ghf.FakeGithubClient{}
	client.LatestCommitsFromPrsReturns([]gh.PRCommit{
		{PullRequest: gh.PullRequest{Number: "1", TargetBranch: "main"}, Ref: "abc123"},
		{PullRequest: gh.PullRequest{Number: "3", TargetBranch: "develop"}, Ref: "def456"},
		{PullRequest: gh.PullRequest{Number: "5", TargetBranch: "develop"}, Ref: "ghi789"},
	}, nil)

	versions := check(req, client)
	assert.Len(versions, 2)
	assert.Equal("3", versions[0].Pr)
	assert.Equal("5", versions[1].Pr)
}

func TestCheckNoMatchingCommits(t *testing.T) {
	assert := require.New(t)
	req := checkRequest{
		Source: Source{
			Config: Config{
				Config: gh.Config{
					Repository: "owner/repo",
				},
			},
		},
	}

	client := &ghf.FakeGithubClient{}
	client.LatestCommitsFromPrsReturns([]gh.PRCommit{}, nil)

	versions := check(req, client)
	assert.Len(versions, 0)
}

func TestCheckPriorVersionNotFound(t *testing.T) {
	assert := require.New(t)
	req := checkRequest{
		Source: Source{
			Config: Config{
				Config: gh.Config{
					Repository: "owner/repo",
				},
			},
		},
		Version: version{
			Ref: "stale_ref",
			Pr:  "99",
		},
	}

	client := &ghf.FakeGithubClient{}
	client.LatestCommitsFromPrsReturns([]gh.PRCommit{
		{PullRequest: gh.PullRequest{Number: "1"}, Ref: "abc123"},
		{PullRequest: gh.PullRequest{Number: "3"}, Ref: "def456"},
	}, nil)

	versions := check(req, client)
	assert.Len(versions, 2, "returns all versions when prior version is not found")
}
