package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents the users table
type User struct {
	ID               uint      `gorm:"primaryKey;autoIncrement"`
	Name             string    `gorm:"type:varchar(255);not null"`
	Email            string    `gorm:"type:varchar(255);unique;not null"`
	PasswordHash     string    `gorm:"type:varchar(255);not null"`
	Role             string    `gorm:"type:varchar(50);not null"`
	RegistrationDate time.Time `gorm:"not null;default:current_timestamp"`
	PhotoURL         *string   `gorm:"type:varchar(255)"`
	Status           string    `gorm:"type:varchar(50);not null"`
}

// Author represents the authors table
type Author struct {
	ID        uint    `gorm:"primaryKey;autoIncrement"`
	Name      string  `gorm:"type:varchar(255);not null"`
	BirthYear *int    `gorm:"type:integer"`
	Country   *string `gorm:"type:varchar(100)"`
	Biography *string `gorm:"type:text"`
}

// Genre represents the genres table
type Genre struct {
	ID          uint    `gorm:"primaryKey;autoIncrement"`
	Name        string  `gorm:"type:varchar(100);not null"`
	Description *string `gorm:"type:text"`
}

// Book represents the books table
type Book struct {
	ID              uint      `gorm:"primaryKey;autoIncrement"`
	Title           string    `gorm:"type:varchar(255);not null"`
	Description     *string   `gorm:"type:text"`
	PublicationYear *int      `gorm:"type:integer"`
	ISBN            string    `gorm:"type:varchar(13);unique;not null"`
	GenreID         uint      `gorm:"not null"`
	Genre           Genre     `gorm:"foreignKey:GenreID;references:ID;constraint:OnDelete:RESTRICT"`
	TotalCopies     int       `gorm:"not null"`
	AvailableCopies int       `gorm:"not null"`
	CoverURL        *string   `gorm:"type:varchar(255)"`
	AddedDate       time.Time `gorm:"not null;default:current_timestamp"`
}

// BookAuthor represents the book_authors junction table
type BookAuthor struct {
	BookID   uint   `gorm:"primaryKey"`
	Book     Book   `gorm:"foreignKey:BookID;references:ID;constraint:OnDelete:CASCADE"`
	AuthorID uint   `gorm:"primaryKey"`
	Author   Author `gorm:"foreignKey:AuthorID;references:ID;constraint:OnDelete:CASCADE"`
}

// Order represents the orders table
type Order struct {
	ID         uint       `gorm:"primaryKey;autoIncrement"`
	UserID     uint       `gorm:"not null"`
	User       User       `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:RESTRICT"`
	BookID     uint       `gorm:"not null"`
	Book       Book       `gorm:"foreignKey:BookID;references:ID;constraint:OnDelete:RESTRICT"`
	OrderDate  time.Time  `gorm:"not null;default:current_timestamp"`
	DueDate    time.Time  `gorm:"not null"`
	ReturnDate *time.Time `gorm:"type:timestamp"`
	Status     string     `gorm:"type:varchar(50);not null"`
}

// Review represents a book review
type Review struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	UserID    uint      `gorm:"not null"`
	User      User      `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:RESTRICT"`
	BookID    uint      `gorm:"not null"`
	Book      Book      `gorm:"foreignKey:BookID;references:ID;constraint:OnDelete:RESTRICT"`
	Rating    int       `gorm:"not null;check:rating >= 1 AND rating <= 5"`
	Comment   *string   `gorm:"type:text"`
	CreatedAt time.Time `gorm:"not null;default:current_timestamp"`
}

func MigrateSchema(db *gorm.DB) error {
	// AutoMigrate all tables
	err := db.AutoMigrate(&User{}, &Author{}, &Genre{}, &Book{}, &BookAuthor{}, &Order{}, &Review{})
	if err != nil {
		return err
	}

	// Add CHECK constraints for User.Role
	err = db.Exec(`ALTER TABLE users ADD CONSTRAINT check_role CHECK (role IN ('admin', 'user'))`).Error
	if err != nil {
		return err
	}

	// Add CHECK constraints for User.Status
	err = db.Exec(`ALTER TABLE users ADD CONSTRAINT check_status CHECK (status IN ('active', 'blocked'))`).Error
	if err != nil {
		return err
	}

	// Add CHECK constraints for Order.Status
	err = db.Exec(`ALTER TABLE orders ADD CONSTRAINT check_order_status CHECK (status IN ('issued', 'returned', 'overdue'))`).Error
	if err != nil {
		return err
	}

	// Add CHECK constraint for Review.Rating
	err = db.Exec(`ALTER TABLE reviews ADD CONSTRAINT check_rating CHECK (rating >= 1 AND rating <= 5)`).Error
	if err != nil {
		return err
	}

	return nil
}
