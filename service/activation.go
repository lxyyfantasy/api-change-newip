package service

import (
    "fmt"
    "crypto/rand"
    "errors"
    "time"

    "github.com/lejianwen/rustdesk-api/v2/model"
    "gorm.io/gorm"
)

type ActivationService struct {
}

const CodeChars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func randomCode(n int) string {
    b := make([]byte, n)
    if _, err := rand.Read(b); err != nil {
        return ""
    }
    for i := range b {
        b[i] = CodeChars[int(b[i])%len(CodeChars)]
    }
    return string(b)
}

// GenLeaseCodes 开通激活码，按需生成，16位，最多20，表内查重
func (as *ActivationService) GenLeaseCodes(days, n int) []string {
    return as.genCodes("lease", days, n, 16)
}

// GenHostCodes 托管激活码，按需生成，18位，最多20，表内查重
func (as *ActivationService) GenHostCodes(days, n int) []string {
    return as.genCodes("host", days, n, 18)
}

// GenInviteCodes 邀请码，按需生成，12位，最多20，表内查重
func (as *ActivationService) GenInviteCodes(n int) []string {
    return as.genInvite(n, 12)
}

// GenAddDeviceCodes 添加设备码，按需生成，14位，最多20，表内查重
func (as *ActivationService) GenAddDeviceCodes(n int) []string {
    return as.genAddDevice(n, 14)
}

func (as *ActivationService) genCodes(kind string, days, n, size int) []string {
    if n <= 0 {
        return nil
    }
    if n > 20 {
        n = 20
    }
    codes := make([]string, 0, n)
    now := time.Now().Unix()
    for i := 0; i < n; i++ {
        for attempt := 0; attempt < 50; attempt++ {
            code := randomCode(size)
            if code == "" || as.codeExists(kind, code) {
                continue
            }
            var err error
            if kind == "lease" {
                err = DB.Create(&model.LeaseCodeUnused{Code: code, Days: days, TakenAt: now}).Error
            } else {
                err = DB.Create(&model.HostCodeUnused{Code: code, Days: days, TakenAt: now}).Error
            }
            if err == nil {
                codes = append(codes, code)
                break
            }
        }
    }
    return codes
}

func (as *ActivationService) genInvite(n, size int) []string {
    if n <= 0 {
        return nil
    }
    if n > 20 {
        n = 20
    }
    codes := make([]string, 0, n)
    now := time.Now().Unix()
    for i := 0; i < n; i++ {
        for attempt := 0; attempt < 50; attempt++ {
            code := randomCode(size)
            if code == "" {
                continue
            }
            var cnt int64
            DB.Model(&model.InviteCode{}).Where("code = ?", code).Count(&cnt)
            if cnt > 0 {
                continue
            }
            if DB.Create(&model.InviteCode{Code: code, TakenAt: now}).Error == nil {
                codes = append(codes, code)
                break
            }
        }
    }
    return codes
}

func (as *ActivationService) genAddDevice(n, size int) []string {
    if n <= 0 {
        return nil
    }
    if n > 20 {
        n = 20
    }
    codes := make([]string, 0, n)
    now := time.Now().Unix()
    for i := 0; i < n; i++ {
        for attempt := 0; attempt < 50; attempt++ {
            code := randomCode(size)
            if code == "" {
                continue
            }
            var cnt int64
            DB.Model(&model.AddDeviceCode{}).Where("code = ?", code).Count(&cnt)
            if cnt > 0 {
                continue
            }
            if DB.Create(&model.AddDeviceCode{Code: code, TakenAt: now}).Error == nil {
                codes = append(codes, code)
                break
            }
        }
    }
    return codes
}

func (as *ActivationService) codeExists(kind, code string) bool {
    var cnt int64
    if kind == "lease" {
        DB.Model(&model.LeaseCodeUnused{}).Where("code = ?", code).Count(&cnt)
        if cnt > 0 {
            return true
        }
        DB.Model(&model.LeaseCodeUsed{}).Where("code = ?", code).Count(&cnt)
        return cnt > 0
    }
    DB.Model(&model.HostCodeUnused{}).Where("code = ?", code).Count(&cnt)
    if cnt > 0 {
        return true
    }
    DB.Model(&model.HostCodeUsed{}).Where("code = ?", code).Count(&cnt)
    return cnt > 0
}

