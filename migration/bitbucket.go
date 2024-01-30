package migration

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	bitbucketv1 "github.com/gfleury/go-bitbucket-v1"
	"github.com/spf13/viper"
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
		bitbucketv1.NewConfiguration(b.server+"/rest"),
	)
	return nil
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
