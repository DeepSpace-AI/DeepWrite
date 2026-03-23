package workspace

import (
	"context"
	"strings"
	"time"

	"github.com/deepwrite/serivces/gateway/models/document"
	"github.com/deepwrite/serivces/gateway/pkg/database"
	"gorm.io/gorm"
)

const (
	dashboardWorkspaceLimit  = 5
	dashboardDocumentLimit   = 6
	dashboardInvitationLimit = 5
)

type DashboardSummary struct {
	WorkspaceCount         int64 `json:"workspace_count"`
	ActiveWorkspaceCount   int64 `json:"active_workspace_count"`
	DocumentCount          int64 `json:"document_count"`
	FileCount              int64 `json:"file_count"`
	PendingInvitationCount int64 `json:"pending_invitation_count"`
	CollaboratorCount      int64 `json:"collaborator_count"`
}

type DashboardWorkspaceItem struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Status        string    `json:"status"`
	Role          string    `json:"role"`
	MemberCount   int64     `json:"member_count"`
	DocumentCount int64     `json:"document_count"`
	FileCount     int64     `json:"file_count"`
	UpdatedAt     time.Time `json:"updated_at"`
	CreatedAt     time.Time `json:"created_at"`
	Public        bool      `json:"public"`
	OwnerID       string    `json:"owner_id"`
}

type DashboardDocumentItem struct {
	ID             string    `json:"id"`
	WorkspaceID    string    `json:"workspace_id"`
	WorkspaceName  string    `json:"workspace_name"`
	Title          string    `json:"title"`
	CurrentVersion int64     `json:"current_version"`
	UpdatedAt      time.Time `json:"updated_at"`
	CreatedAt      time.Time `json:"created_at"`
}

type DashboardInvitationItem struct {
	ID            string    `json:"id"`
	WorkspaceID   string    `json:"workspace_id"`
	WorkspaceName string    `json:"workspace_name"`
	InviteeEmail  string    `json:"invitee_email"`
	Role          string    `json:"role"`
	Status        string    `json:"status"`
	ExpiresAt     time.Time `json:"expires_at"`
	CreatedAt     time.Time `json:"created_at"`
}

type DashboardOverview struct {
	Summary            DashboardSummary          `json:"summary"`
	Workspaces         []DashboardWorkspaceItem  `json:"workspaces"`
	RecentDocuments    []DashboardDocumentItem   `json:"recent_documents"`
	PendingInvitations []DashboardInvitationItem `json:"pending_invitations"`
}

type groupedCountRow struct {
	WorkspaceID string `gorm:"column:workspace_id"`
	Count       int64  `gorm:"column:count"`
}

