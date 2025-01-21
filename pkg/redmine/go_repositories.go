package redmine

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/andrejsstepanovs/andai/pkg/redmine/models"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mattn/go-redmine"
	"github.com/spf13/viper"
)

const (
	queryGetProjectRepository = "SELECT id, project_id, root_url FROM repositories WHERE identifier = ?"
	queryInsertRepository     = "INSERT INTO repositories (project_id, root_url, type, path_encoding, extra_info, identifier, is_default, created_on) VALUES(?, ?, 'Repository::Git', 'UTF-8', ?, ?, 1, NOW())"
	queryUpdateRepository     = "UPDATE repositories SET root_url = ?, created_on = NOW() WHERE id = ?"
)

func (c *Model) DBSaveGit(project redmine.Project, gitPath string) error {
	newURL := fmt.Sprintf("%s/%s", strings.TrimRight(viper.GetString("redmine.repositories"), "/"), strings.TrimLeft(gitPath, "/"))

	repository, err := c.DBGetRepository(project)
	if err != nil {
		return fmt.Errorf("redmine get repository err: %v", err)
	}
	if repository.ID > 0 {
		log.Printf("Repository ID: %d, ProjectID: %s, Url: %s\n", repository.ID, repository.ProjectID, repository.RootURL)
		if repository.RootURL == newURL {
			return nil
		}
		log.Println("Mismatch Repository root_url")
		result, err := c.execDML(queryUpdateRepository, newURL, repository.ID)
		if err != nil {
			return fmt.Errorf("update repository db err: %v", err)
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("rows affected err: %v", err)
		}
		if affected == 0 {
			return errors.New("project repository root_url not changed")
		}
		log.Printf("Project %s repository root_url updated\n", project.Identifier)
		return nil
	}

	result, err := c.execDML(queryInsertRepository, project.Id, newURL, "---\nextra_report_last_commit: '0'\n", project.Identifier)
	if err != nil {
		return fmt.Errorf("error redmine git save: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected err: %v", err)
	}
	if affected == 0 {
		return errors.New("project repository not saved")
	}

	log.Println("project repository inserted")
	return nil
}

func (c *Model) DBGetRepository(project redmine.Project) (models.Repository, error) {
	var repos []models.Repository
	err := c.queryAndScan(queryGetProjectRepository, func(rows *sql.Rows) error {
		var row models.Repository
		if err := rows.Scan(&row.ID, &row.ProjectID, &row.RootURL); err != nil {
			return err
		}
		repos = append(repos, row)
		return nil
	}, project.Identifier)

	if err != nil && !errors.As(err, &sql.ErrNoRows) {
		return models.Repository{}, err
	}
	for _, row := range repos {
		return row, nil
	}
	return models.Repository{}, nil
}
