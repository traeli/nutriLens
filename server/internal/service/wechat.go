package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type WechatService struct {
	AppID     string
	AppSecret string
}

func NewWechatService(appID, appSecret string) *WechatService {
	return &WechatService{AppID: appID, AppSecret: appSecret}
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
