package cmd

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	projectKey  string
	repoSlug    string
	targetOwner string
	targetRepo  string
)

func init() {
	repoCmd.PersistentFlags().StringVar(&projectKey, "project-key", "", "the parent project key")
	repoCmd.PersistentFlags().StringVar(&repoSlug, "repo-slug", "", "the repository slug")
	repoCmd.PersistentFlags().StringVar(&targetOwner, "target-owner", "", "gitea target owner")
	repoCmd.PersistentFlags().StringVar(&targetRepo, "target-repo", "", "gitea target repo")
}

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "migration single repository",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		m, err := NewMigration(ctx)
		if err != nil {
			return err
		}

		if projectKey == "" || repoSlug == "" {
			return errors.New("project-key or repo-slug is empty")
		}

		// check bitbucket project exist
		org, err := m.bitbucket.GetProject(projectKey)
		if err != nil {
			return err
		}
		m.logger.Info("check project success", "name", org.Name)

		projectPermission := make(map[string][]string)
		// check project user permission
		users, err := m.bitbucket.GetUsersPermissionFromProject(projectKey)
		if err != nil {
			return err
		}
		for _, user := range users {
			m.logger.Debug("project permission",
				"display", user.User.DisplayName,
				"account", user.User.Name,
				"permission", user.Permission,
			)
			_, err := m.gitea.GreateOrGetUser(CreateUserOption{
				SourceID:  1,
				LoginName: strings.ToLower(user.User.Name),
				Username:  user.User.Name,
				FullName:  user.User.DisplayName,
				Email:     user.User.EmailAddress,
			})
			if err != nil {
				return err
			}
			projectPermission[user.Permission] = append(projectPermission[user.Permission], strings.ToLower(user.User.Name))
		}

		// check project group permission
		groups, err := m.bitbucket.GetGroupsPermissionFromProject(projectKey)
		if err != nil {
			return err
		}
		for _, group := range groups {
			m.logger.Debug("group permission for project",
				"name", group.Group.Name,
				"permission", group.Permission,
			)

			users, err := m.bitbucket.GetUsersFromGroup(group.Group.Name)
			if err != nil {
				return err
			}
			for _, user := range users {
				m.logger.Debug("user permission in group",
					"display", user.DisplayName,
					"account", user.Name,
					"permission", group.Permission,
					"group", group.Group.Name,
				)
				_, err := m.gitea.GreateOrGetUser(CreateUserOption{
					SourceID:  1,
					LoginName: strings.ToLower(user.Name),
					Username:  user.Name,
					FullName:  user.DisplayName,
					Email:     user.EmailAddress,
				})
				if err != nil {
					return err
				}
				projectPermission[group.Permission] = append(projectPermission[group.Permission], strings.ToLower(user.Name))
			}
		}

		repo, err := m.bitbucket.GetRepo(projectKey, repoSlug)
		if err != nil {
			return err
		}
		m.logger.Info("check repo success", "name", repo.Name)

		// check project group permission
		groups, err = m.bitbucket.GetGroupsPermissionFromRepo(projectKey, repoSlug)
		if err != nil {
			return err
		}

		repoPermission := make(map[string][]string)
		for _, group := range groups {
			m.logger.Debug("group permission for repo",
				"name", group.Group.Name,
				"permission", group.Permission,
			)

			users, err := m.bitbucket.GetUsersFromGroup(group.Group.Name)
			if err != nil {
				return err
			}
			for _, user := range users {
				m.logger.Debug("user permission in repo",
					"display", user.DisplayName,
					"account", user.Name,
					"permission", group.Permission,
					"group", group.Group.Name,
				)
				_, err := m.gitea.GreateOrGetUser(CreateUserOption{
					SourceID:  1,
					LoginName: strings.ToLower(user.Name),
					Username:  user.Name,
					FullName:  user.DisplayName,
					Email:     user.EmailAddress,
				})
				if err != nil {
					return err
				}
				repoPermission[group.Permission] = append(repoPermission[group.Permission], strings.ToLower(user.Name))
			}
		}

		// check repo user permission
		users, err = m.bitbucket.GetUsersPermissionFromRepo(projectKey, repoSlug)
		if err != nil {
			return err
		}
		for _, user := range users {
			m.logger.Debug("repo permission",
				"display", user.User.DisplayName,
				"account", user.User.Name,
				"permission", user.Permission,
			)
			_, err := m.gitea.GreateOrGetUser(CreateUserOption{
				SourceID:  1,
				LoginName: strings.ToLower(user.User.Name),
				Username:  user.User.Name,
				FullName:  user.User.DisplayName,
				Email:     user.User.EmailAddress,
			})
			if err != nil {
				return err
			}
			repoPermission[user.Permission] = append(repoPermission[user.Permission], strings.ToLower(user.User.Name))
		}

		// check gitea owner exist
		if targetOwner == "" {
			targetOwner = org.Name
		}

		// check gitea repository exist
		if targetRepo == "" {
			targetRepo = repo.Name
		}

		m.logger.Info("start create organization", "name", targetOwner)
		newOrg, err := m.gitea.CreateAndGetOrg(CreateOrgOption{
			Name:        targetOwner,
			Description: org.Description,
			Visibility:  org.Public,
		})
		if err != nil {
			return err
		}

		m.logger.Info("start migrate organization permission", "name", targetOwner)
		for permission, users := range projectPermission {
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

		m.logger.Info("start migrate repo", "name", targetRepo, "owner", targetOwner)
		newRepo, err := m.gitea.MigrateRepo(MigrateRepoOption{
			RepoName:     targetRepo,
			RepoOwner:    targetOwner,
			CloneAddr:    repo.Links.Clone[1].Href,
			Private:      !repo.Public,
			Description:  repo.Description,
			AuthUsername: m.bitbucket.Username,
			AuthPassword: m.bitbucket.Token,
		})
		if err != nil {
			return err
		}

		m.logger.Info("start migrate repo permission", "name", newRepo.Name, "owner", newOrg.UserName)
		for permission, users := range repoPermission {
			for _, user := range users {
				_, err := m.gitea.AddCollaborator(targetOwner, targetRepo, user, permission)
				if err != nil {
					return err
				}
			}
		}
		return nil
	},
}
