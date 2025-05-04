package setup

import (
	"fmt"
	"log"

	"github.com/andrejsstepanovs/andai/internal"
	model "github.com/andrejsstepanovs/andai/internal/redmine"
	"github.com/andrejsstepanovs/andai/internal/redmine/models"
	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/spf13/cobra"
)

func newCustomFieldsCommand(deps internal.DependenciesLoader) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "custom-fields",
		Short: "Save (Update) custom fields",
		RunE: func(_ *cobra.Command, _ []string) error {
			log.Println("Processing custom fields sync")

			err := setupCustomFields(deps().Model)
			if err != nil {
				return err
			}

			log.Println("Custom fields repository OK")
			return nil
		},
	}
	return cmd
}

func setupCustomFields(mod *model.Model) error {
	createFields := []models.CustomField{
		{
			Name:        model.CustomFieldBranch,
			Description: "Branch name to work in. If not set will be AI-123 (task id) or if main task and not set `final_branch` value will be used.",
			Type:        "string",
			Default:     "",
			FormatStore: []string{
				"text_formatting: ''",
				"url_pattern: ''",
				"",
			},
			IsFilter: 1,
		},
		{
			Name:        model.CustomFieldSkipMerge,
			Description: "If True on success will NOT merge branch into parent task. If no parent issue exists, then will use andai final_branch value.",
			Type:        "bool",
			Default:     "0",
			FormatStore: []string{
				"url_pattern: ''",
				"edit_tag_style: check_box",
				"",
			},
			IsFilter: 1,
		},
	}

	trackerIDs := make([]int64, 0)
	allTrackers, err := mod.API().Trackers()
	if err != nil {
		log.Println("Redmine Trackers Failed to enable API")
		return fmt.Errorf("error redmine trackers: %v", err)
	}
	for _, tracker := range allTrackers {
		trackerIDs = append(trackerIDs, int64(tracker.Id))
	}
	if len(trackerIDs) == 0 {
		return fmt.Errorf("no redmine trackers found: %v", err)
	}

	currentCustomFields, err := mod.API().CustomFields()
	if err != nil {
		log.Println("Redmine Custom Fields Failed to enable API")
		return fmt.Errorf("error redmine custom fields: %v", err)
	}

	customFieldIDs, err := mod.DBSaveCustomFields(createFields, currentCustomFields)
	if err != nil {
		log.Println("Failed to save Redmine Custom Fields")
		return fmt.Errorf("error redmine custom_fields save: %v", err)
	}

	err = mod.DBInsertCustomFieldTrackers(customFieldIDs, trackerIDs)
	if err != nil {
		log.Println("Failed to save Redmine Custom Fields")
		return fmt.Errorf("error redmine custom_fields save: %v", err)
	}

	return nil
}