// TakeCodes 取码：按需生成后直接返回
func (as *ActivationService) TakeCodes(kind string, days, n int) []string {
    switch kind {
    case "lease":
        return as.GenLeaseCodes(days, n)
    case "host":
        return as.GenHostCodes(days, n)
    case "invite":
        return as.GenInviteCodes(n)
    case "adddevice":
        return as.GenAddDeviceCodes(n)
    }
    return nil
}

// Activate 激活
func (as *ActivationService) Activate(code, kind, deviceId, account string) (err error, expireAt int64) {
    if account == "" {
        account = "__device__"
    }
    // 设备ID必须在地址簿里
    var abCnt int64
    DB.Model(&model.AddressBook{}).Where("id = ?", deviceId).Count(&abCnt)
    if abCnt == 0 {
        return errors.New("DEVICE_NOT_IN_AB"), 0
    }
    tx := DB.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            err = errors.New("panic")
        }
    }()

    var days int
    var expireField string
    now := time.Now().Unix()

    switch kind {
    case "lease":
        expireField = "lease_expire_at"
        var c model.LeaseCodeUnused
        if tx.Where("code = ?", code).First(&c).Error != nil {
            var used model.LeaseCodeUsed
            if tx.Where("code = ?", code).First(&used).Error == nil && used.BindId == deviceId {
                tx.Rollback()
                return errors.New("CODE_USED"), 0
            }
            tx.Rollback()
            return errors.New("INVALID_CODE"), 0
        }
        days = c.Days
        r := tx.Where("code = ?", code).Delete(&model.LeaseCodeUnused{})
        if r.Error != nil || r.RowsAffected != 1 {
            tx.Rollback()
            return errors.New("CODE_USED"), 0
        }
        if err := tx.Create(&model.LeaseCodeUsed{Code: code, Days: days, BindId: deviceId, UsedAt: now}).Error; err != nil {
            tx.Rollback()
            return errors.New("INVALID_CODE"), 0
        }
    case "host":
        expireField = "host_expire_at"
        var c model.HostCodeUnused
        if tx.Where("code = ?", code).First(&c).Error != nil {
            var used model.HostCodeUsed
            if tx.Where("code = ?", code).First(&used).Error == nil && used.BindId == deviceId {
                tx.Rollback()
                return errors.New("CODE_USED"), 0
            }
            tx.Rollback()
            return errors.New("INVALID_CODE"), 0
        }
        days = c.Days
        r := tx.Where("code = ?", code).Delete(&model.HostCodeUnused{})
        if r.Error != nil || r.RowsAffected != 1 {
            tx.Rollback()
            return errors.New("CODE_USED"), 0
        }
        if err := tx.Create(&model.HostCodeUsed{Code: code, Days: days, BindId: deviceId, UsedAt: now}).Error; err != nil {
            tx.Rollback()
            return errors.New("INVALID_CODE"), 0
        }
    default:
        tx.Rollback()
        return errors.New("BAD_KIND"), 0
    }

    pe := AllService.PeerService.FindById(deviceId)
    if pe.RowId == 0 {
        pe = &model.Peer{Id: deviceId, Account: account, LastOnlineTime: time.Now().Unix()}
        tx.Create(pe)
    }
    base := pe.LeaseExpireAt
    if kind == "host" {
        base = pe.HostExpireAt
    }
    if base < now {
        base = now
    }
    expireAt = base + int64(days)*86400

    updates := map[string]interface{}{
        expireField: expireAt,
        "account":   account,
    }
    if kind == "lease" {
        updates["lease_user"] = account
        updates["lease_code"] = code
        updates["lease_days"] = days
        updates["lease_at"] = now
    } else {
        updates["host_user"] = account
        updates["host_code"] = code
        updates["host_days"] = days
        updates["host_at"] = now
    }
    tx.Model(&model.Peer{}).Where("id = ?", deviceId).Updates(updates)
    tx.Commit()
    return nil, expireAt
}

// CleanExpiredPeers 未激活或已到期满3天：移出地址簿，变游离，其他信息不动
func (as *ActivationService) CleanExpiredPeers() {
    cutoff := time.Now().Unix() - 3*86400
    cutoffTime := time.Now().Add(-3 * 24 * time.Hour)
    var list []model.Peer
    DB.Where("account != ? AND (lease_expire_at = 0 AND created_at < ? OR lease_expire_at > 0 AND lease_expire_at < ?)",
        "__device__", cutoffTime, cutoff).Find(&list)
    for _, p := range list {
        DB.Where("id = ?", p.Id).Delete(&model.AddressBook{})
        DB.Model(&model.Peer{}).Where("id = ?", p.Id).Update("account", "__device__")
    }
}

