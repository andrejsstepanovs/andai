package api

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mattn/go-redmine"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewProjectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Create or update project",
		RunE:  runRedmineSaveProject,
	}
	return cmd
}

func runRedmineSaveProject(cmd *cobra.Command, args []string) error {
	fmt.Println("Processing Redmine project save", len(args))

	project := redmine.Project{
		Name:        viper.GetString("project.name"),
		Identifier:  viper.GetString("project.id"),
		Description: viper.GetString("project.description"),
		//CustomFields: nil,
	}

	client := redmine.NewClient(viper.GetString("redmine.url"), viper.GetString("redmine.api_key"))

	err, project := saveProject(client, project)
	if err != nil {
		fmt.Println("Redmine Project Save Fail")
		return fmt.Errorf("error redmine project save: %v", err)
	}

	err = saveWiki(client, project)
	if err != nil {
		fmt.Println("Redmine Project Wiki Save Fail")
		return fmt.Errorf("error redmine project save: %v", err)
	}

	fmt.Printf("ID: %d, Name: %s\n", project.Id, project.Name)

	err = saveGit(project)
	if err != nil {
		fmt.Println("Redmine Git Save Fail")
		return fmt.Errorf("error redmine git save: %v", err)
	}

	return nil
}

func saveWiki(client *redmine.Client, project redmine.Project) error {
	const TITLE = "Wiki"
	page, err := client.WikiPage(project.Id, TITLE)
	if err != nil {
		if "Not Found" != err.Error() {
			fmt.Println("Redmine Wiki Page Fail")
			return fmt.Errorf("error redmine wiki page: %v", err)
		}
		page = &redmine.WikiPage{
			Title: TITLE,
			Text:  viper.GetString("project.wiki"),
		}
		_, err = client.CreateWikiPage(project.Id, *page)
		if err != nil {
			fmt.Println("Redmine Wiki Page Create Fail")
			return fmt.Errorf("error redmine wiki page create: %v", err)
		}
		fmt.Printf("Wiki page created: %s\n", TITLE)
		return nil
	}

	page.Text = strings.TrimSpace(viper.GetString("project.wiki"))
	err = client.UpdateWikiPage(project.Id, *page)
	if err != nil && err.Error() != "EOF" {
		fmt.Println("Redmine Wiki Page Update Fail")
		return fmt.Errorf("error redmine wiki page update: %v", err)
	}
	return nil
}

func saveGit(project redmine.Project) error {
	newUrl := fmt.Sprintf("%s/%s", strings.TrimRight(viper.GetString("redmine.repositories"), "/"), strings.TrimLeft(viper.GetString("project.git_path"), "/"))

	conn, err := sql.Open("mysql", viper.GetString("redmine.db"))
	if err != nil {
		fmt.Println("Redmine Git Save Fail")
		return fmt.Errorf("error redmine git save: %v", err)
	}

	results, err := conn.Query("SELECT id, project_id, root_url FROM repositories WHERE identifier = ?", project.Identifier)
	if err != nil {
		fmt.Println("Redmine Admin Fail")
		return fmt.Errorf("error redmine admin: %v", err)
	}
	defer results.Close()

	type Row struct {
		ID        int
		ProjectID string
		RootUrl   string
	}
	var rows []Row

	// Loop through rows, using Scan to assign column data to struct fields.
	for results.Next() {
		var row Row
		if err := results.Scan(&row.ID, &row.ProjectID, &row.RootUrl); err != nil {
			return err
		}
		rows = append(rows, row)
	}
	if err = results.Err(); err != nil {
		return err
	}

	if len(rows) > 0 {
		for _, row := range rows {
			fmt.Printf("ID: %d, ProjectID: %s, Url: %s\n", row.ID, row.ProjectID, row.RootUrl)

			if row.RootUrl != newUrl {
				fmt.Println("Git repository already exists. Changing URL..")

				result, err := conn.Exec("UPDATE repositories SET root_url = ?, created_on = NOW() WHERE id = ?", newUrl, row.ID)
				if err != nil {
					return fmt.Errorf("update settings db err: %v", err)
				}
				affected, err := result.RowsAffected()
				if err != nil {
					return fmt.Errorf("rows affected err: %v", err)
				}
				if affected > 0 {
					fmt.Println("Git repository url updated")
					return nil
				}
				return errors.New("Git repository url not changed")
			}
		}
		return nil
	}

	query := "INSERT INTO repositories " +
		"(project_id, root_url, type, path_encoding, extra_info, identifier, is_default, created_on) " +
		"VALUES(?, ?, 'Repository::Git', 'UTF-8', ?, ?, 1, NOW())"

	result, err := conn.Exec(query, project.Id, newUrl, "---\nextra_report_last_commit: '0'\n", project.Identifier)
	if err != nil {
		fmt.Println("Redmine Git Save Fail")
		return fmt.Errorf("error redmine git save: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected err: %v", err)
	}

	if affected > 0 {
		fmt.Println("Git repository saved")
	} else {
		return errors.New("Git repository not saved")
	}

	return nil
}

func saveProject(client *redmine.Client, project redmine.Project) (error, redmine.Project) {
	projects, err := client.Projects()
	if err != nil {
		fmt.Println("Redmine Projects Fail")
		return fmt.Errorf("error redmine projects: %v", err), redmine.Project{}
	}

	for _, p := range projects {
		fmt.Println(fmt.Sprintf("ID: %d, Name: %s", p.Id, p.Name))
		if p.Identifier == viper.GetString("project.id") {
			fmt.Printf("Project already exists: %s\n", p.Name)
			project.Id = p.Id
			err = client.UpdateProject(project)
			if err != nil && err.Error() != "EOF" {
				fmt.Println("Redmine Update Project Failed")
				return fmt.Errorf("error redmine update project: %v", err.Error()), redmine.Project{}
			}
			fmt.Printf("Project updated: %s\n", project.Name)
			return nil, project
		}
	}

	response, err := client.CreateProject(project)
	if err != nil {
		fmt.Println("Redmine Create Project Failed")
		return fmt.Errorf("error redmine create project: '%s'", err.Error()), redmine.Project{}
	}

	return nil, *response
}
