package models

type Projects []Project

type Project struct {
	Identifier   string `yaml:"identifier"`
	Name         string `yaml:"name"`
	Description  string `yaml:"description"`
	GitPath      string `yaml:"git_path"`
	LocalGitPath string `yaml:"git_local_dir"`
	Wiki         string `yaml:"wiki"`
}

func (p Projects) Find(identifier string) Project {
	for _, project := range p {
		if project.Identifier == identifier {
			return project
		}
	}
	return Project{}
}
