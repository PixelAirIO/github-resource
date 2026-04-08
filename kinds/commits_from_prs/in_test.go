package commits_from_prs

import (
	"testing"

	"github.com/stretchr/testify/require"

	gh "github.com/PixelAirIO/github-resource"
	ghf "github.com/PixelAirIO/github-resource/github-resourcefakes"
)

func TestInCallsGetPRInfo(t *testing.T) {
	assert := require.New(t)
	req := inRequest{
		Source: Source{
			Config: Config{
				Config: gh.Config{
					Repository: "owner/repo",
				},
			},
		},
		Version: version{
			Ref: "abc123",
			Pr:  "42",
		},
	}

	client := &ghf.FakeGithubClient{}
	client.GetPRInfoReturns(gh.PullRequest{
		Number:       "42",
		Url:          "https://github.com/owner/repo/pull/42",
		TargetBranch: "main",
		Branch:       "feature-branch",
		Author:       "testuser",
	}, nil)

	meta, err := in(req, t.TempDir(), client)
	assert.NoError(err)

	assert.Equal(1, client.GetPRInfoCallCount())
	assert.Equal(42, client.GetPRInfoArgsForCall(0))

	assert.Len(meta, 6)
}

func TestInInvalidPrNumber(t *testing.T) {
	assert := require.New(t)
	req := inRequest{
		Source: Source{
			Config: Config{
				Config: gh.Config{
					Repository: "owner/repo",
				},
			},
		},
		Version: version{
			Ref: "abc123",
			Pr:  "not-a-number",
		},
	}

	client := &ghf.FakeGithubClient{}

	_, err := in(req, t.TempDir(), client)
	assert.ErrorContains(err, "invalid PR number")
}
