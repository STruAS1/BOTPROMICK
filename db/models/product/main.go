package product

import (
	"BOTPROMICK/db/models/user"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID                uint `gorm:"primaryKey"`
	Title             string
	Description       string
	PhotosCount       uint
	StartLink         string
	UserSubID         bool
	EndLink           string
	Prize             uint
	NetworkOwnerPrize uint
	Status            bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
	InputProducts     []InputProduct `gorm:"foreignKey:ProductId"`
}

type InputProduct struct {
	ID        uint `gorm:"primaryKey"`
	ProductId uint
	Title     string
	Optional  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Sale struct {
	ID            uint `gorm:"primaryKey"`
	UserID        uint
	ProductID     uint
	Status        uint
	NetworkID     uint
	OfferResponse string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Photos        []Photo     `gorm:"foreignKey:SaleID"`
	Product       Product     `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE;"`
	InputSales    []InputSale `gorm:"foreignKey:SaleID"`
}

type InputSale struct {
	ID        uint `gorm:"primaryKey"`
	SaleID    uint
	Title     string
	Optional  bool
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Photo struct {
	ID        uint `gorm:"primaryKey"`
	File_ID   string
	SaleID    uint
	CreatedAt time.Time
	UpdatedAt time.Time
}

func GetCounOfSelles(db *gorm.DB, un *user.UserNetwork) (int64, int64) {
	var CounfOfMySalles int64
	var CounfOfNetSalles int64
	timeLimit := time.Now().Add(-24 * time.Hour)
	db.Model(&Sale{}).
		Where("network_id = ? AND user_id = ? AND created_at >= ? AND status > 0 AND status != 2", un.NetworkID, un.UserID, timeLimit).
		Count(&CounfOfMySalles)

	db.Model(&Sale{}).
		Where("network_id = ? AND created_at >= ? AND status > 0 AND status != 2", un.NetworkID, timeLimit).
		Count(&CounfOfNetSalles)
	return CounfOfMySalles, CounfOfNetSalles
}

func GetProducts(db *gorm.DB) ([]Product, error) {
	var Products []Product
	if err := db.Where(&Product{Status: true}).Find(&Products).Error; err != nil {
		return nil, err
	}
	if len(Products) == 0 {
		return nil, errors.New("❌ Нет доступных продуктов")
	}
	return Products, nil
}
func GetProductBtID(db gorm.DB, ID uint) (Product, error) {
	var p Product
	if err := db.Where(&Product{Status: true, ID: ID}).First(&p).Error; err != nil {
		fmt.Print(err)
		fmt.Print(ID)
		return Product{}, err
	}
	return p, nil
}

func (p *Product) NewSale(db *gorm.DB, u *user.UserNetwork) (*Sale, error) {
	New := Sale{
		UserID:    u.UserID,
		ProductID: p.ID,
		NetworkID: u.NetworkID,
		Product:   *p,
	}

	if err := db.Create(&New).Error; err != nil {
		return nil, err
	}

	var inputTemplates []InputProduct
	if err := db.Where("product_id = ?", p.ID).Find(&inputTemplates).Error; err != nil {
		return nil, err
	}

	var saleInputs []InputSale
	for _, input := range inputTemplates {
		saleInputs = append(saleInputs, InputSale{
			SaleID:   New.ID,
			Title:    input.Title,
			Optional: input.Optional,
			Value:    "",
		})
	}

	if len(saleInputs) > 0 {
		if err := db.Create(&saleInputs).Error; err != nil {
			return nil, err
		}
	}

	db.Preload("Product").Preload("InputSales").First(&New, New.ID)

	go func(saleID uint) {
		time.Sleep(30 * time.Minute)
		var sale Sale
		if err := db.First(&sale, saleID).Error; err != nil {
			return
		}
		if sale.Status == 0 {
			db.Model(&sale).Update("Status", 2)
		}
	}(New.ID)

	return &New, nil
}

func (s *Sale) GetLink() string {
	SubIDBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(SubIDBytes, uint64(s.ID+12297829382473034411))
	SubId := hex.EncodeToString(SubIDBytes)
	link := s.Product.StartLink
	if s.Product.UserSubID {
		link += SubId + s.Product.EndLink
	} else {
		link += s.Product.EndLink
	}
	return link
}
func (s *Sale) AddInputValue(db *gorm.DB, i uint, value string) error {
	if len(s.InputSales) <= int(i) {
		return errors.New("Поле не найдено!")
	}
	s.InputSales[i].Value = value
	fmt.Print(s.InputSales[i])
	return db.Save(&s.InputSales[i]).Error

}

func (s *Sale) AddPhoto(db *gorm.DB, fileID string) error {
	if s.Status > 1 {
		return errors.New("🎟️ Данная продажа недоступна")
	}
	photo := Photo{
		File_ID: fileID,
		SaleID:  s.ID,
	}
	if err := db.Create(&photo).Error; err != nil {
		return err
	}
	s.Photos = append(s.Photos, photo)
	return nil
}

func (s *Sale) RemovePhoto(db *gorm.DB, id uint) error {
	if s.Status > 1 {
		return errors.New("🎟️ Данная продажа недоступна")
	}
	for i, photo := range s.Photos {
		if photo.ID == id {
			if err := db.Delete(&photo).Error; err != nil {
				return err
			}
			s.Photos = append(s.Photos[:i], s.Photos[i+1:]...)
			return nil
		}
	}
	return errors.New("📸 Фото не найдено")
}

func (s *Sale) Confirm(db *gorm.DB) error {
	if len(s.Photos) < int(s.Product.PhotosCount) {
		return errors.New("🫥 Недостаточно фотографий")
	}
	if s.Status > 1 {
		return errors.New("🎟️ Данная продажа недоступна")
	}

	for _, input := range s.InputSales {
		if !input.Optional && input.Value == "" {
			return errors.New("Заполните обязательное поле: " + input.Title)
		}
	}

	s.Status = 1
	return db.Save(s).Error
}

func (s *Sale) Cancel(db *gorm.DB) error {
	s.Status = 2
	if db.Save(s).Error != nil {
		return errors.New("❌ Не удалось сохранить продажу")
	}
	return nil
}
