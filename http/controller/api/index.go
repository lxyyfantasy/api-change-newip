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
	if err := c.ShouldBindJSON(info); err != nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	if info.Uuid == "" || info.DeviceCode == "" || info.Id == "" {
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	same := service.AllService.PeerService.FindByDeviceCode(info.DeviceCode)
	p := service.AllService.PeerService.FindById(info.Id)

	switch {
	case same.RowId > 0 && p.RowId > 0:
		// 1. 设备码有、ID 有 → 放行（更新在线/连接类型）
		if time.Now().Unix()-p.LastOnlineTime >= 30 {
			upp := &model.Peer{RowId: p.RowId, LastOnlineTime: time.Now().Unix(), LastOnlineIp: c.ClientIP(), ConnectType: info.ConnectType, DeviceCode: info.DeviceCode}
			service.AllService.PeerService.Update(upp)
		} else {
			if info.ConnectType != "" && info.ConnectType != p.ConnectType {
				service.AllService.PeerService.SetConnectTypeById(info.Id, info.ConnectType)
			}
		}
	case same.RowId > 0 && p.RowId == 0:
		// 2. 设备码有、ID 没有 → 覆盖设备码行的 ID
		service.DB.Model(&model.Peer{}).Where("row_id = ?", same.RowId).Update("id", info.Id)
		service.DB.Model(&model.AddressBook{}).Where("id = ?", same.Id).Update("id", info.Id)
	case same.RowId == 0 && p.RowId > 0:
		// 3. 设备码没、ID 有 → 把设备码写到 ID 那行
		service.DB.Model(&model.Peer{}).Where("row_id = ?", p.RowId).Update("device_code", info.DeviceCode)
	case same.RowId == 0 && p.RowId == 0:
		// 4. 都没有 → 新增
		service.AllService.PeerService.Create(&model.Peer{
			Id: info.Id, Uuid: info.Uuid, DeviceCode: info.DeviceCode,
			ConnectType: info.ConnectType, LastOnlineTime: time.Now().Unix(),
			LastOnlineIp: c.ClientIP(),
		})
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