func (as *ActivationService) removePeerAndAB(id string) {
    DB.Where("id = ?", id).Delete(&model.Peer{})
    DB.Where("id = ?", id).Delete(&model.AddressBook{})
}

func (as *ActivationService) CleanEmptyAccounts() {
	// 1. 扫普通用户，名下无设备记空时间，有设备清空
	var users []model.User
	DB.Where("is_admin = ?", false).Find(&users)
	for _, u := range users {
		var cnt int64
		DB.Model(&model.Peer{}).Where("account = ?", u.Username).Count(&cnt)
		if cnt == 0 {
			if u.EmptySince == 0 {
				DB.Model(&model.User{}).Where("id = ?", u.Id).Update("empty_since", time.Now().Unix())
			}
		} else {
			if u.EmptySince != 0 {
				DB.Model(&model.User{}).Where("id = ?", u.Id).Update("empty_since", 0)
			}
		}
	}

	// 2. 空满 7 天的：删账号及关联，进留存表
	cutoff7 := time.Now().Unix() - 7*86400
	var delUsers []model.User
	DB.Where("is_admin = ? AND empty_since > 0 AND empty_since < ?", false, cutoff7).Find(&delUsers)
	for _, u := range delUsers {
		DB.Where("user_id = ?", u.Id).Delete(&model.UserToken{})
		DB.Where("account = ?", u.Username).Delete(&model.RegAccountInfo{})
		DB.Delete(&u)
		// 每次删除都进留存表，留 30 天
		DB.Create(&model.EmptyAccount{
			Account:    fmt.Sprintf("%s_%d", u.Username, time.Now().UnixMilli()),
			EmptySince: time.Now().Unix(),
		})
	}

	// 3. 留存表满 30 天的，删记录
	cutoff30 := time.Now().Unix() - 30*86400
	DB.Where("empty_since < ?", cutoff30).Delete(&model.EmptyAccount{})
}

// CleanRemovedDevices 已删 ID 留存 30 天
func (as *ActivationService) CleanRemovedDevices() {
    cutoff := time.Now().Unix() - 30*86400
    DB.Where("removed_at < ?", cutoff).Delete(&model.RemovedDevice{})
}

// RefreshAccountColumn 扫地址簿填归属账号，不删行
func (as *ActivationService) RefreshAccountColumn() {
    var peers []model.Peer
    DB.Find(&peers)
    for _, p := range peers {
        var ab model.AddressBook
        var account string
        if DB.Where("id = ?", p.Id).First(&ab).Error == nil && ab.UserId > 0 {
            var u model.User
            if DB.Where("id = ?", ab.UserId).First(&u).Error == nil {
                account = u.Username
            }
        }
        if account == "" {
            account = "__device__"
        }
        DB.Model(&model.Peer{}).Where("id = ?", p.Id).Update("account", account)
    }
}

// ReportIp 上报 IP + 识别码
func (as *ActivationService) ReportIp(deviceId, ip, deviceCode string) error {
	if deviceCode != "" {
		// 按设备码找到这台设备的行，只更新 IP 等，不动 id
		var old model.Peer
		if DB.Where("device_code = ?", deviceCode).First(&old).Error == nil && old.RowId > 0 {
			return DB.Model(&model.Peer{}).Where("row_id = ?", old.RowId).Updates(map[string]interface{}{
				"lan_ip":           ip,
				"last_online_time": time.Now().Unix(),
				"category":         "被控端",
			}).Error
		}
	}
	return DB.Model(&model.Peer{}).Where("id = ?", deviceId).Updates(map[string]interface{}{
		"lan_ip":           ip,
		"device_code":      deviceCode,
		"last_online_time": time.Now().Unix(),
		"category":         "被控端",
	}).Error
}

// CleanUnused 未使用清理：激活码15天，邀请码/添加设备码30天
func (as *ActivationService) CleanUnused() {
    cutoff := time.Now().Unix() - 15*86400
    DB.Where("taken_at > 0 AND taken_at < ?", cutoff).Delete(&model.LeaseCodeUnused{})
    DB.Where("taken_at > 0 AND taken_at < ?", cutoff).Delete(&model.HostCodeUnused{})
    cutoff30 := time.Now().Unix() - 30*86400
    DB.Where("taken_at > 0 AND taken_at < ?", cutoff30).Delete(&model.InviteCode{})
    DB.Where("taken_at > 0 AND taken_at < ?", cutoff30).Delete(&model.AddDeviceCode{})
}

