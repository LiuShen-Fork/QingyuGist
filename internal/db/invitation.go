package db

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type Invitation struct {
	ID        uint   `gorm:"primaryKey"`
	Code      string `gorm:"size:64;uniqueIndex"` // invitations are redeemed by code
	ExpiresAt int64
	NbUsed    uint
	NbMax     uint
}

func GetAllInvitations() ([]*Invitation, error) {
	var invitations []*Invitation
	dialect := db.Name()
	query := db.Model(&Invitation{})

	switch dialect {
	case "sqlite":
		query = query.Order("(((expires_at >= strftime('%s', 'now')) AND ((nb_max <= 0) OR (nb_used < nb_max)))) DESC")
	case "postgres":
		query = query.Order("(((expires_at >= EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)) AND ((nb_max <= 0) OR (nb_used < nb_max)))) DESC")
	case "mysql":
		query = query.Order("(((expires_at >= UNIX_TIMESTAMP()) AND ((nb_max <= 0) OR (nb_used < nb_max)))) DESC")
	default:
		return nil, fmt.Errorf("unsupported database dialect: %s", dialect)
	}

	err := query.Order("id ASC").Find(&invitations).Error

	return invitations, err
}

func GetInvitationByID(id uint) (*Invitation, error) {
	invitation := new(Invitation)
	err := db.
		Where("id = ?", id).
		First(&invitation).Error
	return invitation, err
}

func GetInvitationByCode(code string) (*Invitation, error) {
	invitation := new(Invitation)
	err := db.
		Where("code = ?", code).
		First(&invitation).Error
	return invitation, err
}

func InvitationCodeExists(code string) (bool, error) {
	var count int64
	err := db.Model(&Invitation{}).Where("code = ?", code).Count(&count).Error
	return count > 0, err
}

func (i *Invitation) Create() error {
	code, err := generateRandomCode()
	if err != nil {
		return err
	}
	i.Code = code
	return db.Create(&i).Error
}

func (i *Invitation) Update() error {
	return db.Save(&i).Error
}

func (i *Invitation) Delete() error {
	return db.Delete(&i).Error
}

func (i *Invitation) IsExpired() bool {
	return i.ExpiresAt < time.Now().Unix()
}

func (i *Invitation) IsMaxedOut() bool {
	return i.NbMax > 0 && i.NbUsed >= i.NbMax
}

func (i *Invitation) IsUsable() bool {
	return !i.IsExpired() && !i.IsMaxedOut()
}

func (i *Invitation) Use() error {
	i.NbUsed++
	return i.Update()
}

// generateRandomCode returns 16 hex characters from crypto/rand. Invitation
// codes are the only gate on a closed instance, so they must not be
// predictable from the time they were created.
func generateRandomCode() (string, error) {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return strings.ToUpper(hex.EncodeToString(buf)), nil
}
