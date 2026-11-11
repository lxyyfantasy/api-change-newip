package model

type RegisteredId struct {
    Id        uint   `json:"id" gorm:"primaryKey"`
    PeerId    string `json:"peer_id" gorm:"size:255;not null;index"`
    CreatedAt int64  `json:"created_at" gorm:"not null;index"`
}