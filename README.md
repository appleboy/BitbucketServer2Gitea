# BitbucketServer2Gitea

A command line tool build with Golang to migrate a [Bitbucket Server](https://www.atlassian.com/software/bitbucket/enterprise) (Data Center) Project to Gitea. See the [V1 REST API](https://developer.atlassian.com/server/bitbucket/rest/v815/intro/#about).

## Requirements

The following software is required to run this tool:

* Gitea version: **1.21.3**
* Bitbucket Server version: **8.9.7**

## Initial Setup

Setup the Gitea and Bitbucket config

```bash
bitbucketServer2Gitea config set bitbucket.server https://stash.example.com
bitbucketServer2Gitea config set bitbucket.username admin
bitbucketServer2Gitea config set bitbucket.token xxxxxxxxxxxxxx
bitbucketServer2Gitea config set gitea.server https://gitea.example.com
bitbucketServer2Gitea config set gitea.token xxxxxxxxxxxxxx
```

## Migration Single Repository

```bash
bitbucketServer2Gitea migrate --project-key AIA --repo-slug test \
  --target-owner admin --target-repo test
```