// CleanUsed 已使用 30 天清理
func (as *ActivationService) CleanUsed() {
    cutoff := time.Now().Unix() - 30*86400
    DB.Where("used_at < ?", cutoff).Delete(&model.LeaseCodeUsed{})
    DB.Where("used_at < ?", cutoff).Delete(&model.HostCodeUsed{})
}

// CleanOfflinePeers 30天没入网的设备删掉
func (as *ActivationService) CleanOfflinePeers() {
    cutoff := time.Now().Unix() - 30*86400
    var list []model.Peer
    DB.Where("last_online_time > 0 AND last_online_time < ?", cutoff).Find(&list)
    for _, p := range list {
        DB.Create(&model.RemovedDevice{
            Id: fmt.Sprintf("%s_%d", p.Id, time.Now().UnixMilli()),
            Account:       p.Account,
            LeaseExpireAt: p.LeaseExpireAt,
            HostExpireAt:  p.HostExpireAt,
            Ip:            p.LanIp,
            CreatedAt:     time.Time(p.CreatedAt).Unix(),
            RemovedAt:     time.Now().Unix(),
        })
        DB.Where("id = ?", p.Id).Delete(&model.AddressBook{})
        DB.Where("id = ?", p.Id).Delete(&model.Peer{})
    }
}

// CheckInviteCode 核对邀请码，存在则返回 true
func (as *ActivationService) CheckInviteCode(code string) bool {
    var cnt int64
    DB.Model(&model.InviteCode{}).Where("code = ?", code).Count(&cnt)
    return cnt > 0
}

// UseInviteCode 使用后删除
func (as *ActivationService) UseInviteCode(code string) {
    DB.Where("code = ?", code).Delete(&model.InviteCode{})
}

// CheckAddDeviceCode 核对添加设备码
func (as *ActivationService) CheckAddDeviceCode(code string) bool {
    var cnt int64
    DB.Model(&model.AddDeviceCode{}).Where("code = ?", code).Count(&cnt)
    return cnt > 0
}

// UseAddDeviceCode 使用后删除
func (as *ActivationService) UseAddDeviceCode(code string) {
    DB.Where("code = ?", code).Delete(&model.AddDeviceCode{})
}

// TakePendingDevice 取一台待分配设备并分配到 account
func (as *ActivationService) TakePendingDevice(account string) (ok bool, deviceId string) {
    var d model.PendingDevice
    if DB.Where("account = ?", "").First(&d).Error != nil {
        return false, ""
    }
    DB.Model(&d).Updates(map[string]interface{}{
        "account":     account,
        "assigned_at": time.Now().Unix(),
    })
    return true, d.Id
}

// PendingCount 返回待分配数量
func (as *ActivationService) PendingCount() int64 {
    var cnt int64
    DB.Model(&model.PendingDevice{}).Where("account = ?", "").Count(&cnt)
    return cnt
}

// SaveRegInfo 写入注册答案
func (as *ActivationService) SaveRegInfo(account, a1, a2, a3 string) {
    DB.Create(&model.RegAccountInfo{Account: account, Answer1: a1, Answer2: a2, Answer3: a3})
}

// CheckRegInfo 核对找回密码的4个参数
func (as *ActivationService) CheckRegInfo(account, a1, a2, a3 string) bool {
    var cnt int64
    DB.Model(&model.RegAccountInfo{}).
        Where("account = ? AND answer1 = ? AND answer2 = ? AND answer3 = ?", account, a1, a2, a3).
        Count(&cnt)
    return cnt > 0
}

// CleanAssignedDevices 已分配满30天删除
func (as *ActivationService) CleanAssignedDevices() {
    cutoff := time.Now().Unix() - 30*86400
    DB.Where("id LIKE ? AND assigned_at > 0 AND assigned_at < ?", "已分配_%", cutoff).Delete(&model.PendingDevice{})
}

// RebuildOrphanHost 每天0点重建游离托管设备表：account='__device__' 且 host_expire_at > now
func (as *ActivationService) RebuildOrphanHost() {
    now := time.Now().Unix()
    var ids []string
    DB.Model(&model.Peer{}).
        Where("account = ? AND host_expire_at > ?", "__device__", now).
        Pluck("id", &ids)
    tx := DB.Begin()
    tx.Where("1 = 1").Delete(&model.OrphanHostDevice{})
    for _, id := range ids {
        tx.Create(&model.OrphanHostDevice{Id: id, TakenAt: 0})
    }
    tx.Commit()
}

