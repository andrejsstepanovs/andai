package redmine

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/andrejsstepanovs/andai/internal/redmine/models"
	_ "github.com/go-sql-driver/mysql" // mysql driver
)

const (
	// TokenActionAPI is a constant for API token action
	TokenActionAPI = "api"
	// SettingRestAPIEnabled setting for enabling REST API
	SettingRestAPIEnabled = "rest_api_enabled"
	// SettingEnableSCM setting for enabling SCM (source control)
	SettingEnableSCM = "enabled_scm"
	// SettingAutofetchChengesets setting name for auto repo reading
	SettingAutofetchChengesets = "autofetch_changesets"
	// SettingSysAPIEnabled setting name to enable system api resource
	SettingSysAPIEnabled = "sys_api_enabled"
	// SettingSysAdminKey defines system api key value setting
	SettingSysAdminKey = "sys_api_key"

	settingsValueEnabled = "1"
	autoIncrementDefault = 100
)

const (
	queryUpdateAdminNoChangePass = "UPDATE users SET must_change_passwd = 0 WHERE login = ?"                   // nolint:gosec
	queryGetSettings             = "SELECT id, name, value FROM settings WHERE name = ?"                       // nolint:gosec
	queryInsertSettings          = "INSERT INTO settings (name, value, updated_on) VALUES (?, ?, NOW())"       // nolint:gosec
	queryUpdateSettingsValue     = "UPDATE settings SET value = ?, updated_on = NOW() WHERE name = ?"          // nolint:gosec
	queryGetAutoIncrement        = "SELECT AUTO_INCREMENT FROM information_schema.TABLES WHERE TABLE_NAME = ?" // nolint:gosec
	querySetAutoIncrement        = "ALTER TABLE %s AUTO_INCREMENT = %d"                                        // nolint:gosec
)

func (c *Model) DBGetSettings(settingName string) ([]models.Settings, error) {
	var settings []models.Settings
	err := c.queryAndScan(queryGetSettings, func(rows *sql.Rows) error {
		var row models.Settings
		if err := rows.Scan(&row.ID, &row.Name, &row.Value); err != nil {
			return err
		}
		settings = append(settings, row)
		return nil
	}, settingName)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return settings, nil
}

func (c *Model) DBGetAutoIncrementValue(tableName string) (int, error) {
	var autoIncrementValue []int
	err := c.queryAndScan(queryGetAutoIncrement, func(rows *sql.Rows) error {
		var row int
		if err := rows.Scan(&row); err != nil {
			return err
		}
		autoIncrementValue = append(autoIncrementValue, row)
		return nil
	}, tableName)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	return autoIncrementValue[0], nil
}

func (c *Model) DBSetAutoIncrementValue(tableName string, newValue int) error {
	if newValue < 0 {
		return fmt.Errorf("new value %d is less than 0", newValue)
	}
	result, err := c.execDML(fmt.Sprintf(querySetAutoIncrement, tableName, newValue))
	if err != nil {
		return fmt.Errorf("set auto increment value db err: %v", err)
	}
	_, err = result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected err: %v", err)
	}
	return nil
}

func (c *Model) DBSettingsInsert(settingName, value string) error {
	result, err := c.execDML(queryInsertSettings, settingName, value)
	if err != nil {
		return fmt.Errorf("update settings db err: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected err: %v", err)
	}
	if affected == 0 {
		return errors.New("admin rest_api_enabled not changed")
	}
	return nil
}

func (c *Model) DBSettingsUpdate(settingName, value string) error {
	result, err := c.execDML(queryUpdateSettingsValue, value, settingName)
	if err != nil {
		return fmt.Errorf("update settings db err: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected err: %v", err)
	}
	if affected > 0 {
		return nil
	}
	return nil
}

func (c *Model) DBSetAutoIncrements() error {
	for _, table := range []string{"users", "projects", "issues"} {
		value, err := c.DBGetAutoIncrementValue(table)
		if err != nil {
			return fmt.Errorf("get auto increment db err: %v", err)
		}
		if value >= autoIncrementDefault {
			log.Printf("Auto Increment for %s is %d\n", table, value)
			continue
		}

		err = c.DBSetAutoIncrementValue(table, autoIncrementDefault)
		if err != nil {
			return fmt.Errorf("set auto increment db err: %v", err)
		}
	}
	return nil
}

func (c *Model) DBSettingsEnableAPI() error {
	rows, err := c.DBGetSettings(SettingRestAPIEnabled)
	if err != nil {
		return fmt.Errorf("get settings db err: %v", err)
	}
	if len(rows) > 0 {
		row := rows[0]
		log.Printf("Setting Identifier: %d, Name: %s, Value: %s\n", row.ID, row.Name, row.Value)
		if row.Value != settingsValueEnabled {
			log.Printf("Setting %s is not enabled\n", SettingRestAPIEnabled)
			err = c.DBSettingsUpdate(SettingRestAPIEnabled, settingsValueEnabled)
			if err != nil {
				return fmt.Errorf("update settings db err: %v", err)
			}
			log.Printf("Setting %s updated to %s\n", SettingRestAPIEnabled, settingsValueEnabled)
			return nil
		}
		log.Printf("Setting %s is OK\n", SettingRestAPIEnabled)
		return nil
	}

	log.Printf("Setting %s is not present\n", SettingRestAPIEnabled)
	err = c.DBSettingsInsert(SettingRestAPIEnabled, settingsValueEnabled)
	if err != nil {
		return fmt.Errorf("insert settings db err: %v", err)
	}
	log.Printf("Setting %s created with value: %s\n", SettingRestAPIEnabled, settingsValueEnabled)
	return nil
}

