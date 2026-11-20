package db

import (
	"encoding/json"
	"time"

	"gorm.io/gorm/clause"
)

// Notification is a message shown to users in the header bell. A row with
// UserID 0 is a broadcast that everybody sees; a row with a UserID belongs to
// that account only. Read state is tracked per user so a broadcast costs one
// row no matter how many accounts exist.
type Notification struct {
	ID   uint   `gorm:"primaryKey"`
	Kind string `gorm:"size:16;index"` // broadcast | moderation | system
	// UserID 0 means "every logged-in user".
	UserID uint `gorm:"index"`
	// System notifications carry i18n keys plus positional arguments so each
	// viewer reads them in their own language. Broadcasts carry plain text
	// typed by an administrator and ignore these fields.
	TitleKey string `gorm:"size:120"`
	BodyKey  string `gorm:"size:120"`
	ArgsJSON string `gorm:"type:text"`
	Title    string `gorm:"size:200"`
	Body     string `gorm:"type:text"`
	Link     string `gorm:"size:300"`
	// CreatedAt is a unix timestamp rather than GORM's time.Time so the admin
	// list can render it the same way as every other timestamp in the app.
	CreatedAt int64 `gorm:"index"`
}

// NotificationRead marks one notification as read for one user.
type NotificationRead struct {
	UserID         uint `gorm:"primaryKey"`
	NotificationID uint `gorm:"primaryKey;index"` // deleting a broadcast clears its reads by notification id
	ReadAt         int64
}

func (Notification) TableName() string     { return "notifications" }
func (NotificationRead) TableName() string { return "notification_reads" }

// NotificationArgs decodes the stored positional arguments.
func (n *Notification) NotificationArgs() []any {
	if n.ArgsJSON == "" {
		return nil
	}
	var parts []string
	if err := json.Unmarshal([]byte(n.ArgsJSON), &parts); err != nil {
		return nil
	}
	args := make([]any, len(parts))
	for i, p := range parts {
		args[i] = p
	}
	return args
}

// CreateNotification stores a notification for a single user.
func CreateNotification(kind string, userID uint, titleKey, bodyKey string, args []string, link string) error {
	encoded := ""
	if len(args) > 0 {
		raw, err := json.Marshal(args)
		if err != nil {
			return err
		}
		encoded = string(raw)
	}
	return db.Create(&Notification{
		Kind:      kind,
		UserID:    userID,
		TitleKey:  titleKey,
		BodyKey:   bodyKey,
		ArgsJSON:  encoded,
		Link:      link,
		CreatedAt: time.Now().Unix(),
	}).Error
}

// CreateBroadcast stores a site-wide announcement for every user.
func CreateBroadcast(title, body, link string) (*Notification, error) {
	n := &Notification{
		Kind:      "broadcast",
		UserID:    0,
		Title:     title,
		Body:      body,
		Link:      link,
		CreatedAt: time.Now().Unix(),
	}
	return n, db.Create(n).Error
}

// GetVisibleNotifications returns the notifications a user can see, newest
// first, with their read timestamp attached (0 when unread).
type UserNotification struct {
	Notification
	ReadAt int64
}

func GetVisibleNotifications(userID uint, limit, offset int) ([]*UserNotification, error) {
	rows := make([]*UserNotification, 0, limit)
	err := db.Table("notifications AS n").
		Select("n.*, COALESCE(r.read_at, 0) AS read_at").
		Joins("LEFT JOIN notification_reads AS r ON r.notification_id = n.id AND r.user_id = ?", userID).
		Where("n.user_id = 0 OR n.user_id = ?", userID).
		Order("n.created_at DESC, n.id DESC").
		Limit(limit).Offset(offset).
		Scan(&rows).Error
	return rows, err
}

// CountUnreadNotifications counts what the bell badge shows.
func CountUnreadNotifications(userID uint) (int64, error) {
	var count int64
	err := db.Table("notifications AS n").
		Where("(n.user_id = 0 OR n.user_id = ?)", userID).
		Where("NOT EXISTS (SELECT 1 FROM notification_reads AS r WHERE r.notification_id = n.id AND r.user_id = ?)", userID).
		Count(&count).Error
	return count, err
}

// MarkNotificationRead records that a user opened one notification. Visible
// notifications only: a user must not be able to mark somebody else's row.
func MarkNotificationRead(userID, notificationID uint) error {
	var count int64
	if err := db.Model(&Notification{}).
		Where("id = ? AND (user_id = 0 OR user_id = ?)", notificationID, userID).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return nil
	}
	// GORM translates this upsert into the ON CONFLICT / ON DUPLICATE KEY
	// syntax the active dialect understands.
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "notification_id"}},
		DoUpdates: clause.Assignments(map[string]any{"read_at": time.Now().Unix()}),
	}).Create(&NotificationRead{
		UserID:         userID,
		NotificationID: notificationID,
		ReadAt:         time.Now().Unix(),
	}).Error
}

// MarkAllNotificationsRead marks everything currently visible as read, in
// batches small enough to stay under every driver's parameter limit.
func MarkAllNotificationsRead(userID uint) error {
	var ids []uint
	if err := db.Table("notifications AS n").
		Where("(n.user_id = 0 OR n.user_id = ?)", userID).
		Where("NOT EXISTS (SELECT 1 FROM notification_reads AS r WHERE r.notification_id = n.id AND r.user_id = ?)", userID).
		Pluck("n.id", &ids).Error; err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}

	now := time.Now().Unix()
	const batch = 200
	for start := 0; start < len(ids); start += batch {
		end := min(start+batch, len(ids))
		rows := make([]NotificationRead, 0, end-start)
		for _, id := range ids[start:end] {
			rows = append(rows, NotificationRead{UserID: userID, NotificationID: id, ReadAt: now})
		}
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error; err != nil {
			return err
		}
	}
	return nil
}

// GetBroadcastNotifications lists what administrators have sent, newest first,
// for the admin page.
func GetBroadcastNotifications(limit int) ([]*Notification, error) {
	rows := make([]*Notification, 0, limit)
	err := db.Where("kind = ?", "broadcast").Order("created_at DESC, id DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

// DeleteNotification removes a broadcast (read records cascade by hand).
func DeleteNotification(id uint) error {
	if err := db.Where("notification_id = ?", id).Delete(&NotificationRead{}).Error; err != nil {
		return err
	}
	return db.Where("id = ? AND kind = ?", id, "broadcast").Delete(&Notification{}).Error
}
