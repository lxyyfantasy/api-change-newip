package model

// ===== 激活码 4 表 =====

type LeaseCodeUnused struct {
    Code     string `json:"code" gorm:"primaryKey;"`
    Days     int    `json:"days" gorm:"not null;"`
    ExpireAt int64  `json:"expire_at" gorm:"default:0;not null;"`
    TakenAt  int64  `json:"taken_at" gorm:"default:0;not null;"`
}
type LeaseCodeUsed struct {
    Code   string `json:"code" gorm:"primaryKey;"`
    Days   int    `json:"days" gorm:"not null;"`
    BindId string `json:"bind_id" gorm:"not null;index"`
    UsedAt int64  `json:"used_at" gorm:"not null;"`
}
type HostCodeUnused struct {
    Code     string `json:"code" gorm:"primaryKey;"`
    Days     int    `json:"days" gorm:"not null;"`
    ExpireAt int64  `json:"expire_at" gorm:"default:0;not null;"`
    TakenAt  int64  `json:"taken_at" gorm:"default:0;not null;"`
}
type HostCodeUsed struct {
    Code   string `json:"code" gorm:"primaryKey;"`
    Days   int    `json:"days" gorm:"not null;"`
    BindId string `json:"bind_id" gorm:"not null;index"`
    UsedAt int64  `json:"used_at" gorm:"not null;"`
}

// ===== 已删 ID 留存 =====

type RemovedDevice struct {
    Id            string `json:"id" gorm:"primaryKey;"`
    Account       string `json:"account" gorm:"not null;"`
    LeaseExpireAt int64  `json:"lease_expire_at" gorm:"default:0;not null;"`
    HostExpireAt  int64  `json:"host_expire_at" gorm:"default:0;not null;"`
    Ip            string `json:"ip" gorm:"default:'';not null;"`
    CreatedAt     int64  `json:"created_at" gorm:"default:0;not null;"`
    RemovedAt     int64  `json:"removed_at" gorm:"not null;"`
}

// ===== 已删账号留存 =====

type EmptyAccount struct {
    Account    string `json:"account" gorm:"primaryKey;"`
    EmptySince int64  `json:"empty_since" gorm:"not null;"`
}

// ===== 服务器配置（Key + ip1 + ip2） =====

type ServerConfig struct {
    Id         uint   `json:"id" gorm:"primaryKey;"`
    ServerKey  string `json:"server_key" gorm:"default:'';not null;"`
    Ip1        string `json:"ip1" gorm:"default:'';not null;"`
    Ip2        string `json:"ip2" gorm:"default:'';not null;"`
    UpdatedAt  int64  `json:"updated_at" gorm:"default:0;not null;"`
}

// ===== 注册账号信息 =====

type RegAccountInfo struct {
    Account string `json:"account" gorm:"primaryKey;"`
    Answer1 string `json:"answer1" gorm:"default:'';not null;"`
    Answer2 string `json:"answer2" gorm:"default:'';not null;"`
    Answer3 string `json:"answer3" gorm:"default:'';not null;"`
    TimeModel
}

// ===== 邀请码 =====

type InviteCode struct {
    Code    string `json:"code" gorm:"primaryKey;"`
    TakenAt int64  `json:"taken_at" gorm:"default:0;not null;"`
    TimeModel
}

// ===== 添加设备码 =====

type AddDeviceCode struct {
    Code    string `json:"code" gorm:"primaryKey;"`
    TakenAt int64  `json:"taken_at" gorm:"default:0;not null;"`
    TimeModel
}

// ===== 分配设备（账号为空=待分配；账号非空=已分配） =====

type PendingDevice struct {
    Id         string `json:"id" gorm:"primaryKey;"`
    Account    string `json:"account" gorm:"default:'';not null;index"`
    AssignedAt int64  `json:"assigned_at" gorm:"default:0;not null;"`
    TimeModel
}

// ===== 游离托管设备（每天0点从peers重建，分发后打标记） =====

type OrphanHostDevice struct {
    Id        string `json:"id" gorm:"primaryKey;"`
    TakenAt   int64  `json:"taken_at" gorm:"default:0;not null;index"`
    TimeModel
}