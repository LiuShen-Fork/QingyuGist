package db

import (
	"gorm.io/gorm"
)

// ModerationReview is one attempt at checking a gist, kept so administrators
// can see what the provider was asked and what it answered. Gists hold only
// the latest state; the review log keeps the history, including failures that
// were retried later.
type ModerationReview struct {
	ID uint `gorm:"primaryKey"`
	// Result is "approved", "rejected", "error" or "deferred" (daily quota
	// reached, so the gist was left pending for a later attempt).
	Result string `gorm:"size:16;index"`
	// GistID keeps the row useful after the gist itself is deleted; the
	// snapshot below is what the admin page shows.
	GistID      uint   `gorm:"index"`
	UserID      uint   `gorm:"index"`
	Username    string `gorm:"size:64"`
	GistURL     string `gorm:"size:255"`
	GistTitle   string `gorm:"size:255"`
	Reason      string `gorm:"type:text"`
	Model       string `gorm:"size:120"`
	LatencyMs   int
	ProviderID  string `gorm:"size:120"`
	AttemptedAt int64  `gorm:"index"`
}

// TableName pins the table name so a future rename of the struct does not
// silently orphan the rows.
func (ModerationReview) TableName() string { return "moderation_reviews" }

// RecordModerationReview stores one attempt.
func RecordModerationReview(r *ModerationReview) error {
	return db.Create(r).Error
}

// GetModerationReviews returns the most recent reviews first.
func GetModerationReviews(limit, offset int) ([]*ModerationReview, error) {
	var reviews []*ModerationReview
	err := db.Order("attempted_at DESC, id DESC").Limit(limit).Offset(offset).Find(&reviews).Error
	return reviews, err
}

// CountModerationReviews returns how many attempts were logged in total.
func CountModerationReviews() (int64, error) {
	var count int64
	err := db.Model(&ModerationReview{}).Count(&count).Error
	return count, err
}

// CountModerationReviewsSince counts provider calls made at or after the given
// unix timestamp. Only real calls count against the daily quota: rows written
// for a deferred attempt never reached the provider.
func CountModerationReviewsSince(ts int64) (int64, error) {
	var count int64
	err := db.Model(&ModerationReview{}).
		Where("attempted_at >= ? AND result <> ?", ts, "deferred").
		Count(&count).Error
	return count, err
}

// CountModerationReviewsForGistSince counts the attempts logged for one gist at
// or after the given unix timestamp, optionally filtered by result.
func CountModerationReviewsForGistSince(gistID uint, result string, ts int64) (int64, error) {
	var count int64
	q := db.Model(&ModerationReview{}).Where("gist_id = ? AND attempted_at >= ?", gistID, ts)
	if result != "" {
		q = q.Where("result = ?", result)
	}
	err := q.Count(&count).Error
	return count, err
}

// GetModerationReviewStats summarises the log for the admin dashboard.
func GetModerationReviewStats(since int64) (map[string]int64, error) {
	type row struct {
		Result string
		Count  int64
	}
	rows := make([]row, 0, 4)
	err := db.Model(&ModerationReview{}).
		Select("result, count(*) as count").
		Where("attempted_at >= ?", since).
		Group("result").
		Scan(&rows).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	stats := make(map[string]int64, len(rows))
	for _, r := range rows {
		stats[r.Result] = r.Count
	}
	return stats, nil
}
