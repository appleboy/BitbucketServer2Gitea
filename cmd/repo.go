package cmd

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"code.gitea.io/sdk/gitea"
	bitbucketv1 "github.com/gfleury/go-bitbucket-v1"
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
		response, err := m.bitbucketClient.DefaultApi.GetProject(projectKey)
		if err != nil {
			return err
		}

		org, err := bitbucketv1.GetRepositoryResponse(response)
		if err != nil {
			return err
		}
		slog.Info("check project success", "name", org.Name)

		// check project user permission
		response, err = m.bitbucketClient.DefaultApi.GetUsersWithAnyPermission_23(projectKey, map[string]interface{}{})
		if err != nil {
			return err
		}

		users, err := bitbucketv1.GetUsersPermissionResponse(response)
		if err != nil {
			return err
		}

		for _, user := range users {
			slog.Info("project permission",
				"display", user.User.DisplayName,
				"account", user.User.Name,
				"permission", user.Permission,
			)
		}

		// check project group permission
		response, err = m.bitbucketClient.DefaultApi.GetGroupsWithAnyPermission_12(projectKey, map[string]interface{}{})
		if err != nil {
			return err
		}

		groups, err := bitbucketv1.GetGroupsPermissionResponse(response)
		if err != nil {
			return err
		}

		for _, group := range groups {
			slog.Info("group permission",
				"name", group.Group.Name,
				"permission", group.Permission,
			)

			response, err = m.bitbucketClient.DefaultApi.FindUsersInGroup(
				map[string]interface{}{
					"context": group.Group.Name,
				},
			)
			if err != nil {
				return err
			}

			users, err := bitbucketv1.GetUsersResponse(response)
			if err != nil {
				return err
			}
			for _, user := range users {
				slog.Info("user permission in group",
					"display", user.DisplayName,
					"account", user.Name,
					"permission", group.Permission,
				)
			}
		}

		response, err = m.bitbucketClient.DefaultApi.GetRepository(projectKey, repoSlug)
		if err != nil {
			return err
		}

		repo, err := bitbucketv1.GetRepositoryResponse(response)
		if err != nil {
			return err
		}
		slog.Info("check repo success", "name", repo.Name)

		// check repo user permission
		response, err = m.bitbucketClient.DefaultApi.GetUsersWithAnyPermission_24(projectKey, repoSlug, map[string]interface{}{})
		if err != nil {
			return err
		}

		users, err = bitbucketv1.GetUsersPermissionResponse(response)
		if err != nil {
			return err
		}

		for _, user := range users {
			slog.Info("repo permission",
				"display", user.User.DisplayName,
				"account", user.User.Name,
				"permission", user.Permission,
			)
		}

		// check gitea owner exist
		if targetOwner == "" {
			targetOwner = org.Name
		}
		newOrg, reponse, err := m.giteaClient.GetOrg(targetOwner)
		if reponse.StatusCode == http.StatusNotFound {
			visible := gitea.VisibleTypePublic
			if !org.Public {
				visible = gitea.VisibleTypePrivate
			}
			newOrg, _, err = m.giteaClient.CreateOrg(gitea.CreateOrgOption{
				Name:        targetOwner,
				Description: org.Description,
				Visibility:  visible,
			})
			if err != nil {
				return err
			}
			slog.Info("create new org success", "name", newOrg.UserName)
		} else if err != nil {
			return err
		}

		if targetRepo == "" {
			targetRepo = repo.Name
		}

		slog.Info("start migrate repo", "name", targetRepo, "owner", targetOwner)
		newRepo, _, err := m.giteaClient.MigrateRepo(gitea.MigrateRepoOption{
			RepoName:     targetRepo,
			RepoOwner:    targetOwner,
			CloneAddr:    repo.Links.Clone[1].Href,
			Private:      !repo.Public,
			Description:  repo.Description,
			AuthUsername: m.bitbucketUsername,
			AuthPassword: m.bitbucketToken,
		})
		if err != nil {
			return err
		}
		slog.Info("migrate repo success", "name", newRepo.Name, "owner", newOrg.UserName)

		return nil
	},
}
