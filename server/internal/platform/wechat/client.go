package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Client struct {
	appID          string
	appSecret      string
	httpClient     *http.Client
	mu             sync.Mutex
	accessToken    string
	tokenExpiresAt time.Time
	allowMockLogin bool
}

type PhoneResult struct {
	PhoneNumber     string
	PurePhoneNumber string
	CountryCode     string
}

type TextSafetyResult struct {
	Suggest   string
	RiskLabel int
}

func (c *Client) CheckText(ctx context.Context, openID, content string) (TextSafetyResult, error) {
	token, err := c.getAccessToken(ctx)
	if err != nil {
		return TextSafetyResult{}, err
	}
	body, _ := json.Marshal(map[string]any{"content": content, "version": 2, "scene": 2, "openid": openID})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.weixin.qq.com/wxa/msg_sec_check?access_token="+url.QueryEscape(token), strings.NewReader(string(body)))
	if err != nil {
		return TextSafetyResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return TextSafetyResult{}, err
	}
	defer resp.Body.Close()
	var result struct {
		Code   int    `json:"errcode"`
		Msg    string `json:"errmsg"`
		Result struct {
			Suggest string `json:"suggest"`
			Label   int    `json:"label"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return TextSafetyResult{}, err
	}
	if result.Code != 0 {
		return TextSafetyResult{}, fmt.Errorf("wechat text safety failed: code=%d message=%s", result.Code, result.Msg)
	}
	return TextSafetyResult{Suggest: result.Result.Suggest, RiskLabel: result.Result.Label}, nil
}

func (c *Client) GetPhoneNumber(ctx context.Context, code string) (PhoneResult, error) {
	if strings.TrimSpace(code) == "" {
		return PhoneResult{}, fmt.Errorf("phone code is required")
	}
	token, err := c.getAccessToken(ctx)
	if err != nil {
		return PhoneResult{}, err
	}
	body, _ := json.Marshal(map[string]string{"code": code})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token="+url.QueryEscape(token), strings.NewReader(string(body)))
	if err != nil {
		return PhoneResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return PhoneResult{}, fmt.Errorf("request wechat phone: %w", err)
	}
	defer resp.Body.Close()
	var result struct {
		Code      int    `json:"errcode"`
		Msg       string `json:"errmsg"`
		PhoneInfo struct {
			PhoneNumber     string `json:"phoneNumber"`
			PurePhoneNumber string `json:"purePhoneNumber"`
			CountryCode     string `json:"countryCode"`
		} `json:"phone_info"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return PhoneResult{}, err
	}
	if result.Code != 0 || result.PhoneInfo.PurePhoneNumber == "" {
		return PhoneResult{}, fmt.Errorf("wechat phone failed: code=%d message=%s", result.Code, result.Msg)
	}
	return PhoneResult{PhoneNumber: result.PhoneInfo.PhoneNumber, PurePhoneNumber: result.PhoneInfo.PurePhoneNumber, CountryCode: result.PhoneInfo.CountryCode}, nil
}

func (c *Client) getAccessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.accessToken != "" && time.Now().Before(c.tokenExpiresAt) {
		return c.accessToken, nil
	}
	if c.appID == "" || c.appSecret == "" {
		return "", fmt.Errorf("wechat credentials are not configured")
	}
	query := url.Values{"grant_type": {"client_credential"}, "appid": {c.appID}, "secret": {c.appSecret}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.weixin.qq.com/cgi-bin/token?"+query.Encode(), nil)
	if err != nil {
		return "", err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Code        int    `json:"errcode"`
		Msg         string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Code != 0 || result.AccessToken == "" {
		return "", fmt.Errorf("wechat access token failed: code=%d message=%s", result.Code, result.Msg)
	}
	c.accessToken = result.AccessToken
	c.tokenExpiresAt = time.Now().Add(time.Duration(result.ExpiresIn-300) * time.Second)
	return c.accessToken, nil
}

func NewClient(appID, appSecret string, allowMock ...bool) *Client {
	client := &Client{appID: appID, appSecret: appSecret, httpClient: &http.Client{Timeout: 8 * time.Second}}
	if len(allowMock) > 0 {
		client.allowMockLogin = allowMock[0]
	}
	return client
}

// Code2Session 将小程序 uni.login 获得的临时 code 交给微信验证，返回本小程序下的 OpenID。
// 仅显式开启开发模拟登录时接受固定 mock code；请求取消、网络或微信校验失败均返回错误。
// 请求地址含应用密钥和临时 code，不应记录完整 URL。
func (c *Client) Code2Session(ctx context.Context, code string) (string, error) {
	if c.allowMockLogin && code == "the code is a mock one" {
		return "mock_openid_dev", nil
	}
	if c.appID == "" || c.appSecret == "" {
		return "", fmt.Errorf("wechat credentials are not configured")
	}
	query := url.Values{"appid": {c.appID}, "secret": {c.appSecret}, "js_code": {code}, "grant_type": {"authorization_code"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.weixin.qq.com/sns/jscode2session?"+query.Encode(), nil)
	if err != nil {
		return "", fmt.Errorf("create wechat request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request wechat session: %w", err)
	}
	defer resp.Body.Close()
	var result struct {
		OpenID string `json:"openid"`
		Code   int    `json:"errcode"`
		Msg    string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode wechat response: %w", err)
	}
	if result.Code != 0 || result.OpenID == "" {
		return "", fmt.Errorf("wechat code2session failed: code=%d message=%s", result.Code, result.Msg)
	}
	return result.OpenID, nil
}

// Init validates configured WeChat credentials without consuming a login code.
func (c *Client) Init(ctx context.Context) error {
	if c.allowMockLogin && c.appID == "" && c.appSecret == "" {
		return nil
	}
	if c.appID == "" || c.appSecret == "" {
		return fmt.Errorf("wechat credentials are not configured")
	}
	if _, err := c.getAccessToken(ctx); err != nil {
		return fmt.Errorf("wechat startup check failed")
	}
	return nil
}
