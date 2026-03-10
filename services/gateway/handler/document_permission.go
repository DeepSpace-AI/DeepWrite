package handler

import (
	"context"
	"strings"

	"github.com/deepwrite/serivces/gateway/models/workspace"
)

func canViewWorkspace(role string) bool {
	switch role {
	case workspace.RoleViewer, workspace.RoleEditor, workspace.RoleAdmin, workspace.RoleOwner:
		return true
	default:
		return false
	}
}

func canEditWorkspace(role string) bool {
	switch role {
	case workspace.RoleEditor, workspace.RoleAdmin, workspace.RoleOwner:
		return true
	default:
		return false
	}
}

func getWorkspaceRole(ctx context.Context, workspaceID, userID string) (workspace.Workspace, string, error) {
	ws, err := workspace.GetWorkSpaceByID(ctx, strings.TrimSpace(workspaceID))
	if err != nil {
		return workspace.Workspace{}, "", err
	}

	role := ws.GetUserRole(strings.TrimSpace(userID))
	if ws.Public && role == "" {
		role = workspace.RoleViewer
	}
	return ws, role, nil
}
