package deps

import (
	"database/sql"

	"github.com/andrejsstepanovs/andai/pkg/llm"
	"github.com/andrejsstepanovs/andai/pkg/models"
	"github.com/andrejsstepanovs/andai/pkg/redmine"
	apiredmine "github.com/mattn/go-redmine"
	"github.com/spf13/viper"
)

type AppDependencies struct {
	Model *redmine.Model
	LLM   *llm.LLM
}

var Container *AppDependencies

func NewAppDependencies(models models.LlmModels) (*AppDependencies, error) {
	if Container != nil {
		return Container, nil
	}

	db, err := sql.Open("mysql", viper.GetString("redmine.db"))
	if err != nil {
		return nil, err
	}
	api := apiredmine.NewClient(viper.GetString("redmine.url"), viper.GetString("redmine.api_key"))

	llm := llm.NewLLM(models)

	//api.Project()
	Container = &AppDependencies{
		Model: redmine.NewModel(db, api),
		LLM:   llm,
	}

	return Container, nil
}
