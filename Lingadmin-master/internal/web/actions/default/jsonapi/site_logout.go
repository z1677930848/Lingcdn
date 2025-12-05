package jsonapi

// SiteLogoutAction 登出接口
type SiteLogoutAction struct {
	BaseAPIAction
}

func (this *SiteLogoutAction) RunPost(params struct{}) {
	token := this.GetToken()
	if token != "" {
		RemoveToken(token)
	}
	this.SuccessOK()
}
