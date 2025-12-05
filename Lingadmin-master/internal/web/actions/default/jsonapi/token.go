package jsonapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"
)

// JWT 密钥（生产环境应从配置文件读取）
var jwtSecretKey = []byte("lingcdn-admin-jwt-secret-key-2024")

// TokenStore Token 存储（简单内存存储，生产环境建议使用 Redis）
var tokenStore = &TokenStorage{
	tokens: make(map[string]*TokenInfo),
}

// TokenInfo Token 信息
type TokenInfo struct {
	AdminId   int64     `json:"adminId"`
	Username  string    `json:"username"`
	ExpireAt  time.Time `json:"expireAt"`
	CreatedAt time.Time `json:"createdAt"`
}

// TokenStorage Token 存储器
type TokenStorage struct {
	tokens map[string]*TokenInfo
	mu     sync.RWMutex
}

// JWTClaims JWT 声明
type JWTClaims struct {
	AdminId  int64  `json:"adminId"`
	Username string `json:"username"`
	Exp      int64  `json:"exp"`
	Iat      int64  `json:"iat"`
}

// GenerateToken 生成 JWT Token
func GenerateToken(adminId int64, username string) (string, int64, error) {
	now := time.Now()
	expireAt := now.Add(24 * time.Hour * 7) // 7天过期

	claims := JWTClaims{
		AdminId:  adminId,
		Username: username,
		Exp:      expireAt.Unix(),
		Iat:      now.Unix(),
	}

	// 创建 JWT Header
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}
	headerBytes, _ := json.Marshal(header)
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerBytes)

	// 创建 JWT Payload
	payloadBytes, _ := json.Marshal(claims)
	payloadEncoded := base64.RawURLEncoding.EncodeToString(payloadBytes)

	// 创建签名
	signatureInput := headerEncoded + "." + payloadEncoded
	h := hmac.New(sha256.New, jwtSecretKey)
	h.Write([]byte(signatureInput))
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	token := signatureInput + "." + signature

	// 存储 Token 信息
	tokenStore.mu.Lock()
	tokenStore.tokens[token] = &TokenInfo{
		AdminId:   adminId,
		Username:  username,
		ExpireAt:  expireAt,
		CreatedAt: now,
	}
	tokenStore.mu.Unlock()

	return token, expireAt.Unix(), nil
}

// ValidateToken 验证 Token
func ValidateToken(token string) (*TokenInfo, error) {
	if token == "" {
		return nil, errors.New("token is empty")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	// 验证签名
	signatureInput := parts[0] + "." + parts[1]
	h := hmac.New(sha256.New, jwtSecretKey)
	h.Write([]byte(signatureInput))
	expectedSignature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	if parts[2] != expectedSignature {
		return nil, errors.New("invalid token signature")
	}

	// 解析 Payload
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("invalid token payload")
	}

	var claims JWTClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, errors.New("invalid token claims")
	}

	// 检查过期时间
	if claims.Exp < time.Now().Unix() {
		return nil, errors.New("token expired")
	}

	return &TokenInfo{
		AdminId:   claims.AdminId,
		Username:  claims.Username,
		ExpireAt:  time.Unix(claims.Exp, 0),
		CreatedAt: time.Unix(claims.Iat, 0),
	}, nil
}

// RemoveToken 移除 Token
func RemoveToken(token string) {
	tokenStore.mu.Lock()
	delete(tokenStore.tokens, token)
	tokenStore.mu.Unlock()
}

// CleanExpiredTokens 清理过期 Token（可定期调用）
func CleanExpiredTokens() {
	tokenStore.mu.Lock()
	defer tokenStore.mu.Unlock()

	now := time.Now()
	for token, info := range tokenStore.tokens {
		if info.ExpireAt.Before(now) {
			delete(tokenStore.tokens, token)
		}
	}
}
