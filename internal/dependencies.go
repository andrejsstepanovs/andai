package internal

import (
	"database/sql"
	"fmt"

	"github.com/andrejsstepanovs/andai/internal/redmine"
	"github.com/andrejsstepanovs/andai/internal/settings"
	apiredmine "github.com/mattn/go-redmine"
	"github.com/spf13/viper"
)

type AppDependencies struct {
	Config  *settings.Config
	Model   *redmine.Model
	LlmPool *settings.LlmModels
}

type DependenciesLoader func() *AppDependencies

var Container *AppDependencies

func NewAppDependencies(config *settings.Config) (*AppDependencies, error) {
	if Container != nil {
		return Container, nil
	}

	params, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load settings err: %v", err)
	}

	db, err := sql.Open("mysql", viper.GetString("redmine.db"))
	if err != nil {
		return nil, err
	}
	api := apiredmine.NewClient(viper.GetString("redmine.url"), viper.GetString("redmine.api_key"))

	Container = &AppDependencies{
		Config:  config,
		Model:   redmine.NewModel(db, api),
		LlmPool: &params.LlmModels,
	}

	return Container, nil
}
