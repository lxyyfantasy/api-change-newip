package api

/*
*

	message LoginRequest {
	  string username = 1;
	  bytes password = 2;
	  string my_id = 4;
	  string my_name = 5;
	  OptionMessage option = 6;
	  oneof union {
	    FileTransfer file_transfer = 7;
	    PortForward port_forward = 8;
	  }
	  bool video_ack_required = 9;
	  uint64 session_id = 10;
	  string version = 11;
	  OSLogin os_login = 12;
	  string my_platform = 13;
	  bytes hwid = 14;
	}
*/

type DeviceInfoInLogin struct {
	Name string `json:"name" label:"name"`
	Os   string `json:"os" label:"os"`
	Type string `json:"type" label:"type"`
}

type LoginForm struct {
	AutoLogin  bool              `json:"autoLogin" label:"自动登录"`
	DeviceInfo DeviceInfoInLogin `json:"deviceInfo" label:"设备信息"`
	Id         string            `json:"id"  label:"id"`
	Type       string            `json:"type"  label:"type"`
	Uuid       string            `json:"uuid"  label:"uuid"`
	Username   string            `json:"username" validate:"required,gte=2,lte=32" label:"用户名"`
	Password   string            `json:"password,omitempty" validate:"gte=4,lte=32" label:"密码"`
}

type UserListQuery struct {
	Page       uint   `json:"page" form:"page" validate:"required" label:"页码"`
	PageSize   uint   `json:"pageSize" form:"pageSize" validate:"required" label:"每页数量"`
	Status     int    `json:"status" form:"status" label:"状态"`
	Accessible string `json:"accessible" form:"accessible"`
}

type PeerListQuery struct {
	Page       uint   `json:"page" form:"page" validate:"required" label:"页码"`
	PageSize   uint   `json:"pageSize" form:"pageSize" validate:"required" label:"每页数量"`
	Status     int    `json:"status" form:"status" label:"状态"`
	Accessible string `json:"accessible" form:"accessible"`
}

// RegisterForm 客户端 /register 注册请求
type RegisterForm struct {
	Username string `json:"username" validate:"required,gte=2,lte=32" label:"用户名"`
	Password string `json:"password" validate:"required,gte=4,lte=32" label:"密码"`
	Answer1  string `json:"answer1" label:"回答1"`
	Answer2  string `json:"answer2" label:"回答2"`
	Answer3  string `json:"answer3" label:"回答3"`
	InviteCode string `json:"invite_code" label:"邀请码"`
	RustdeskId string `json:"rustdesk_id" label:"RustDesk ID"`
}

// ChangePasswordForm 客户端 /change-password 修改密码请求
type ChangePasswordForm struct {
	Username    string `json:"username" validate:"required" label:"用户名"`
	OldPassword string `json:"old_password" validate:"required" label:"旧密码"`
	NewPassword string `json:"new_password" validate:"required,gte=4,lte=32" label:"新密码"`
}

// RecoverPasswordForm 找回密码请求
type RecoverPasswordForm struct {
	Username string `json:"username" validate:"required" label:"用户名"`
	Answer1  string `json:"answer1" validate:"required" label:"回答1"`
	Answer2  string `json:"answer2" validate:"required" label:"回答2"`
	Answer3  string `json:"answer3" validate:"required" label:"回答3"`
	NewPassword string `json:"new_password" validate:"required,gte=4,lte=32" label:"新密码"`
}

// AddDeviceForm 添加设备请求
type AddDeviceForm struct {
	Username   string `json:"username" validate:"required" label:"用户名"`
	DeviceCode string `json:"device_code" validate:"required" label:"添加设备码"`
}