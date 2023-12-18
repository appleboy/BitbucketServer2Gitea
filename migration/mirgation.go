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

	permission := make(map[string][]string)
	// check project user permission
	users, err := m.Bitbucket.GetUsersPermissionFromProject(projectKey)
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		m.Logger.Debug("project permission",
			"display", user.User.DisplayName,
			"account", user.User.Name,
			"permission", user.Permission,
		)
		_, err := m.Gitea.GreateOrGetUser(CreateUserOption{
			SourceID:  m.Gitea.sourceID,
			LoginName: strings.ToLower(user.User.Name),
			Username:  user.User.Name,
			FullName:  user.User.DisplayName,
			Email:     user.User.EmailAddress,
		})
		if err != nil {
			return nil, err
		}
		permission[user.Permission] = append(permission[user.Permission], strings.ToLower(user.User.Name))
	}

	// check project group permission
	groups, err := m.Bitbucket.GetGroupsPermissionFromProject(projectKey)
	if err != nil {
		return nil, err
	}
	for _, group := range groups {
		m.Logger.Debug("group permission for project",
			"name", group.Group.Name,
			"permission", group.Permission,
		)

		users, err := m.Bitbucket.GetUsersFromGroup(group.Group.Name)
		if err != nil {
			return nil, err
		}
		for _, user := range users {
			m.Logger.Debug("user permission in group",
				"display", user.DisplayName,
				"account", user.Name,
				"permission", group.Permission,
				"group", group.Group.Name,
			)
			_, err := m.Gitea.GreateOrGetUser(CreateUserOption{
				// SourceID:  sourceID,
				LoginName: strings.ToLower(user.Name),
				Username:  user.Name,
				FullName:  user.DisplayName,
				Email:     user.EmailAddress,
			})
			if err != nil {
				return nil, err
			}
			permission[group.Permission] = append(permission[group.Permission], strings.ToLower(user.Name))
		}
	}

	return &ProjectResponse{
		Project:    org,
		Permission: permission,
	}, nil
}

// RepositoryResponse repository response
type RepositoryResponse struct {
	Repository bitbucketv1.Repository
	Permission map[string][]string
}

// GetRepositoryData get repository data
func (m *migration) GetRepositoryData(projectKey, repoSlug string) (*RepositoryResponse, error) {
	repo, err := m.Bitbucket.GetRepo(projectKey, repoSlug)
	if err != nil {
		return nil, err
	}

	// check project group permission
	groups, err := m.Bitbucket.GetGroupsPermissionFromRepo(projectKey, repoSlug)
	if err != nil {
		return nil, err
	}

	permission := make(map[string][]string)
	for _, group := range groups {
		m.Logger.Debug("group permission for repo",
			"name", group.Group.Name,
			"permission", group.Permission,
		)

		users, err := m.Bitbucket.GetUsersFromGroup(group.Group.Name)
		if err != nil {
			return nil, err
		}
		for _, user := range users {
			m.Logger.Debug("user permission in repo",
				"display", user.DisplayName,
				"account", user.Name,
				"permission", group.Permission,
				"group", group.Group.Name,
			)
			_, err := m.Gitea.GreateOrGetUser(CreateUserOption{
				SourceID:  m.Gitea.sourceID,
				LoginName: strings.ToLower(user.Name),
				Username:  user.Name,
				FullName:  user.DisplayName,
				Email:     user.EmailAddress,
			})
			if err != nil {
				return nil, err
			}
			permission[group.Permission] = append(permission[group.Permission], strings.ToLower(user.Name))
		}
	}

	// check repo user permission
	users, err := m.Bitbucket.GetUsersPermissionFromRepo(projectKey, repoSlug)
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		m.Logger.Debug("repo permission",
			"display", user.User.DisplayName,
			"account", user.User.Name,
			"permission", user.Permission,
		)
		_, err := m.Gitea.GreateOrGetUser(CreateUserOption{
			SourceID:  m.Gitea.sourceID,
			LoginName: strings.ToLower(user.User.Name),
			Username:  user.User.Name,
			FullName:  user.User.DisplayName,
			Email:     user.User.EmailAddress,
		})
		if err != nil {
			return nil, err
		}
		permission[user.Permission] = append(permission[user.Permission], strings.ToLower(user.User.Name))
	}

	return &RepositoryResponse{
		Repository: repo,
		Permission: permission,
	}, nil
}
