package db

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type MigrationVersion struct {
	ID      uint `gorm:"primaryKey"`
	Version uint
}

func applyMigrations(dbInfo *databaseInfo) error {
	switch dbInfo.Type {
	case SQLite, PostgreSQL, MySQL:
		return applyAllMigrations(dbInfo.Type)
	default:
		return fmt.Errorf("unknown database type: %s", dbInfo.Type)
	}
}
func applyAllMigrations(dbType databaseType) error {
	if err := db.AutoMigrate(&MigrationVersion{}); err != nil {
		log.Fatal().Err(err).Msg("Error creating migration version table")
		return err
	}

	var currentVersion MigrationVersion
	db.First(&currentVersion)
	migrations := []struct {
		Version uint
		DBTypes []databaseType // nil = all types
		Func    func(*gorm.DB) error
	}{
		{1, []databaseType{SQLite}, v1_modifyConstraintToSSHKeys},
		{2, []databaseType{SQLite}, v2_lowercaseEmails},
		{3, nil, v3_normalizedColumns},
		{4, nil, v4_uniqueUsername},
		{5, nil, v5_gistCommentsCount},
	}

	for _, m := range migrations {
		if m.Version <= currentVersion.Version {
			continue
		}
		// Skip migrations not intended for this DB type
		if len(m.DBTypes) > 0 {
			applicable := false
			for _, t := range m.DBTypes {
				if t == dbType {
					applicable = true
					break
				}
			}
			if !applicable {
				// Advance version so we don't retry on next startup
				currentVersion.Version = m.Version
				db.Save(&currentVersion)
				continue
			}
		}

		tx := db.Begin()
		if err := tx.Error; err != nil {
			log.Fatal().Err(err).Msg("Error starting transaction")
			return err
		}
		if err := m.Func(tx); err != nil {
			tx.Rollback()
			log.Fatal().Err(err).Msg(fmt.Sprintf("Error applying migration %d:", m.Version))
			return err
		}

		if err := tx.Commit().Error; err != nil {
			log.Fatal().Err(err).Msg(fmt.Sprintf("Error committing migration %d:", m.Version))
			return err
		}

		currentVersion.Version = m.Version
		db.Save(&currentVersion)
		log.Info().Msg(fmt.Sprintf("Migration %d applied successfully", m.Version))
	}

	return nil
}

// Modify the constraint on the ssh_keys table to use ON DELETE CASCADE
func v1_modifyConstraintToSSHKeys(tx *gorm.DB) error {
	createSQL := `
	CREATE TABLE ssh_keys_temp (
		id integer primary key,
		title text,
		content text,
		sha text,
		created_at integer,
		last_used_at integer,
		user_id integer
		constraint fk_users_ssh_keys references users(id) on update cascade on delete cascade
	);
	`

	if err := tx.Exec(createSQL).Error; err != nil {
		return err
	}
	// Copy data from the old table to the new table
	copySQL := `INSERT INTO ssh_keys_temp SELECT * FROM ssh_keys;`
	if err := tx.Exec(copySQL).Error; err != nil {
		return err
	}

	// Drop the old table
	dropSQL := `DROP TABLE ssh_keys;`
	if err := tx.Exec(dropSQL).Error; err != nil {
		return err
	}

	// Rename the new table to the original table name
	renameSQL := `ALTER TABLE ssh_keys_temp RENAME TO ssh_keys;`
	return tx.Exec(renameSQL).Error
}
func v2_lowercaseEmails(tx *gorm.DB) error {
	// Copy the lowercase emails into the new column
	copySQL := `UPDATE users SET email = lower(email);`
	return tx.Exec(copySQL).Error
}
func v3_normalizedColumns(tx *gorm.DB) error {
	if err := tx.Model(&User{}).Where("username_normalized = '' OR username_normalized IS NULL").
		Updates(map[string]interface{}{"username_normalized": gorm.Expr("LOWER(username)")}).Error; err != nil {
		return err
	}
	return tx.Model(&Gist{}).Where("url_normalized = '' OR url_normalized IS NULL").
		Updates(map[string]interface{}{"url_normalized": gorm.Expr("LOWER(url)")}).Error
}

// v4_uniqueUsername turns the plain index on users.username_normalized into a
// unique one. Usernames are compared case-insensitively everywhere (login,
// registration, profile URLs), so two accounts whose names differ only by case
// must not exist; until now that was only enforced in the registration handler,
// which two simultaneous sign-ups can slip past.
//
// Older databases cannot be repaired automatically, so duplicates win: the
// migration keeps the plain index and logs the offending names instead of
// refusing to start the server.
func v4_uniqueUsername(tx *gorm.DB) error {
	var duplicates []string
	if err := tx.Raw(`SELECT username_normalized FROM users
		GROUP BY username_normalized HAVING COUNT(*) > 1`).Scan(&duplicates).Error; err != nil {
		return err
	}
	if len(duplicates) > 0 {
		log.Warn().Strs("usernames", duplicates).
			Msg("Duplicate usernames exist; keeping the non-unique index. Rename the accounts listed to enable the unique index.")
		return nil
	}

	if tx.Migrator().HasIndex(&User{}, "idx_users_username_normalized") {
		if err := tx.Migrator().DropIndex(&User{}, "idx_users_username_normalized"); err != nil {
			return err
		}
	}
	return tx.Exec("CREATE UNIQUE INDEX idx_users_username_normalized ON users (username_normalized)").Error
}

// v5_gistCommentsCount sets the comments counter of the gists created before
// comments existed, which AutoMigrate added as NULL, so that it can be
// incremented.
func v5_gistCommentsCount(tx *gorm.DB) error {
	return tx.Model(&Gist{}).Where("nb_comments IS NULL").
		UpdateColumn("nb_comments", 0).Error
}
