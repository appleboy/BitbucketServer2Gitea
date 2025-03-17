module github.com/appleboy/BitbucketServer2Gitea

go 1.22.0

toolchain go1.24.1

require (
	code.gitea.io/sdk/gitea v0.20.0
	github.com/appleboy/com v0.3.0
	github.com/fatih/color v1.18.0
	github.com/gfleury/go-bitbucket-v1 v0.0.0-20230830121038-6e30c5760c87
	github.com/spf13/cobra v1.8.1
	github.com/spf13/viper v1.20.0
)

require (
	github.com/42wim/httpsig v1.2.2 // indirect
	github.com/davidmz/go-pageant v1.0.2 // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/go-fed/httpsig v1.1.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/sagikazarmark/locafero v0.7.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.12.0 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.33.0 // indirect
	golang.org/x/net v0.35.0 // indirect
	golang.org/x/oauth2 v0.26.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/gfleury/go-bitbucket-v1 => github.com/appleboy/go-bitbucket-v1 v0.0.0-20231216080418-bafb48ca1464
