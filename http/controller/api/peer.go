package api

import (
    "fmt"
    "github.com/gin-gonic/gin"
    "github.com/gin-gonic/gin/binding"
    requstform "github.com/lejianwen/rustdesk-api/v2/http/request/api"
    "github.com/lejianwen/rustdesk-api/v2/http/response"
    "github.com/lejianwen/rustdesk-api/v2/service"
	"github.com/lejianwen/rustdesk-api/v2/model"
    "net/http"
)

type Peer struct {
}

// SysInfo
// @Tags System
// @Summary 提交系统信息
// @Description 提交系统信息
// @Accept  json
// @Produce  json
// @Param body body requstform.PeerForm true "系统信息表单"
// @Success 200 {string} string "SYSINFO_UPDATED,ID_NOT_FOUND"
// @Failure 500 {object} response.ErrorResponse
// @Router /sysinfo [post]
func (p *Peer) SysInfo(c *gin.Context) {
	f := &requstform.PeerForm{}
	err := c.ShouldBindBodyWith(f, binding.JSON)
	if err != nil {
		response.Error(c, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	fpe := f.ToPeer()
	// device_code 空：不建不写，等心跳带 device_code 再处理
	if fpe.DeviceCode == "" {
		c.String(http.StatusOK, "SYSINFO_UPDATED")
		return
	}
	pe := service.AllService.PeerService.FindById(f.Id)
	if pe.RowId == 0 {
		// 先按 device_code 查，避免同设备建两行
		if same := service.AllService.PeerService.FindByDeviceCode(fpe.DeviceCode); same.RowId > 0 {
			service.DB.Model(&model.Peer{}).Where("row_id = ?", same.RowId).Update("id", fpe.Id)
			service.DB.Model(&model.AddressBook{}).Where("id = ?", same.Id).Update("id", fpe.Id)
			c.String(http.StatusOK, "SYSINFO_UPDATED")
			return
		}
		pe = fpe
		pe.UserId = service.AllService.UserService.FindLatestUserIdFromLoginLogByUuid(pe.Uuid, pe.Id)
		err = service.AllService.PeerService.Create(pe)
		if err != nil {
			response.Error(c, response.TranslateMsg(c, "OperationFailed")+err.Error())
			return
		}
	} else {
		if pe.DeviceCode != "" && pe.DeviceCode != fpe.DeviceCode {
			c.String(http.StatusOK, "SYSINFO_UPDATED")
			return
		}
		fpe.RowId = pe.RowId
		fpe.UserId = pe.UserId
		err = service.AllService.PeerService.Update(fpe)
		if err != nil {
			response.Error(c, response.TranslateMsg(c, "OperationFailed")+err.Error())
			return
		}
	}
	c.String(http.StatusOK, "SYSINFO_UPDATED")
}

// SysInfoVer
// @Tags System
// @Summary 获取系统版本信息
// @Description 获取系统版本信息
// @Accept  json
// @Produce  json
// @Success 200 {string} string ""
// @Failure 500 {object} response.ErrorResponse
// @Router /sysinfo_ver [post]
func (p *Peer) SysInfoVer(c *gin.Context) {
    //读取resources/version文件
    v := service.AllService.AppService.GetAppVersion()
    // 加上启动时间，方便client上传信息
    v = fmt.Sprintf("%s\n%s", v, service.AllService.AppService.GetStartTime())
    c.String(http.StatusOK, v)
}

// ServerIp
// @Tags System
// @Summary 获取设备服务器ip
// @Description 根据设备ID获取归属服务器ip
// @Accept  json
// @Produce  json
// @Param id query string true "设备ID"
// @Success 200 {object} response.Response{data=string}
// @Failure 500 {object} response.ErrorResponse
// @Router /peer/server_ip [get]
func (p *Peer) ServerIp(c *gin.Context) {
    id := c.Query("id")
    if id == "" {
        response.Error(c, response.TranslateMsg(c, "ParamsError"))
        return
    }
    ip := service.AllService.PeerService.GetServerIpById(id)
    response.Success(c, ip)
}