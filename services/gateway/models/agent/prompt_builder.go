package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	usermodel "github.com/deepwrite/serivces/gateway/models/user"
	workspacemodel "github.com/deepwrite/serivces/gateway/models/workspace"
)

type PromptLayer int

const (
	LayerIdentity PromptLayer = iota
	LayerCapability
	LayerUserContext
	LayerMemory
	LayerSession
	LayerInstruction
	LayerSafety
)

func (l PromptLayer) String() string {
	switch l {
	case LayerIdentity:
		return "Identity"
	case LayerCapability:
		return "Capability"
	case LayerUserContext:
		return "UserContext"
	case LayerMemory:
		return "Memory"
	case LayerSession:
		return "Session"
	case LayerInstruction:
		return "Instruction"
	case LayerSafety:
		return "Safety"
	default:
		return "Unknown"
	}
}

type PromptContext struct {
	UserDisplayName   string
	UserEmail         string
	UserLanguage      string
	UserTimezone      string
	WorkspaceName     string
	CurrentTime       string
	SessionTitle      string
	SessionSummary    string
	RetrievedMemories []RetrievedMemory
}

type AssembledPrompt struct {
	FullPrompt string
	Layers     map[PromptLayer]string
}

type PromptBuilder struct {
	agent   *Agent
	context *PromptContext
}

func NewPromptBuilder(agent *Agent, ctx *PromptContext) *PromptBuilder {
	return &PromptBuilder{
		agent:   agent,
		context: ctx,
	}
}

func (b *PromptBuilder) BuildIdentityLayer() string {
	if b.agent.IdentityPrompt != "" {
		return b.agent.IdentityPrompt
	}

	if b.agent.SystemPrompt != "" {
		return b.agent.SystemPrompt
	}

	return fmt.Sprintf("You are %s, %s", b.agent.Name, b.agent.Description)
}

func (b *PromptBuilder) BuildCapabilityLayer() string {
	if b.agent.CapabilityPrompt != "" {
		return b.agent.CapabilityPrompt
	}

	var parts []string

	var tools []AgentTool
	if len(b.agent.Tools) > 0 {
		_ = json.Unmarshal(b.agent.Tools, &tools)
	}

	if len(tools) > 0 {
		var toolDescs []string
		for _, tool := range tools {
			if tool.Enabled {
				toolDescs = append(toolDescs, fmt.Sprintf("- **%s**: %s", tool.Name, tool.Description))
			}
		}
		if len(toolDescs) > 0 {
			parts = append(parts, "# Available Tools\n\n"+strings.Join(toolDescs, "\n"))
		}
	}

	var skills []AgentSkill
	if len(b.agent.Skills) > 0 {
		_ = json.Unmarshal(b.agent.Skills, &skills)
	}

	if len(skills) > 0 {
		var skillDescs []string
		for _, skill := range skills {
			skillDescs = append(skillDescs, fmt.Sprintf("- **%s**: %s", skill.Name, skill.Description))
		}
		if len(skillDescs) > 0 {
			parts = append(parts, "# Skills\n\n"+strings.Join(skillDescs, "\n"))
		}
	}

	return strings.Join(parts, "\n\n")
}

func (b *PromptBuilder) BuildUserContextLayer() string {
	if !b.agent.InjectUserContext {
		return ""
	}

	var lines []string

	if b.context.UserDisplayName != "" {
		lines = append(lines, fmt.Sprintf("- **Name**: %s", b.context.UserDisplayName))
	}

	if b.context.UserEmail != "" {
		lines = append(lines, fmt.Sprintf("- **Email**: %s", b.context.UserEmail))
	}

	if b.context.UserLanguage != "" {
		lines = append(lines, fmt.Sprintf("- **Preferred Language**: %s", b.context.UserLanguage))
	}

	if b.context.UserTimezone != "" {
		lines = append(lines, fmt.Sprintf("- **Timezone**: %s", b.context.UserTimezone))
	}

	if b.agent.InjectWorkspace && b.context.WorkspaceName != "" {
		lines = append(lines, fmt.Sprintf("- **Current Workspace**: %s", b.context.WorkspaceName))
	}

	if b.agent.InjectTime && b.context.CurrentTime != "" {
		lines = append(lines, fmt.Sprintf("- **Current Time**: %s", b.context.CurrentTime))
	}

	if len(lines) == 0 {
		return ""
	}

	return "# User Context\n\n" + strings.Join(lines, "\n")
}

