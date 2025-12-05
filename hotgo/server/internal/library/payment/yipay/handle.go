// Package yipay
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2023 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE
package yipay

import (
	"context"
	"crypto/md5"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"

	"hotgo/internal/consts"
	"hotgo/internal/model"
	"hotgo/internal/model/input/payin"
)

type yiPay struct {
	config *model.PayConfig
}

func New(config *model.PayConfig) *yiPay {
	return &yiPay{config: config}
}

// Refund 易支付大部分站点无官方退款接口，直接返回不支持。
func (h *yiPay) Refund(ctx context.Context, in payin.RefundInp) (res *payin.RefundModel, err error) {
	return nil, gerror.New("易支付暂不支持在线退款，请线下处理")
}

// Notify 异步通知
func (h *yiPay) Notify(ctx context.Context, in payin.NotifyInp) (res *payin.NotifyModel, err error) {
	r := ghttp.RequestFromCtx(ctx)
	if err = r.ParseForm(); err != nil {
		return
	}

	form := r.Form
	sign := form.Get("sign")
	tradeStatus := form.Get("trade_status")
	if sign == "" || tradeStatus == "" {
		err = gerror.New("签名或交易状态为空")
		return
	}

	// 验签
	signParams := map[string]string{
		"money":        form.Get("money"),
		"name":         form.Get("name"),
		"out_trade_no": form.Get("out_trade_no"),
		"pid":          form.Get("pid"),
		"trade_no":     form.Get("trade_no"),
		"trade_status": tradeStatus,
		"type":         form.Get("type"),
	}

	if !h.verify(signParams, sign) {
		err = gerror.New("易支付验签失败")
		return
	}

	if !strings.EqualFold(tradeStatus, "TRADE_SUCCESS") {
		err = gerror.New("非交易成功状态，无需处理")
		return
	}

	payAmount := gconv.Float64(form.Get("money"))

	res = new(payin.NotifyModel)
	res.TransactionId = form.Get("trade_no")
	res.OutTradeNo = form.Get("out_trade_no")
	res.PayAt = gtime.Now()
	res.ActualAmount = payAmount
	return
}

// CreateOrder 生成支付链接
func (h *yiPay) CreateOrder(ctx context.Context, in payin.CreateOrderInp) (res *payin.CreateOrderModel, err error) {
	if h.config == nil {
		err = gerror.New("易支付配置未设置")
		return
	}

	gateway := strings.TrimRight(h.config.YiPayGateway, "/")
	if gateway == "" {
		err = gerror.New("请先配置易支付网关地址")
		return
	}

	payType := h.config.YiPayType
	if payType == "" {
		payType = "alipay" // 易支付默认类型
	}

	params := map[string]string{
		"pid":          h.config.YiPayMchId,
		"type":         payType,
		"out_trade_no": in.Pay.OutTradeNo,
		"notify_url":   in.Pay.NotifyUrl,
		"return_url":   in.Pay.ReturnUrl,
		"name":         in.Pay.Subject,
		"money":        fmt.Sprintf("%.2f", in.Pay.PayAmount),
		"sign_type":    "MD5",
	}

	sign := h.sign(params)
	params["sign"] = sign

	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}

	res = new(payin.CreateOrderModel)
	res.TradeType = in.Pay.TradeType
	res.OutTradeNo = in.Pay.OutTradeNo
	res.PayURL = fmt.Sprintf("%s/submit.php?%s", gateway, values.Encode())
	return
}

func (h *yiPay) verify(params map[string]string, sign string) bool {
	return strings.EqualFold(sign, h.sign(params))
}

// sign 生成易支付签名（key 已在末尾追加）
func (h *yiPay) sign(params map[string]string) string {
	var keys []string
	for k, v := range params {
		if v == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var builder strings.Builder
	for i, k := range keys {
		if i > 0 {
			builder.WriteString("&")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(params[k])
	}
	builder.WriteString(h.config.YiPayKey)

	hash := md5.Sum([]byte(builder.String()))
	return gstr.ToUpper(fmt.Sprintf("%x", hash))
}
