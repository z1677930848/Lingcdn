package tickets

import (
	"github.com/TeaOSLab/EdgeAdmin/internal/web/actions/actionutils"
	"github.com/TeaOSLab/EdgeAdmin/internal/web/helpers"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/maps"
)

type CategoriesAction struct {
	actionutils.ParentAction
}

func (this *CategoriesAction) Init() {
	this.Nav("", "ticket", "")
}

func (this *CategoriesAction) RunGet(params struct {
	All  bool
	Auth *helpers.UserShouldAuth
}) {
	adminCtx := this.AdminContext()

	var resp *pb.FindAllUserTicketCategoriesResponse
	var err error
	if params.All {
		resp, err = this.RPC().UserTicketCategoryService().FindAllUserTicketCategories(adminCtx, &pb.FindAllUserTicketCategoriesRequest{})
	} else {
		availableResp, errFind := this.RPC().UserTicketCategoryService().FindAllAvailableUserTicketCategories(adminCtx, &pb.FindAllAvailableUserTicketCategoriesRequest{})
		err = errFind
		if err == nil {
			// 转换为统一结构
			resp = &pb.FindAllUserTicketCategoriesResponse{UserTicketCategories: availableResp.UserTicketCategories}
		}
	}
	if err != nil {
		this.ErrorPage(err)
		return
	}

	result := []maps.Map{}
	for _, c := range resp.UserTicketCategories {
		result = append(result, maps.Map{
			"id":   c.Id,
			"name": c.Name,
			"isOn": c.IsOn,
		})
	}
	this.Data["categories"] = result
	this.Success()
}
