package agent

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/deepwrite/serivces/gateway/pkg/database"
	"gorm.io/gorm"
)

type MemoryRetriever struct {
	db *gorm.DB
}

func NewMemoryRetriever() *MemoryRetriever {
	return &MemoryRetriever{db: database.DB}
}

func NewMemoryRetrieverWithDB(db *gorm.DB) *MemoryRetriever {
	return &MemoryRetriever{db: db}
}

type RetrievedMemory struct {
	Memory    Memory
	Relevance float64
}

func (r *MemoryRetriever) RetrieveRelevant(
	ctx context.Context,
	agentID, userID string,
	query string,
	limit int,
	minRelevance float64,
) ([]RetrievedMemory, error) {
	var memories []Memory

	err := r.db.WithContext(ctx).
		Where("agent_id = ? AND user_id = ?", agentID, userID).
		Where("(expires_at IS NULL OR expires_at > ?)", time.Now()).
		Order("importance DESC, access_count DESC, created_at DESC").
		Limit(limit * 3).
		Find(&memories).Error

	if err != nil {
		return nil, err
	}

	results := make([]RetrievedMemory, 0, len(memories))
	queryLower := strings.ToLower(query)

	for _, mem := range memories {
		relevance := calculateRelevance(mem, queryLower)
		if relevance >= minRelevance {
			results = append(results, RetrievedMemory{
				Memory:    mem,
				Relevance: relevance,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Relevance > results[j].Relevance
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

func (r *MemoryRetriever) RetrieveByRecency(
	ctx context.Context,
	agentID, userID string,
	limit int,
) ([]Memory, error) {
	var memories []Memory

	err := r.db.WithContext(ctx).
		Where("agent_id = ? AND user_id = ?", agentID, userID).
		Where("(expires_at IS NULL OR expires_at > ?)", time.Now()).
		Order("created_at DESC").
		Limit(limit).
		Find(&memories).Error

	return memories, err
}

func (r *MemoryRetriever) RetrieveByImportance(
	ctx context.Context,
	agentID, userID string,
	limit int,
) ([]Memory, error) {
	var memories []Memory

	err := r.db.WithContext(ctx).
		Where("agent_id = ? AND user_id = ?", agentID, userID).
		Where("(expires_at IS NULL OR expires_at > ?)", time.Now()).
		Order("importance DESC, created_at DESC").
		Limit(limit).
		Find(&memories).Error

	return memories, err
}

func (r *MemoryRetriever) RetrieveBySession(
	ctx context.Context,
	sessionID string,
	limit int,
) ([]Memory, error) {
	var memories []Memory

	err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at DESC").
		Limit(limit).
		Find(&memories).Error

	return memories, err
}

func (r *MemoryRetriever) RetrieveByWorkspace(
	ctx context.Context,
	workspaceID string,
	limit int,
) ([]Memory, error) {
	var memories []Memory

	err := r.db.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Where("(expires_at IS NULL OR expires_at > ?)", time.Now()).
		Order("importance DESC, created_at DESC").
		Limit(limit).
		Find(&memories).Error

	return memories, err
}

func calculateRelevance(mem Memory, queryLower string) float64 {
	score := 0.0

	contentLower := strings.ToLower(mem.Content)
	summaryLower := strings.ToLower(mem.Summary)

	if strings.Contains(contentLower, queryLower) {
		score += 0.5
	}
	if strings.Contains(summaryLower, queryLower) {
		score += 0.3
	}

	if len(mem.Keywords) > 0 {
		var keywords []string
		if err := json.Unmarshal(mem.Keywords, &keywords); err == nil {
			for _, kw := range keywords {
				if strings.Contains(strings.ToLower(kw), queryLower) {
					score += 0.2
					break
				}
			}
		}
	}

	score += mem.Importance * 0.2

	if mem.AccessCount > 0 {
		score += 0.1
	}

	if score > 1.0 {
		score = 1.0
	}

	return score
}
