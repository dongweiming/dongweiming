package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gocolly/colly"
)

const (
	MAX_LIMIT int = 56
)

type Author_ struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ResharedStatus_ struct {
	Author Author_ `json:"Author"`
}

type Card_ struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

type Status_ struct {
	ID             string          `json:"id"`
	Card           Card_           `json:"card,omitempty"`
	Text           string          `json:"text,omitempty"`
	Activity       string          `json:"activity"`
	SharingURL     string          `json:"sharing_url"`
	ResharedStatus ResharedStatus_ `json:"reshared_status,omitempty"`
}

type Item struct {
	Status Status_ `json:"status"`
}

type Result struct {
	Count int    `json:"count"`
	Items []Item `json:"items"`
}

func WriteToFile(path string, items []string) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err = scanner.Err(); err != nil {
		log.Fatal(err)
	}

	f.Close()

	f, err = os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "<!-- DOUBAN-ACTIVITIES:START -->" {
			for _, item := range items {
				fmt.Fprintln(w, item)
			}
		}
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

func main() {
	items := []string{}
	url := "https://m.douban.com/rexxar/api/v2/status/user_timeline/62943420?for_mobile=1"
	RefererURL := "https://m.douban.com/people/62943420/statuses"
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"),
	)

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Referer", RefererURL)
	})

	c.OnError(func(_ *colly.Response, err error) {
		fmt.Println("Something went wrong:", err)
	})

	c.OnResponse(func(r *colly.Response) {
		rs := &Result{}
		if err := json.Unmarshal(r.Body, rs); err != nil {
			log.Fatal(err)
		}
		var (
			text                 []rune
			url, activity, item_ string
		)

		for _, item := range rs.Items {
			ResharedStatus := item.Status.ResharedStatus
			url = item.Status.SharingURL
			activity = item.Status.Activity
			if ResharedStatus != (ResharedStatus_{}) {
				item_ = fmt.Sprintf("- [%s %s 的动态](%s)", activity, ResharedStatus.Author.Name, url)
			} else {
				if activity == "说" {
					text = []rune(item.Status.Text)
					if len(text) > MAX_LIMIT {
						activity = fmt.Sprintf("说: %s...", string(text[:MAX_LIMIT]))
					} else {
						activity = fmt.Sprintf("说: %s", item.Status.Text)
					}
				} else if !strings.HasPrefix(activity, "转发") {
					activity = string([]rune(activity)[:2])
				}
				item_ = fmt.Sprintf("- [%s %s](%s)", activity, item.Status.Card.Title, url)
			}
			items = append(items, item_)
		}
	})

	c.OnScraped(func(r *colly.Response) {
		WriteToFile("README.md", items)
	})

	c.Visit(url)
}
