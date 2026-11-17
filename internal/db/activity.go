package db

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Activity is a per-user, per-day counter used by the contribution heatmaps.
// It is intentionally a tiny aggregate table rather than an event log: it is
// cheap to query for 365 days and survives upgrades without migrations beyond
// the automatic column creation.
type Activity struct {
	ID     uint   `gorm:"primaryKey"`
	UserID uint   `gorm:"uniqueIndex:idx_activity_user_day"`
	Day    string `gorm:"size:10;uniqueIndex:idx_activity_user_day"` // YYYY-MM-DD, server local time
	Count  int
}

// ActivityDay is the aggregate shape returned to handlers.
type ActivityDay struct {
	Day   string
	Count int
}

func activityDay(t time.Time) string {
	return t.Format("2006-01-02")
}

// IncrementActivity adds delta to the user's counter for the given day.
// Negative deltas are ignored; a zero delta is a no-op.
func IncrementActivity(userID uint, t time.Time, delta int) error {
	if userID == 0 || delta <= 0 {
		return nil
	}
	row := Activity{UserID: userID, Day: activityDay(t), Count: delta}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "day"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"count": gorm.Expr("count + ?", delta)}),
	}).Create(&row).Error
}

// SetActivityDay replaces a user's counter for a day (used by the backfill
// action, which recomputes values from git history).
func SetActivityDay(userID uint, day string, count int) error {
	if userID == 0 {
		return nil
	}
	row := Activity{UserID: userID, Day: day, Count: count}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "day"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"count": count}),
	}).Create(&row).Error
}

// GetUserActivity returns the daily counts for one user between from and to
// (inclusive, YYYY-MM-DD). A userID of 0 returns the global aggregate.
func GetUserActivity(userID uint, from, to string) ([]ActivityDay, error) {
	var rows []ActivityDay
	q := db.Model(&Activity{}).
		Select("day, SUM(count) as count").
		Where("day >= ? AND day <= ?", from, to).
		Group("day").
		Order("day")
	if userID != 0 {
		q = q.Where("user_id = ?", userID)
	}
	err := q.Scan(&rows).Error
	return rows, err
}

// GetActivityTotals returns the number of events recorded for a user, or for
// the whole instance when userID is 0.
func GetActivityTotals(userID uint) (int64, error) {
	var total int64
	q := db.Model(&Activity{}).Select("COALESCE(SUM(count), 0)")
	if userID != 0 {
		q = q.Where("user_id = ?", userID)
	}
	err := q.Scan(&total).Error
	return total, err
}

// ResetAllActivity clears every stored counter. Used before a full backfill
// from existing gists so deleted gists do not keep contributing.
func ResetAllActivity() error {
	return db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Activity{}).Error
}

// DeleteActivityForUser removes a deleted user's heatmap data.
func DeleteActivityForUser(userID uint) error {
	return db.Where("user_id = ?", userID).Delete(&Activity{}).Error
}
