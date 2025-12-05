package jsonapi

import (
	"github.com/iwind/TeaGo"
)

func init() {
	TeaGo.BeforeStart(func(server *TeaGo.Server) {
		// JSON API 路由注册
		server.Prefix("/admin").
			// 站点相关（无需认证）
			Get("/site/captcha", new(SiteCaptchaAction)).
			Get("/site/config", new(SiteConfigAction)).
			Post("/site/login", new(SiteLoginAction)).
			// 需要认证的接口
			Post("/site/logout", new(SiteLogoutAction)).
			Get("/member/info", new(MemberInfoAction)).
			Get("/console/stat", new(ConsoleStatAction)).
			// 菜单路由
			Get("/menu/list", new(MenuListAction)).
			EndAll()
	})
}
