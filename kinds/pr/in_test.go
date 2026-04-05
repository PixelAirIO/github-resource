package pr

import (
	"testing"

	"github.com/stretchr/testify/require"

	gh "github.com/PixelAirIO/github-resource"
	ghf "github.com/PixelAirIO/github-resource/github-resourcefakes"
)

func TestInSucceedsWithMergeStrategy(t *testing.T) {
	dest := t.TempDir()
	assert := require.New(t)
	req := inRequest{
		Source: Source{
			Config: Config{
				Config: gh.Config{
					Repository: "owner/repo",
				},
				Number: 60,
			},
		},
		Version: version{
			Ref: "some-ref",
		},
	}

	client := &ghf.FakeGithubClient{}
	client.GetPRInfoReturns(gh.PullRequest{
		Number:       "60",
		Url:          "http://example/pull/60",
		TargetBranch: "main",
		Author:       "some-author",
		Branch:       "pr-branch",
	}, nil)

	meta, err := in(req, dest, client)
	assert.NoError(err)
	assert.NotNil(meta)
	assert.Len(meta, 6)
	assert.Equal(1, client.GetPRInfoCallCount())
	assert.Equal(1, client.InitRepoCallCount())
	assert.Equal(1, client.FetchPrCallCount())
	assert.Equal(1, client.PullBranchCallCount())
	assert.Equal(1, client.MergePrCallCount())
	actualRef, _ := client.MergePrArgsForCall(0)
	assert.Equal(req.Version.Ref, actualRef)
	assert.Equal(0, client.RebasePrCallCount(), "should not be called")
}

func TestInSucceedsWithRebaseStrategy(t *testing.T) {
	dest := t.TempDir()
	assert := require.New(t)
	req := inRequest{
		Source: Source{
			Config: Config{
				Config: gh.Config{
					Repository: "owner/repo",
				},
				Number:        60,
				MergeStrategy: "rebase",
			},
		},
		Version: version{
			Ref: "some-ref",
		},
	}

	client := &ghf.FakeGithubClient{}
	client.GetPRInfoReturns(gh.PullRequest{
		Number:       "60",
		Url:          "http://example/pull/60",
		TargetBranch: "main",
		Author:       "some-author",
		Branch:       "pr-branch",
	}, nil)

	meta, err := in(req, dest, client)
	assert.NoError(err)
	assert.NotNil(meta)
	assert.Len(meta, 6)
	assert.Equal(1, client.GetPRInfoCallCount())
	assert.Equal(1, client.InitRepoCallCount())
	assert.Equal(1, client.FetchPrCallCount())
	assert.Equal(1, client.PullBranchCallCount())
	assert.Equal(1, client.RebasePrCallCount())
	assert.Equal(0, client.MergePrCallCount(), "should not be called")
}

func TestInSucceedsWithCheckoutStrategy(t *testing.T) {
	dest := t.TempDir()
	assert := require.New(t)
	req := inRequest{
		Source: Source{
			Config: Config{
				Config: gh.Config{
					Repository: "owner/repo",
				},
				Number:        60,
				MergeStrategy: "checkout",
			},
		},
		Version: version{
			Ref: "some-ref",
		},
	}

	client := &ghf.FakeGithubClient{}
	client.GetPRInfoReturns(gh.PullRequest{
		Number:       "60",
		Url:          "http://example/pull/60",
		TargetBranch: "main",
		Author:       "some-author",
		Branch:       "pr-branch",
	}, nil)

	meta, err := in(req, dest, client)
	assert.NoError(err)
	assert.NotNil(meta)
	assert.Len(meta, 6)
	assert.Equal(1, client.GetPRInfoCallCount())
	assert.Equal(1, client.InitRepoCallCount())
	assert.Equal(1, client.FetchPrCallCount())
	assert.Equal(1, client.CheckoutPrCallCount())
	_, actualRef, _ := client.CheckoutPrArgsForCall(0)
	assert.Equal(req.Version.Ref, actualRef)
	assert.Equal(0, client.PullBranchCallCount(), "should not be called")
	assert.Equal(0, client.RebasePrCallCount(), "should not be called")
	assert.Equal(0, client.MergePrCallCount(), "should not be called")
}
