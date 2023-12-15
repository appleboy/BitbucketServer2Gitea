package cmd

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"strings"

	"code.gitea.io/sdk/gitea"
	bitbucketv1 "github.com/gfleury/go-bitbucket-v1"
	"github.com/spf13/viper"
)

type migration struct {
	ctx               context.Context
	bitbucketServer   string
	bitbucketToken    string
	bitbucketUsername string
	bitbucketClient   *bitbucketv1.APIClient
	giteaServer       string
	giteaToken        string
	giteaSkipVerify   bool
	giteaClient       *gitea.Client
}

func (m *migration) initBitbucket() error {
	if m.bitbucketServer == "" || m.bitbucketToken == "" {
		return errors.New("mission bitbucket server or token")
	}

	m.bitbucketServer = strings.TrimRight(m.bitbucketServer, "/")

	ctx := context.WithValue(m.ctx, bitbucketv1.ContextAccessToken, m.bitbucketToken)
	if m.bitbucketClient == nil {
		m.bitbucketClient = bitbucketv1.NewAPIClient(
			ctx,
			bitbucketv1.NewConfiguration(m.bitbucketServer+"/rest"),
		)
	}
	return nil
}

func (m *migration) initGitea() error {
	if m.giteaServer == "" || m.giteaToken == "" {
		return errors.New("mission gitea server or token")
	}

	opts := []gitea.ClientOption{
		gitea.SetToken(m.giteaToken),
	}

	if m.giteaSkipVerify {
		// add new http client for skip verify
		httpClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
		opts = append(opts, gitea.SetHTTPClient(httpClient))
	}

	if m.giteaClient == nil {
		client, err := gitea.NewClient(m.giteaServer, opts...)
		if err != nil {
			return err
		}
		m.giteaClient = client
	}
	return nil
}

// NewMigration creates a new instance of the migration struct.
// It initializes the bitbucketServer, bitbucketToken, giteaServer, and giteaToken fields
// with values from the viper configuration.
// It also initializes the bitbucket and gitea clients.
// Returns a pointer to the migration struct and any error encountered during initialization.
func NewMigration(ctx context.Context) (*migration, error) {
	m := &migration{
		ctx:               ctx,
		bitbucketServer:   viper.GetString("bitbucket.server"),
		bitbucketUsername: viper.GetString("bitbucket.username"),
		bitbucketToken:    viper.GetString("bitbucket.token"),
		giteaServer:       viper.GetString("gitea.server"),
		giteaSkipVerify:   viper.GetBool("gitea.skip-verify"),
		giteaToken:        viper.GetString("gitea.token"),
	}

	if err := m.initBitbucket(); err != nil {
		return nil, err
	}
	if err := m.initGitea(); err != nil {
		return nil, err
	}

	return m, nil
}
