package githubresource

//go:generate go run github.com/Khan/genqlient

import (
	"context"
	"net/http"
	"net/url"
	"strings"

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
	GetRepositories(owner string, ownerType string, getForks bool, getArchived bool, visibility RepositoryVisibility) ([]Repository, error)
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

type Repository struct {
	Name string  `json:"name"`
	Url  url.URL `json:"url"`
}

// Returns repositories for a user or organization
func (g *githubClient) GetRepositories(owner string, ownerType string, getForks bool, getArchived bool, visibility RepositoryVisibility) ([]Repository, error) {
	_ = `# @genqlient
	query getRepositories(
    $owner: String!
    $isUser: Boolean!
    $isOrg: Boolean!
    $isFork: Boolean!
    $isArchived: Boolean!
    $visibility: RepositoryVisibility!
    $cursor: String
) {
    # For an organization
    organization(login: $owner) @include(if: $isOrg) {
        repositories(
            first: 100
            after: $cursor
            isFork: $isFork
            isArchived: $isArchived
            visibility: $visibility
            orderBy: { field: NAME, direction: ASC }
        ) {
            pageInfo {
                hasNextPage
                endCursor
            }
            nodes {
                name
                url
            }
        }
    }

    # For a user
    user(login: $owner) @include(if: $isUser) {
        repositories(
            first: 100
            after: $cursor
            isFork: $isFork
            isArchived: $isArchived
            visibility: $visibility
            orderBy: { field: NAME, direction: ASC }
        ) {
            pageInfo {
                hasNextPage
                endCursor
            }
            nodes {
                name
                url
            }
        }
    }
}`
	repos := []Repository{}
	ctx := context.Background()
	hasNextPage := true
	endCursor := ""

	var isUser, isOrg bool
	switch strings.ToLower(ownerType) {
	case "user":
		isUser = true
	case "organization":
		isOrg = true
	}

	for hasNextPage {
		resp, err := getRepositories(ctx, g.client, owner, isUser, isOrg, getForks, getArchived, visibility, endCursor)
		if err != nil {
			return nil, err
		}

		if isUser {
			for _, v := range resp.User.Repositories.Nodes {
				repos = append(repos, Repository{
					Name: v.Name,
					Url:  v.Url,
				})
			}

			hasNextPage = resp.User.Repositories.PageInfo.HasNextPage
			endCursor = resp.User.Repositories.PageInfo.EndCursor
		} else {
			for _, v := range resp.Organization.Repositories.Nodes {
				repos = append(repos, Repository{
					Name: v.Name,
					Url:  v.Url,
				})
			}

			hasNextPage = resp.Organization.Repositories.PageInfo.HasNextPage
			endCursor = resp.Organization.Repositories.PageInfo.EndCursor
		}
	}

	return repos, nil
}
