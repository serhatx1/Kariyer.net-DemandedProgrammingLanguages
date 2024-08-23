package main

import (
	"github.com/gocolly/colly/v2"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	c := colly.NewCollector(colly.AllowURLRevisit())

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "en-US,en;q=0.5")

	})

	topLanguages := []string{
		"JavaScript", "Python", "Java", "C#", "PHP", "C++", "TypeScript", "Ruby", "Go",
	}

	var validLinks []string
	var validLanguages map[string][]string = make(map[string][]string)

	var mu sync.Mutex

	c.OnHTML(".job-detail-content", func(e *colly.HTMLElement) {
		content := e.Text
		log.Printf("Job Details: %s\n", content)

		foundLanguages := []string{}
		for _, lang := range topLanguages {
			if strings.Contains(content, lang) {
				foundLanguages = append(foundLanguages, lang)
				log.Printf("Found language: %s\n", lang)
			}
		}

		if len(foundLanguages) > 0 {
			url := e.Request.URL.String()
			mu.Lock()
			validLinks = append(validLinks, url)
			validLanguages[url] = foundLanguages
			mu.Unlock()
		}
	})
	c.OnHTML("div.list-items", func(e *colly.HTMLElement) {
		e.ForEach("a", func(aIndex int, aElement *colly.HTMLElement) {
			href := aElement.Attr("href")
			if href != "" {
				fullURL := "https://www.kariyer.net" + href
				log.Println("Found link:", fullURL)

				go func(url string) {
					err := c.Visit(url)
					if err != nil {
						log.Printf("Error visiting link %s: %v\n", url, err)
					}
				}(fullURL)
			}
		})
	})
	var wg sync.WaitGroup
	page := 1
	for {
		time.Sleep(time.Second * 2)

		url := "https://www.kariyer.net/is-ilanlari?cp=" + strconv.Itoa(page)
		log.Printf("Visiting page %d: %s\n", page, url)

		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			err := c.Visit(u)
			if err != nil {
				log.Printf("Error visiting page %d: %v\n", page, err)
			}
		}(url)

		c.Wait()

		if page < 2 {
			page++
		} else {
			break
		}
	}
	wg.Wait()
	log.Println("Valid job links and found languages:")
	for link, languages := range validLanguages {
		log.Printf("Link: %s, Languages: %v\n", link, languages)
	}
}
