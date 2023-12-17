package migration

import (
	"context"
	"log/slog"
	"os"
)

type migration struct {
	ctx       context.Context
	Bitbucket *bitbucket
	Gitea     *gitea
	Logger    *slog.Logger
}

func NewMigration(ctx context.Context) (*migration, error) {
	logLevel := &slog.LevelVar{} // INFO
	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)

	// if debug {
	// 	logLevel.Set(slog.LevelDebug)
	// }

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
		Bitbucket: b,
		Gitea:     g,
		Logger:    slog.New(handler),
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
	m.Logger.Info("start create organization", "name", opts.Name)
	_, err := m.Gitea.CreateAndGetOrg(CreateOrgOption{
		Name:        opts.Name,
		Description: opts.Description,
		Visibility:  opts.Public,
	})
	if err != nil {
		return err
	}

	m.Logger.Info("start migrate organization permission", "name", opts.Name)
	for permission, users := range opts.Permission {
		team, err := m.Gitea.CreateOrGetTeam(opts.Name, permission)
		if err != nil {
			return err
		}
		for _, user := range users {
			err := m.Gitea.AddTeamMember(team.ID, user)
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

	m.Logger.Info("start migrate repo permission",
		"owner", opts.Owner,
		"name", opts.Name,
	)
	for permission, users := range opts.Permission {
		for _, user := range users {
			_, err := m.Gitea.AddCollaborator(opts.Owner, opts.Name, user, permission)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
