package migration

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	bitbucketv1 "github.com/gfleury/go-bitbucket-v1"
	"github.com/spf13/viper"
)

const (
	// project permission
	ProjectAdmin = "PROJECT_ADMIN"
	ProjectWrite = "PROJECT_WRITE"
	ProjectRead  = "PROJECT_READ"
	// repo permission
	RepoAdmin  = "REPO_ADMIN"
	RepoWrite  = "REPO_WRITE"
	RepoRead   = "REPO_READ"
	RepoCreate = "REPO_CREATE"
)

// NewBitbucket creates a new instance of the bitbucket struct.
func NewBitbucket(ctx context.Context, logger *slog.Logger) (*bitbucket, error) {
	b := &bitbucket{
		ctx:      ctx,
		server:   viper.GetString("bitbucket.server"),
		Token:    viper.GetString("bitbucket.token"),
		Username: viper.GetString("bitbucket.username"),
		logger:   logger,
	}

	err := b.init()
	if err != nil {
		return nil, err
	}

	return b, nil
}

// bitbucket is a struct that holds the bitbucket client.
type bitbucket struct {
	ctx      context.Context
	server   string
	Token    string
	Username string
	client   *bitbucketv1.APIClient
	logger   *slog.Logger
}

// init initializes the bitbucket client.
func (b *bitbucket) init() error {
	if b.server == "" || b.Username == "" || b.Token == "" {
		return errors.New("mission bitbucket server, username or token")
	}

	b.server = strings.TrimRight(b.server, "/")

	ctx := context.WithValue(b.ctx, bitbucketv1.ContextAccessToken, b.Token)
	b.client = bitbucketv1.NewAPIClient(
		ctx,
		bitbucketv1.NewConfiguration(
			b.server+"/rest",
			func(cfg *bitbucketv1.Configuration) {
				certs, _ := x509.SystemCertPool()
				// add new http client for skip verify
				httpClient := &http.Client{
					Transport: &http.Transport{
						TLSClientConfig: &tls.Config{
							RootCAs:            certs,
							InsecureSkipVerify: true,
						},
					},
				}
				cfg.HTTPClient = httpClient
			},
		),
	)

	return nil
}

// GetUsersPermissionFromProject get users permission from project
func (b *bitbucket) GetUsersPermissionFromProject(projectKey string) ([]bitbucketv1.UserPermission, error) {
	// check project user permission
	response, err := b.client.DefaultApi.GetUsersWithAnyPermission_23(
		projectKey,
		map[string]interface{}{
			"limit": 200,
		})
	if err != nil {
		return nil, err
	}

	return bitbucketv1.GetUsersPermissionResponse(response)
}

// GetUsersPermissionFromRepo get users permission from repo
func (b *bitbucket) GetUsersPermissionFromRepo(projectKey, repoSlug string) ([]bitbucketv1.UserPermission, error) {
	// check project user permission
	response, err := b.client.DefaultApi.GetUsersWithAnyPermission_24(
		projectKey,
		repoSlug,
		map[string]interface{}{
			"limit": 200,
		})
	if err != nil {
		return nil, err
	}

	return bitbucketv1.GetUsersPermissionResponse(response)
}

// GetGroupsPermissionFromProject get groups permission from project
func (b *bitbucket) GetGroupsPermissionFromProject(projectKey string) ([]bitbucketv1.GroupPermission, error) {
	// check project group permission
	response, err := b.client.DefaultApi.GetGroupsWithAnyPermission_12(
		projectKey,
		map[string]interface{}{
			"limit": 200,
		})
	if err != nil {
		return nil, err
	}

	return bitbucketv1.GetGroupsPermissionResponse(response)
}

// GetGroupsPermissionFromRepo get groups permission from repo
func (b *bitbucket) GetGroupsPermissionFromRepo(projectKey, repoSlug string) ([]bitbucketv1.GroupPermission, error) {
	// check project group permission
	response, err := b.client.DefaultApi.GetGroupsWithAnyPermission_13(
		projectKey,
		repoSlug,
		map[string]interface{}{
			"limit": 200,
		})
	if err != nil {
		return nil, err
	}

	return bitbucketv1.GetGroupsPermissionResponse(response)
}

// GetUsersFromGroup get users from group
func (b *bitbucket) GetUsersFromGroup(g string) ([]bitbucketv1.User, error) {
	response, err := b.client.DefaultApi.FindUsersInGroup(
		map[string]interface{}{
			"context": g,
			"limit":   200,
		},
	)
	if err != nil {
		return nil, err
	}

	return bitbucketv1.GetUsersResponse(response)
}

// GetProject get project
func (b *bitbucket) GetProject(projectKey string) (bitbucketv1.Project, error) {
	response, err := b.client.DefaultApi.GetProject(projectKey)
	if err != nil {
		return bitbucketv1.Project{}, err
	}

	return bitbucketv1.GetRrojectResponse(response)
}

// GetRepo get repo
func (b *bitbucket) GetRepo(projectKey, repoSlug string) (bitbucketv1.Repository, error) {
	response, err := b.client.DefaultApi.GetRepository(projectKey, repoSlug)
	if err != nil {
		return bitbucketv1.Repository{}, err
	}

	return bitbucketv1.GetRepositoryResponse(response)
}

// GetRepositories get repositories from project
func (b *bitbucket) GetRepositories(projectKey string) ([]bitbucketv1.Repository, error) {
	response, err := b.client.DefaultApi.GetRepositoriesWithOptions(projectKey, map[string]interface{}{
		"limit": 200,
	})
	if err != nil {
		return nil, err
	}

	return bitbucketv1.GetRepositoriesResponse(response)
}
