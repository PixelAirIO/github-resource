package pr

import (
	"testing"

	"github.com/stretchr/testify/require"

	gh "github.com/PixelAirIO/github-resource"
)

func TestValidateErrors(t *testing.T) {
	assert := require.New(t)
	src := Source{
		Config: Config{
			Config: gh.Config{
				Repository: "owner/repo",
			},
		},
	}

	err := validateSource(&src)
	assert.ErrorContains(err, "'number' field is required and should be set to the PR's number")
}
