package service

import (
	"errors"
	"github.com/lejianwen/rustdesk-api/v2/model"
)

type ServerInfoService struct {
}

type ServerInfoWithCount struct {
	model.ServerInfo
	DeviceCount int64 `json:"device_count"`
}

// List 服务器列表, 带设备数量
func (s *ServerInfoService) List() []*ServerInfoWithCount {
	var list []*model.ServerInfo
	DB.Order("id asc").Find(&list)
	res := make([]*ServerInfoWithCount, 0, len(list))
	for _, item := range list {
		var cnt int64
		DB.Model(&model.Peer{}).Where("server_ip = ?", item.Ip).Count(&cnt)
		res = append(res, &ServerInfoWithCount{
			ServerInfo:  *item,
			DeviceCount: cnt,
		})
	}
	return res
}

// Create 添加服务器
func (s *ServerInfoService) Create(ip, bandwidth string) error {
	var cnt int64
	DB.Model(&model.ServerInfo{}).Where("ip = ?", ip).Count(&cnt)
	if cnt > 0 {
		return errors.New("该IP已存在")
	}
	return DB.Create(&model.ServerInfo{Ip: ip, Bandwidth: bandwidth}).Error
}

// DeleteById 删除服务器, 并清空peers表对应server_ip
func (s *ServerInfoService) DeleteById(id uint) error {
	item := &model.ServerInfo{}
	DB.Where("id = ?", id).First(item)
	if item.Id == 0 {
		return nil
	}
	if err := DB.Delete(item).Error; err != nil {
		return err
	}
	//只清空server_ip格子, 不删行
	return DB.Model(&model.Peer{}).Where("server_ip = ?", item.Ip).Update("server_ip", "").Error
}