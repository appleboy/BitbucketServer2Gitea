package cmd

import (
	"github.com/spf13/cobra"
)

// configCmd represents the command for custom configuration,
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "custom config (Bitbucket and Gitea server URL and Token)",
}
