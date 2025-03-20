package fetcher

import (
	"io"
	"log"
	"net/http"
	"news-aggregator/models"
	"strings"
)

func FetchNews(feedURL string) []models.NewsItem {
	resp, err := http.Get(feedURL)
	if err != nil {
		log.Println("Error fetching feed:", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading feed:", err)
		return nil
	}

	content := string(body)
	//log.Println("Feed content:", content)
	if strings.Contains(content, "<rss") {
		log.Println("RSS feed detected:")
		return ParseRSS(body)
	} else if strings.Contains(content, "<feed") {
		log.Println("Atom feed detected")
		return ParseAtom(body)
	} else {
		log.Println("Unknown feed format")
		return nil
	}
}
