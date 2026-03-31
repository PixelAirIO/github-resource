package githubresource

//go:generate go run github.com/Khan/genqlient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/Khan/genqlient/graphql"
	"github.com/google/go-github/v84/github"
)

//go:generate go tool counterfeiter -generate

type Config struct {
	AccessToken   string `json:"access_token"`
	APIEndpointV4 string `json:"api_endpoint_v4"`
	APIEndpointV3 string `json:"api_endpoint_v3"`
	HostEndpoint  string `json:"host_endpoint"`
	Repository    string `json:"repository"`
}

//counterfeiter:generate . GithubClient
type GithubClient interface {
	// Returns the REST and GraphQL endpoints
	APIEndpoints() (string, string)

	HostEndpoint() string
	AccessToken() string

	// Returns pull requests matching the states and labels provided.
	//
	// If you want to match against no labels, pass in nil.
	ListPullRequests(states []PullRequestState, labels []string) ([]PullRequest, error)

	// Returns the latest commit SHA for a given PR
	LatestCommitForPR(int) (string, error)

	// Updates the status for a given ref
	UpdatePRStatus(ref string, name string, status string) error
}

type githubClient struct {
	gqlClient  graphql.Client
	restClient *github.Client
	owner      string
	repo       string
	config     Config
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
const DefaultRestEndpoint = "https://api.github.com/"
const DefaultHostEndpoint = "https://github.com"

func NewGithubClient(cfg Config) (GithubClient, error) {
	if cfg.APIEndpointV4 == "" {
		cfg.APIEndpointV4 = DefaultGraphqlEndpoint
	}

	if cfg.APIEndpointV3 == "" {
		cfg.APIEndpointV3 = DefaultRestEndpoint
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

	gqlClient := graphql.NewClient(cfg.APIEndpointV4, &http.Client{
		Transport: &authedTransport{
			accessToken: cfg.AccessToken,
			transport:   http.DefaultTransport,
		},
	})

	ghc := github.NewClient(nil)
	if cfg.AccessToken != "" {
		ghc = ghc.WithAuthToken(cfg.AccessToken)
	}

	if cfg.APIEndpointV3 != DefaultRestEndpoint {
		u, err := url.Parse(cfg.APIEndpointV3)
		if err != nil {
			return nil, err
		}
		ghc.BaseURL = u
	}

	return &githubClient{
		gqlClient:  gqlClient,
		restClient: ghc,
		owner:      repository[0],
		repo:       repository[1],
		config:     cfg,
	}, nil
}

func (g *githubClient) APIEndpoints() (string, string) {
	return g.restClient.BaseURL.String(), g.config.APIEndpointV4
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
		resp, err := getPullRequests(ctx, g.gqlClient, g.owner, g.repo, states, labels, endCursor)
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
	resp, err := latestCommitForPr(ctx, g.gqlClient, g.owner, g.repo, prNumber)
	if err != nil {
		return "", err
	}

	if len(resp.Repository.PullRequest.Commits.Nodes) < 1 {
		return "", errors.New("no commits found for the given PR")
	}

	return resp.Repository.PullRequest.Commits.Nodes[0].Commit.Oid, nil
}

func (g *githubClient) UpdatePRStatus(ref string, name string, status string) error {
	targetUrl := os.Getenv("BUILD_URL_SHORT")
	if targetUrl == "" {
		targetUrl = fmt.Sprintf("%s/builds/%s", os.Getenv("ATC_EXTERNAL_URL"), os.Getenv("BUILD_ID"))
	}

	_, _, err := g.restClient.Repositories.CreateStatus(context.TODO(), g.owner, g.repo, ref, github.RepoStatus{
		State:     &status,
		Context:   &name,
		TargetURL: &targetUrl,
	})

	return err
}
