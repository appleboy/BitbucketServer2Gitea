package cmd

import (
	"context"
	"crypto/tls"
	"errors"
	"log"
	"net/http"
	"time"

	"code.gitea.io/sdk/gitea"
	bitbucketv1 "github.com/gfleury/go-bitbucket-v1"
	"github.com/spf13/viper"
)

type migration struct {
	bitbucketServer string
	bitbucketToken  string
	bitbucketClient *bitbucketv1.APIClient
	giteaServer     string
	giteaToken      string
	giteaSkipVerify bool
	giteaClient     *gitea.Client
}

func (m *migration) initBitbucket() error {
	if m.bitbucketServer == "" || m.bitbucketToken == "" {
		return errors.New("mission bitbucket server or token")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Millisecond)
	ctx = context.WithValue(ctx, bitbucketv1.ContextAccessToken, m.bitbucketToken)
	defer cancel()
	if m.bitbucketClient == nil {
		m.bitbucketClient = bitbucketv1.NewAPIClient(
			ctx,
			bitbucketv1.NewConfiguration(m.bitbucketServer),
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
		if client, err := gitea.NewClient(
			m.giteaServer,
			opts...,
		); err != nil {
			log.Fatal(err)
		} else {
			m.giteaClient = client
		}
	}
	return nil
}

// NewMigration creates a new instance of the migration struct.
// It initializes the bitbucketServer, bitbucketToken, giteaServer, and giteaToken fields
// with values from the viper configuration.
// It also initializes the bitbucket and gitea clients.
// Returns a pointer to the migration struct and any error encountered during initialization.
func NewMigration() (*migration, error) {
	m := &migration{
		bitbucketServer: viper.GetString("bitbucket.server"),
		bitbucketToken:  viper.GetString("bitbucket.token"),
		giteaServer:     viper.GetString("gitea.server"),
		giteaSkipVerify: viper.GetBool("gitea.skip-verify"),
		giteaToken:      viper.GetString("gitea.token"),
	}

	if err := m.initBitbucket(); err != nil {
		return nil, err
	}
	if err := m.initGitea(); err != nil {
		return nil, err
	}

	return m, nil
}
