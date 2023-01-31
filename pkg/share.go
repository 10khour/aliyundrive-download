package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"path/filepath"
	"regexp"

	"github.com/tidwall/gjson"
)

var (
	httpClient http.Client
)

func init() {
	// init http client
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Got error while creating cookie jar %s", err.Error())
	}
	httpClient = http.Client{
		Jar: jar,
	}
}

func GetShareInfo(shareURL string) (shareID string, parentFileID string, creatorID string, err error) {
	shareID, err = getShareID(shareURL)
	if err != nil {
		return "", "", "", err
	}
	parentFileID, creatorID, err = getShareByAnonymous(shareID)
	return shareID, parentFileID, creatorID, err
}

func ListFiles(shareID string, parentFile string, token string) ([]AliyunDriverFile, error) {
	return listFiles("", shareID, parentFile, token)
}

func listFiles(dir string, shareID string, parentFile string, token string) ([]AliyunDriverFile, error) {
	var values = make(map[string]interface{})
	values["share_id"] = shareID
	values["image_thumbnail_process"] = "image/resize,w_256/format,jpeg"
	values["image_url_process"] = "image/resize,w_1920/format,jpeg/interlace,1"
	values["limit"] = 200
	values["order_by"] = "name"
	values["order_direction"] = "DESC"
	values["parent_file_id"] = parentFile
	values["video_thumbnail_process"] = "video/snapshot,t_1000,f_jpg,ar_auto,w_256"
	body, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(body)
	url := "https://api.aliyundrive.com/adrive/v3/file/list"
	method := "POST"

	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Add("accept-language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Add("content-type", "application/json;charset=UTF-8")
	req.Header.Add("x-share-token", token)
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	content, err := readJson(res)
	if err != nil {
		return nil, err
	}
	var files []AliyunDriverFile
	var results = gjson.Get(content, "items").Array()
	for _, result := range results {
		file := AliyunDriverFile{
			Name:         result.Get("name").String(),
			ShareID:      result.Get("share_id").String(),
			DomainID:     result.Get("domain_id").String(),
			FileID:       result.Get("file_id").String(),
			Type:         result.Get("type").String(),
			CreateAt:     result.Get("create_at").String(),
			UpdateAt:     result.Get("update_at").String(),
			ParentFileID: result.Get("parent_file_id").String(),
			RevisionID:   result.Get("revision_id").String(),
			FromShareID:  result.Get("from_share_id").String(),
		}
		if file.Type == "folder" {
			subFiles, err := listFiles(filepath.Join(dir, file.Name), shareID, file.FileID, token)
			if err != nil {
				return nil, err
			}
			files = append(files, subFiles...)
		} else {
			file.Name = filepath.Join(dir, file.Name)
			files = append(files, file)
		}
	}

	return files, err
}

func getShareByAnonymous(shareID string) (string, string, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(fmt.Sprintf(`{
		"share_id": "%s"
	  }`, shareID))
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("https://api.aliyundrive.com/adrive/v3/share_link/get_share_by_anonymous?share_id=%s", shareID), buf)

	header := http.Header{}
	header.Add("accept", "application/json, text/plain, */*")
	header.Add("referer", "https://www.aliyundrive.com/")
	header.Add("origin", "https://www.aliyundrive.com/")
	header.Add("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36")
	header.Add("content-type", "application/json;charset=UTF-8")
	req.Header = header
	res, err := httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	body, err := readJson(res)
	if err != nil {
		log.Fatal(err)
	}
	displayName := gjson.Get(body, "display_name")
	updateTime := gjson.Get(body, "updated_at")
	log.Printf("%s 最近更新时间: %s", displayName.String(), updateTime.String())
	results := gjson.GetMany(body, "file_infos")
	return results[0].Get("file_id").String(), gjson.Get(body, "creator_id").String(), nil
}

func getShareID(shareURL string) (string, error) {
	r := regexp.MustCompile(`www.aliyundrive.com/s/(.*)/?`)
	matches := r.FindStringSubmatch(shareURL)
	if len(matches) > 1 {
		return matches[1], nil
	}
	return "", fmt.Errorf("%s 不是一个合法的分享网址", shareURL)
}

func readJson(res *http.Response) (string, error) {
	buf, err := io.ReadAll(res.Body)
	return string(buf), err
}

func ShareToken(shareID string, passwd string) (string, error) {
	buf := bytes.NewBuffer(nil)
	if passwd == "" {
		buf.WriteString(fmt.Sprintf(`{"share_id":"%s"}`, shareID))

	} else {
		buf.WriteString(fmt.Sprintf(`{"share_id":"%s","share_pwd":"%s"}`, shareID, passwd))

	}
	req, _ := http.NewRequest(http.MethodPost, "https://api.aliyundrive.com/v2/share_link/get_share_token", buf)
	res, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, err := readJson(res)
	if err != nil {
		return "", err
	}
	return gjson.Get(body, "share_token").String(), nil
}

func GetShareDownloadURL(driveID string, fileID string, shareID string, shareToken string, accessToken string) (string, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(fmt.Sprintf(`{"drive_id":"%s","file_id":"%s","expire_sec":600,"share_id":"%s"}`, driveID, fileID, shareID))

	req, _ := http.NewRequest(http.MethodPost, "https://api.aliyundrive.com/v2/file/get_share_link_download_url", buf)
	req.Header.Add("content-type", "application/json")
	req.Header.Add("Authorization", "Bearer\t"+accessToken)
	req.Header.Add("x-share-token", shareToken)
	req.Host = "api.aliyundrive.com"
	res, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, err := readJson(res)
	if err != nil {
		return "", err
	}
	return getRealDownloadURL(gjson.Get(body, "download_url").String())
}

func getRealDownloadURL(directURL string) (string, error) {
	log.Printf("获取 %s 重定向后的地址", directURL)
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest(http.MethodGet, directURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("user-agent", "curl/7.87.0")
	req.Header.Add("accept", "*/*")
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	printHttpBody(res)
	log.Printf("http %d", res.StatusCode)
	realURL := res.Header.Get("location")
	if len(realURL) == 0 {
		return "", fmt.Errorf("重定向地址为空")
	}
	log.Printf("下载地址 %s\n", realURL)
	return realURL, nil
}
