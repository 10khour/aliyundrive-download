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
	flag.StringVar(&url, "url", "https://www.aliyundrive.com/s/ydCwbMCNqgG", "分享网址")
	flag.Parse()
}

func main() {
	log.Printf("%s\n", url)
	shareID, _, _, err := pkg.GetShareInfo(url)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("抓取文件列表...")
	token, err := pkg.ShareToken(shareID, "")
	if err != nil {
		log.Fatal(err)
	}
	files, err := pkg.ListFiles(shareID, "root", token)
	if err != nil {
		log.Fatal(err)
	}
	for index, file := range files {
		log.Printf("[%d/%d] %s", index+1, len(files), file.Name)
	}
}