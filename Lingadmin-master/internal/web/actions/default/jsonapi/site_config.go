package jsonapi

import (
	"github.com/TeaOSLab/EdgeAdmin/internal/configloaders"
	teaconst "github.com/TeaOSLab/EdgeAdmin/internal/const"
)

// SiteConfigAction 获取站点配置
type SiteConfigAction struct {
	BaseAPIAction
}

func (this *SiteConfigAction) RunGet(params struct{}) {
	uiConfig, err := configloaders.LoadAdminUIConfig()

	systemName := "LingCDN"
	version := teaconst.Version
	showVersion := true

	if err == nil && uiConfig != nil {
		if uiConfig.AdminSystemName != "" {
			systemName = uiConfig.AdminSystemName
		}
		if uiConfig.Version != "" {
			version = uiConfig.Version
		}
		showVersion = uiConfig.ShowVersion
	}

	this.SuccessData(map[string]interface{}{
		"projectName":         systemName,
		"version":             version,
		"showVersion":         showVersion,
		"loginCaptchaSwitch":  1, // 开启验证码
		"loginRegisterSwitch": 0, // 关闭注册
		"i18nSwitch":          false,
		"defaultLanguage":     "zh-CN",
	})
}