func GetDashboardOverview(ctx context.Context, userID, userEmail string) (DashboardOverview, error) {
	overview := DashboardOverview{
		Summary:            DashboardSummary{},
		Workspaces:         make([]DashboardWorkspaceItem, 0),
		RecentDocuments:    make([]DashboardDocumentItem, 0),
		PendingInvitations: make([]DashboardInvitationItem, 0),
	}

	workspaces, err := ListByUser(ctx, userID, 0, 0)
	if err != nil {
		return DashboardOverview{}, err
	}

	overview.Summary.WorkspaceCount = int64(len(workspaces))
	workspaceIDs := make([]string, 0, len(workspaces))
	for _, ws := range workspaces {
		workspaceIDs = append(workspaceIDs, ws.ID)
		if strings.TrimSpace(ws.Status) == "archived" {
			continue
		}
		overview.Summary.ActiveWorkspaceCount++
	}

	documentCounts := map[string]int64{}
	fileCounts := map[string]int64{}
	memberCounts := map[string]int64{}

	if len(workspaceIDs) > 0 {
		var docRows []groupedCountRow
		if err := database.DB.WithContext(ctx).
			Model(&document.Document{}).
			Select("workspace_id, COUNT(*) as count").
			Where("workspace_id IN ?", workspaceIDs).
			Group("workspace_id").
			Scan(&docRows).Error; err != nil {
			return DashboardOverview{}, err
		}
		for _, row := range docRows {
			documentCounts[row.WorkspaceID] = row.Count
			overview.Summary.DocumentCount += row.Count
		}

		var fileRows []groupedCountRow
		if err := database.DB.WithContext(ctx).
			Model(&WorkspaceFile{}).
			Select("workspace_id, COUNT(*) as count").
			Where("workspace_id IN ?", workspaceIDs).
			Group("workspace_id").
			Scan(&fileRows).Error; err != nil {
			return DashboardOverview{}, err
		}
		for _, row := range fileRows {
			fileCounts[row.WorkspaceID] = row.Count
			overview.Summary.FileCount += row.Count
		}

		var memberRows []groupedCountRow
		if err := database.DB.WithContext(ctx).
			Model(&Members{}).
			Select("workspace_id, COUNT(DISTINCT user_id) as count").
			Where("workspace_id IN ?", workspaceIDs).
			Group("workspace_id").
			Scan(&memberRows).Error; err != nil {
			return DashboardOverview{}, err
		}
		for _, row := range memberRows {
			memberCounts[row.WorkspaceID] = row.Count
		}

		if err := database.DB.WithContext(ctx).
			Model(&Members{}).
			Where("workspace_id IN ? AND user_id <> ?", workspaceIDs, strings.TrimSpace(userID)).
			Distinct("user_id").
			Count(&overview.Summary.CollaboratorCount).Error; err != nil {
			return DashboardOverview{}, err
		}

		if err := database.DB.WithContext(ctx).
			Table("documents").
			Select("documents.id, documents.workspace_id, workspaces.name as workspace_name, documents.title, documents.current_version, documents.updated_at, documents.created_at").
			Joins("left join workspaces on workspaces.id = documents.workspace_id").
			Where("documents.workspace_id IN ?", workspaceIDs).
			Order("documents.updated_at desc").
			Limit(dashboardDocumentLimit).
			Scan(&overview.RecentDocuments).Error; err != nil {
			return DashboardOverview{}, err
		}
	}

	for index, ws := range workspaces {
		if index >= dashboardWorkspaceLimit {
			break
		}

		role := ws.GetUserRole(userID)
		if role == "" && ws.OwnerID == strings.TrimSpace(userID) {
			role = RoleOwner
		}
		if role == "" {
			role = RoleViewer
		}

		overview.Workspaces = append(overview.Workspaces, DashboardWorkspaceItem{
			ID:            ws.ID,
			Name:          ws.Name,
			Description:   ws.Description,
			Status:        ws.Status,
			Role:          role,
			MemberCount:   memberCounts[ws.ID],
			DocumentCount: documentCounts[ws.ID],
			FileCount:     fileCounts[ws.ID],
			UpdatedAt:     ws.UpdatedAt,
			CreatedAt:     ws.CreatedAt,
			Public:        ws.Public,
			OwnerID:       ws.OwnerID,
		})
	}

	normalizedEmail := strings.ToLower(strings.TrimSpace(userEmail))
	pendingInvitationBaseQuery := database.DB.WithContext(ctx).
		Table("workspace_invitations").
		Joins("left join workspaces on workspaces.id = workspace_invitations.workspace_id").
		Where("workspace_invitations.status = ?", InvitationStatusPending)

	if normalizedEmail != "" {
		pendingInvitationBaseQuery = pendingInvitationBaseQuery.Where("workspace_invitations.invitee_user_id = ? OR workspace_invitations.invitee_email = ?", strings.TrimSpace(userID), normalizedEmail)
	} else {
		pendingInvitationBaseQuery = pendingInvitationBaseQuery.Where("workspace_invitations.invitee_user_id = ?", strings.TrimSpace(userID))
	}

	if err := pendingInvitationBaseQuery.Count(&overview.Summary.PendingInvitationCount).Error; err != nil {
		return DashboardOverview{}, err
	}

	if err := pendingInvitationBaseQuery.Session(&gorm.Session{}).
		Select("workspace_invitations.id, workspace_invitations.workspace_id, workspaces.name as workspace_name, workspace_invitations.invitee_email, workspace_invitations.role, workspace_invitations.status, workspace_invitations.expires_at, workspace_invitations.created_at").
		Order("workspace_invitations.created_at desc").
		Limit(dashboardInvitationLimit).
		Scan(&overview.PendingInvitations).Error; err != nil {
		return DashboardOverview{}, err
	}

	return overview, nil
}
