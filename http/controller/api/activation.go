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
        if err.Error() == "BAD_KIND" {
            code = "BAD_REQUEST"
        }
        c.JSON(http.StatusOK, gin.H{"code": code, "msg": "激活失败"})
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
    prefix := strings.TrimSpace(c.Query("prefix"))
    if prefix == "" {
        c.String(http.StatusOK, "")
        return
    }
    now := time.Now().Unix()
    var ips []string
    service.DB.Model(&model.Peer{}).
        Where("lan_ip LIKE ? AND lease_expire_at > ? AND host_expire_at > ?", prefix+".%", now, now).
        Where("lan_ip <> ''").
        Pluck("lan_ip", &ips)
    c.String(http.StatusOK, strings.Join(ips, "\n"))
}

func (a *Activation) ReportConnect(c *gin.Context) {
    f := &struct {
        DeviceId string `json:"device_id" binding:"required"`
        Type     string `json:"type" binding:"required"`
    }{}
    if err := c.ShouldBindJSON(f); err != nil {
        c.JSON(http.StatusOK, gin.H{"code": "BAD_REQUEST", "msg": "参数错误"})
        return
    }
    if f.Type != "relay" && f.Type != "direct" {
        c.JSON(http.StatusOK, gin.H{"code": "BAD_REQUEST", "msg": "参数错误"})
        return
    }
    service.DB.Model(&model.Peer{}).Where("id = ?", f.DeviceId).Update("connect_type", f.Type)
    c.JSON(http.StatusOK, gin.H{"code": "OK"})
}