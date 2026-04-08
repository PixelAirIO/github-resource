package factory_test

import (
	"encoding/json"
	"testing"

	ghr "github.com/PixelAirIO/github-resource"
	"github.com/PixelAirIO/github-resource/factory"
	cfp "github.com/PixelAirIO/github-resource/kinds/commits_from_prs"
	"github.com/PixelAirIO/github-resource/kinds/pr"
	"github.com/PixelAirIO/github-resource/kinds/prs"
	"github.com/stretchr/testify/require"
)

func TestFactoryCreatesPRs(t *testing.T) {
	assert := require.New(t)
	payload := ghr.BaseRequest{
		Source: ghr.Source{
			Kind: "prs",
		},
	}
	stdin, err := json.Marshal(payload)
	assert.NoError(err)

	k := factory.NewKind(stdin)
	assert.IsType(&prs.Prs{}, k)
}

func TestFactoryCreatesPR(t *testing.T) {
	assert := require.New(t)
	payload := ghr.BaseRequest{
		Source: ghr.Source{
			Kind: "pr",
		},
	}
	stdin, err := json.Marshal(payload)
	assert.NoError(err)

	k := factory.NewKind(stdin)
	assert.IsType(&pr.Pr{}, k)
}

func TestFactoryCreatesCommitsFromPrs(t *testing.T) {
	assert := require.New(t)
	payload := ghr.BaseRequest{
		Source: ghr.Source{
			Kind: "commits-from-prs",
		},
	}
	stdin, err := json.Marshal(payload)
	assert.NoError(err)

	k := factory.NewKind(stdin)
	assert.IsType(&cfp.CommitsFromPrs{}, k)
}
