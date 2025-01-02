package deps

import (
	"database/sql"

	"github.com/andrejsstepanovs/andai/pkg/redmine"
	apiredmine "github.com/mattn/go-redmine"
	"github.com/spf13/viper"
)

type AppDependencies struct {
	Model *redmine.Model
}

var Container *AppDependencies

func NewAppDependencies() (*AppDependencies, error) {
	if Container != nil {
		return Container, nil
	}

	db, err := sql.Open("mysql", viper.GetString("redmine.db"))
	if err != nil {
		return nil, err
	}
	api := apiredmine.NewClient(viper.GetString("redmine.url"), viper.GetString("redmine.api_key"))

	//api.Trackers()
	Container = &AppDependencies{
		Model: redmine.NewModel(db, api),
	}

	return Container, nil
}
