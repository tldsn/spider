package main

import (
	"fmt"
	"log"
	"spider/utils/httpclient"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	url0 := "https://www.993dy.com/vod-detail-id-64459.html#"
	headers0 := httpclient.HGetHTML()
	headers0["Host"] = "https://www.993dy.com/vod-detail-id-64459.html#"
	headers0["Cookie"] = "https://www.993dy.com/vod-detail-id-64459.html#"
	result, err := httpclient.Get(url0, "", headers0, 20000)
	if err != nil {
		log.Println(err.Error())
	}
	fmt.Println(result["body"])
	dom, err := goquery.NewDocumentFromReader(strings.NewReader(result["body"]))
	if err != nil {
		log.Println(err.Error())
	}
	dom.Find(".downurl li").Each(func(i int, selection *goquery.Selection) {
		val, _ := selection.Find(".1addr").Eq(0).Attr("value")
		fmt.Println(val)
	})
}
