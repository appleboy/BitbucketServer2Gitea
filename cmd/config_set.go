package cmd

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	configCmd.AddCommand(configSetCmd)
	configSetCmd.Flags().StringP("bitbucket-token", "", "", "access token for Bitbucket API access")
	configSetCmd.Flags().StringP("bitbucket-server", "", "", "Bitbucket server URL with a trailing slash (https://stash.example.com/rest/)")
	configSetCmd.Flags().StringP("bitbucket-username", "", "", "username for Bitbucket API access")
	configSetCmd.Flags().StringP("gitea-token", "", "", "token for Gitea API access")
	configSetCmd.Flags().StringP("gitea-server", "", "", "Gitea server URL (https://gitea.example.com/)")
	configSetCmd.Flags().BoolP("gitea-skip-verify", "", false, "Skip SSL verification for Gitea server")
	_ = viper.BindPFlag("bitbucket.token", configSetCmd.Flags().Lookup("bitbucket-token"))
	_ = viper.BindPFlag("bitbucket.server", configSetCmd.Flags().Lookup("bitbucket-server"))
	_ = viper.BindPFlag("bitbucket.username", configSetCmd.Flags().Lookup("bitbucket-username"))
	_ = viper.BindPFlag("gitea.token", configSetCmd.Flags().Lookup("gitea-token"))
	_ = viper.BindPFlag("gitea.server", configSetCmd.Flags().Lookup("gitea-server"))
	_ = viper.BindPFlag("gitea.skip-verify", configSetCmd.Flags().Lookup("gitea-skip-verify"))
}

// configSetCmd updates the config value.
// It takes at least two arguments, the first one being the key and the second one being the value.
// If the key is not available, it returns an error message.
// If the key is "git.exclude_list", it sets the value as a slice of strings.
// It writes the config to file and prints a success message with the config file location.
var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "update the config value",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.Set(args[0], args[1])

		// Write config to file
		if err := viper.WriteConfig(); err != nil {
			return err
		}

		// Print success message with config file location
		color.Green("you can see the config file: %s", viper.ConfigFileUsed())
		return nil
	},
}
