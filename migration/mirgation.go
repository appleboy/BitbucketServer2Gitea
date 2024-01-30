package migration

import (
	"context"
	"log/slog"
	"os"
	"strings"

	bitbucketv1 "github.com/gfleury/go-bitbucket-v1"
)

type migration struct {
	ctx       context.Context
	Bitbucket *bitbucket
	Gitea     *gitea
	Logger    *slog.Logger
}

// Option migration option
type Option struct {
	Debug bool
}

// NewMigration creates a new instance of the migration struct.
func NewMigration(ctx context.Context, opts Option) (*migration, error) {
	logLevel := &slog.LevelVar{} // INFO
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})

	if opts.Debug {
		logLevel.Set(slog.LevelDebug)
	}

	l := slog.New(handler)

	// initial bitbucket client
	b, err := NewBitbucket(ctx, l)
	if err != nil {
		return nil, err
	}

	g, err := NewGitea(ctx, l)
	if err != nil {
		return nil, err
	}

	m := &migration{
		ctx:       ctx,
		Bitbucket: b,
		Gitea:     g,
		Logger:    l,
	}

	return m, nil
}

// CreateNewOrgOption create new organization option
type CreateNewOrgOption struct {
	Name        string
	Description string
	Public      bool
}

// CreateNewOrg create new organization
func (m *migration) CreateNewOrg(opts CreateNewOrgOption) error {
	m.Logger.Info("start create organization", "name", opts.Name)
	_, err := m.Gitea.CreateAndGetOrg(CreateOrgOption{
		Name:        opts.Name,
		Description: opts.Description,
		Visibility:  opts.Public,
	})
	if err != nil {
		return err
	}
	return nil
}

// MigrateNewRepoOption migrate repository option
type MigrateNewRepoOption struct {
	Owner       string
	Name        string
	CloneAddr   string
	Description string
	Private     bool
}

// MigrateNewRepo migrate repository
func (m *migration) MigrateNewRepo(opts MigrateNewRepoOption) error {
	m.Logger.Info("start migrate repo",
		"owner", opts.Owner,
		"name", opts.Name,
	)
	_, err := m.Gitea.MigrateRepo(MigrateRepoOption{
		RepoName:     opts.Name,
		RepoOwner:    opts.Owner,
		CloneAddr:    opts.CloneAddr,
		Private:      opts.Private,
		Description:  opts.Description,
		AuthUsername: m.Bitbucket.Username,
		AuthPassword: m.Bitbucket.Token,
	})
	if err != nil {
		return err
	}
	return nil
}

// ProjectResponse project response
type ProjectResponse struct {
	Project    bitbucketv1.Project
	Permission map[string][]string
}

// GetProjectData get project data
func (m *migration) GetProjectData(projectKey string) (*ProjectResponse, error) {
	org, err := m.Bitbucket.GetProject(projectKey)
	if err != nil {
		return nil, err
	}

	return &ProjectResponse{
		Project:    org,
	}, nil
}

// RepositoryResponse repository response
type RepositoryResponse struct {
	Repository bitbucketv1.Repository
}

// GetRepositoryData get repository data
func (m *migration) GetRepositoryData(projectKey, repoSlug string) (*RepositoryResponse, error) {
	repo, err := m.Bitbucket.GetRepo(projectKey, repoSlug)
	if err != nil {
		return nil, err
	}

	return &RepositoryResponse{
		Repository: repo,
	}, nil
}
