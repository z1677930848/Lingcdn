package pay

import (
	"context"

	v1 "hotgo/api/api/pay/v1"
	"hotgo/internal/consts"
	"hotgo/internal/library/response"
	"hotgo/internal/model/input/payin"
	"hotgo/internal/service"

	"github.com/gogf/gf/v2/frame/g"
)

// NotifyYiPay 易支付回调
func (c *ControllerV1) NotifyYiPay(ctx context.Context, req *v1.NotifyYiPayReq) (res *v1.NotifyYiPayRes, err error) {
	if _, err = service.Pay().Notify(ctx, &payin.PayNotifyInp{PayType: consts.PayTypeYiPay}); err != nil {
		return
	}

	response.RText(g.RequestFromCtx(ctx), "success")
	return
}
