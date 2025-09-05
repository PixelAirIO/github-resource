package githubresource

import (
	"errors"
	"net/http"

	"github.com/Khan/genqlient/graphql"
)

type Config struct {
	AccessToken string `json:"access_token"`
	APIEndpoint string `json:"api_endpoint"`
}

type Github interface {
	ListPullRequests()
}

type githubClient struct {
	client graphql.Client
}

type authedTransport struct {
	accessToken string
	transport   http.RoundTripper
}

func (a *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "bearer "+a.accessToken)
	return a.transport.RoundTrip(req)
}

func NewGithubClient(cfg Config) (*githubClient, error) {
	if cfg.APIEndpoint == "" {
		cfg.APIEndpoint = "https://api.github.com/graphql"
	}
	if cfg.AccessToken == "" {
		return nil, errors.New("access_token is required")
	}

	client := graphql.NewClient(cfg.APIEndpoint, &http.Client{
		Transport: &authedTransport{
			accessToken: cfg.AccessToken,
			transport:   http.DefaultTransport,
		},
	})

	return &githubClient{client: client}, nil
}
