package jsonapi

// MenuListAction 获取菜单列表
type MenuListAction struct {
	BaseAPIAction
}

func (this *MenuListAction) RunGet(params struct{}) {
	token := this.GetToken()
	if token == "" {
		this.Unauthorized("请先登录")
		return
	}

	_, err := ValidateToken(token)
	if err != nil {
		this.Unauthorized("登录已过期，请重新登录")
		return
	}

	// 返回菜单列表
	menus := []map[string]interface{}{
		{
			"id":        1,
			"pid":       0,
			"title":     "仪表盘",
			"name":      "dashboard",
			"path":      "/dashboard",
			"icon":      "DashboardOutlined",
			"type":      1,
			"sort":      1,
			"status":    1,
			"component": "/dashboard/console/console.vue",
		},
		{
			"id":        2,
			"pid":       0,
			"title":     "网站管理",
			"name":      "servers",
			"path":      "/servers",
			"icon":      "GlobalOutlined",
			"type":      1,
			"sort":      2,
			"status":    1,
			"component": "LAYOUT",
			"children": []map[string]interface{}{
				{
					"id":        21,
					"pid":       2,
					"title":     "网站列表",
					"name":      "servers-list",
					"path":      "/servers/list",
					"icon":      "",
					"type":      1,
					"sort":      1,
					"status":    1,
					"component": "/servers/list/index.vue",
				},
				{
					"id":        22,
					"pid":       2,
					"title":     "证书管理",
					"name":      "servers-certs",
					"path":      "/servers/certs",
					"icon":      "",
					"type":      1,
					"sort":      2,
					"status":    1,
					"component": "/servers/certs/index.vue",
				},
			},
		},
		{
			"id":        3,
			"pid":       0,
			"title":     "节点管理",
			"name":      "clusters",
			"path":      "/clusters",
			"icon":      "ClusterOutlined",
			"type":      1,
			"sort":      3,
			"status":    1,
			"component": "LAYOUT",
			"children": []map[string]interface{}{
				{
					"id":        31,
					"pid":       3,
					"title":     "集群列表",
					"name":      "clusters-list",
					"path":      "/clusters/list",
					"icon":      "",
					"type":      1,
					"sort":      1,
					"status":    1,
					"component": "/clusters/list/index.vue",
				},
				{
					"id":        32,
					"pid":       3,
					"title":     "节点列表",
					"name":      "nodes-list",
					"path":      "/nodes/list",
					"icon":      "",
					"type":      1,
					"sort":      2,
					"status":    1,
					"component": "/nodes/list/index.vue",
				},
			},
		},
		{
			"id":        4,
			"pid":       0,
			"title":     "系统设置",
			"name":      "settings",
			"path":      "/settings",
			"icon":      "SettingOutlined",
			"type":      1,
			"sort":      10,
			"status":    1,
			"component": "LAYOUT",
			"children": []map[string]interface{}{
				{
					"id":        41,
					"pid":       4,
					"title":     "管理员",
					"name":      "admins",
					"path":      "/settings/admins",
					"icon":      "",
					"type":      1,
					"sort":      1,
					"status":    1,
					"component": "/settings/admins/index.vue",
				},
				{
					"id":        42,
					"pid":       4,
					"title":     "系统配置",
					"name":      "system-config",
					"path":      "/settings/config",
					"icon":      "",
					"type":      1,
					"sort":      2,
					"status":    1,
					"component": "/settings/config/index.vue",
				},
			},
		},
	}

	this.SuccessData(map[string]interface{}{
		"list": menus,
	})
}
