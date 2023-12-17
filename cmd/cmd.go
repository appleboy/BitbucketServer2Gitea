package cmd

import (
	"context"
	"log"
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/appleboy/com/file"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Short:         "A command line tool build with Golang to migrate a Bitbucket to Gitea.",
	SilenceUsage:  true,
	Args:          cobra.MaximumNArgs(1),
	SilenceErrors: true,
}

// Used for flags.
var (
	cfgFile  string
	debug    bool
	replacer = strings.NewReplacer("-", "_", ".", "_")
)

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/bitbucketServer2Gitea/.config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "debug mode")
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(repoCmd)

	// hide completion command
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
		if !file.IsFile(cfgFile) {
			// Config file not found; ignore error if desired
			_, err := os.Create(cfgFile)
			if err != nil {
				log.Fatal(err)
			}
		}
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".cobra" (without extension).
		configFolder := path.Join(home, ".config", "bitbucketServer2Gitea")
		viper.AddConfigPath(configFolder)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".config")
		cfgFile = path.Join(configFolder, ".config.yaml")

		if !file.IsDir(configFolder) {
			if err := os.MkdirAll(configFolder, os.ModePerm); err != nil {
				log.Fatal(err)
			}
		}
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(replacer)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			_, err := os.Create(cfgFile)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			// Config file was found but another error was produced
			slog.Error("read config error", "msg", err)
		}
	}
}

func Execute(ctx context.Context) error {
	if _, err := rootCmd.ExecuteContextC(ctx); err != nil {
		return err
	}

	return nil
}
