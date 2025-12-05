package jsonapi

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"strings"
	"sync"
	"time"
)

// SiteCaptchaAction 获取验证码
type SiteCaptchaAction struct {
	BaseAPIAction
}

// 验证码存储
var captchaStore = &CaptchaStorage{
	codes: make(map[string]*CaptchaInfo),
}

type CaptchaInfo struct {
	Code      string
	ExpireAt  time.Time
}

type CaptchaStorage struct {
	codes map[string]*CaptchaInfo
	mu    sync.RWMutex
}

func (this *SiteCaptchaAction) RunGet(params struct{}) {
	// 生成验证码ID
	cid := generateRandomString(32)

	// 生成4位验证码
	code := generateCaptchaCode(4)

	// 存储验证码
	captchaStore.mu.Lock()
	captchaStore.codes[cid] = &CaptchaInfo{
		Code:     code,
		ExpireAt: time.Now().Add(5 * time.Minute),
	}
	captchaStore.mu.Unlock()

	// 生成验证码图片
	b64 := generateCaptchaImage(code)

	this.SuccessData(map[string]interface{}{
		"cid":  cid,
		"b64s": b64,
	})
}

// ValidateCaptcha 验证验证码
func ValidateCaptcha(cid, code string) bool {
	if cid == "" || code == "" {
		return false
	}

	captchaStore.mu.Lock()
	defer captchaStore.mu.Unlock()

	info, ok := captchaStore.codes[cid]
	if !ok {
		return false
	}

	// 删除已使用的验证码
	delete(captchaStore.codes, cid)

	// 检查是否过期
	if info.ExpireAt.Before(time.Now()) {
		return false
	}

	// 不区分大小写比较
	return strings.EqualFold(info.Code, code)
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func generateCaptchaCode(length int) string {
	const charset = "0123456789abcdefghjkmnpqrstuvwxyz" // 去掉容易混淆的字符
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func generateCaptchaImage(code string) string {
	width := 100
	height := 40

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// 填充背景色
	bgColor := color.RGBA{240, 240, 240, 255}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, bgColor)
		}
	}

	// 添加干扰点
	for i := 0; i < 100; i++ {
		x := rand.Intn(width)
		y := rand.Intn(height)
		img.Set(x, y, color.RGBA{
			uint8(rand.Intn(256)),
			uint8(rand.Intn(256)),
			uint8(rand.Intn(256)),
			255,
		})
	}

	// 简单绘制文字（实际项目中应使用字体库）
	textColor := color.RGBA{50, 50, 150, 255}
	startX := 15
	for i, c := range code {
		drawChar(img, startX+i*20, 10, byte(c), textColor)
	}

	// 编码为 base64
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

// 简单的字符绘制（5x7 点阵）
func drawChar(img *image.RGBA, x, y int, c byte, col color.Color) {
	// 简化版本：只画一个矩形代表字符
	for dy := 0; dy < 20; dy++ {
		for dx := 0; dx < 12; dx++ {
			if (dx+dy)%3 == 0 {
				img.Set(x+dx, y+dy, col)
			}
		}
	}
}

