package pkg

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/mattn/go-runewidth"
	"github.com/schollz/progressbar/v3"
)

func DownloadFile(downloadURL string, filename string, accessToken string) error {
	req, err := http.NewRequest(http.MethodGet, downloadURL, nil)
	if err != nil {
		return err
	}
	req.Header.Add("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://www.aliyundrive.com/")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Add("Authorization", "Bearer\t"+accessToken)
	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, os.ModeDir|os.ModePerm); err != nil {
		return err
	}
	fh, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer fh.Close()
	state, err := fh.Stat()
	if err == nil {
		if state.Size() == res.ContentLength {
			log.Printf("%s 下载完成\n", filename)
			return nil
		}
	}
	bar := progressbar.NewOptions(
		int(res.ContentLength),
		progressbar.OptionSetWriter(os.Stdout),
		progressbar.OptionSetDescription(runewidth.FillRight(path.Base(filename), 40)),
		progressbar.OptionOnCompletion(func() {
			fmt.Printf("\n")
		}),
	)

	if _, err := io.Copy(io.MultiWriter(fh, bar), res.Body); err != nil {
		return err
	}
	return nil
}
