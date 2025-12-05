package jsonapi

import (
	"github.com/TeaOSLab/EdgeAdmin/internal/rpc"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// ConsoleStatAction 获取控制台统计数据
type ConsoleStatAction struct {
	BaseAPIAction
}

func (this *ConsoleStatAction) RunGet(params struct{}) {
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

	// 获取统计数据
	rpcClient, err := rpc.SharedRPC()
	if err != nil {
		this.FailMsg("服务连接失败")
		return
	}

	// 获取仪表盘数据
	dashboardResp, err := rpcClient.AdminRPC().ComposeAdminDashboard(rpcClient.Context(tokenInfo.AdminId), &pb.ComposeAdminDashboardRequest{})

	var countServers int64 = 0
	var countNodes int64 = 0
	var countOfflineNodes int64 = 0
	var countUsers int64 = 0
	var countClusters int64 = 0
	var dailyTrafficStats []map[string]interface{}
	var hourlyTrafficStats []map[string]interface{}
	var topDomainStats []map[string]interface{}

	if err == nil && dashboardResp != nil {
		countServers = dashboardResp.CountServers
		countNodes = dashboardResp.CountNodes
		countOfflineNodes = dashboardResp.CountOfflineNodes
		countUsers = dashboardResp.CountUsers
		countClusters = dashboardResp.CountNodeClusters

		// 转换每日流量统计
		for _, stat := range dashboardResp.DailyTrafficStats {
			dailyTrafficStats = append(dailyTrafficStats, map[string]interface{}{
				"day":                 stat.Day,
				"bytes":               stat.Bytes,
				"cachedBytes":         stat.CachedBytes,
				"countRequests":       stat.CountRequests,
				"countCachedRequests": stat.CountCachedRequests,
				"countAttackRequests": stat.CountAttackRequests,
				"attackBytes":         stat.AttackBytes,
			})
		}

		// 转换每小时流量统计
		for _, stat := range dashboardResp.HourlyTrafficStats {
			hourlyTrafficStats = append(hourlyTrafficStats, map[string]interface{}{
				"hour":                stat.Hour,
				"bytes":               stat.Bytes,
				"cachedBytes":         stat.CachedBytes,
				"countRequests":       stat.CountRequests,
				"countCachedRequests": stat.CountCachedRequests,
				"countAttackRequests": stat.CountAttackRequests,
				"attackBytes":         stat.AttackBytes,
			})
		}

		// 转换域名排行
		for _, stat := range dashboardResp.TopDomainStats {
			topDomainStats = append(topDomainStats, map[string]interface{}{
				"serverId":      stat.ServerId,
				"domain":        stat.Domain,
				"countRequests": stat.CountRequests,
				"bytes":         stat.Bytes,
			})
		}
	}

	this.SuccessData(map[string]interface{}{
		"countServers":       countServers,
		"countNodes":         countNodes,
		"countOfflineNodes":  countOfflineNodes,
		"countUsers":         countUsers,
		"countClusters":      countClusters,
		"dailyTrafficStats":  dailyTrafficStats,
		"hourlyTrafficStats": hourlyTrafficStats,
		"topDomainStats":     topDomainStats,
	})
}
