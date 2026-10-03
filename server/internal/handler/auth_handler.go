package handler

import (
	"net/http"
	"net/mail"
	"strings"

	"shijibu/internal/platform/httpx"

	"github.com/gin-gonic/gin"
)

// WxLogin 校验微信登录请求，将请求上下文和临时 code 交给账户服务。
// 服务失败时映射 HTTP 错误；成功时将用户资料及双 Token 序列化为 JSON。
func (h *Handler) WxLogin(c *gin.Context) {
	var input struct {
		Code string `json:"code" binding:"required"` // 登陆凭证
	}
	if c.ShouldBindJSON(&input) != nil {
		httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "登录凭证不能为空")
		return
	}
	result, err := h.Accounts.Login(c.Request.Context(), input.Code) // 完成微信登录，取得用户资料和双 Token。
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, result)
}

// EmailLogin 接收邮箱与六位数字验证码，认证后返回与微信登录相同的双 Token。
func (h *Handler) EmailLogin(c *gin.Context) {
	var input struct {
		Email     string `json:"email" binding:"required,max=254"`
		EmailCode string `json:"email_code" binding:"required,len=6,numeric"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "邮箱或验证码格式错误")
		return
	}

	if !validAuthEmail(input.Email) {
		httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "邮箱格式错误")
		return
	}
	result, err := h.Accounts.EmailLogin(c.Request.Context(), strings.TrimSpace(input.Email), input.EmailCode, c.RemoteIP())
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, result)
}

// SendEmailCode 使用连接对端 IP 限流，避免客户端伪造转发头绕过限制。
func (h *Handler) SendEmailCode(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,max=254"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "邮箱格式错误")
		return
	}

	if !validAuthEmail(input.Email) {
		httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "邮箱格式错误")
		return
	}
	err := h.Accounts.SendEmailCode(c.Request.Context(), strings.TrimSpace(input.Email), c.RemoteIP())
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, gin.H{"message": "验证码已发送"})
}

func (h *Handler) RefreshToken(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token" binding:"required,max=4096"`
	}
	if c.ShouldBindJSON(&input) != nil {
		httpx.Error(c, http.StatusBadRequest, "INVALID_ARGUMENT", "刷新凭证不能为空")
		return
	}
	result, err := h.Accounts.Refresh(c.Request.Context(), input.RefreshToken)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	httpx.OK(c, http.StatusOK, result)
}

//func (h *Handler) authlogin(c *gin.Context) {
//
//}

// validAuthEmail 在进入业务层前校验单个邮箱地址，拒绝带显示名或换行的输入。
func validAuthEmail(value string) bool {
	value = strings.TrimSpace(value)
	address, err := mail.ParseAddress(value)
	return err == nil && address.Address == value && !strings.ContainsAny(value, "\r\n")
}
