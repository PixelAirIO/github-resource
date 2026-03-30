package pr

import (
	"testing"

	"github.com/stretchr/testify/require"

	gh "github.com/PixelAirIO/github-resource"
	ghf "github.com/PixelAirIO/github-resource/github-resourcefakes"
)

func TestInternalCheckNoPriorVersion(t *testing.T) {
	assert := require.New(t)
	req := checkRequest{
		Source: Source{
			Config: Config{
				Config: gh.Config{
					Repository: "owner/repo",
				},
				Number: 60,
			},
		},
	}

	client := &ghf.FakeGithubClient{}
	client.LatestCommitForPRReturns("some-sha", nil)

	versions := check(req, client)
	assert.Len(versions, 1)
	assert.Equal("some-sha", versions[0].SHA)
}

func TestInternalCheckPriorVersionIsTheSame(t *testing.T) {
	assert := require.New(t)
	req := checkRequest{
		Source: Source{
			Config: Config{
				Config: gh.Config{
					Repository: "owner/repo",
				},
				Number: 60,
			},
		},
		Version: version{
			SHA: "some-sha",
		},
	}

	client := &ghf.FakeGithubClient{}
	client.LatestCommitForPRReturns("some-sha", nil)

	versions := check(req, client)
	assert.Len(versions, 0)
}
