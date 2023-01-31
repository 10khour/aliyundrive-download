package main

import (
	"flag"
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
	log.Printf("%s\n", url)
	shareID, _, _, err := pkg.GetShareInfo(url)
	if err != nil {
		log.Fatal(err)
	}
	shareToken, err := pkg.ShareToken(shareID, "")
	log.Printf("登录成功")

	if err != nil {
		log.Fatal(err)
	}
	files, err := pkg.ListFiles(shareID, "root", shareToken)
	if err != nil {
		log.Fatal(err)
	}
	for index, file := range files {
		downloadURL, err := pkg.GetShareDownloadURL(file.DriveID, file.FileID, file.ShareID, shareToken, accessToken)
		if err != nil {
			log.Fatal(err)
		}
		println(downloadURL)
		if err := pkg.DownloadFile(downloadURL, file.Name, refreshToken); err != nil {
			log.Fatal(err)
		}
		log.Printf("%s 下载完成 [%d/%d] ", file.Name, index+1, len(files))
	}
}
