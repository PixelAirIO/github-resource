package githubresource

//go:generate go run github.com/Khan/genqlient

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Khan/genqlient/graphql"
)

//go:generate go tool counterfeiter -generate

type Config struct {
	AccessToken  string `json:"access_token"`
	APIEndpoint  string `json:"api_endpoint"`
	HostEndpoint string `json:"host_endpoint"`
	Repository   string `json:"repository"`
}

//counterfeiter:generate . GithubClient
type GithubClient interface {
	APIEndpoint() string
	HostEndpoint() string
	AccessToken() string

	// Returns pull requests matching the states and labels provided.
	//
	// If you want to match against no labels, pass in nil.
	ListPullRequests(states []PullRequestState, labels []string) ([]PullRequest, error)

	// Returns the latest commit SHA for a given PR
	//
	LatestCommitForPR(int) (string, error)
}

type githubClient struct {
	client graphql.Client
	owner  string
	repo   string
	config Config
}

var _ GithubClient = (*githubClient)(nil)

type authedTransport struct {
	accessToken string
	transport   http.RoundTripper
}

func (a *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if a.accessToken != "" {
		req.Header.Set("Authorization", "bearer "+a.accessToken)
	}
	return a.transport.RoundTrip(req)
}

const DefaultGraphqlEndpoint = "https://api.github.com/graphql"
const DefaultHostEndpoint = "https://github.com"

func NewGithubClient(cfg Config) (GithubClient, error) {
	if cfg.APIEndpoint == "" {
		cfg.APIEndpoint = DefaultGraphqlEndpoint
	}

	if cfg.HostEndpoint == "" {
		cfg.HostEndpoint = DefaultHostEndpoint
	}

	if cfg.Repository == "" {
		return nil, errors.New("repository is blank and must be set. Expected format is 'OWNER/REPO'.")
	}

	repository := strings.Split(cfg.Repository, "/")
	if len(repository) != 2 {
		return nil, errors.New("unexpected format for 'repository'. Expected format is 'OWNER/REPO'.")
	}

	client := graphql.NewClient(cfg.APIEndpoint, &http.Client{
		Transport: &authedTransport{
			accessToken: cfg.AccessToken,
			transport:   http.DefaultTransport,
		},
	})

	return &githubClient{
		client: client,
		owner:  repository[0],
		repo:   repository[1],
		config: cfg,
	}, nil
}

func (g *githubClient) APIEndpoint() string {
	return g.config.APIEndpoint
}

func (g *githubClient) HostEndpoint() string {
	return g.config.HostEndpoint
}

func (g *githubClient) AccessToken() string {
	return g.config.AccessToken
}

type PullRequest struct {
	Number       string `json:"number"`
	Url          string `json:"url"`
	IsDraft      bool   `json:"-"`
	TargetBranch string `json:"target_branch"`
}

func (g *githubClient) ListPullRequests(states []PullRequestState, labels []string) ([]PullRequest, error) {
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
		resp, err := getPullRequests(ctx, g.client, g.owner, g.repo, states, labels, endCursor)
		if err != nil {
			return nil, err
		}

		for _, v := range resp.Repository.PullRequests.Nodes {
			prs = append(prs, PullRequest{
				Number:       strconv.Itoa(v.Number),
				Url:          v.Permalink,
				IsDraft:      v.IsDraft,
				TargetBranch: v.BaseRefName,
			})
		}

		hasNextPage = resp.Repository.PullRequests.PageInfo.HasNextPage
		endCursor = resp.Repository.PullRequests.PageInfo.EndCursor
	}

	return prs, nil
}

func (g *githubClient) LatestCommitForPR(prNumber int) (string, error) {
	_ = `# @genqlient
query latestCommitForPr(
    $owner: String!
    $name: String!
    $number: Int!
) {
    repository(owner: $owner, name: $name) {
        pullRequest(number: $number) {
            baseRefName
            permalink
            commits(last: 1) {
                nodes {
                    commit {
                        oid
                    }
                }
            }
        }
    }
}`

	ctx := context.Background()
	resp, err := latestCommitForPr(ctx, g.client, g.owner, g.repo, prNumber)
	if err != nil {
		return "", err
	}

	if len(resp.Repository.PullRequest.Commits.Nodes) < 1 {
		return "", errors.New("no commits found for the given PR")
	}

	return resp.Repository.PullRequest.Commits.Nodes[0].Commit.Oid, nil
}
