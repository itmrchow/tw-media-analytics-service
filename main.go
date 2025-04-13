package main

import (
	"encoding/json"
	"fmt"
	"log"

	"itmrchow/tw-media-analytics-service/domain/spider"
)

func main() {
	// 設定新聞ID
	// newsID := 1638406 // 這裡可以改成您要抓取的新聞ID

	spider := spider.NewSetnSpider()

	newsIDList, err := spider.GetNewsIdList()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(newsIDList)

	// TODO: 判斷哪些id沒有處理過

	newsDataList, err := spider.GetNewsList(newsIDList)
	if err != nil {
		log.Fatal(err)
	}

	jsonData, err := json.Marshal(newsDataList)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(jsonData))

	// newsData, err := spider.GetNews(newsID)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// jsonData, err := json.Marshal(newsData)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(string(jsonData))
}
