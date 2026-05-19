package service

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"nutrilens/internal/cache"
)

type WechatService struct {
	AppID     string
	AppSecret string
	Token     string
	cache     *cache.Client
}

func NewWechatService(appID, appSecret, token string, c *cache.Client) *WechatService {
	return &WechatService{AppID: appID, AppSecret: appSecret, Token: token, cache: c}
}

// Code2Session exchanges wx.login code for openid and session_key.
func (s *WechatService) Code2Session(code string) (openID, sessionKey string, err error) {
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		s.AppID, s.AppSecret, code,
	)
	resp, err := http.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var data struct {
		OpenID     string `json:"openid"`
		SessionKey string `json:"session_key"`
		ErrCode    int    `json:"errcode"`
		ErrMsg     string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", "", err
	}
	if data.ErrCode != 0 {
		return "", "", fmt.Errorf("wechat error: %d %s", data.ErrCode, data.ErrMsg)
	}
	return data.OpenID, data.SessionKey, nil
}

// GetAccessToken returns a cached wechat access_token, fetching a new one if expired/missing.
func (s *WechatService) GetAccessToken(ctx context.Context) (string, error) {
	const cacheKey = "wechat:access_token"

	if s.cache != nil {
		token, err := s.cache.Get(ctx, cacheKey)
		if err == nil && token != "" {
			return token, nil
		}
	}

	url := fmt.Sprintf(
		"https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		s.AppID, s.AppSecret,
	)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("get access_token failed: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var data struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", fmt.Errorf("parse access_token response failed: %w", err)
	}
	if data.ErrCode != 0 {
		return "", fmt.Errorf("wechat token error: %d %s", data.ErrCode, data.ErrMsg)
	}

	// Cache with 7000s TTL (official expiry is 7200s, leave buffer)
	if s.cache != nil {
		if err := s.cache.Set(ctx, cacheKey, data.AccessToken, 7000*time.Second); err != nil {
			log.Printf("[WechatService] failed to cache access_token: %v", err)
		}
	}

	return data.AccessToken, nil
}

// SendSubscribeMessage sends a WeChat subscribe message to the specified user.
func (s *WechatService) SendSubscribeMessage(ctx context.Context, openID, templateID, page string, data map[string]map[string]string) error {
	token, err := s.GetAccessToken(ctx)
	if err != nil {
		return err
	}

	err = s.sendSubscribeMessageWithToken(ctx, token, openID, templateID, page, data)
	if err != nil && s.cache != nil {
		log.Printf("[WechatService] send failed, clearing token cache and retrying: %v", err)
		s.cache.Del(ctx, "wechat:access_token")
		token, tokenErr := s.GetAccessToken(ctx)
		if tokenErr != nil {
			return tokenErr
		}
		return s.sendSubscribeMessageWithToken(ctx, token, openID, templateID, page, data)
	}
	return err
}

// VerifySignature checks the WeChat server verification signature.
// It sorts token, timestamp, nonce lexicographically, joins them, SHA1 hashes,
// and compares with the given signature.
func (s *WechatService) VerifySignature(timestamp, nonce, signature string) bool {
	arr := []string{s.Token, timestamp, nonce}
	sort.Strings(arr)
	str := strings.Join(arr, "")

	h := sha1.New()
	h.Write([]byte(str))
	calculated := fmt.Sprintf("%x", h.Sum(nil))

	return calculated == signature
}

func (s *WechatService) sendSubscribeMessageWithToken(ctx context.Context, token, openID, templateID, page string, data map[string]map[string]string) error {
	payload := map[string]interface{}{
		"touser":            openID,
		"template_id":       templateID,
		"page":              page,
		"data":              data,
		"miniprogram_state": "formal",
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/subscribe/send?access_token=%s", token)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send subscribe message failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("parse subscribe message response failed: %s", string(respBody))
	}
	if result.ErrCode != 0 {
		return fmt.Errorf("wechat subscribe message error: %d %s", result.ErrCode, result.ErrMsg)
	}
	return nil
}