func (b *PromptBuilder) BuildMemoryLayer() string {
	if !b.agent.InjectMemory || len(b.context.RetrievedMemories) == 0 {
		return ""
	}

	var memLines []string
	for i, rm := range b.context.RetrievedMemories {
		mem := rm.Memory
		summary := mem.Summary
		if summary == "" {
			if len(mem.Content) > 200 {
				summary = mem.Content[:200] + "..."
			} else {
				summary = mem.Content
			}
		}
		memLines = append(memLines, fmt.Sprintf("%d. %s", i+1, summary))
	}

	return fmt.Sprintf("# Relevant Memories\n\n%s", strings.Join(memLines, "\n"))
}

func (b *PromptBuilder) BuildSessionLayer() string {
	if b.context.SessionSummary == "" && b.context.SessionTitle == "" {
		return ""
	}

	var parts []string

	if b.context.SessionTitle != "" {
		parts = append(parts, fmt.Sprintf("**Session**: %s", b.context.SessionTitle))
	}

	if b.context.SessionSummary != "" {
		parts = append(parts, b.context.SessionSummary)
	}

	if len(parts) == 0 {
		return ""
	}

	return "# Session Context\n\n" + strings.Join(parts, "\n\n")
}

func (b *PromptBuilder) BuildInstructionLayer() string {
	if b.agent.InstructionPrompt != "" {
		return b.agent.InstructionPrompt
	}
	return ""
}

func (b *PromptBuilder) BuildSafetyLayer() string {
	if b.agent.SafetyPrompt != "" {
		return b.agent.SafetyPrompt
	}
	return GetDefaultSafetyPrompt()
}

func (b *PromptBuilder) Assemble() *AssembledPrompt {
	layers := make(map[PromptLayer]string)
	var parts []string

	if layer := b.BuildIdentityLayer(); layer != "" {
		layers[LayerIdentity] = layer
		parts = append(parts, layer)
	}

	if layer := b.BuildCapabilityLayer(); layer != "" {
		layers[LayerCapability] = layer
		parts = append(parts, layer)
	}

	if layer := b.BuildUserContextLayer(); layer != "" {
		layers[LayerUserContext] = layer
		parts = append(parts, layer)
	}

	if layer := b.BuildMemoryLayer(); layer != "" {
		layers[LayerMemory] = layer
		parts = append(parts, layer)
	}

	if layer := b.BuildSessionLayer(); layer != "" {
		layers[LayerSession] = layer
		parts = append(parts, layer)
	}

	if layer := b.BuildInstructionLayer(); layer != "" {
		layers[LayerInstruction] = layer
		parts = append(parts, layer)
	}

	if layer := b.BuildSafetyLayer(); layer != "" {
		layers[LayerSafety] = layer
		parts = append(parts, layer)
	}

	return &AssembledPrompt{
		FullPrompt: strings.Join(parts, "\n\n---\n\n"),
		Layers:     layers,
	}
}

func BuildPromptContext(ctx context.Context, userID string, session *Session) PromptContext {
	promptCtx := PromptContext{
		CurrentTime:  GetCurrentTimeString(),
		SessionTitle: session.Title,
	}

	user, err := usermodel.GetUserByID(ctx, userID)
	if err == nil {
		promptCtx.UserEmail = user.Email
		promptCtx.UserDisplayName = user.UserProfile.DisplayName
		promptCtx.UserLanguage = user.UserProfile.Language
		promptCtx.UserTimezone = user.UserProfile.Timezone
	}

	if session.WorkspaceID != nil && *session.WorkspaceID != "" {
		workspace, err := workspacemodel.GetWorkSpaceByID(ctx, *session.WorkspaceID)
		if err == nil {
			promptCtx.WorkspaceName = workspace.Name
		}
	}

	return promptCtx
}

func GetCurrentTimeString() string {
	return time.Now().Format("2006-01-02 15:04:05 MST")
}

func GetCurrentTimeStringInTimezone(timezone string) string {
	if timezone == "" {
		timezone = "UTC"
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}

	return time.Now().In(loc).Format("2006-01-02 15:04:05 MST")
}
