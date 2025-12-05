package tickets

import (
	"github.com/TeaOSLab/EdgeAdmin/internal/web/actions/actionutils"
	"github.com/TeaOSLab/EdgeAdmin/internal/web/helpers"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/userconfigs"
)

type CreateLogAction struct {
	actionutils.ParentAction
}

func (this *CreateLogAction) Init() {
	this.Nav("", "ticket", "")
}

func (this *CreateLogAction) RunPost(params struct {
	TicketId int64
	Status   string
	Comment  string
	Auth     *helpers.UserShouldAuth
	Must     *actionutils.Must
}) {
	adminCtx := this.AdminContext()

	if len(params.Status) > 0 {
		switch params.Status {
		case userconfigs.UserTicketStatusNone, userconfigs.UserTicketStatusSolved, userconfigs.UserTicketStatusClosed:
		default:
			this.Fail("非法状态")
			return
		}
	}

	_, err := this.RPC().UserTicketLogService().CreateUserTicketLog(adminCtx, &pb.CreateUserTicketLogRequest{
		UserTicketId: params.TicketId,
		Status:       params.Status,
		Comment:      params.Comment,
	})
	if err != nil {
		this.ErrorPage(err)
		return
	}

	this.Success()
}
