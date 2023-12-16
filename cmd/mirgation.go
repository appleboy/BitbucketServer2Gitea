package cmd

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"

	"code.gitea.io/sdk/gitea"
	"github.com/spf13/viper"
)

type migration struct {
	ctx context.Context
	// bitbucketServer   string
	// bitbucketToken    string
	// bitbucketUsername string
	// bitbucketClient   *bitbucketv1.APIClient
	bitbucket       *bitbucket
	giteaServer     string
	giteaToken      string
	giteaSkipVerify bool
	giteaClient     *gitea.Client
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
	// initial bitbucket client
	b, err := NewBitbucket(ctx)
	if err != nil {
		return nil, err
	}

	m := &migration{
		ctx:             ctx,
		bitbucket:       b,
		giteaServer:     viper.GetString("gitea.server"),
		giteaSkipVerify: viper.GetBool("gitea.skip-verify"),
		giteaToken:      viper.GetString("gitea.token"),
	}

	if err := m.initGitea(); err != nil {
		return nil, err
	}

	return m, nil
}
