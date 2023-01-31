package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/10khour/aliyundrive-download/pkg"
)

var (
	url string
)

func init() {
	flag.StringVar(&url, "url", "https://www.aliyundrive.com/s/Q7RLN7WEbrx", "分享网址")
	flag.Parse()
}

func main() {
	accessToken, refreshToken, err := pkg.Login()
	if err != nil {
		log.Fatalf("登录失败 %s", err)
	}
	log.Printf("登录成功")
	log.Printf("%s\n", url)
	shareID, _, _, err := pkg.GetShareInfo(url)
	if err != nil {
		log.Fatal(err)
	}
	shareToken, err := pkg.ShareToken(shareID, "")

	if err != nil {
		log.Fatal(err)
	}
	files, err := pkg.ListFiles(shareID, "root", shareToken)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("解析到 %d 个文件，即将开始下载", len(files))
	for index, file := range files {
		downloadURL, err := pkg.GetShareDownloadURL(file.DriveID, file.FileID, file.ShareID, shareToken, accessToken)
		if err != nil {
			log.Fatal(err)
		}
		if err := pkg.DownloadFile(fmt.Sprintf("[%d/%d] %s", index+1, len(files), file.Name), downloadURL, file.Name, refreshToken); err != nil {
			log.Fatal(err)
		}
	}
}
