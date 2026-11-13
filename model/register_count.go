package model

type RegisterCount struct {
	Id         uint   `gorm:"primaryKey;autoIncrement"`
	RustdeskId string `gorm:"index;size:64;not null"`
	Date       string `gorm:"index;size:10;not null"`
	Count      int    `gorm:"default:0"`
}