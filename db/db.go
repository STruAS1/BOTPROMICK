package db

import (
	"BOTPROMICK/config"
	"BOTPROMICK/db/models/product"
	"BOTPROMICK/db/models/user"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect(cfg *config.Config) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		cfg.Database.Host, cfg.Database.User, cfg.Database.Password, cfg.Database.Name, cfg.Database.Port)
	var err error

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Connected to database.")

	if err := DB.AutoMigrate(&user.User{}, &user.Network{}, &user.UserNetwork{}, &user.Invite{}, &product.Product{}, &product.Sale{}, &product.InputProduct{}, &product.InputSale{}, &product.Photo{}); err != nil {
		log.Fatalf("Error creating tables: %v", err)
	} else {
		log.Println("Tables created successfully.")
	}
}
