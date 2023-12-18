package cmd

import (
	"context"
	"errors"
	"time"

	"github.com/appleboy/BitbucketServer2Gitea/migration"

	"github.com/spf13/cobra"
)

var (
	projectKey  string
	repoSlug    string
	targetOwner string
	targetRepo  string
)

func init() {
	migrateCmd.PersistentFlags().StringVar(&projectKey, "project-key", "", "the parent project key")
	migrateCmd.PersistentFlags().StringVar(&repoSlug, "repo-slug", "", "the repository slug")
	migrateCmd.PersistentFlags().StringVar(&targetOwner, "target-owner", "", "gitea target owner")
	migrateCmd.PersistentFlags().StringVar(&targetRepo, "target-repo", "", "gitea target repo")
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "migrate organization repository",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		m, err := migration.NewMigration(
			ctx,
			migration.Option{
				Debug: debug,
			})
		if err != nil {
			return err
		}

		if projectKey == "" || repoSlug == "" {
			return errors.New("project-key or repo-slug is empty")
		}

		orgResp, err := m.GetProjectData(projectKey)
		if err != nil {
			return err
		}

		repoResp, err := m.GetRepositoryData(projectKey, repoSlug)
		if err != nil {
			return err
		}

		// check gitea owner exist
		if targetOwner == "" {
			targetOwner = orgResp.Project.Name
		}

		// check gitea repository exist
		if targetRepo == "" {
			targetRepo = repoResp.Repository.Name
		}

		// create new gitea organization
		err = m.CreateNewOrg(migration.CreateNewOrgOption{
			Name:        targetOwner,
			Description: orgResp.Project.Description,
			Public:      orgResp.Project.Public,
			Permission:  orgResp.Permission,
		})
		if err != nil {
			return err
		}

		// create new gitea repository
		err = m.MigrateNewRepo(migration.MigrateNewRepoOption{
			Owner:       targetOwner,
			Name:        targetRepo,
			CloneAddr:   repoResp.Repository.Links.Clone[1].Href,
			Description: repoResp.Repository.Description,
			Private:     !repoResp.Repository.Public,
			Permission:  repoResp.Permission,
		})
		if err != nil {
			return err
		}

		return nil
	},
}
