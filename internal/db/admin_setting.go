package db

import (
	"sync/atomic"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// settingsVersion is bumped whenever a setting row changes. The settings
// package uses it to detect writes made outside its API (tests, the legacy
// set-config route, a second process sharing the database) without querying
// the database on every request.
var settingsVersion atomic.Uint64

// SettingsVersion returns the current in-process settings generation.
func SettingsVersion() uint64 { return settingsVersion.Load() }

func bumpSettingsVersion() { settingsVersion.Add(1) }

type AdminSetting struct {
	Key   string `gorm:"index:,unique"`
	Value string
}

const (
	SettingDisableSignup          = "disable-signup"
	SettingRequireLogin           = "require-login"
	SettingAllowGistsWithoutLogin = "allow-gists-without-login"
	SettingDisableLoginForm       = "disable-login-form"
	SettingDisableGravatar        = "disable-gravatar"
)

func GetSetting(key string) (string, error) {
	var setting AdminSetting
	var err error
	switch db.Name() {
	case "mysql", "sqlite":
		err = db.Where("`key` = ?", key).First(&setting).Error
	case "postgres":
		err = db.Where("key = ?", key).First(&setting).Error
	}
	return setting.Value, err
}

func GetSettings() (map[string]string, error) {
	var settings []AdminSetting
	err := db.Find(&settings).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, setting := range settings {
		result[setting.Key] = setting.Value
	}

	return result, nil
}

func UpdateSetting(key string, value string) error {
	err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}}, // key column
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&AdminSetting{
		Key:   key,
		Value: value,
	}).Error
	if err == nil {
		bumpSettingsVersion()
	}
	return err
}

// UpdateSettings writes several settings in one transaction. Callers that
// need all-or-nothing semantics (the admin panel) should validate first and
// then use this instead of looping over UpdateSetting.
func UpdateSettings(settings map[string]string) error {
	if len(settings) == 0 {
		return nil
	}
	return db.Transaction(func(tx *gorm.DB) error {
		for key, value := range settings {
			err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "key"}},
				DoUpdates: clause.AssignmentColumns([]string{"value"}),
			}).Create(&AdminSetting{Key: key, Value: value}).Error
			if err != nil {
				return err
			}
		}
		bumpSettingsVersion()
		return nil
	})
}

func setSetting(key string, value string) error {
	err := db.FirstOrCreate(&AdminSetting{Key: key, Value: value}, &AdminSetting{Key: key}).Error
	if err == nil {
		bumpSettingsVersion()
	}
	return err
}

func initAdminSettings(settings map[string]string) error {
	for key, value := range settings {
		if err := setSetting(key, value); err != nil {
			if !IsUniqueConstraintViolation(err) {
				return err
			}
		}
	}

	return nil
}

type AuthInfo struct{}

func (auth AuthInfo) RequireLogin() (bool, error) {
	s, err := GetSetting(SettingRequireLogin)
	if err != nil {
		return true, err
	}
	return s == "1", nil
}

func (auth AuthInfo) AllowGistsWithoutLogin() (bool, error) {
	s, err := GetSetting(SettingAllowGistsWithoutLogin)
	if err != nil {
		return false, err
	}
	return s == "1", nil
}
