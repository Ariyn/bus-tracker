package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func ExampleScrape() {
	// Request the HTML page.
	res, err := http.Get("http://metalsucks.net")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Find the review items
	doc.Find(".left-content article .post-title").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the title
		title := s.Find("a").Text()
		fmt.Printf("Review %d: %s\n", i, title)
	})

	// raw XPATH not works
	xpath := "/html/body/section[3]/div/div/article[5]/div[2]/span[1]/a"
	link := doc.Find(xpath).First()
	log.Println(link.Text())

	// converted selector works
	convertedSelector := "html > body > section:nth-of-type(3) > div > div > article:nth-of-type(5) > div:nth-of-type(2) > span:first-of-type > a"
	link2 := doc.Find(convertedSelector)
	log.Println(link2.Text())
}

func main() {
	ExampleScrape()
}
