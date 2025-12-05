package services

import (
	"context"

	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/userconfigs"
)

// UserTicketLogService 工单日志服务
type UserTicketLogService struct {
	BaseService
}

// CreateUserTicketLog 创建日志
func (this *UserTicketLogService) CreateUserTicketLog(ctx context.Context, req *pb.CreateUserTicketLogRequest) (*pb.CreateUserTicketLogResponse, error) {
	adminId, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if len(req.Status) > 0 {
		switch req.Status {
		case userconfigs.UserTicketStatusNone, userconfigs.UserTicketStatusSolved, userconfigs.UserTicketStatusClosed:
		default:
			return nil, errors.New("invalid status '" + req.Status + "'")
		}
	}

	var tx = this.NullTx()

	if userId > 0 {
		allowed, err := models.SharedUserTicketDAO.CheckUserTicket(tx, userId, req.UserTicketId)
		if err != nil {
			return nil, err
		}
		if !allowed {
			return nil, this.PermissionError()
		}
	}

	logId, err := models.SharedUserTicketLogDAO.CreateUserTicketLog(tx, req.UserTicketId, adminId, userId, req.Status, req.Comment, false)
	if err != nil {
		return nil, err
	}

	return &pb.CreateUserTicketLogResponse{UserTicketLogId: logId}, nil
}

// DeleteUserTicketLog 删除日志
func (this *UserTicketLogService) DeleteUserTicketLog(ctx context.Context, req *pb.DeleteUserTicketLogRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedUserTicketLogDAO.DeleteUserTicketLog(tx, req.UserTicketLogId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// CountUserTicketLogs 查询日志数量
func (this *UserTicketLogService) CountUserTicketLogs(ctx context.Context, req *pb.CountUserTicketLogsRequest) (*pb.RPCCountResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		allowed, err := models.SharedUserTicketDAO.CheckUserTicket(tx, userId, req.UserTicketId)
		if err != nil {
			return nil, err
		}
		if !allowed {
			return nil, this.PermissionError()
		}
	}

	count, err := models.SharedUserTicketLogDAO.CountUserTicketLogs(tx, req.UserTicketId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListUserTicketLogs 列出单页日志
func (this *UserTicketLogService) ListUserTicketLogs(ctx context.Context, req *pb.ListUserTicketLogsRequest) (*pb.ListUserTicketLogsResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		allowed, err := models.SharedUserTicketDAO.CheckUserTicket(tx, userId, req.UserTicketId)
		if err != nil {
			return nil, err
		}
		if !allowed {
			return nil, this.PermissionError()
		}
	}

	logs, err := models.SharedUserTicketLogDAO.ListUserTicketLogs(tx, req.UserTicketId, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	var cacheMap = utils.NewCacheMap()
	result := make([]*pb.UserTicketLog, 0, len(logs))
	for _, log := range logs {
		pbLog, err := buildPBUserTicketLog(tx, log, cacheMap)
		if err != nil {
			return nil, err
		}
		if pbLog != nil {
			result = append(result, pbLog)
		}
	}

	return &pb.ListUserTicketLogsResponse{UserTicketLogs: result}, nil
}
