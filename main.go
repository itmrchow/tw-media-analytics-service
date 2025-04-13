package main

import (
	"fmt"
	"log"

	"itmrchow/tw-media-analytics-service/domain/spider"
)

func main() {
	spider := spider.NewCtiNewsSpider()

	// news, err := spider.GetNews("rbW4LBXoWL")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println(news)

	// spider := spider.NewSetnSpider()

	// news, err := spider.GetNews("1639490")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println(news)

	newsIDs, err := spider.GetNewsIdList()
	if err != nil {
		log.Fatal(err)
	}

	newsList, err := spider.GetNewsList(newsIDs)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(newsList)
}
