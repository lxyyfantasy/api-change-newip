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
	// 1. 参数有空拒绝
	if info.Uuid == "" || info.DeviceCode == "" || info.Id == "" {
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	if same := service.AllService.PeerService.FindByDeviceCode(info.DeviceCode); same.RowId > 0 {
		// 2/3. 设备码查到
		if p := service.AllService.PeerService.FindById(info.Id); p.RowId > 0 {
			// 3. ID 也查到，放行
			if time.Now().Unix()-p.LastOnlineTime >= 30 {
				upp := &model.Peer{RowId: p.RowId, LastOnlineTime: time.Now().Unix(), LastOnlineIp: c.ClientIP(), ConnectType: info.ConnectType, DeviceCode: info.DeviceCode}
				service.AllService.PeerService.Update(upp)
			} else {
				if info.ConnectType != "" && info.ConnectType != p.ConnectType {
					service.AllService.PeerService.SetConnectTypeById(info.Id, info.ConnectType)
				}
				service.AllService.PeerService.SetDeviceCodeById(info.Id, info.DeviceCode)
			}
			c.JSON(http.StatusOK, gin.H{})
			return
		}
		// 2. ID 没查到，覆盖 ID
		service.DB.Model(&model.Peer{}).Where("row_id = ?", same.RowId).Update("id", info.Id)
		service.DB.Model(&model.AddressBook{}).Where("id = ?", same.Id).Update("id", info.Id)
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	// 设备码没查到
	if p := service.AllService.PeerService.FindById(info.Id); p.RowId > 0 {
		// 5. ID 查到，拒绝
		c.JSON(http.StatusOK, gin.H{"error": "id_duplicate"})
		return
	}
	// 4. ID 没查到，新建
	service.AllService.PeerService.Create(&model.Peer{
		Id: info.Id, Uuid: info.Uuid, DeviceCode: info.DeviceCode,
		ConnectType: info.ConnectType, LastOnlineTime: time.Now().Unix(),
		LastOnlineIp: c.ClientIP(),
	})
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
