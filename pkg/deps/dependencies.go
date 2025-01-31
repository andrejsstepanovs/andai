package deps

import (
	"database/sql"
	"fmt"

	"github.com/andrejsstepanovs/andai/pkg/ai"
	"github.com/andrejsstepanovs/andai/pkg/models"
	"github.com/andrejsstepanovs/andai/pkg/redmine"
	apiredmine "github.com/mattn/go-redmine"
	"github.com/spf13/viper"
)

type AppDependencies struct {
	Model   *redmine.Model
	LlmNorm *ai.AI
}

var Container *AppDependencies

func NewAppDependencies(mod models.LlmModels) (*AppDependencies, error) {
	if Container != nil {
		return Container, nil
	}

	db, err := sql.Open("mysql", viper.GetString("redmine.db"))
	if err != nil {
		return nil, err
	}
	api := apiredmine.NewClient(viper.GetString("redmine.url"), viper.GetString("redmine.api_key"))

	p := mod.Get(models.LlmModelNormal)
	aiNorm, err := ai.NewAI(p.Provider, p.Model, p.APIKey, p.Temperature)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI (normal) err: %v", err)
	}

	Container = &AppDependencies{
		Model:   redmine.NewModel(db, api),
		LlmNorm: aiNorm,
	}

	return Container, nil
}
