package ecode

//nolint: golint
var (
	// Common errors
	Success                = &ErrCode{Code: 0, Msg: "Success"}
	InternalServerError    = &ErrCode{Code: 10001, Msg: "系统错误"}
	ErrParam               = &ErrCode{Code: 10003, Msg: "参数有误"}
	ErrIDNotFound          = &ErrCode{Code: 10004, Msg: "ID不存在"}
	ErrCertInUse           = &ErrCode{Code: 10005, Msg: "删除失败，有站点正在使用此证书"}
	ErrAtLeastOneOrigin    = &ErrCode{Code: 10006, Msg: "删除源站失败，至少需要一个源站"}
	ErrRecordAlreadyExists = &ErrCode{Code: 10007, Msg: "记录已存在"}

	ErrIPInvalid           = &ErrCode{Code: 10009, Msg: "无效的IP地址或CIDR"}
	ErrIPv6NotSupportedYet = &ErrCode{Code: 10010, Msg: "暂不支持IPv6地址或CIDR"}
	ErrIPAlreadyExisted    = &ErrCode{Code: 10011, Msg: "IP地址或CIDR已经存在"}

	ErrSiteNotFound = &ErrCode{Code: 10013, Msg: "站点不存在"}
	ErrUpdate       = &ErrCode{Code: 10014, Msg: "更新失败"}

	ErrCertNotFound           = &ErrCode{Code: 20001, Msg: "该证书未找到"}
	ErrCertInvalid            = &ErrCode{Code: 20002, Msg: "证书格式不正确"}
	ErrCertPriKeyInvalid      = &ErrCode{Code: 20003, Msg: "证书私钥格式不正确"}
	ErrIPorCIDREmpty          = &ErrCode{Code: 20004, Msg: "请输入IP或CIDR"}
	ErrUsernameOrPwdIncorrect = &ErrCode{Code: 20005, Msg: "用户名或密码不正确"}
)
