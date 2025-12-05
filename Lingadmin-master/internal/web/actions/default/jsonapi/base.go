package jsonapi

import (
	"encoding/json"
	"github.com/TeaOSLab/EdgeAdmin/internal/web/actions/actionutils"
	"net/http"
	"time"
)

// BaseAPIAction JSON API 基础 Action
type BaseAPIAction struct {
	actionutils.ParentAction
}

// Response JSON 响应结构
type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// WriteJSON 写入 JSON 响应
func (this *BaseAPIAction) WriteJSON(code int, message string, data interface{}) {
	resp := Response{
		Code:      code,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	this.ResponseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	this.ResponseWriter.Header().Set("Access-Control-Allow-Origin", "*")
	this.ResponseWriter.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	this.ResponseWriter.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

	jsonBytes, _ := json.Marshal(resp)
	this.ResponseWriter.Write(jsonBytes)
}

// SuccessData 成功响应（带数据）
func (this *BaseAPIAction) SuccessData(data interface{}) {
	this.WriteJSON(0, "success", data)
}

// SuccessMsg 成功响应（带消息）
func (this *BaseAPIAction) SuccessMsg(message string) {
	this.WriteJSON(0, message, nil)
}

// SuccessOK 成功响应
func (this *BaseAPIAction) SuccessOK() {
	this.WriteJSON(0, "success", nil)
}

// FailMsg 失败响应
func (this *BaseAPIAction) FailMsg(message string) {
	this.WriteJSON(-1, message, nil)
}

// FailCode 失败响应（带错误码）
func (this *BaseAPIAction) FailCode(code int, message string) {
	this.WriteJSON(code, message, nil)
}

// Unauthorized 未授权响应
func (this *BaseAPIAction) Unauthorized(message string) {
	this.ResponseWriter.WriteHeader(http.StatusUnauthorized)
	this.WriteJSON(61, message, nil)
}

// GetToken 从请求头获取 Token
func (this *BaseAPIAction) GetToken() string {
	auth := this.Request.Header.Get("Authorization")
	if len(auth) > 7 && auth[:7] == "Bearer " {
		return auth[7:]
	}
	return auth
}
