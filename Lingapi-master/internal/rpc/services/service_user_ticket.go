package services

import (
	"context"

	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// UserTicketService 工单服务
type UserTicketService struct {
	BaseService
}

// CreateUserTicket 创建工单
func (this *UserTicketService) CreateUserTicket(ctx context.Context, req *pb.CreateUserTicketRequest) (*pb.CreateUserTicketResponse, error) {
	userId, err := this.ValidateUserNode(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if req.UserTicketCategoryId > 0 {
		category, err := models.SharedUserTicketCategoryDAO.FindEnabledUserTicketCategory(tx, req.UserTicketCategoryId)
		if err != nil {
			return nil, err
		}
		if category == nil || category.IsOn != 1 {
			return nil, errors.New("invalid ticket category")
		}
	}

	ticketId, err := models.SharedUserTicketDAO.CreateUserTicket(tx, userId, req.UserTicketCategoryId, req.Subject, req.Body)
	if err != nil {
		return nil, err
	}
	return &pb.CreateUserTicketResponse{UserTicketId: ticketId}, nil
}

// UpdateUserTicket 修改工单
func (this *UserTicketService) UpdateUserTicket(ctx context.Context, req *pb.UpdateUserTicketRequest) (*pb.RPCSuccess, error) {
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

	if req.UserTicketCategoryId > 0 {
		category, err := models.SharedUserTicketCategoryDAO.FindEnabledUserTicketCategory(tx, req.UserTicketCategoryId)
		if err != nil {
			return nil, err
		}
		if category == nil || category.IsOn != 1 {
			return nil, errors.New("invalid ticket category")
		}
	}

	err = models.SharedUserTicketDAO.UpdateUserTicket(tx, req.UserTicketId, req.UserTicketCategoryId, req.Subject, req.Body)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// DeleteUserTicket 删除工单
func (this *UserTicketService) DeleteUserTicket(ctx context.Context, req *pb.DeleteUserTicketRequest) (*pb.RPCSuccess, error) {
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

	err = models.SharedUserTicketDAO.DeleteUserTicket(tx, req.UserTicketId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// CountUserTickets 计算工单数量
func (this *UserTicketService) CountUserTickets(ctx context.Context, req *pb.CountUserTicketsRequest) (*pb.RPCCountResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}
	if userId > 0 {
		req.UserId = userId
	}

	var tx = this.NullTx()
	count, err := models.SharedUserTicketDAO.CountUserTickets(tx, req.UserId, req.UserTicketCategoryId, req.Status)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListUserTickets 列出单页工单
func (this *UserTicketService) ListUserTickets(ctx context.Context, req *pb.ListUserTicketsRequest) (*pb.ListUserTicketsResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}
	if userId > 0 {
		req.UserId = userId
	}

	var tx = this.NullTx()
	tickets, err := models.SharedUserTicketDAO.ListUserTickets(tx, req.UserId, req.UserTicketCategoryId, req.Status, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	var cacheMap = utils.NewCacheMap()
	result := make([]*pb.UserTicket, 0, len(tickets))
	for _, ticket := range tickets {
		pbTicket, err := buildPBUserTicket(tx, ticket, cacheMap)
		if err != nil {
			return nil, err
		}
		if pbTicket != nil {
			result = append(result, pbTicket)
		}
	}
	return &pb.ListUserTicketsResponse{UserTickets: result}, nil
}

// FindUserTicket 查找单个工单
func (this *UserTicketService) FindUserTicket(ctx context.Context, req *pb.FindUserTicketRequest) (*pb.FindUserTicketResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	ticket, err := models.SharedUserTicketDAO.FindEnabledUserTicket(tx, req.UserTicketId)
	if err != nil {
		return nil, err
	}
	if ticket == nil {
		return &pb.FindUserTicketResponse{UserTicket: nil}, nil
	}

	if userId > 0 && int64(ticket.UserId) != userId {
		return nil, this.PermissionError()
	}

	pbTicket, err := buildPBUserTicket(tx, ticket, utils.NewCacheMap())
	if err != nil {
		return nil, err
	}

	return &pb.FindUserTicketResponse{UserTicket: pbTicket}, nil
}
