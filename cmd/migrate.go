package cmd

import (
	"context"
	"errors"
	"time"

	"github.com/appleboy/BitbucketServer2Gitea/migration"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	migrateCmd.Flags().StringP("timeout", "t", "10m", "timeout for migration")
	_ = viper.BindPFlag("timeout", migrateCmd.Flags().Lookup("timeout"))
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "migrate organization repository",
	RunE: func(cmd *cobra.Command, args []string) error {
		// check timeout format
		timeout, err := time.ParseDuration(viper.GetString("timeout"))
		if err != nil {
			return err
		}

		// command timeout
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		m, err := migration.NewMigration(
			ctx,
			migration.Option{
				Debug: debug,
			})
		if err != nil {
			return err
		}

		repoList := []string{}

		if projectKey == "" {
			return errors.New("project-key can't be empty")
		}

		orgResp, err := m.GetProjectData(projectKey)
		if err != nil {
			return err
		}

		if repoSlug != "" {
			repoList = append(repoList, repoSlug)
		} else {
			// get all repository list
			repos, err := m.Bitbucket.GetRepositories(projectKey)
			if err != nil {
				return err
			}

			for _, repo := range repos {
				repoList = append(repoList, repo.Slug)
			}
		}

		// check gitea owner exist
		if targetOwner == "" {
			targetOwner = orgResp.Project.Name
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

		for _, repoSlug := range repoList {
			repoResp, err := m.GetRepositoryData(projectKey, repoSlug)
			if err != nil {
				return err
			}

			repoName := repoResp.Repository.Name
			// check gitea repository exist
			if targetRepo != "" && len(repoList) == 1 {
				repoName = targetRepo
			}

			cloneAddr := ""
			for _, link := range repoResp.Repository.Links.Clone {
				if link.Name == "http" {
					cloneAddr = link.Href
					break
				}
			}

			// create new gitea repository
			err = m.MigrateNewRepo(migration.MigrateNewRepoOption{
				Owner:       targetOwner,
				Name:        repoName,
				CloneAddr:   cloneAddr,
				Description: repoResp.Repository.Description,
				Private:     !repoResp.Repository.Public,
				Permission:  repoResp.Permission,
			})
			if err != nil {
				m.Logger.Error("migration repository error", "error", err)
			}
		}

		return nil
	},
}
