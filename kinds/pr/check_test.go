package pr

import (
	"testing"

	"github.com/stretchr/testify/require"

	gh "github.com/PixelAirIO/github-resource"
	ghf "github.com/PixelAirIO/github-resource/github-resourcefakes"
)

func TestInternalCheckNoPriorVersion(t *testing.T) {
	assert := require.New(t)
	n := 60
	req := checkRequest{
		Source: Source{
			Config: Config{
				Config: gh.Config{
					Repository: "owner/repo",
				},
				Number: &n,
			},
		},
	}

	client := &ghf.FakeGithubClient{}
	client.LatestCommitForPRReturns(gh.PullRequestCommit{
		LatestSHA:    "some-sha",
		TargetBranch: "main",
	}, nil)

	versions := check(req, client)
	assert.Len(versions, 1)
	assert.Equal("some-sha", versions[0].SHA)
	assert.Equal("main", versions[0].TargetBranch)
}

func TestInternalCheckPriorVersionIsTheSame(t *testing.T) {
	assert := require.New(t)
	n := 60
	req := checkRequest{
		Source: Source{
			Config: Config{
				Config: gh.Config{
					Repository: "owner/repo",
				},
				Number: &n,
			},
		},
		Version: version{
			SHA:          "some-sha",
			TargetBranch: "main",
		},
	}

	client := &ghf.FakeGithubClient{}
	client.LatestCommitForPRReturns(gh.PullRequestCommit{
		LatestSHA:    "some-sha",
		TargetBranch: "main",
	}, nil)

	versions := check(req, client)
	assert.Len(versions, 0)
}

func TestInternalCheckTagetBranchChanged(t *testing.T) {
	assert := require.New(t)
	n := 60
	req := checkRequest{
		Source: Source{
			Config: Config{
				Config: gh.Config{
					Repository: "owner/repo",
				},
				Number: &n,
			},
		},
		Version: version{
			SHA:          "some-sha",
			TargetBranch: "main",
		},
	}

	client := &ghf.FakeGithubClient{}
	client.LatestCommitForPRReturns(gh.PullRequestCommit{
		LatestSHA:    "some-sha",
		TargetBranch: "other-branch",
	}, nil)

	versions := check(req, client)
	assert.Len(versions, 1)
	assert.Equal("some-sha", versions[0].SHA)
	assert.Equal("other-branch", versions[0].TargetBranch)
}
