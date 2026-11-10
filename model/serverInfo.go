package model

type ServerInfo struct {
	Id        uint   `json:"id" gorm:"primaryKey;"`
	Ip        string `json:"ip" gorm:"default:'';not null;uniqueIndex"`
	Bandwidth string `json:"bandwidth" gorm:"default:'';not null;"`
	TimeModel
}

type ServerInfoList struct {
	List []*ServerInfo `json:"list"`
}