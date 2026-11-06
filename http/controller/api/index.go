package api

import (
    "github.com/gin-gonic/gin"
    requstform "github.com/lejianwen/rustdesk-api/v2/http/request/api"
    "github.com/lejianwen/rustdesk-api/v2/http/response"
    "github.com/lejianwen/rustdesk-api/v2/model"
    "github.com/lejianwen/rustdesk-api/v2/service"
    "net/http"
    "time"
)

type Index struct {
}

// Index 首页
// @Tags 首页
// @Summary 首页
// @Description 首页
// @Accept  json
// @Produce  json
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router / [get]
func (i *Index) Index(c *gin.Context) {
    response.Success(
        c,
        "Hello Gwen",
    )
}

// Heartbeat 心跳
// @Tags 首页
// @Summary 心跳
// @Description 心跳
// @Accept  json
// @Produce  json
// @Success 200 {object} nil
// @Failure 500 {object} response.Response
// @Router /heartbeat [post]
func (i *Index) Heartbeat(c *gin.Context) {
    info := &requstform.PeerInfoInHeartbeat{}
    err := c.ShouldBindJSON(info)
    if err != nil {
        c.JSON(http.StatusOK, gin.H{})
        return
    }
    if info.Uuid == "" {
        c.JSON(http.StatusOK, gin.H{})
        return
    }
    if info.DeviceCode != "" {
        // 同一台设备（Android ID 相同）不算重复
        sameDevice := service.AllService.PeerService.FindByDeviceCode(info.DeviceCode)
        if sameDevice.RowId > 0 && sameDevice.Id == info.Id {
            // 同设备同 ID，老设备，正常
        } else {
            // 不是同一台设备，再看 ID 是否已被占用
            if p := service.AllService.PeerService.FindById(info.Id); p.RowId > 0 && p.DeviceCode != info.DeviceCode {
                c.JSON(http.StatusOK, gin.H{"error": "id_duplicate"})
                return
            }
        }
    }
    peer := service.AllService.PeerService.FindById(info.Id)
    if peer == nil || peer.RowId == 0 {
        service.AllService.PeerService.Create(&model.Peer{
            Id:             info.Id,
            Uuid:           info.Uuid,
            DeviceCode:     info.DeviceCode,
            ConnectType:    info.ConnectType,
            LastOnlineTime: time.Now().Unix(),
            LastOnlineIp:   c.ClientIP(),
        })
        c.JSON(http.StatusOK, gin.H{})
        return
    }
    // 如果在30s以内则不更新
    if time.Now().Unix()-peer.LastOnlineTime >= 30 {
        upp := &model.Peer{RowId: peer.RowId, LastOnlineTime: time.Now().Unix(), LastOnlineIp: c.ClientIP(), ConnectType: info.ConnectType, DeviceCode: info.DeviceCode}
        service.AllService.PeerService.Update(upp)
    } else {
        if info.ConnectType != "" && info.ConnectType != peer.ConnectType {
            service.AllService.PeerService.SetConnectTypeById(info.Id, info.ConnectType)
        }
        service.AllService.PeerService.SetDeviceCodeById(info.Id, info.DeviceCode)
    }
    c.JSON(http.StatusOK, gin.H{})
}

// Version 版本
// @Tags 首页
// @Summary 版本
// @Description 版本
// @Accept  json
// @Produce  json
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /version [get]
func (i *Index) Version(c *gin.Context) {
    //读取resources/version文件
    v := service.AllService.AppService.GetAppVersion()
    response.Success(
        c,
        v,
    )
}
