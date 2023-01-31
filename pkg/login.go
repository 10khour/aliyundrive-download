package pkg

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	qrcode2 "github.com/skip2/go-qrcode"
	"github.com/tidwall/gjson"
)

func Login() (string, string, error) {
	if err := preLogin(); err != nil {
		return "", "", err
	}
	qrCode, err := getQrCode()
	if err != nil {
		return "", "", err
	}
	obj, err := qrcode2.New(qrCode.CodeContent, qrcode2.Low)
	if err != nil {
		return "", "", err
	}
	// 终端输出二维码
	fmt.Print(obj.ToSmallString(false))
	log.Print("请使用阿里云盘手机客户端扫码登录")
	for {
		body, err := queryQrCodeState(strconv.FormatInt(qrCode.T, 10), qrCode.Ck)
		if err != nil {
			log.Fatal(err)
			break
		}
		state := gjson.Get(body, "content.data.qrCodeStatus").String()
		switch state {
		case "SCANED":
			log.Print("已扫描，登录中...\n")
		case "NEW":
		case "EXPIRED":
			log.Print("二维码过期\n")
			return "", "", fmt.Errorf("二维码过期")
		case "CANCELED":
			log.Print("取消登录\n")
		case "CONFIRMED":
			bizExt, err := base64.StdEncoding.DecodeString(gjson.Get(body, "content.data.bizExt").String())
			if err != nil {
				return "", "", err
			}
			var query queryQrCodeBizAction
			err = json.Unmarshal(bizExt, &query)
			if err != nil {
				return "", "", err
			}
			return confirmLogin(query.PdsLoginResult.AccessToken)
		}
		time.Sleep(time.Second)
	}
	return "", "", fmt.Errorf("未登录")
}

func queryQrCodeState(t string, ck string) (string, error) {
	form := url.Values{}
	form.Add("t", t)
	form.Add("ck", ck)
	form.Add("appName", "aliyun_drive")
	form.Add("appEntrance", "web")
	form.Add("isMobile", "false")
	form.Add("lang", "zh_CN")
	form.Add("returnUrl", "")
	form.Add("fromSite", "52")
	form.Add("bizParams", "")
	form.Add("navlanguage", "zh-CN")
	form.Add("navPlatform", "MacIntel")

	req, err := http.NewRequest(http.MethodPost, "https://passport.aliyundrive.com/newlogin/qrcode/query.do?appName=aliyun_drive&fromSite=52&_bx-v=2.0.31", strings.NewReader(form.Encode()))
	if err != nil {
		return "ERROR", err
	}
	req.Header.Add("Accept", "application/json, text/plain")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, err := readJson(res)
	return body, err
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

func preLogin() error {
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

func confirmLogin(accessToken string) (string, string, error) {
	// var buf = bytes.NewBuffer(nil)
	// buf.WriteString(fmt.Sprintf(`{"token":"%s"}`, accessToken))
	var hash = make(map[string]string)
	hash["token"] = accessToken
	buf, _ := json.Marshal(hash)
	req, err := http.NewRequest(http.MethodPost, "https://auth.aliyundrive.com/v2/oauth/token_login", strings.NewReader(string(buf)))
	if err != nil {
		return "", "", err
	}
	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	res, err := httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()
	body, err := readJson(res)
	if err != nil {
		return "", "", err
	}
	callBackUrl := gjson.Get(body, "goto").String()
	_, err = httpClient.Get(callBackUrl)
	if err != nil {
		return "", "", err
	}
	u, _ := url.Parse(callBackUrl)
	return login(u.Query().Get("code"))
}

func login(code string) (string, string, error) {
	hash := make(map[string]string)
	hash["code"] = code
	hash["loginType"] = "normal"
	hash["deviceId"] = "aliyundrive"
	buf, _ := json.Marshal(hash)
	req, err := http.NewRequest(http.MethodPost, "https://api.aliyundrive.com/token/get", bytes.NewBuffer(buf))
	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	if err != nil {
		return "", "", err
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	body, _ := readJson(res)
	if res.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("login http %d", res.StatusCode)
	}
	return gjson.Get(body, "access_token").String(), gjson.Get(body, "refresh_token").String(), nil

}

type qrCode struct {
	TitleMsg    string `json:"title_msg"`
	T           int64  `json:"t"`
	CodeContent string `json:"codeContent"`
	Ck          string `json:"ck"`
	ResultCode  int    `json:"resultCode"`
}

type queryQrCodeBizAction struct {
	PdsLoginResult struct {
		Role           string        `json:"role"`
		IsFirstLogin   bool          `json:"isFirstLogin"`
		NeedLink       bool          `json:"needLink"`
		LoginType      string        `json:"loginType"`
		NickName       string        `json:"nickName"`
		NeedRpVerify   bool          `json:"needRpVerify"`
		Avatar         string        `json:"avatar"`
		AccessToken    string        `json:"accessToken"`
		UserName       string        `json:"userName"`
		UserID         string        `json:"userId"`
		DefaultDriveID string        `json:"defaultDriveId"`
		ExistLink      []interface{} `json:"existLink"`
		ExpiresIn      int           `json:"expiresIn"`
		ExpireTime     time.Time     `json:"expireTime"`
		RequestID      string        `json:"requestId"`
		DataPinSetup   bool          `json:"dataPinSetup"`
		State          string        `json:"state"`
		TokenType      string        `json:"tokenType"`
		DataPinSaved   bool          `json:"dataPinSaved"`
		RefreshToken   string        `json:"refreshToken"`
		Status         string        `json:"status"`
	} `json:"pds_login_result"`
}
