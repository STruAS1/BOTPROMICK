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
	Secret       string `gorm:"size:12;not null;uniqueIndex"`
	ByUserID     uint64 `gorm:"not null"`
	ForNetworkID uint64 `gorm:"not null"`
	UsedByUserID uint64 `gorm:"default:0"`
	Nonce        int64  `gorm:"not null;uniqueIndex"`
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

func generateSecretKey(networkID, byUserID uint32, nonce int64) string {
	data := fmt.Sprintf("%d:%d:%d", networkID, byUserID, nonce)

	h := hmac.New(sha256.New, []byte(data))
	return hex.EncodeToString(h.Sum(nil)[:6])
}
func (n *Network) CreateInvite(db *gorm.DB, byUserID uint32) (string, error) {
	networkID := uint32(n.ID)
	nonce := time.Now().UnixNano()

	secretKey := generateSecretKey(networkID, byUserID, nonce)

	h := hmac.New(sha256.New, []byte(secretKey))
	data := fmt.Sprintf("%d:%d:%d:%s", networkID, byUserID, nonce, secretKey)
	h.Write([]byte(data))
	hashed := h.Sum(nil)[:12]

	invite := Invite{
		Secret:       secretKey,
		ByUserID:     uint64(byUserID),
		ForNetworkID: uint64(n.ID),
		Nonce:        nonce,
	}
	if err := db.Create(&invite).Error; err != nil {
		return "", err
	}

	inviteID := uint32(invite.ID) + 1_000_000_000
	idBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(idBytes, inviteID)
	inviteStr := hex.EncodeToString(append(idBytes, hashed...))

	return inviteStr, nil
}

func (n *Network) NewUser(db *gorm.DB, u *User, confirmed bool) error {
	var countOfOwner int64
	if err := db.Model(&Network{}).Where("owner_id = ?", u.ID).Count(&countOfOwner).Error; err != nil {
		return err
	}
	if countOfOwner != 0 {
		return errors.New("❌ Вы не можете вступить в сеть")
	}
	if u.UserNetwork != nil && u.UserNetwork.NetworkID == n.ID {
		return errors.New("🫥 Вы уже в этой сети")
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
		return errors.New("Владелец сети не может ее покинуть 🤷")
	}

	if u.UserNetwork == nil || u.UserNetwork.NetworkID != n.ID {
		return errors.New("🫥 Вы не состоите в данной сети")
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
	if err != nil || len(bytes) != 16 {
		return errors.New("❌ некорректное приглашение")
	}

	inviteIDPlusBillion := binary.BigEndian.Uint32(bytes[:4])
	if inviteIDPlusBillion < 1_000_000_000 {
		return errors.New("❌ Некорректный ID приглашения")
	}
	inviteID := inviteIDPlusBillion - 1_000_000_000

	var invite Invite
	if err := db.Where("id = ? AND used_by_user_id = 0", inviteID).First(&invite).Error; err != nil {
		return errors.New("🫥 Приглашение не найдено или уже использовано")
	}

	if time.Since(invite.CreatedAt) > 24*time.Hour {
		return errors.New("❌ Приглашение истекло")
	}

	networkID := uint32(invite.ForNetworkID)
	byUserID := uint32(invite.ByUserID)
	nonce := invite.Nonce

	expectedSecretKey := generateSecretKey(networkID, byUserID, nonce)

	h := hmac.New(sha256.New, []byte(expectedSecretKey))
	data := fmt.Sprintf("%d:%d:%d:%s", networkID, byUserID, nonce, expectedSecretKey)
	h.Write([]byte(data))
	expectedHash := h.Sum(nil)[:12]

	if !hmac.Equal(expectedHash, bytes[4:]) {
		return errors.New("❌ Невалидный хеш приглашения")
	}

	invite.UsedByUserID = uint64(u.ID)

	network := GetNetworkById(db, uint(networkID))
	if network == nil {
		return errors.New("🦋 Неизвестная ошибка")
	}

	if err := network.NewUser(db, u, true); err != nil {
		return err
	}

	if err := db.Save(&invite).Error; err != nil {
		return err
	}

	return nil
}
