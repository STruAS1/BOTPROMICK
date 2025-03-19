package user

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

type Network struct {
	ID          uint `gorm:"primaryKey"`
	Title       string
	OwnerID     uint  `gorm:"uniqueIndex"`
	CountOfUser int64 `gorm:"-"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
type Invite struct {
	ID           uint64 `gorm:"primaryKey"`
	Secret       string
	ByUserID     uint64
	ForNetworkID uint64
	UsedByUserID uint64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

var (
	networks   = make(map[uint]*Network)
	networksMu sync.RWMutex
)

func GetNetworkById(DB *gorm.DB, ID uint) *Network {
	networksMu.RLock()
	if n, ok := networks[ID]; ok {
		networksMu.RUnlock()
		return n
	}
	networksMu.RUnlock()

	networksMu.Lock()
	defer networksMu.Unlock()

	if n, ok := networks[ID]; ok {
		return n
	}

	var n *Network
	err := DB.Where(&Network{ID: ID}).First(&n).Error
	if err != nil {
		return nil
	}
	DB.Model(&UserNetwork{}).Where(&UserNetwork{NetworkID: n.ID}).Count(&n.CountOfUser)
	networks[ID] = n
	return n
}

func (un *UserNetwork) Network(DB *gorm.DB) *Network {
	networksMu.RLock()
	if n, ok := networks[un.NetworkID]; ok {
		networksMu.RUnlock()
		return n
	}
	networksMu.RUnlock()

	networksMu.Lock()
	defer networksMu.Unlock()

	if n, ok := networks[un.NetworkID]; ok {
		return n
	}

	var n *Network
	err := DB.Where(&Network{ID: un.NetworkID}).First(&n).Error
	if err != nil {
		return nil
	}
	DB.Model(&UserNetwork{}).Where(&UserNetwork{NetworkID: n.ID}).Count(&n.CountOfUser)
	networks[un.NetworkID] = n
	return n
}

func (Network) TableName() string {
	return "Networks"
}

func (n *Network) GetAllUsers(DB *gorm.DB, confirmed bool) ([]struct {
	UserID   uint
	Username string
	FullName string
	TgID     int64
	IsOwner  bool
}, error) {
	var usersData []struct {
		UserID   uint
		Username string
		FullName string
		TgID     int64
		IsOwner  bool
	}
	err := DB.Raw(`
		SELECT u.id AS "UserID", u.username AS "Username", u.full_name AS "FullName", u.telegram_id AS "TgID",
		       CASE WHEN n.owner_id = un.user_id THEN TRUE ELSE FALSE END AS "IsOwner"
		FROM "Users" u
		JOIN "UserNetworks" un ON u.id = un.user_id
		JOIN "Networks" n ON un.network_id = n.id
		WHERE un.network_id = ? AND un.confirmed = ?
	`, n.ID, confirmed).Scan(&usersData).Error

	if err != nil {
		return nil, err
	}
	return usersData, nil
}

func generateSecretKey(networkID, byUserID uint64) string {
	nonce := time.Now().UnixNano()
	data := fmt.Sprintf("%d:%d:%d", networkID, byUserID, nonce)

	h := hmac.New(sha256.New, []byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func (n *Network) CreateInvite(db *gorm.DB, byUserID uint64) (string, error) {

	secretKey := generateSecretKey(uint64(n.ID), byUserID)

	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(fmt.Sprintf("%d:%d:%s", uint64(n.ID), byUserID, secretKey)))
	hashed := h.Sum(nil)

	invite := Invite{
		Secret:       secretKey,
		ByUserID:     byUserID,
		ForNetworkID: uint64(n.ID),
	}
	if err := db.Create(&invite).Error; err != nil {
		return "", err
	}

	idBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(idBytes, invite.ID)

	inviteStr := hex.EncodeToString(append(idBytes, hashed...))

	return inviteStr, nil
}

func (n *Network) NewUser(db *gorm.DB, u *User, confirmed bool) error {
	var countOfOwner int64
	if err := db.Model(&Network{}).Where("owner_id = ?", u.ID).Count(&countOfOwner).Error; err != nil {
		return err
	}
	if countOfOwner != 0 {
		return errors.New("вы не можете вступить в сеть")
	}
	if u.UserNetwork != nil && u.UserNetwork.NetworkID == n.ID {
		return errors.New("вы уже в этой сети")
	}
	if u.UserNetwork != nil {
		if err := db.Where("user_id = ?", u.ID).Delete(&UserNetwork{}).Error; err != nil {
			return err
		}
	}
	netUserNetwork := UserNetwork{
		UserID:    u.ID,
		NetworkID: n.ID,
		CanSell:   true,
		Confirmed: confirmed,
	}

	if err := db.Create(&netUserNetwork).Error; err != nil {
		return err
	}
	n.CountOfUser++
	u.UserNetwork = &netUserNetwork
	return nil
}

func (n *Network) RemoveUser(db *gorm.DB, u *User, BotAPI *tgbotapi.BotAPI, message string) error {
	if n.OwnerID == u.ID {
		return errors.New("владелец сети не может покинуть её")
	}

	if u.UserNetwork == nil || u.UserNetwork.NetworkID != n.ID {
		return errors.New("вы не состоите в этой сети")
	}

	if err := db.Where("user_id = ? AND network_id = ?", u.ID, n.ID).Delete(&UserNetwork{}).Error; err != nil {
		return err
	}

	u.UserNetwork = nil
	if n.CountOfUser > 0 {
		n.CountOfUser--
	}
	BotAPI.Send(tgbotapi.NewMessage(u.TelegramID, message))
	return nil
}

func (u *User) UseInvite(db *gorm.DB, inviteStr string) error {
	bytes, err := hex.DecodeString(inviteStr)
	if err != nil || len(bytes) < 40 {
		return errors.New("некорректное приглашение")
	}
	inviteID := binary.BigEndian.Uint64(bytes[:8])
	var invite Invite
	if err := db.Where("id = ? AND used_by_user_id = 0", inviteID).First(&invite).Error; err != nil {
		return errors.New("приглашение не найдено или уже использовано")
	}
	expectedSecretKey := generateSecretKey(invite.ForNetworkID, invite.ByUserID)
	h := hmac.New(sha256.New, []byte(expectedSecretKey))
	h.Write([]byte(fmt.Sprintf("%d:%d:%s", invite.ForNetworkID, invite.ByUserID, expectedSecretKey)))
	expectedHash := h.Sum(nil)
	if !hmac.Equal(expectedHash, bytes[8:]) {
		return errors.New("невалидный хеш приглашения")
	}
	invite.UsedByUserID = uint64(u.ID)
	network := u.UserNetwork.Network(db)
	if network == nil {
		return errors.New("неизвестная ошибка")
	}
	if err := network.NewUser(db, u, true); err != nil {
		return err
	}
	if err := db.Save(&invite).Error; err != nil {
		return err
	}

	return nil
}
