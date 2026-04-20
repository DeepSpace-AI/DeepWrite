package agent

const (
	AgentTypeOfficial = "official"
	AgentTypeUser     = "user"

	AgentCategoryResearch   = "research"
	AgentCategoryWriting    = "writing"
	AgentCategoryData       = "data"
	AgentCategoryPublishing = "publishing"
	AgentCategoryGeneral    = "general"

	SessionStatusActive   = "active"
	SessionStatusArchived = "archived"

	MessageRoleUser      = "user"
	MessageRoleAssistant = "assistant"
	MessageRoleSystem    = "system"

	MemoryTypePreference = "preference"
	MemoryTypeFact       = "fact"
	MemoryTypeTask       = "task"
	MemoryTypeInsight    = "insight"

	DefaultAgentTableName   = "dw_agents"
	DefaultSessionTableName = "dw_agent_sessions"
	DefaultMessageTableName = "dw_agent_messages"
	DefaultMemoryTableName  = "dw_agent_memories"
)
