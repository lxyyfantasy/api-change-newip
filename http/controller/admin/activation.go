package admin

import (
    "net/http"
    "strconv"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/lejianwen/rustdesk-api/v2/model"
    "github.com/lejianwen/rustdesk-api/v2/service"
)

type Activation struct {
}

func (ct *Activation) PeerAll(c *gin.Context) {
    list := make([]*model.Peer, 0)
    service.DB.Find(&list)
    c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"list": list}})
}

func (ct *Activation) ServerConfigGet(c *gin.Context) {
    sc := &model.ServerConfig{}
    service.DB.Where("id = 1").First(sc)
    c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
        "server_key": sc.ServerKey,
        "ip1":        sc.Ip1,
        "ip2":        sc.Ip2,
    }})
}

func (ct *Activation) ServerConfigPut(c *gin.Context) {
    f := &struct {
        Key string `json:"key"`
        Ip1 string `json:"ip1"`
        Ip2 string `json:"ip2"`
    }{}
    if err := c.ShouldBindJSON(f); err != nil {
        c.JSON(http.StatusOK, gin.H{"code": 101, "message": "参数错误"})
        return
    }
    service.DB.Model(&model.ServerConfig{}).Where("id = 1").Updates(map[string]interface{}{
        "server_key": f.Key,
        "ip1":        f.Ip1,
        "ip2":        f.Ip2,
        "updated_at": time.Now().Unix(),
    })
    c.JSON(http.StatusOK, gin.H{"code": 0})
}

func (ct *Activation) TakeCodes(c *gin.Context) {
    kind := c.Query("kind")
    days, _ := strconv.Atoi(c.Query("days"))
    n, _ := strconv.Atoi(c.Query("n"))
    if n <= 0 || n > 20 {
        c.JSON(http.StatusOK, gin.H{"code": 101, "message": "参数错误"})
        return
    }
    if kind == "lease" || kind == "host" {
        if days <= 0 {
            c.JSON(http.StatusOK, gin.H{"code": 101, "message": "参数错误"})
            return
        }
    }
    codes := service.AllService.ActivationService.TakeCodes(kind, days, n)
    c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"list": codes}})
}

func (ct *Activation) RemovedDevices(c *gin.Context) {
    list := make([]*model.RemovedDevice, 0)
    service.DB.Order("removed_at desc").Find(&list)
    c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"list": list}})
}

func (ct *Activation) RemovedAccounts(c *gin.Context) {
    list := make([]*model.EmptyAccount, 0)
    service.DB.Order("empty_since desc").Find(&list)
    c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"list": list}})
}

func (ct *Activation) UsedCodes(c *gin.Context) {
    kind := c.Query("kind")
    if kind == "lease" {
        list := make([]*model.LeaseCodeUsed, 0)
        service.DB.Order("used_at desc").Find(&list)
        c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"list": list}})
        return
    }
    list := make([]*model.HostCodeUsed, 0)
    service.DB.Order("used_at desc").Find(&list)
    c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"list": list}})
}

func (ct *Activation) PendingList(c *gin.Context) {
    list := make([]*model.PendingDevice, 0)
    service.DB.Where("account = ?", "").Order("created_at desc").Find(&list)
    c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"list": list}})
}

func (ct *Activation) PendingAdd(c *gin.Context) {
    f := &struct {
        Id string `json:"id" binding:"required"`
    }{}
    if err := c.ShouldBindJSON(f); err != nil {
        c.JSON(http.StatusOK, gin.H{"code": 101, "message": "参数错误"})
        return
    }

    pe := service.AllService.PeerService.FindById(f.Id)
    if pe.RowId == 0 {
        c.JSON(http.StatusOK, gin.H{"code": 101, "message": "ID不存在，不能添加"})
        return
    }

    // 不在任何账号的地址簿里
    var abCnt int64
    service.DB.Model(&model.AddressBook{}).Where("id = ?", f.Id).Count(&abCnt)
    if abCnt > 0 {
        c.JSON(http.StatusOK, gin.H{"code": 101, "message": "设备已在地址簿中，不能添加"})
        return
    }

    var cnt int64
    service.DB.Model(&model.PendingDevice{}).Where("id = ? AND account = ?", f.Id, "").Count(&cnt)
    if cnt > 0 {
        c.JSON(http.StatusOK, gin.H{"code": 101, "message": "ID已存在"})
        return
    }
    service.DB.Create(&model.PendingDevice{Id: f.Id})
    c.JSON(http.StatusOK, gin.H{"code": 0})
}

func (ct *Activation) PendingDelete(c *gin.Context) {
    f := &struct {
        Id string `json:"id" binding:"required"`
    }{}
    if err := c.ShouldBindJSON(f); err != nil {
        c.JSON(http.StatusOK, gin.H{"code": 101, "message": "参数错误"})
        return
    }
    service.DB.Where("id = ? AND account = ?", f.Id, "").Delete(&model.PendingDevice{})
    c.JSON(http.StatusOK, gin.H{"code": 0})
}

func (ct *Activation) AssignedList(c *gin.Context) {
    list := make([]*model.PendingDevice, 0)
    service.DB.Where("account != ?", "").Order("assigned_at desc").Find(&list)
    c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"list": list}})
}

func (ct *Activation) PullOrphan(c *gin.Context) {
    n, _ := strconv.Atoi(c.Query("n"))
    ids := service.AllService.ActivationService.PullOrphanHost(n)
    c.String(http.StatusOK, strings.Join(ids, "\n"))
}