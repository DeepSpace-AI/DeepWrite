package agent

import (
	"time"

	"gorm.io/datatypes"
)

type Agent struct {
	ID           string         `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name         string         `json:"name" gorm:"type:varchar(100);not null"`
	Description  string         `json:"description" gorm:"type:text"`
	Type         string         `json:"type" gorm:"type:varchar(20);not null;default:'user';index"`
	Category     string         `json:"category" gorm:"type:varchar(50);index"`
	IconURL      string         `json:"icon_url" gorm:"type:varchar(512)"`
	SystemPrompt string         `json:"system_prompt" gorm:"type:text"`
	DefaultModel string         `json:"default_model" gorm:"type:varchar(100)"`
	Skills       datatypes.JSON `json:"skills" gorm:"type:jsonb;default:'[]'::jsonb"`
	Tools        datatypes.JSON `json:"tools" gorm:"type:jsonb;default:'[]'::jsonb"`
	Temperature  float64        `json:"temperature" gorm:"default:0.7"`
	MaxTokens    int            `json:"max_tokens" gorm:"default:4096"`
	OwnerID      *string        `json:"owner_id" gorm:"type:uuid;index"`
	Public       bool           `json:"public" gorm:"default:false;index"`
	Rating       float64        `json:"rating" gorm:"default:0"`
	UsageCount   int            `json:"usage_count" gorm:"default:0"`
	Enabled      bool           `json:"enabled" gorm:"default:true;index"`

	IdentityPrompt    string `json:"identity_prompt" gorm:"type:text"`
	CapabilityPrompt  string `json:"capability_prompt" gorm:"type:text"`
	InstructionPrompt string `json:"instruction_prompt" gorm:"type:text"`
	SafetyPrompt      string `json:"safety_prompt" gorm:"type:text"`

	InjectUserContext bool `json:"inject_user_context" gorm:"default:true"`
	InjectMemory      bool `json:"inject_memory" gorm:"default:true"`
	InjectTime        bool `json:"inject_time" gorm:"default:true"`
	InjectWorkspace   bool `json:"inject_workspace" gorm:"default:true"`

	MemoryRetrievalCount int     `json:"memory_retrieval_count" gorm:"default:5"`
	MemoryMinRelevance   float64 `json:"memory_min_relevance" gorm:"default:0.7"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Agent) TableName() string {
	return DefaultAgentTableName
}

type AgentSkill struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type AgentTool struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}
