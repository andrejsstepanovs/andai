package models

type CustomField struct {
	Name        string
	Description string
	Type        string
	Default     string
	IsFilter    int
	FormatStore []string
}
