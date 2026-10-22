package githubresource

//go:generate go run github.com/Khan/genqlient

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/google/go-github/v84/github"
)

//go:generate go tool counterfeiter -generate

type Config struct {
	AccessToken         string `json:"access_token"`
	AppClientID         string `json:"app_client_id"`
	AppPrivateKey       string `json:"app_private_key"`
	AppInstallID        string `json:"app_install_id"`
	GraphqlEndpoint     string `json:"graphql_endpoint"`
	RestEndpoint        string `json:"rest_endpoint"`
	HostEndpoint        string `json:"host_endpoint"`
	Repository          string `json:"repository"`
	DisableGitLFS       bool   `json:"disable_git_lfs"`
	SkipSSLVerification bool   `json:"skip_ssl_verification"`
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
	// PullRequest.FilesChanged is NOT populated.
	ListPullRequests(states []PullRequestState, labels []string) ([]PullRequest, error)

	// Returns the latest commit SHA for a given PR
	LatestCommitForPR(int) (string, error)

	// Returns information about the Pull Request. Only the first 100 files are
	// listed.
	GetPRInfo(int) (PullRequest, error)

	// Returns the latest commits for matching PRs
	LatestCommitsFromPrs(states []PullRequestState, labels []string) ([]PRCommit, error)

	// Updates the status for a given ref
	UpdatePRStatus(ref, name, status, descr string) error

	// Configures the repo, initializing at the specified branch
	InitRepo(uri, branch string) error

	// Does `git fetch` to download the PR's data
	FetchPr(uri, number string, depth int, fetchTags, submodules bool) error

	// Pulls a known branch from origin
	PullBranch(branch string, depth int, fetchTags, submodules bool) error

	CheckoutPr(prBranch, ref string, submodules bool) error
	RebasePr(baseRef, prRef string, submodules bool) error
	MergePr(prRef string, submodules bool) error
}