func (c *Model) DBSettingsEnableSysAPIEnabled() error {
	rows, err := c.DBGetSettings(SettingSysAPIEnabled)
	if err != nil {
		return fmt.Errorf("get settings db err: %v", err)
	}
	if len(rows) > 0 {
		row := rows[0]
		log.Printf("Setting Identifier: %d, Name: %s, Value: %s\n", row.ID, row.Name, row.Value)
		if row.Value != settingsValueEnabled {
			log.Printf("Setting %s is not enabled\n", SettingSysAPIEnabled)
			err = c.DBSettingsUpdate(SettingSysAPIEnabled, settingsValueEnabled)
			if err != nil {
				return fmt.Errorf("update settings db err: %v", err)
			}
			log.Printf("Setting %s updated to %s\n", SettingSysAPIEnabled, settingsValueEnabled)
			return nil
		}
		log.Printf("Setting %s is OK\n", SettingSysAPIEnabled)
		return nil
	}

	log.Printf("Setting %s is not present\n", SettingSysAPIEnabled)
	err = c.DBSettingsInsert(SettingSysAPIEnabled, settingsValueEnabled)
	if err != nil {
		return fmt.Errorf("insert settings db err: %v", err)
	}
	log.Printf("Setting %s created with value: %s\n", SettingSysAPIEnabled, settingsValueEnabled)
	return nil
}

func (c *Model) DBSettingsEnableAutofetchChengesets() error {
	rows, err := c.DBGetSettings(SettingAutofetchChengesets)
	if err != nil {
		return fmt.Errorf("get settings db err: %v", err)
	}
	if len(rows) > 0 {
		row := rows[0]
		log.Printf("Setting Identifier: %d, Name: %s, Value: %s\n", row.ID, row.Name, row.Value)
		if row.Value != settingsValueEnabled {
			log.Printf("Setting %s is not enabled\n", SettingAutofetchChengesets)
			err = c.DBSettingsUpdate(SettingAutofetchChengesets, settingsValueEnabled)
			if err != nil {
				return fmt.Errorf("update settings db err: %v", err)
			}
			log.Printf("Setting %s updated to %s\n", SettingAutofetchChengesets, settingsValueEnabled)
			return nil
		}
		log.Printf("Setting %s is OK\n", SettingAutofetchChengesets)
		return nil
	}

	log.Printf("Setting %s is not present\n", SettingAutofetchChengesets)
	err = c.DBSettingsInsert(SettingAutofetchChengesets, settingsValueEnabled)
	if err != nil {
		return fmt.Errorf("insert settings db err: %v", err)
	}
	log.Printf("Setting %s created with value: %s\n", SettingAutofetchChengesets, settingsValueEnabled)
	return nil
}

func (c *Model) DBSettingsEnableSCM() error {
	rows, err := c.DBGetSettings(SettingEnableSCM)
	if err != nil {
		return fmt.Errorf("get settings db err: %v", err)
	}

	val := "---\n- Subversion\n- Mercurial\n- Cvs\n- Bazaar\n- Git\n"
	if len(rows) > 0 {
		row := rows[0]
		log.Printf("Setting Identifier: %d, Name: %s, Value: %s\n", row.ID, row.Name, row.Value)
		if row.Value != val {
			log.Printf("Setting %s is not enabled\n", SettingEnableSCM)
			err = c.DBSettingsUpdate(SettingEnableSCM, val)
			if err != nil {
				return fmt.Errorf("update settings db err: %v", err)
			}
			log.Printf("Setting %s updated to %s\n", SettingEnableSCM, val)
			return nil
		}
		log.Printf("Setting %s is OK\n", SettingEnableSCM)
		return nil
	}

	log.Printf("Setting %s is not present\n", SettingEnableSCM)
	err = c.DBSettingsInsert(SettingEnableSCM, val)
	if err != nil {
		return fmt.Errorf("insert settings db err: %v", err)
	}
	log.Printf("Setting %s created with value: %s\n", SettingEnableSCM, val)
	return nil
}

func (c *Model) DBSettingsSetSysAPIKey(tokenValue string) error {
	rows, err := c.DBGetSettings(SettingSysAdminKey)
	if err != nil {
		return fmt.Errorf("get settings db err: %v", err)
	}

	if len(rows) > 0 {
		row := rows[0]
		log.Printf("Setting Identifier: %d, Name: %s, Value: %s\n", row.ID, row.Name, row.Value)
		if row.Value != tokenValue {
			log.Printf("Setting %s is not enabled\n", SettingSysAdminKey)
			err = c.DBSettingsUpdate(SettingSysAdminKey, tokenValue)
			if err != nil {
				return fmt.Errorf("update settings db err: %v", err)
			}
			log.Printf("Setting %s updated to %s\n", SettingSysAdminKey, tokenValue)
			return nil
		}
		log.Printf("Setting %s is OK\n", SettingSysAdminKey)
		return nil
	}

	log.Printf("Setting %s is not present\n", SettingSysAdminKey)
	err = c.DBSettingsInsert(SettingSysAdminKey, tokenValue)
	if err != nil {
		return fmt.Errorf("insert settings db err: %v", err)
	}
	log.Printf("Setting %s created with value: %s\n", SettingSysAdminKey, tokenValue)
	return nil
}

func (c *Model) DBSettingsAdminMustNotChangePassword() error {
	result, err := c.execDML(queryUpdateAdminNoChangePass, "admin")
	if err != nil {
		return fmt.Errorf("update users db err: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected err: %v", err)
	}

	if affected > 0 {
		log.Println("Admin must_change_passwd set to 0")
	} else {
		log.Println("Admin must_change_passwd not changed")
	}

	return nil
}
