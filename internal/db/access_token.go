package db

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

const (
	ScopeGist = iota
	ScopeUser
)

const (
	NoPermission = iota
	ReadPermission
	ReadWritePermission
)

const (
	// TokenPrefix marks a personal access token issued by QingyuGist.
	TokenPrefix = "qyg_"
	// legacyTokenPrefix was issued before the project was renamed. Such tokens
	// stay valid so existing integrations keep working.
	legacyTokenPrefix = "og_"
)

// IsAccessToken reports whether s looks like a personal access token, either
// the current format or the legacy one. It says nothing about validity.
func IsAccessToken(s string) bool {
	return strings.HasPrefix(s, TokenPrefix) || strings.HasPrefix(s, legacyTokenPrefix)
}

type AccessToken struct {
	ID         uint `gorm:"primaryKey"`
	Name       string
	TokenHash  string `gorm:"size:64;uniqueIndex"` // SHA-256 hash of the token, looked up on every API call
	CreatedAt  int64
	ExpiresAt  int64 // 0 means no expiration
	LastUsedAt int64
	UserID     uint `gorm:"index"`
	User       User `validate:"-"`

	ScopeGist uint // 0 = none, 1 = read, 2 = read+write
	ScopeUser uint // 0 = none, 1 = read, 2 = read+write
}

// GenerateToken creates a new random token and returns the plain text token.
// The token hash is stored in the AccessToken struct.
// The plain text token should be shown to the user once and never stored.
func (t *AccessToken) GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	plainToken := TokenPrefix + hex.EncodeToString(bytes)

	hash := sha256.Sum256([]byte(plainToken))
	t.TokenHash = hex.EncodeToString(hash[:])

	return plainToken, nil
}

func GetAccessTokenByID(tokenID uint) (*AccessToken, error) {
	token := new(AccessToken)
	err := db.
		Where("id = ?", tokenID).
		First(&token).Error
	return token, err
}

func GetAccessTokenByToken(plainToken string) (*AccessToken, error) {
	hash := sha256.Sum256([]byte(plainToken))
	tokenHash := hex.EncodeToString(hash[:])

	token := new(AccessToken)
	err := db.
		Preload("User").
		Where("token_hash = ?", tokenHash).
		First(&token).Error
	return token, err
}

func GetAccessTokensByUserID(userID uint) ([]*AccessToken, error) {
	var tokens []*AccessToken
	err := db.
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&tokens).Error
	return tokens, err
}

func (t *AccessToken) Create() error {
	t.CreatedAt = time.Now().Unix()
	return db.Create(t).Error
}

func (t *AccessToken) Delete() error {
	return db.Delete(t).Error
}

func (t *AccessToken) UpdateLastUsed() error {
	return db.Model(t).Update("last_used_at", time.Now().Unix()).Error
}

func (t *AccessToken) IsExpired() bool {
	if t.ExpiresAt == 0 {
		return false
	}
	return time.Now().Unix() > t.ExpiresAt
}

func (t *AccessToken) HasGistReadPermission() bool {
	return t.ScopeGist >= ReadPermission
}

func (t *AccessToken) HasUserReadPermission() bool {
	return t.ScopeUser >= ReadPermission
}

func (t *AccessToken) CheckForPermission(scope, permission uint) bool {
	if scope == ScopeGist {
		return t.ScopeGist >= permission
	}
	if scope == ScopeUser {
		return t.ScopeUser >= permission
	}
	return false
}

// -- DTO -- //

type AccessTokenDTO struct {
	Name      string `form:"name" validate:"required,max=255"`
	ScopeGist uint   `form:"scope_gist" validate:"min=0,max=2"`
	ScopeUser uint   `form:"scope_user" validate:"min=0,max=2"`
	ExpiresAt string `form:"expires_at"` // empty means no expiration, otherwise date format (YYYY-MM-DD)
}

func (dto *AccessTokenDTO) ToAccessToken() *AccessToken {
	var expiresAt int64
	if dto.ExpiresAt != "" {
		// date input format: 2006-01-02, expires at end of day
		if t, err := time.ParseInLocation("2006-01-02", dto.ExpiresAt, time.Local); err == nil {
			expiresAt = t.Add(24*time.Hour - time.Second).Unix()
		}
	}

	return &AccessToken{
		Name:      dto.Name,
		ScopeGist: dto.ScopeGist,
		ScopeUser: dto.ScopeUser,
		ExpiresAt: expiresAt,
	}
}

// SaveAccessTokenForTest is exported for tests only; saves the entire AccessToken row.
func SaveAccessTokenForTest(t *AccessToken) error {
	return db.Save(t).Error
}
