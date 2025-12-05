package jsonapi

import (
	"github.com/TeaOSLab/EdgeAdmin/internal/rpc"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// MemberInfoAction 获取当前用户信息
type MemberInfoAction struct {
	BaseAPIAction
}

func (this *MemberInfoAction) RunGet(params struct{}) {
	token := this.GetToken()
	if token == "" {
		this.Unauthorized("请先登录")
		return
	}

	tokenInfo, err := ValidateToken(token)
	if err != nil {
		this.Unauthorized("登录已过期，请重新登录")
		return
	}

	// 获取管理员详细信息
	rpcClient, err := rpc.SharedRPC()
	if err != nil {
		this.FailMsg("服务连接失败")
		return
	}

	adminResp, err := rpcClient.AdminRPC().FindEnabledAdmin(rpcClient.Context(tokenInfo.AdminId), &pb.FindEnabledAdminRequest{
		AdminId: tokenInfo.AdminId,
	})
	if err != nil || adminResp.Admin == nil {
		this.Unauthorized("用户不存在")
		return
	}

	admin := adminResp.Admin

	this.SuccessData(map[string]interface{}{
		"id":          admin.Id,
		"username":    admin.Username,
		"realName":    admin.Fullname,
		"avatar":      "",
		"permissions": []string{"*"},
		"deptName":    "管理员",
		"deptType":    "admin",
		"roleName":    "超级管理员",
	})
}