type githubClient struct {
	gqlClient  graphql.Client
	restClient *github.Client
	owner      string
	repo       string
	config     Config
	cliEnv     []string
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
	_, err := exec.LookPath("git")
	if err != nil {
		return nil, fmt.Errorf("error checking for the git cli: %w", err)
	}

	if cfg.GraphqlEndpoint == "" {
		cfg.GraphqlEndpoint = DefaultGraphqlEndpoint
	}

	if cfg.RestEndpoint == "" {
		cfg.RestEndpoint = DefaultRestEndpoint
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

	var httpClient *http.Client
	var transport http.RoundTripper
	if cfg.SkipSSLVerification {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
		httpClient = &http.Client{
			Transport: transport,
		}
	} else {
		httpClient = http.DefaultClient
		transport = http.DefaultTransport
	}

	accessToken, err := getAccessToken(cfg, httpClient)
	if err != nil {
		return nil, err
	}
	cfg.AccessToken = accessToken

	gqlClient := graphql.NewClient(cfg.GraphqlEndpoint, &http.Client{
		Transport: &authedTransport{
			accessToken: cfg.AccessToken,
			transport:   transport,
		},
	})

	ghc := github.NewClient(httpClient)
	if cfg.AccessToken != "" {
		ghc = ghc.WithAuthToken(cfg.AccessToken)
	}

	if cfg.RestEndpoint != DefaultRestEndpoint {
		u, err := url.Parse(cfg.RestEndpoint)
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
		cliEnv: []string{
			fmt.Sprintf("%s=%s", "X_OAUTH_BASIC_TOKEN", cfg.AccessToken),
			fmt.Sprintf("%s=%t", "GIT_LFS_SKIP_SMUDGE", cfg.DisableGitLFS),
			"GIT_ASKPASS=/usr/local/bin/gitpass.sh",
			"GIT_TERMINAL_PROMPT=0",
		},
	}, nil
}

func getAccessToken(cfg Config, httpClient *http.Client) (string, error) {
	if cfg.AccessToken == "" &&
		(cfg.AppClientID == "" && cfg.AppPrivateKey == "" && cfg.AppInstallID == "") {
		return "", nil
	}

	if cfg.AccessToken != "" &&
		(cfg.AppClientID != "" || cfg.AppPrivateKey != "" || cfg.AppInstallID != "") {
		return "", errors.New("both access_token and app_* settings are provided. Only one auth method can be used. Please fix the resource's source configuration.")
	}

	if cfg.AccessToken != "" {
		return cfg.AccessToken, nil
	}

	var missing []string
	for _, field := range []struct{ name, value string }{
		{"app_client_id", cfg.AppClientID},
		{"app_private_key", cfg.AppPrivateKey},
		{"app_install_id", cfg.AppInstallID},
	} {
		if field.value == "" {
			missing = append(missing, field.name)
		}
	}
	if len(missing) > 0 {
		return "", fmt.Errorf("missing required GitHub App settings: %s", strings.Join(missing, ", "))
	}

	installationID, err := strconv.ParseInt(cfg.AppInstallID, 10, 64)
	if err != nil || installationID <= 0 {
		return "", errors.New("app_install_id must be a positive integer")
	}

	block, _ := pem.Decode([]byte(cfg.AppPrivateKey))
	if block == nil {
		return "", errors.New("app_private_key must contain a PEM-encoded RSA private key")
	}

	var key *rsa.PrivateKey
	switch block.Type {
	case "RSA PRIVATE KEY":
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)

	case "PRIVATE KEY":
		var parsed any
		parsed, err = x509.ParsePKCS8PrivateKey(block.Bytes)
		if err == nil {
			var ok bool
			key, ok = parsed.(*rsa.PrivateKey)
			if !ok {
				return "", errors.New("app_private_key must be an RSA private key")
			}
		}

	default:
		return "", errors.New("app_private_key must contain a PKCS#1 or PKCS#8 RSA private key")
	}
	if err != nil {
		return "", fmt.Errorf("parsing app_private_key: %w", err)
	}

	// GitHub recommends backdating iat by a minute to allow for clock drift.
	now := time.Now()
	claims, err := json.Marshal(struct {
		Issuer    string `json:"iss"`
		IssuedAt  int64  `json:"iat"`
		ExpiresAt int64  `json:"exp"`
	}{cfg.AppClientID, now.Add(-time.Minute).Unix(), now.Add(10 * time.Minute).Unix()})
	if err != nil {
		return "", fmt.Errorf("encoding GitHub App JWT: %w", err)
	}
	unsigned := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`)) + "." + base64.RawURLEncoding.EncodeToString(claims)
	digest := sha256.Sum256([]byte(unsigned))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		return "", fmt.Errorf("signing GitHub App JWT: %w", err)
	}
	jwt := unsigned + "." + base64.RawURLEncoding.EncodeToString(signature)

	client := github.NewClient(httpClient).WithAuthToken(jwt)
	if cfg.RestEndpoint != "" {
		client.BaseURL, err = url.Parse(cfg.RestEndpoint)
		if err != nil {
			return "", fmt.Errorf("parsing rest_endpoint: %w", err)
		}
	}
	// Repository APIs, GraphQL, and Git need an installation token, not the app JWT.
	token, _, err := client.Apps.CreateInstallationToken(context.Background(), installationID, nil)
	if err != nil {
		return "", fmt.Errorf("creating GitHub App installation token: %w", err)
	}
	if token.GetToken() == "" {
		return "", errors.New("GitHub App installation token response contained an empty token")
	}
	return token.GetToken(), nil
}

func (g *githubClient) APIEndpoints() (string, string) {
	return g.restClient.BaseURL.String(), g.config.GraphqlEndpoint
}

func (g *githubClient) HostEndpoint() string {
	return g.config.HostEndpoint
}

func (g *githubClient) AccessToken() string {
	return g.config.AccessToken
}

type PullRequest struct {
	Number        string   `json:"number"`
	Url           string   `json:"url,omitempty"`
	IsDraft       bool     `json:"-"`
	TargetBranch  string   `json:"target_branch,omitempty"`
	FilesChanged  []string `json:"changed_files,omitempty"`
	ParentRepoUrl string   `json:"parent_url,omitempty"`
	Branch        string   `json:"branch,omitempty"`
	Author        string   `json:"author,omitempty"`
	Title         string   `json:"title,omitempty"`
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
                headRefName
                author {
                    login
                }
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
				Branch:       v.HeadRefName,
				Author:       v.Author.GetLogin(),
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

func (g *githubClient) GetPRInfo(prNumber int) (PullRequest, error) {
	_ = `# @genqlient
query getPullRequest(
   $owner: String!
   $name: String!
   $number: Int!
) {
    repository(owner: $owner, name: $name) {
        url
        pullRequest(number: $number) {
            number
            title
            isDraft
            permalink
            baseRefName
            headRefName
            author {
                login
            }
            files(first: 100) {
                nodes {
                    path
                }
            }
        }
    }
}
`
	ctx := context.Background()
	resp, err := getPullRequest(ctx, g.gqlClient, g.owner, g.repo, prNumber)
	if err != nil {
		return PullRequest{}, err
	}

	files := []string{}
	for _, p := range resp.Repository.PullRequest.Files.Nodes {
		files = append(files, p.Path)
	}

	return PullRequest{
		Number:        strconv.Itoa(prNumber),
		Url:           resp.Repository.PullRequest.Permalink,
		IsDraft:       resp.Repository.PullRequest.IsDraft,
		TargetBranch:  resp.Repository.PullRequest.BaseRefName,
		FilesChanged:  files,
		ParentRepoUrl: resp.Repository.Url,
		Branch:        resp.Repository.PullRequest.HeadRefName,
		Author:        resp.Repository.PullRequest.Author.GetLogin(),
		Title:         resp.Repository.PullRequest.Title,
	}, nil
}

type PRCommit struct {
	PullRequest
	Ref      string `json:"ref"`
	Date     string `json:"date"`
	Headline string `json:"headline"`
}

func (g *githubClient) LatestCommitsFromPrs(states []PullRequestState, labels []string) ([]PRCommit, error) {
	_ = `# @genqlient
query latestCommitsFromPrs(
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
			orderBy: {field: UPDATED_AT, direction: ASC}
		) {
            nodes {
                number
                isDraft
                title
                permalink
                baseRefName
                headRefName
                url
                author {
                    login
                }
                commits(last: 1) {
                  nodes {
                    commit {
                      oid
                      messageHeadline
                      committedDate
                    }
                  }
                }
            }
            pageInfo {
	            endCursor
                hasNextPage
            }
        }
    }
}`
	prs := []PRCommit{}
	ctx := context.Background()
	hasNextPage := true
	endCursor := ""

	for hasNextPage {
		resp, err := latestCommitsFromPrs(ctx, g.gqlClient, g.owner, g.repo, states, labels, endCursor)
		if err != nil {
			return nil, err
		}

		for _, v := range resp.Repository.PullRequests.Nodes {
			if len(v.Commits.Nodes) > 0 {
				prs = append(prs, PRCommit{
					PullRequest: PullRequest{
						Number:       strconv.Itoa(v.Number),
						Url:          v.Permalink,
						IsDraft:      v.IsDraft,
						TargetBranch: v.BaseRefName,
						Branch:       v.HeadRefName,
						Author:       v.Author.GetLogin(),
					},
					Ref:      v.Commits.Nodes[0].Commit.Oid,
					Date:     v.Commits.Nodes[0].Commit.CommittedDate,
					Headline: v.Commits.Nodes[0].Commit.MessageHeadline,
				})
			} else {
				log.Printf("PR '%d' has no commits\n", v.Number)
			}
		}

		hasNextPage = resp.Repository.PullRequests.PageInfo.HasNextPage
		endCursor = resp.Repository.PullRequests.PageInfo.EndCursor
	}

	return prs, nil
}

func (g *githubClient) UpdatePRStatus(ref, name, status, descr string) error {
	targetUrl := os.Getenv("BUILD_URL_SHORT")
	if targetUrl == "" {
		targetUrl = fmt.Sprintf("%s/builds/%s", os.Getenv("ATC_EXTERNAL_URL"), os.Getenv("BUILD_ID"))
	}

	_, _, err := g.restClient.Repositories.CreateStatus(context.TODO(), g.owner, g.repo, ref, github.RepoStatus{
		State:       &status,
		Context:     &name,
		Description: &descr,
		TargetURL:   &targetUrl,
	})

	return err
}

func (g *githubClient) git(args ...string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, g.cliEnv...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd
}

func (g *githubClient) endpoint(uri string) (string, error) {
	endpoint, err := url.Parse(uri)
	if err != nil {
		return "", fmt.Errorf("failed to parse uri: %w", err)
	}
	endpoint.User = url.UserPassword("x-oauth-basic", g.config.AccessToken)
	return endpoint.String(), nil
}

func (g *githubClient) InitRepo(uri, branch string) error {
	err := g.git("init", "--initial-branch", branch).Run()
	if err != nil {
		return fmt.Errorf("git init error: %w", err)
	}

	err = g.git("config", "user.name", "concourse-ci").Run()
	if err != nil {
		return fmt.Errorf("git config user.name error: %w", err)
	}

	err = g.git("config", "user.email", "concourse@local").Run()
	if err != nil {
		return fmt.Errorf("git config user.email error: %w", err)
	}

	err = g.git("config", "url.https://x-oauth-basic@github.com/.insteadOf", "git@github.com:").Run()
	if err != nil {
		return fmt.Errorf("git config url-1 error: %w", err)
	}

	err = g.git("config", "url.https://.insteadOf", "git://").Run()
	if err != nil {
		return fmt.Errorf("git config url-2 error: %w", err)
	}

	remoteUri, err := g.endpoint(uri)
	if err != nil {
		return err
	}

	err = g.git("remote", "add", "origin", remoteUri).Run()
	if err != nil {
		return fmt.Errorf("error setting origin: %w", err)
	}

	return nil
}

func (g *githubClient) PullBranch(branch string, depth int, fetchTags, submodules bool) error {
	pullArgs := []string{"pull", "origin", branch}
	if depth > 0 {
		pullArgs = append(pullArgs, "--depth", strconv.Itoa(depth))
	}
	if fetchTags {
		pullArgs = append(pullArgs, "--tags")
	}
	if submodules {
		pullArgs = append(pullArgs, "--recurse-submodules")
	}

	err := g.git(pullArgs...).Run()
	if err != nil {
		return fmt.Errorf("error pulling origin: %w", err)
	}

	if submodules {
		err = g.git("submodule", "update", "--init", "--recursive").Run()
		if err != nil {
			return fmt.Errorf("error updating submodules: %w", err)
		}
	}

	return nil
}

func (g *githubClient) FetchPr(uri, number string, depth int, fetchTags, submodules bool) error {
	remoteUri, err := g.endpoint(uri)
	if err != nil {
		return err
	}

	args := []string{"fetch", remoteUri, fmt.Sprintf("pull/%s/head", number)}
	if depth > 0 {
		args = append(args, "--depth", strconv.Itoa(depth))
	}
	if fetchTags {
		args = append(args, "--tags")
	}
	if submodules {
		args = append(args, "--recurse-submodules")
	}

	err = g.git(args...).Run()
	if err != nil {
		return fmt.Errorf("error fetching PR: %w", err)
	}

	return nil
}

func (g *githubClient) CheckoutPr(prBranch, ref string, submodules bool) error {
	err := g.git("checkout", "-b", prBranch, ref).Run()
	if err != nil {
		return fmt.Errorf("error checking out PR: %w", err)
	}

	if submodules {
		err = g.git("submodule", "update", "--init", "--recursive", "--checkout").Run()
		if err != nil {
			return fmt.Errorf("error updating submodules: %w", err)
		}
	}

	return nil
}

func (g *githubClient) RebasePr(baseRef, prRef string, submodules bool) error {
	err := g.git("rebase", baseRef, prRef).Run()
	if err != nil {
		return fmt.Errorf("error rebasing PR: %w", err)
	}

	if submodules {
		err = g.git("submodule", "update", "--init", "--recursive", "--rebase").Run()
		if err != nil {
			return fmt.Errorf("error updating submodules: %w", err)
		}
	}

	return nil
}

func (g *githubClient) MergePr(prRef string, submodules bool) error {
	err := g.git("merge", prRef, "--no-stat").Run()
	if err != nil {
		return fmt.Errorf("error merging PR: %w", err)
	}

	if submodules {
		err = g.git("submodule", "update", "--init", "--recursive", "--merge").Run()
		if err != nil {
			return fmt.Errorf("error updating submodules: %w", err)
		}
	}

	return nil
}
