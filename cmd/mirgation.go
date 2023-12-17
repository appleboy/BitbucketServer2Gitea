package cmd

import (
	"context"
	"log/slog"
	"os"
)

type migration struct {
	ctx       context.Context
	bitbucket *bitbucket
	gitea     *gitea
	logger    *slog.Logger
}

func NewMigration(ctx context.Context) (*migration, error) {
	logLevel := &slog.LevelVar{} // INFO
	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)

	if debug {
		logLevel.Set(slog.LevelDebug)
	}

	// initial bitbucket client
	b, err := NewBitbucket(ctx)
	if err != nil {
		return nil, err
	}

	g, err := NewGitea(ctx)
	if err != nil {
		return nil, err
	}

	m := &migration{
		ctx:       ctx,
		bitbucket: b,
		gitea:     g,
		logger:    slog.New(handler),
	}

	return m, nil
}

// CreateNewOrgOption create new organization option
type CreateNewOrgOption struct {
	Name        string
	Description string
	Public      bool
	Permission  map[string][]string
}

// CreateNewOrg create new organization
func (m *migration) CreateNewOrg(opts CreateNewOrgOption) error {
	m.logger.Info("start create organization", "name", opts.Name)
	_, err := m.gitea.CreateAndGetOrg(CreateOrgOption{
		Name:        targetOwner,
		Description: opts.Description,
		Visibility:  opts.Public,
	})
	if err != nil {
		return err
	}

	m.logger.Info("start migrate organization permission", "name", opts.Name)
	for permission, users := range opts.Permission {
		team, err := m.gitea.CreateOrGetTeam(targetOwner, permission)
		if err != nil {
			return err
		}
		for _, user := range users {
			err := m.gitea.AddTeamMember(team.ID, user)
			if err != nil {
				return err
			}
		}
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
	Permission  map[string][]string
}

// MigrateNewRepo migrate repository
func (m *migration) MigrateNewRepo(opts MigrateNewRepoOption) error {
	m.logger.Info("start migrate repo",
		"owner", opts.Owner,
		"name", opts.Name,
	)
	_, err := m.gitea.MigrateRepo(MigrateRepoOption{
		RepoName:     opts.Name,
		RepoOwner:    opts.Owner,
		CloneAddr:    opts.CloneAddr,
		Private:      opts.Private,
		Description:  opts.Description,
		AuthUsername: m.bitbucket.Username,
		AuthPassword: m.bitbucket.Token,
	})
	if err != nil {
		return err
	}

	m.logger.Info("start migrate repo permission",
		"owner", opts.Owner,
		"name", opts.Name,
	)
	for permission, users := range opts.Permission {
		for _, user := range users {
			_, err := m.gitea.AddCollaborator(opts.Owner, opts.Name, user, permission)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
