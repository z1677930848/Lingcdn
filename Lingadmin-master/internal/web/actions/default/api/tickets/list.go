package tickets

import (
	"github.com/TeaOSLab/EdgeAdmin/internal/web/actions/actionutils"
	"github.com/TeaOSLab/EdgeAdmin/internal/web/helpers"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/userconfigs"
	"github.com/iwind/TeaGo/maps"
)

type ListAction struct {
	actionutils.ParentAction
}

func (this *ListAction) Init() {
	this.Nav("", "ticket", "")
}

func (this *ListAction) RunGet(params struct {
	Status     string
	UserId     int64
	CategoryId int64
	Offset     int64
	Size       int64
	Auth       *helpers.UserShouldAuth
}) {
	adminCtx := this.AdminContext()

	if params.Size <= 0 {
		params.Size = 20
	}

	countResp, err := this.RPC().UserTicketService().CountUserTickets(adminCtx, &pb.CountUserTicketsRequest{
		UserId:               params.UserId,
		UserTicketCategoryId: params.CategoryId,
		Status:               params.Status,
	})
	if err != nil {
		this.ErrorPage(err)
		return
	}

	ticketsResp, err := this.RPC().UserTicketService().ListUserTickets(adminCtx, &pb.ListUserTicketsRequest{
		UserId:               params.UserId,
		UserTicketCategoryId: params.CategoryId,
		Status:               params.Status,
		Offset:               params.Offset,
		Size:                 params.Size,
	})
	if err != nil {
		this.ErrorPage(err)
		return
	}

	list := []maps.Map{}
	for _, ticket := range ticketsResp.UserTickets {
		var category maps.Map = nil
		if ticket.UserTicketCategory != nil {
			category = maps.Map{
				"id":   ticket.UserTicketCategory.Id,
				"name": ticket.UserTicketCategory.Name,
				"isOn": ticket.UserTicketCategory.IsOn,
			}
		}
		var user maps.Map = nil
		if ticket.User != nil {
			user = maps.Map{
				"id":       ticket.User.Id,
				"username": ticket.User.Username,
				"fullname": ticket.User.Fullname,
				"isOn":     ticket.User.IsOn,
			}
		}
		var latestLog maps.Map = nil
		if ticket.LatestUserTicketLog != nil {
			latestLog = maps.Map{
				"id":        ticket.LatestUserTicketLog.Id,
				"adminId":   ticket.LatestUserTicketLog.AdminId,
				"userId":    ticket.LatestUserTicketLog.UserId,
				"status":    ticket.LatestUserTicketLog.Status,
				"comment":   ticket.LatestUserTicketLog.Comment,
				"createdAt": ticket.LatestUserTicketLog.CreatedAt,
			}
			if ticket.LatestUserTicketLog.Admin != nil {
				latestLog["admin"] = maps.Map{
					"id":       ticket.LatestUserTicketLog.Admin.Id,
					"username": ticket.LatestUserTicketLog.Admin.Username,
					"fullname": ticket.LatestUserTicketLog.Admin.Fullname,
				}
			}
			if ticket.LatestUserTicketLog.User != nil {
				latestLog["user"] = maps.Map{
					"id":       ticket.LatestUserTicketLog.User.Id,
					"username": ticket.LatestUserTicketLog.User.Username,
					"fullname": ticket.LatestUserTicketLog.User.Fullname,
				}
			}
		}

		list = append(list, maps.Map{
			"id":         ticket.Id,
			"categoryId": ticket.CategoryId,
			"userId":     ticket.UserId,
			"subject":    ticket.Subject,
			"status":     ticket.Status,
			"statusName": userconfigs.UserTicketStatusName(ticket.Status),
			"createdAt":  ticket.CreatedAt,
			"lastLogAt":  ticket.LastLogAt,
			"category":   category,
			"user":       user,
			"latestLog":  latestLog,
		})
	}

	this.Data["tickets"] = list
	this.Data["count"] = countResp.Count
	this.Data["size"] = params.Size
	this.Data["offset"] = params.Offset
	this.Success()
}
