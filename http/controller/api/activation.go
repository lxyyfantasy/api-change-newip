package api

import (
    "github.com/gin-gonic/gin"
    "github.com/lejianwen/rustdesk-api/v2/http/response"
    "github.com/lejianwen/rustdesk-api/v2/service"
    "github.com/lejianwen/rustdesk-api/v2/model"
    "net/http"
    "strconv"
    "strings"
    "time"
)

type Activation struct {
}

// Activate 激活
func (a *Activation) Activate(c *gin.Context) {
    f := &struct {
        Code     string `json:"code" binding:"required"`
        Kind     string `json:"kind" binding:"required"`
        DeviceId string `json:"device_id" binding:"required"`
        Account  string `json:"account"`
    }{}
    if err := c.ShouldBindJSON(f); err != nil {
        c.JSON(http.StatusOK, gin.H{"code": "BAD_REQUEST", "msg": "参数错误"})
        return
    }
    err, expireAt := service.AllService.ActivationService.Activate(f.Code, f.Kind, f.DeviceId, f.Account)
    if err != nil {
        code := "INVALID_CODE"
        msg := "激活失败"
        switch err.Error() {
        case "BAD_KIND":
            code = "BAD_REQUEST"
            msg = "参数错误"
        case "CODE_USED":
            code = "CODE_USED"
            msg = "激活码已被使用"
        }
        c.JSON(http.StatusOK, gin.H{"code": code, "msg": msg})
        return
    }
    c.JSON(http.StatusOK, gin.H{"code": "OK", "msg": "激活成功", "expire_at": expireAt})
}

// Report IP + 识别码上报
func (a *Activation) Report(c *gin.Context) {
    f := &struct {
        DeviceId   string `json:"device_id" binding:"required"`
        Ip         string `json:"ip" binding:"required"`
        DeviceCode string `json:"device_code"`
    }{}
    if err := c.ShouldBindJSON(f); err != nil {
        c.JSON(http.StatusOK, gin.H{"code": "BAD_REQUEST", "msg": "参数错误"})
        return
    }
    err := service.AllService.ActivationService.ReportIp(f.DeviceId, f.Ip, f.DeviceCode)
    if err != nil {
        c.JSON(http.StatusOK, gin.H{"code": "SERVER_ERROR", "msg": "服务器异常"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"code": "OK"})
}

var _ = response.Success

func (a *Activation) PeerAll(c *gin.Context) {
	list := make([]*model.Peer, 0)
	service.DB.Find(&list)
	for _, p := range list {
		var ab model.AddressBook
		account := "__device__"
		if service.DB.Where("id = ?", p.Id).First(&ab).Error == nil && ab.UserId > 0 {
			var u model.User
			if service.DB.Where("id = ?", ab.UserId).First(&u).Error == nil {
				account = u.Username
			}
		}
		p.Account = account
	}
	c.JSON(http.StatusOK, gin.H{"list": list})
}

func (a *Activation) ServerConfig(c *gin.Context) {
    sc := &model.ServerConfig{}
    service.DB.Where("id = 1").First(sc)
    c.JSON(http.StatusOK, gin.H{"server_key": sc.ServerKey, "ip1": sc.Ip1, "ip2": sc.Ip2})
}
func (a *Activation) DeviceInfo(c *gin.Context) {
	id := c.Query("device_id")
	if id == "" {
		c.JSON(http.StatusOK, gin.H{"code": "BAD_REQUEST", "msg": "参数错误"})
		return
	}
	pe := service.AllService.PeerService.FindById(id)
	if pe.RowId == 0 {
		c.JSON(http.StatusOK, gin.H{"code": "NOT_FOUND", "msg": "设备不存在"})
		return
	}
	now := time.Now().Unix()
	c.JSON(http.StatusOK, gin.H{
		"code":          "OK",
		"lease_expire":  pe.LeaseExpireAt,
		"host_expire":   pe.HostExpireAt,
		"ip":            pe.LanIp,
		"lease_expired": pe.LeaseExpireAt < now,
	})
}

func (a *Activation) PullOrphanHost(c *gin.Context) {
    n, _ := strconv.Atoi(c.Query("n"))
    ids := service.AllService.ActivationService.PullOrphanHost(n)
    c.String(http.StatusOK, strings.Join(ids, "\n"))
}

func (a *Activation) LeaseHostIps(c *gin.Context) {
    prefix := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(c.Query("prefix")), "."))
    parts := strings.Split(prefix, ".")
    if len(parts) != 3 {
        c.String(http.StatusOK, "")
        return
    }
    for _, p := range parts {
        if p == "" || len(p) > 3 {
            c.String(http.StatusOK, "")
            return
        }
        for _, ch := range p {
            if ch < '0' || ch > '9' {
                c.String(http.StatusOK, "")
                return
            }
        }
    }
    now := time.Now().Unix()

    // 1. 按 IP 段过滤
    type peerRow struct {
        Id           string
        LanIp        string
        HostExpireAt int64
    }
    var peers []peerRow
    service.DB.Model(&model.Peer{}).
        Select("id, lan_ip, host_expire_at").
        Where("lan_ip LIKE ?", prefix+".%").
        Where("lan_ip <> ''").
        Scan(&peers)

    if len(peers) == 0 {
        c.String(http.StatusOK, "")
        return
    }

    // 2. 先按有效期分：有效直接收，过期进 expiredIds 待查归属
    ips := make([]string, 0, len(peers))
    expiredIds := make([]string, 0)
    for _, p := range peers {
        if p.HostExpireAt > now {
            ips = append(ips, p.LanIp)
        } else {
            expiredIds = append(expiredIds, p.Id)
        }
    }

    // 3. 过期设备：查地址簿归属，是管理员的也收
    if len(expiredIds) > 0 {
        type abRow struct {
            Id     string
            UserId uint
        }
        var abs []abRow
        service.DB.Model(&model.AddressBook{}).
            Select("id, user_id").
            Where("id IN ?", expiredIds).
            Where("user_id > 0").
            Scan(&abs)

        if len(abs) > 0 {
            userIds := make([]uint, 0, len(abs))
            for _, ab := range abs {
                userIds = append(userIds, ab.UserId)
            }
            var adminIds []uint
            service.DB.Model(&model.User{}).
                Where("id IN ? AND is_admin = ?", userIds, true).
                Pluck("id", &adminIds)
            adminSet := make(map[uint]struct{}, len(adminIds))
            for _, id := range adminIds {
                adminSet[id] = struct{}{}
            }

            ipById := make(map[string]string, len(peers))
            for _, p := range peers {
                ipById[p.Id] = p.LanIp
            }
            for _, ab := range abs {
                if _, isAdmin := adminSet[ab.UserId]; isAdmin {
                    ips = append(ips, ipById[ab.Id])
                }
            }
        }
    }

    c.String(http.StatusOK, strings.Join(ips, "\n"))
}
