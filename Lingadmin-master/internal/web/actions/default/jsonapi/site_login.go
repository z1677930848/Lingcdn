package jsonapi

import (
	"github.com/TeaOSLab/EdgeAdmin/internal/rpc"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// SiteLoginAction 登录接口
type SiteLoginAction struct {
	BaseAPIAction
}

func (this *SiteLoginAction) RunPost(params struct {
	Username string
	Password string
	Cid      string
	Code     string
}) {
	// 验证参数
	if params.Username == "" {
		this.FailMsg("请输入用户名")
		return
	}
	if params.Password == "" {
		this.FailMsg("请输入密码")
		return
	}

	// 验证验证码（如果提供了）
	if params.Cid != "" {
		if !ValidateCaptcha(params.Cid, params.Code) {
			this.FailMsg("验证码错误")
			return
		}
	}

	// 调用 RPC 验证登录
	rpcClient, err := rpc.SharedRPC()
	if err != nil {
		this.FailMsg("服务连接失败")
		return
	}

	resp, err := rpcClient.AdminRPC().LoginAdmin(rpcClient.Context(0), &pb.LoginAdminRequest{
		Username: params.Username,
		Password: params.Password,
	})
	if err != nil {
		this.FailMsg("用户名或密码错误")
		return
	}

	adminId := resp.AdminId
	if adminId <= 0 {
		this.FailMsg("用户名或密码错误")
		return
	}

	// 生成 JWT Token
	token, expireAt, err := GenerateToken(adminId, params.Username)
	if err != nil {
		this.FailMsg("登录失败，请重试")
		return
	}

	// 获取管理员信息
	adminResp, err := rpcClient.AdminRPC().FindEnabledAdmin(rpcClient.Context(adminId), &pb.FindEnabledAdminRequest{
		AdminId: adminId,
	})
	if err != nil || adminResp.Admin == nil {
		this.FailMsg("获取用户信息失败")
		return
	}

	admin := adminResp.Admin

	this.SuccessData(map[string]interface{}{
		"token":       token,
		"sid":         token, // 兼容 hotgo 前端
		"expireAt":    expireAt,
		"id":          adminId,
		"username":    admin.Username,
		"realName":    admin.Fullname,
		"avatar":      "",
		"permissions": []string{"*"}, // 简化权限，后续可扩展
	})
}
