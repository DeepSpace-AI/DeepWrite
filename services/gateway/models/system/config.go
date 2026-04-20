package system

import (
	"time"
)

const ConfigTableName = "system_configs"

type SystemConfig struct {
	Key         string    `json:"key" gorm:"primaryKey;type:varchar(100)"`
	Value       string    `json:"value" gorm:"type:text"`
	Description string    `json:"description" gorm:"type:varchar(255)"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	UpdatedBy   *string   `json:"updated_by" gorm:"type:uuid"`
}

func (SystemConfig) TableName() string {
	return ConfigTableName
}

const (
	ConfigKeyDefaultModel          = "default_ai_model"
	ConfigKeyTitleGenModel         = "title_generation_model"
	ConfigKeyMemoryExtractionModel = "memory_extraction_model"
	ConfigKeySummaryGenModel       = "summary_generation_model"
)

type ConfigDefinition struct {
	Key         string `json:"key"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Type        string `json:"type"`
}

var ConfigDefinitions = []ConfigDefinition{
	{Key: ConfigKeyDefaultModel, Description: "System default AI model for general tasks", Category: "ai", Type: "model_select"},
	{Key: ConfigKeyTitleGenModel, Description: "AI model for title generation (optional, uses default if not set)", Category: "ai", Type: "model_select"},
	{Key: ConfigKeyMemoryExtractionModel, Description: "AI model for memory extraction (optional)", Category: "ai", Type: "model_select"},
	{Key: ConfigKeySummaryGenModel, Description: "AI model for summary generation (optional)", Category: "ai", Type: "model_select"},
}
