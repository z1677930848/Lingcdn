Tea.context(function () {
    var actionData = (window.TEA && window.TEA.ACTION && window.TEA.ACTION.data) ? window.TEA.ACTION.data : {}
    var summary = window.CCSummary || actionData.ccSummary || {}
    var ccIsOn = false
    if (typeof summary.isOn != "undefined") {
        ccIsOn = !!summary.isOn
    } else if (this.ccConfig != null && typeof this.ccConfig.isOn != "undefined") {
        ccIsOn = !!this.ccConfig.isOn
    }

    this.ccSummary = {
        templateName: summary.templateName || "自定义",
        isOn: ccIsOn,
        block5m: summary.block5m || 0,
        captcha5m: summary.captcha5m || 0,
        ban5m: summary.ban5m || 0,
        keys: summary.keys || 0,
        maxKeys: summary.maxKeys || 0,
        memMB: summary.memMB || 0
    }

    this.summaryLoading = false

    this.loadSummary = function () {
        var that = this
        this.summaryLoading = true
        this.$get("/servers/server/settings/cc/summary")
            .params({
                webId: this.webId
            })
            .success(function (resp) {
                var data = resp.data || {}
                var s = data.summary || data
                var isOn = that.ccSummary.isOn
                if (typeof s.isOn != "undefined") {
                    isOn = !!s.isOn
                }
                that.ccSummary = {
                    templateName: s.templateName || that.ccSummary.templateName,
                    isOn: isOn,
                    block5m: s.block5m || 0,
                    captcha5m: s.captcha5m || 0,
                    ban5m: s.ban5m || 0,
                    keys: s.keys || 0,
                    maxKeys: s.maxKeys || that.ccSummary.maxKeys,
                    memMB: s.memMB || that.ccSummary.memMB
                }
            })
            .done(function () {
                that.summaryLoading = false
            })
            .fail(function () {
                that.summaryLoading = false
            })
    }

    this.$delay(function () {
        this.loadSummary()
    })

    this.success = NotifyReloadSuccess("保存成功")
})
