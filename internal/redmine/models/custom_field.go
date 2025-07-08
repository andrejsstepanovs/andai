package models

type CustomField struct {
	Name        string
	Description string
	Type        string
	Default     string
	IsFilter    int
	Visible     int
	Editable    int
	FormatStore []string
}
