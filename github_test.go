package githubresource

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewGithubClientReturnsAClient(t *testing.T) {
	assert := require.New(t)

	cfg := Config{}
	client, err := NewGithubClient(cfg)
	assert.NoError(err)
	assert.NotNil(client)
	assert.Equal(DefaultGraphqlEndpoint, client.APIEndpoint())
	assert.Equal(cfg.AccessToken, client.AccessToken())
}

func TestNewGithubClientDoesNotOverrideGivenEndpoint(t *testing.T) {
	assert := require.New(t)

	cfg := Config{
		APIEndpoint:  "https://custom.endpoint/graphql",
		HostEndpoint: "https://custom.host/",
		AccessToken:  "some-access-token",
	}
	client, err := NewGithubClient(cfg)
	assert.NoError(err)
	assert.NotNil(client)
	assert.Equal(cfg.APIEndpoint, client.APIEndpoint())
	assert.Equal(cfg.HostEndpoint, client.HostEndpoint())
	assert.Equal(cfg.AccessToken, client.AccessToken())
}
