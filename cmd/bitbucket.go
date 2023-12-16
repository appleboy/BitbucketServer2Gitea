package cmd

import (
	bitbucketv1 "github.com/gfleury/go-bitbucket-v1"
)

// GetUsersPermissionFromProject get users permission from project
func (m *migration) GetUsersPermissionFromProject(projectKey string) ([]bitbucketv1.UserPermission, error) {
	// check project user permission
	response, err := m.bitbucketClient.DefaultApi.GetUsersWithAnyPermission_23(
		projectKey,
		map[string]interface{}{
			"limit": 200,
		})
	if err != nil {
		return nil, err
	}

	return bitbucketv1.GetUsersPermissionResponse(response)
}

// GetUsersPermissionFromRepo get users permission from repo
func (m *migration) GetUsersPermissionFromRepo(projectKey, repoSlug string) ([]bitbucketv1.UserPermission, error) {
	// check project user permission
	response, err := m.bitbucketClient.DefaultApi.GetUsersWithAnyPermission_24(
		projectKey,
		repoSlug,
		map[string]interface{}{
			"limit": 200,
		})
	if err != nil {
		return nil, err
	}

	return bitbucketv1.GetUsersPermissionResponse(response)
}

// GetGroupsPermissionFromProject get groups permission from project
func (m *migration) GetGroupsPermissionFromProject(projectKey string) ([]bitbucketv1.GroupPermission, error) {
	// check project group permission
	response, err := m.bitbucketClient.DefaultApi.GetGroupsWithAnyPermission_12(
		projectKey,
		map[string]interface{}{
			"limit": 200,
		})
	if err != nil {
		return nil, err
	}

	return bitbucketv1.GetGroupsPermissionResponse(response)
}

// GetGroupsPermissionFromRepo get groups permission from repo
func (m *migration) GetGroupsPermissionFromRepo(projectKey, repoSlug string) ([]bitbucketv1.GroupPermission, error) {
	// check project group permission
	response, err := m.bitbucketClient.DefaultApi.GetGroupsWithAnyPermission_13(
		projectKey,
		repoSlug,
		map[string]interface{}{
			"limit": 200,
		})
	if err != nil {
		return nil, err
	}

	return bitbucketv1.GetGroupsPermissionResponse(response)
}

// GetUsersFromGroup get users from group
func (m *migration) GetUsersFromGroup(g string) ([]bitbucketv1.User, error) {
	response, err := m.bitbucketClient.DefaultApi.FindUsersInGroup(
		map[string]interface{}{
			"context": g,
			"limit":   200,
		},
	)
	if err != nil {
		return nil, err
	}

	return bitbucketv1.GetUsersResponse(response)
}
