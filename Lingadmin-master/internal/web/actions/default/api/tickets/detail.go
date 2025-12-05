package tickets

import (
	"github.com/TeaOSLab/EdgeAdmin/internal/web/actions/actionutils"
	"github.com/TeaOSLab/EdgeAdmin/internal/web/helpers"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/userconfigs"
	"github.com/iwind/TeaGo/maps"
)

type DetailAction struct {
	actionutils.ParentAction
}

func (this *DetailAction) Init() {
	this.Nav("", "ticket", "")
}

func (this *DetailAction) RunGet(params struct {
	TicketId int64
	Auth     *helpers.UserShouldAuth
}) {
	adminCtx := this.AdminContext()

	ticketResp, err := this.RPC().UserTicketService().FindUserTicket(adminCtx, &pb.FindUserTicketRequest{UserTicketId: params.TicketId})
	if err != nil {
		this.ErrorPage(err)
		return
	}
	if ticketResp.UserTicket == nil {
		this.NotFound("ticket", params.TicketId)
		return
	}
	ticket := ticketResp.UserTicket

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

	logsResp, err := this.RPC().UserTicketLogService().ListUserTicketLogs(adminCtx, &pb.ListUserTicketLogsRequest{
		UserTicketId: params.TicketId,
		Offset:       0,
		Size:         50,
	})
	if err != nil {
		this.ErrorPage(err)
		return
	}

	logs := []maps.Map{}
	for _, log := range logsResp.UserTicketLogs {
		item := maps.Map{
			"id":         log.Id,
			"adminId":    log.AdminId,
			"userId":     log.UserId,
			"status":     log.Status,
			"statusName": userconfigs.UserTicketStatusName(log.Status),
			"comment":    log.Comment,
			"createdAt":  log.CreatedAt,
			"isReadonly": log.IsReadonly,
		}
		if log.Admin != nil {
			item["admin"] = maps.Map{
				"id":       log.Admin.Id,
				"username": log.Admin.Username,
				"fullname": log.Admin.Fullname,
			}
		}
		if log.User != nil {
			item["user"] = maps.Map{
				"id":       log.User.Id,
				"username": log.User.Username,
				"fullname": log.User.Fullname,
			}
		}
		logs = append(logs, item)
	}

	this.Data["ticket"] = maps.Map{
		"id":         ticket.Id,
		"categoryId": ticket.CategoryId,
		"userId":     ticket.UserId,
		"subject":    ticket.Subject,
		"body":       ticket.Body,
		"status":     ticket.Status,
		"statusName": userconfigs.UserTicketStatusName(ticket.Status),
		"createdAt":  ticket.CreatedAt,
		"lastLogAt":  ticket.LastLogAt,
		"category":   category,
		"user":       user,
	}
	this.Data["logs"] = logs
	this.Success()
}