// SyncOrphanHost 每小时兜底扫描：全扫 Peer 表，把符合游离条件但游离表中没有的 ID 补进去
// 筛选逻辑与 RebuildOrphanHost 完全一致：account='__device__' 且 host_expire_at > now
// 新增记录一律按未分配写入：taken_at = 0
func (as *ActivationService) SyncOrphanHost() {
    now := time.Now().Unix()

    var ids []string
    DB.Model(&model.Peer{}).
        Where("account = ? AND host_expire_at > ?", "__device__", now).
        Pluck("id", &ids)
    if len(ids) == 0 {
        return
    }

    // 查重：不带 taken_at 条件，已分配 + 未分配全部纳入
    var exists []string
    DB.Model(&model.OrphanHostDevice{}).
        Where("id in ?", ids).
        Pluck("id", &exists)
    existSet := make(map[string]struct{}, len(exists))
    for _, id := range exists {
        existSet[id] = struct{}{}
    }

    // 只补缺失的，统一按未分配(taken_at=0)写入
    tx := DB.Begin()
    for _, id := range ids {
        if _, ok := existSet[id]; ok {
            continue
        }
        tx.Create(&model.OrphanHostDevice{Id: id, TakenAt: 0})
    }
    tx.Commit()
}

// PullOrphanHost 分发游离托管设备：取 n 个未分发的，标记已分发，返回ID
func (as *ActivationService) PullOrphanHost(n int) []string {
    if n <= 0 {
        return nil
    }
    if n > 100 {
        n = 100
    }
    var list []model.OrphanHostDevice
    DB.Where("taken_at = 0").Limit(n).Find(&list)
    if len(list) == 0 {
        return nil
    }
    ids := make([]string, 0, len(list))
    now := time.Now().Unix()
    for _, d := range list {
        ids = append(ids, d.Id)
    }
    DB.Model(&model.OrphanHostDevice{}).Where("id in ?", ids).Update("taken_at", now)
    return ids
}

// 消费码，返回影响行数，调用方判断 == 1
func (as *ActivationService) UseInviteCodeTx(tx *gorm.DB, code string) (int64, error) {
    r := tx.Where("code = ?", code).Delete(&model.InviteCode{})
    return r.RowsAffected, r.Error
}

func (as *ActivationService) UseAddDeviceCodeTx(tx *gorm.DB, code string) (int64, error) {
    r := tx.Where("code = ?", code).Delete(&model.AddDeviceCode{})
    return r.RowsAffected, r.Error
}

func (as *ActivationService) SaveRegInfoTx(tx *gorm.DB, account, a1, a2, a3 string) error {
    return tx.Create(&model.RegAccountInfo{Account: account, Answer1: a1, Answer2: a2, Answer3: a3}).Error
}

// 原子取设备：UPDATE ... WHERE account='' LIMIT 1，看影响行数
func (as *ActivationService) TakePendingDeviceTx(tx *gorm.DB, account string) (bool, string) {
    var d model.PendingDevice
    // 先原子占位
    r := tx.Exec(`UPDATE pending_devices SET account=?, assigned_at=? 
                  WHERE account='' AND id = (SELECT id FROM pending_devices WHERE account='' LIMIT 1)`,
        account, time.Now().Unix())
    if r.Error != nil || r.RowsAffected != 1 {
        return false, ""
    }
    // 再取出 id（同一事务内，安全）
    if tx.Where("account = ?", account).Order("assigned_at desc").First(&d).Error != nil {
        return false, ""
    }
    return true, d.Id
}

// SaveRegisteredId 记录已注册 ID，存在则跳过（不刷新时间戳）
func (as *ActivationService) SaveRegisteredId(peerId string) error {
    var cnt int64
    DB.Model(&model.RegisteredId{}).Where("peer_id = ?", peerId).Count(&cnt)
    if cnt > 0 {
        return nil
    }
    return DB.Create(&model.RegisteredId{PeerId: peerId, CreatedAt: time.Now().Unix()}).Error
}

// CleanRegisteredIds 清理 14 天前的已注册 ID
func (as *ActivationService) CleanRegisteredIds() {
    cutoff := time.Now().Unix() - 14*86400
    DB.Where("created_at < ?", cutoff).Delete(&model.RegisteredId{})
}