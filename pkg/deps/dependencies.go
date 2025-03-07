package deps

import (
	"database/sql"
	"fmt"

	"github.com/andrejsstepanovs/andai/pkg/ai"
	"github.com/andrejsstepanovs/andai/pkg/config"
	"github.com/andrejsstepanovs/andai/pkg/models"
	"github.com/andrejsstepanovs/andai/pkg/redmine"
	apiredmine "github.com/mattn/go-redmine"
	"github.com/spf13/viper"
)

type AppDependencies struct {
	Config  *config.Config
	Model   *redmine.Model
	LlmNorm *ai.AI
}

var Container *AppDependencies

func NewAppDependencies(config *config.Config) (*AppDependencies, error) {
	if Container != nil {
		return Container, nil
	}

	settings, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load settings err: %v", err)
	}

	db, err := sql.Open("mysql", viper.GetString("redmine.db"))
	if err != nil {
		return nil, err
	}
	api := apiredmine.NewClient(viper.GetString("redmine.url"), viper.GetString("redmine.api_key"))

	p := settings.LlmModels.Get(models.LlmModelNormal)
	aiNorm, err := ai.NewAI(p)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI (normal) err: %v", err)
	}

	Container = &AppDependencies{
		Config:  config,
		Model:   redmine.NewModel(db, api),
		LlmNorm: aiNorm,
	}

	return Container, nil
}
