package settings

import "time"

type Aider struct {
	Timeout           time.Duration `yaml:"timeout"`
	Config            string        `yaml:"config"`
	ConfigFallback    string        `yaml:"config_fallback"`
	MapTokens         int           `yaml:"map_tokens"`
	ModelMetadataFile string        `yaml:"model_metadata_file"`
	TaskSummaryPrompt string        `yaml:"task_summary_prompt"`
}
