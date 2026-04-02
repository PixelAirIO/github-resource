package githubresource

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewGithubClientReturnsAClient(t *testing.T) {
	assert := require.New(t)

	cfg := Config{
		Repository: "owner/repo",
	}
	client, err := NewGithubClient(cfg)
	assert.NoError(err)
	assert.NotNil(client)
	rest, gql := client.APIEndpoints()
	assert.Equal(DefaultRestEndpoint, rest)
	assert.Equal(DefaultGraphqlEndpoint, gql)
	assert.Equal(cfg.AccessToken, client.AccessToken())
}

func TestNewGithubClientDoesNotOverrideGivenEndpoint(t *testing.T) {
	assert := require.New(t)

	cfg := Config{
		Repository:      "owner/repo",
		RestEndpoint:    "https://custom.endpoint/",
		GraphqlEndpoint: "https://custom.endpoint/graphql",
		HostEndpoint:    "https://custom.host/",
		AccessToken:     "some-access-token",
	}
	client, err := NewGithubClient(cfg)
	assert.NoError(err)
	assert.NotNil(client)
	rest, gql := client.APIEndpoints()
	assert.Equal(cfg.RestEndpoint, rest)
	assert.Equal(cfg.GraphqlEndpoint, gql)
	assert.Equal(cfg.HostEndpoint, client.HostEndpoint())
	assert.Equal(cfg.AccessToken, client.AccessToken())
}

func TestNewGithubClientErrorsWhenRepositoryIsBlank(t *testing.T) {
	assert := require.New(t)

	cfg := Config{}
	client, err := NewGithubClient(cfg)
	assert.Error(err)
	assert.EqualError(err, "repository is blank and must be set. Expected format is 'OWNER/REPO'.")
	assert.Nil(client)
}

func TestNewGithubClientErrorsWhenRepositoryIsMalformed(t *testing.T) {
	assert := require.New(t)

	cfg := Config{
		Repository: "owner-repo",
	}
	client, err := NewGithubClient(cfg)
	assert.Error(err)
	assert.EqualError(err, "unexpected format for 'repository'. Expected format is 'OWNER/REPO'.")
	assert.Nil(client)
}
