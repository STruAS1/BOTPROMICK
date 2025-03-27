package user

import (
	"errors"
	"sync"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID          uint   `gorm:"primaryKey"`
	TelegramID  int64  `gorm:"uniqueIndex"`
	Balance     uint64 `gorm:"index;not null;default:0"`
	Username    string `gorm:"size:100"`
	FirstName   string `gorm:"size:100"`
	LastName    string `gorm:"size:100"`
	Registered  bool
	FullName    string
	Experience  int
	Phone       string
	City        string
	Lat         float64
	Lon         float64
	UserNetwork *UserNetwork `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type UserNetwork struct {
	ID              uint `gorm:"primaryKey"`
	UserID          uint `gorm:"uniqueIndex;not null"`
	NetworkID       uint
	Confirmed       bool
	CanSell         bool
	CanInviteUser   bool
	CanEditUser     bool
	CanEditNetwork  bool
	CanViewAllSales bool `gorm:"column:CanViewAllSales"`
}

func (User) TableName() string {
	return "Users"
}

func (UserNetwork) TableName() string {
	return "UserNetworks"
}

var (
	users   = make(map[int64]*User)
	usersMu sync.RWMutex
)

func (u *User) RegisterData(db *gorm.DB, FullName, Phone string, Experience int, lat, lon float64) {
	u.FullName = FullName
	u.Phone = Phone
	u.Experience = Experience
	u.Registered = true
	u.Lat = lat
	u.Lon = lon
	db.Save(u)
}

func GetUser(db *gorm.DB, TelegramID int64, Username, FirstName, LastName string) (*User, error) {
	usersMu.RLock()
	if u, ok := users[TelegramID]; ok {
		usersMu.RUnlock()
		return u, nil
	}
	usersMu.RUnlock()

	usersMu.Lock()
	defer usersMu.Unlock()

	if u, ok := users[TelegramID]; ok {
		return u, nil
	}

	var u User
	err := db.Preload("UserNetwork").Where(&User{TelegramID: TelegramID}).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			u = User{
				TelegramID: TelegramID,
				Username:   Username,
				FirstName:  FirstName,
				LastName:   LastName,
			}
			if err := db.Create(&u).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	newUser := &u
	users[TelegramID] = newUser
	return newUser, nil
}
