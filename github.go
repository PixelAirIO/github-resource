package githubresource

//go:generate go run github.com/Khan/genqlient

import (
	"context"
	"net/http"

	"github.com/Khan/genqlient/graphql"
)

//go:generate go tool counterfeiter -generate

type Config struct {
	AccessToken string `json:"access_token"`
	APIEndpoint string `json:"api_endpoint"`
}

//counterfeiter:generate . GithubClient
type GithubClient interface {
	APIEndpoint() string
	AccessToken() string
	ListPullRequests(owner string, repo string, states []PullRequestState, labels []string) ([]PullRequest, error)
}

type githubClient struct {
	client graphql.Client
	config Config
}

var _ GithubClient = (*githubClient)(nil)

type authedTransport struct {
	accessToken string
	transport   http.RoundTripper
}

func (a *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "bearer "+a.accessToken)
	return a.transport.RoundTrip(req)
}

const DefaultEndpoint = "https://api.github.com/graphql"

func NewGithubClient(cfg Config) (GithubClient, error) {
	if cfg.APIEndpoint == "" {
		cfg.APIEndpoint = DefaultEndpoint
	}

	client := graphql.NewClient(cfg.APIEndpoint, &http.Client{
		Transport: &authedTransport{
			accessToken: cfg.AccessToken,
			transport:   http.DefaultTransport,
		},
	})

	return &githubClient{client: client, config: cfg}, nil
}

func (g *githubClient) APIEndpoint() string {
	return g.config.APIEndpoint
}

func (g *githubClient) AccessToken() string {
	return g.config.AccessToken
}

type PullRequest struct {
	ID int
}

// Returns pull requests matching the states and labels provided.
//
// If you want to match against no labels, pass in nil.
func (g *githubClient) ListPullRequests(owner string, repo string, states []PullRequestState, labels []string) ([]PullRequest, error) {
	_ = `# @genqlient
query getPullRequests(
    $owner: String!
    $name: String!
    $states: [PullRequestState!]
    $labels: [String!]
    $endCursor: String
) {
    repository(owner: $owner, name: $name) {
        pullRequests(
	        first: 100,
			after: $endCursor,
			states: $states,
			labels: $labels,
			orderBy: {field: CREATED_AT, direction: ASC}
		) {
            nodes {
                number
                isDraft
                permalink
                baseRefName
            }
            pageInfo {
	            endCursor
                hasNextPage
            }
        }
    }
}`
	prs := []PullRequest{}
	ctx := context.Background()
	hasNextPage := true
	endCursor := ""

	for hasNextPage {
		resp, err := getPullRequests(ctx, g.client, owner, repo, states, labels, endCursor)
		if err != nil {
			return nil, err
		}

		for _, v := range resp.Repository.PullRequests.Nodes {
			prs = append(prs, PullRequest{ID: v.Number})
		}

		hasNextPage = resp.Repository.PullRequests.PageInfo.HasNextPage
		endCursor = resp.Repository.PullRequests.PageInfo.EndCursor
	}

	return prs, nil
}
