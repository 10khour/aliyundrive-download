package pkg

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tidwall/gjson"
)

func Login() (string, error) {
	if err := reLogin(); err != nil {
		return "", err
	}
	qrCode, err := getQrCode()
	if err != nil {
		return "", err
	}
	return qrCode.CodeContent, nil
}

func getQrCode() (*qrCode, error) {
	req, err := http.NewRequest(http.MethodGet, "https://passport.aliyundrive.com/newlogin/qrcode/generate.do?"+
		"appName=aliyun_drive"+
		"&fromSite=52"+
		"&appName=aliyun_drive"+
		"&appEntrance=web"+
		"&isMobile=false"+
		"&lang=zh_CN"+
		"&returnUrl="+
		"&fromSite=52"+
		"&bizParams="+
		"&_bx-v=2.0.31", nil)
	if err != nil {
		return nil, err
	}
	response, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := readJson(response)
	if err != nil {
		return nil, err
	}
	var qrCode = new(qrCode)
	var content = gjson.Get(body, "content.data").String()
	err = json.Unmarshal([]byte(content), qrCode)
	return qrCode, err
}

func reLogin() error {
	req, err := http.NewRequest(http.MethodGet, "https://auth.aliyundrive.com/v2/oauth/authorize?client_id=25dzX3vbYqktVxyX"+
		"&redirect_uri=https%3A%2F%2Fwww.aliyundrive.com%2Fsign%2Fcallback"+
		"&response_type=code"+
		"&login_type=custom"+
		"&state=%7B%22origin%22%3A%22https%3A%2F%2Fwww.aliyundrive.com%22%7D", nil)
	if err != nil {
		return err
	}
	response, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("server %d", response.StatusCode)
	}
	return nil
}

type qrCode struct {
	TitleMsg    string `json:"title_msg"`
	T           int64  `json:"t"`
	CodeContent string `json:"codeContent"`
	Ck          string `json:"ck"`
	ResultCode  int    `json:"resultCode"`
}
