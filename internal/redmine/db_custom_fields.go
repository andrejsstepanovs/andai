package redmine

import (
	"errors"
	"fmt"
	"log"
	"strings"

	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/mattn/go-redmine"
)

// CustomFieldIssue is the type of custom field for issues
const CustomFieldIssue = "IssueCustomField"

func (c *Model) DBSaveCustomFields(customFields []redmine.CustomField, current []redmine.CustomField) ([]int64, error) {
	newCustomFields := make([]redmine.CustomField, 0)
	for _, field := range customFields {
		exists := false
		for _, s := range current {
			if s.Name == field.Name {
				exists = true
				break
			}
		}
		if !exists {
			newCustomFields = append(newCustomFields, field)
		}
	}

	if len(newCustomFields) == 0 {
		return nil, nil
	}

	newIDs := make([]int64, 0)
	for i, field := range newCustomFields {
		log.Printf("Creating New Custom Field: %s\n\n", field.Name)
		id, err := c.DBInsertCustomField(CustomFieldIssue, "string", field, i+1)
		if err != nil {
			return nil, fmt.Errorf("redmine issue status insert err: %v", err)
		}
		newIDs = append(newIDs, id)
	}

	return newIDs, nil
}

func (c *Model) DBInsertCustomField(customFieldType, fieldFormat string, field redmine.CustomField, position int) (int64, error) {
	formatStore := []string{
		fmt.Sprintf("text_formatting: ''"),
		fmt.Sprintf("url_pattern: ''"),
	}
	formatStoreParts := []string{"---"}
	for _, format := range formatStore {
		formatStoreParts = append(formatStoreParts, fmt.Sprintf("%s", format))
	}
	formatStoreParts = append(formatStoreParts, "")
	formatStoreStr := strings.Join(formatStoreParts, "\n")

	queryInsertCustomField := fmt.Sprintf("INSERT INTO custom_fields " +
		"(`type`, `name`, field_format, possible_values, `regexp`, min_length, max_length, is_required, is_for_all, is_filter, position, searchable, default_value, editable, visible, multiple, format_store, description) " +
		"VALUES (?, ?, ?, NULL, '', NULL, NULL, ?, 1, 0, ?, 0, '', 1, 1, 0, ?, ?)")

	result, err := c.execDML(
		queryInsertCustomField,
		customFieldType,   // type
		field.Name,        // name
		fieldFormat,       // field_format
		1,                 // is_required
		position,          // position
		formatStoreStr,    // format_store
		field.Description, // description
	)
	if err != nil {
		return 0, fmt.Errorf("error redmine custom field insert: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected err: %v", err)
	}
	if affected == 0 {
		return 0, errors.New("custom field not created")
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id err: %v", err)
	}

	return id, nil
}

func (c *Model) DBInsertCustomFieldTrackers(customFieldIDs, trackerIDs []int64) error {
	values := make([]string, 0)
	for _, customFieldID := range customFieldIDs {
		for _, trackerID := range trackerIDs {
			values = append(values, fmt.Sprintf("(%d, %d)", customFieldID, trackerID))
		}
	}
	if len(values) == 0 {
		return nil
	}
	sql := "INSERT INTO custom_fields_trackers (custom_field_id, tracker_id) VALUES " + strings.Join(values, ", ")

	result, err := c.execDML(sql)
	if err != nil {
		return fmt.Errorf("error redmine custom field trackers insert: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected err: %v", err)
	}
	if affected == 0 {
		return errors.New("custom field trackers not created")
	}
	return nil
}
