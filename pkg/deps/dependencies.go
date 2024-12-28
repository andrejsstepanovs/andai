package deps

import (
	"github.com/andrejsstepanovs/andai/pkg/redmine"
)

type AppDependencies struct {
	Model *redmine.Model
}

var Container *AppDependencies

func NewAppDependencies() *AppDependencies {
	if Container != nil {
		return Container
	}

	Container = &AppDependencies{
		Model: redmine.NewModel(),
	}

	return Container
}
