package model

type Peer struct {
    RowId          uint   `json:"row_id" gorm:"primaryKey;"`
    Id             string `json:"id"  gorm:"default:'';not null;index"`
    Cpu            string `json:"cpu"  gorm:"default:'';not null;"`
    Hostname       string `json:"hostname"  gorm:"default:'';not null;"`
    Memory         string `json:"memory"  gorm:"default:'';not null;"`
    Os             string `json:"os"  gorm:"default:'';not null;"`
    Username       string `json:"username"  gorm:"default:'';not null;"`
    Uuid           string `json:"uuid"  gorm:"default:'';not null;index"`
    Version        string `json:"version"  gorm:"default:'';not null;"`
    UserId         uint   `json:"user_id"  gorm:"default:0;not null;index"`
    User           *User  `json:"user,omitempty"`
    LastOnlineTime int64  `json:"last_online_time"  gorm:"default:0;not null;"`
    LastOnlineIp   string `json:"last_online_ip"  gorm:"default:'';not null;"`
    GroupId        uint   `json:"group_id"  gorm:"default:0;not null;index"`
    Alias          string `json:"alias" gorm:"default:'';not null;index"`
	Account       string `json:"account" gorm:"default:'';not null;index"`
	Category      string `json:"category" gorm:"default:'';not null;index"`
    LanIp         string `json:"lan_ip" gorm:"default:'';not null;"`
    DeviceCode    string `json:"device_code" gorm:"default:'';not null;index"`
    LeaseExpireAt int64  `json:"lease_expire_at" gorm:"default:0;not null;"`
    LeaseUser     string `json:"lease_user" gorm:"default:'';not null;"`
    LeaseCode     string `json:"lease_code" gorm:"default:'';not null;"`
    LeaseDays     int    `json:"lease_days" gorm:"default:0;not null;"`
    LeaseAt       int64  `json:"lease_at" gorm:"default:0;not null;"`
    HostExpireAt  int64  `json:"host_expire_at" gorm:"default:0;not null;"`
    HostUser      string `json:"host_user" gorm:"default:'';not null;"`
    HostCode      string `json:"host_code" gorm:"default:'';not null;"`
    HostDays      int    `json:"host_days" gorm:"default:0;not null;"`
    HostAt        int64  `json:"host_at" gorm:"default:0;not null;"`
    TimeModel
}

type PeerList struct {
    Peers []*Peer `json:"list"`
    Pagination
}
