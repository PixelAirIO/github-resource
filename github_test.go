package githubresource

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Khan/genqlient/graphql"
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

func TestGetAccessTokenValidation(t *testing.T) {
	tests := []struct {
		name           string
		cfg            Config
		token, message string
	}{
		{name: "anonymous"},
		{name: "explicit token", cfg: Config{AccessToken: "secret"}, token: "secret"},
		{name: "mixed", cfg: Config{AccessToken: "secret", AppClientID: "client"}, message: "both access_token and app_*"},
		{name: "missing key and installation", cfg: Config{AppClientID: "client"}, message: "app_private_key, app_install_id"},
		{name: "missing client", cfg: Config{AppPrivateKey: "key", AppInstallID: "1"}, message: "app_client_id"},
		{name: "missing key", cfg: Config{AppClientID: "client", AppInstallID: "1"}, message: "app_private_key"},
		{name: "missing installation", cfg: Config{AppClientID: "client", AppPrivateKey: "key"}, message: "app_install_id"},
	}

	for _, id := range []string{"abc", "0", "-1", "9223372036854775808"} {
		tests = append(tests, struct {
			name           string
			cfg            Config
			token, message string
		}{
			name: "invalid ID " + id,
			cfg: Config{
				AppClientID:   "client",
				AppPrivateKey: "key",
				AppInstallID:  id},
			message: "app_install_id must be a positive integer",
		})
	}
	for _, key := range []string{
		"not PEM",
		string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: []byte("invalid")})),
		string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("invalid")})),
		string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: []byte("invalid")})),
	} {
		tests = append(tests, struct {
			name           string
			cfg            Config
			token, message string
		}{
			name: "invalid key " + key,
			cfg: Config{
				AppClientID:   "client",
				AppPrivateKey: key,
				AppInstallID:  "1"},
			message: "app_private_key",
		})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := getAccessToken(tt.cfg, http.DefaultClient)
			require.Equal(t, tt.token, token)
			if tt.message == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.message)
			}
		})
	}
	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	der, err := x509.MarshalPKCS8PrivateKey(ecKey)
	require.NoError(t, err)
	_, err = getAccessToken(Config{
		AppClientID:   "client",
		AppInstallID:  "1",
		AppPrivateKey: string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})),
	}, http.DefaultClient)
	require.ErrorContains(t, err, "must be an RSA private key")
}

func TestGitHubAppAuthentication(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	pkcs8, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	for _, format := range []struct {
		name string
		der  []byte
	}{
		{name: "RSA PRIVATE KEY", der: x509.MarshalPKCS1PrivateKey(key)},
		{name: "PRIVATE KEY", der: pkcs8},
	} {
		t.Run(format.name, func(t *testing.T) {
			started := time.Now().Unix()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/graphql" || r.URL.Path == "/api/v3/user" {
					require.Equal(t, "bearer installation-token", strings.ToLower(r.Header.Get("Authorization")))
					w.Header().Set("Content-Type", "application/json")
					if r.URL.Path == "/graphql" {
						_, _ = w.Write([]byte(`{"data":{}}`))
					} else {
						_, _ = w.Write([]byte(`{}`))
					}
					return
				}
				require.Equal(t, "POST", r.Method)
				require.Equal(t, "/api/v3/app/installations/42/access_tokens", r.URL.Path)
				require.True(t, strings.HasPrefix(r.Header.Get("Authorization"), "Bearer "))
				parts := strings.Split(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "), ".")
				require.Len(t, parts, 3)
				header, err := base64.RawURLEncoding.DecodeString(parts[0])
				require.NoError(t, err)
				require.JSONEq(t, `{"alg":"RS256","typ":"JWT"}`, string(header))
				payload, err := base64.RawURLEncoding.DecodeString(parts[1])
				require.NoError(t, err)
				var claims struct {
					Issuer    string `json:"iss"`
					IssuedAt  int64  `json:"iat"`
					ExpiresAt int64  `json:"exp"`
				}
				require.NoError(t, json.Unmarshal(payload, &claims))
				require.Equal(t, "client-id", claims.Issuer)
				require.GreaterOrEqual(t, claims.IssuedAt, started-60)
				require.LessOrEqual(t, claims.IssuedAt, time.Now().Unix()-60)
				require.Equal(t, int64(660), claims.ExpiresAt-claims.IssuedAt)
				signature, err := base64.RawURLEncoding.DecodeString(parts[2])
				require.NoError(t, err)
				digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
				require.NoError(t, rsa.VerifyPKCS1v15(&key.PublicKey, crypto.SHA256, digest[:], signature))
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"token":"installation-token"}`))
			}))
			defer server.Close()
			cfg := Config{Repository: "owner/repo", AppClientID: "client-id", AppInstallID: "42", AppPrivateKey: string(pem.EncodeToMemory(&pem.Block{Type: format.name, Bytes: format.der})), RestEndpoint: server.URL + "/api/v3/", GraphqlEndpoint: server.URL + "/graphql"}
			client, err := NewGithubClient(cfg)
			require.NoError(t, err)
			require.Equal(t, "installation-token", client.AccessToken())
			concrete := client.(*githubClient)
			require.Contains(t, concrete.cliEnv, "X_OAUTH_BASIC_TOKEN=installation-token")
			_, _, err = concrete.restClient.Users.Get(context.Background(), "")
			require.NoError(t, err)
			require.NoError(t, concrete.gqlClient.MakeRequest(context.Background(), &graphql.Request{Query: "{ viewer { login } }"}, &graphql.Response{}))
		})
	}
}

func TestGitHubAppTokenExchangeErrors(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	cfg := Config{AppClientID: "client", AppInstallID: "1", AppPrivateKey: string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}))}
	for _, tt := range []struct {
		name          string
		status        int
		body, message string
	}{
		{name: "API error", status: http.StatusUnauthorized, body: `{"message":"Bad credentials"}`, message: "401"},
		{name: "empty token", status: http.StatusCreated, body: `{}`, message: "empty token"},
		{name: "invalid JSON", status: http.StatusCreated, body: `invalid`, message: "creating GitHub App installation token"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()
			cfg.RestEndpoint = server.URL + "/"
			token, err := getAccessToken(cfg, http.DefaultClient)
			require.Empty(t, token)
			require.ErrorContains(t, err, tt.message)
		})
	}
	cfg.RestEndpoint = ":invalid"
	_, err = getAccessToken(cfg, http.DefaultClient)
	require.ErrorContains(t, err, "parsing rest_endpoint")
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"token":"tls-token"}`))
	}))
	defer server.Close()
	cfg.RestEndpoint = server.URL + "/"
	_, err = getAccessToken(cfg, http.DefaultClient)
	require.Error(t, err)
	server.Close()
	_, err = getAccessToken(cfg, http.DefaultClient)
	require.ErrorContains(t, err, "creating GitHub App installation token")
}

func TestNewGithubClientPropagatesAuthenticationError(t *testing.T) {
	client, err := NewGithubClient(Config{Repository: "owner/repo", AppClientID: "client"})
	require.Nil(t, client)
	require.ErrorContains(t, err, "app_private_key, app_install_id")
}
