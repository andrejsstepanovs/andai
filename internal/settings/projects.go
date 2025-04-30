package settings

import "fmt"

type Projects []Project

type ProjectCommand struct {
	Name                   string   `yaml:"name"`
	Command                []string `yaml:"command"`
	IgnoreError            bool     `yaml:"ignore_err"`
	IgnoreStdOutIfNoStdErr bool     `yaml:"ignore_stdout_if_no_stderr"`
	SuccessIfNoOutput      bool     `yaml:"success_if_no_output"`
}

type ProjectCommands []ProjectCommand

type Project struct {
	Identifier             string          `yaml:"identifier"`
	Name                   string          `yaml:"name"`
	Description            string          `yaml:"description"`
	GitPath                string          `yaml:"git_path"`
	LocalGitPath           string          `yaml:"git_local_dir"`
	FinalBranch            string          `yaml:"final_branch"`
	DeleteBranchAfterMerge bool            `yaml:"delete_branch_after_merge"` // will delete source (child) branch after merge into parent
	Wiki                   string          `yaml:"wiki"`
	Commands               ProjectCommands `yaml:"commands"`
}

func (p Projects) Find(identifier string) Project {
	for _, project := range p {
		if project.Identifier == identifier {
			return project
		}
	}
	return Project{}
}

func (p ProjectCommands) Find(identifier string) (ProjectCommand, error) {
	for _, command := range p {
		if command.Name == identifier {
			return command, nil
		}
	}
	return ProjectCommand{}, fmt.Errorf("%q project commands not found", identifier)
}
