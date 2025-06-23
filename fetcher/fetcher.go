package fetcher

import (
	"io"
	"log"
	"net/http"
	"news-aggregator/models"
	"strings"
	"time"
)

type ErrorURL struct {
	URL   string
	Error string
	Time  time.Time
}

var ErrorURLs []ErrorURL

func FetchNews(feedURL string, category string) []models.NewsItem {
	if feedURL == "" {
		log.Println("Empty feed URL")
	}
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, _ := http.NewRequest("GET", feedURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error fetching feed:", err)
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		ErrorURLs = append(ErrorURLs, ErrorURL{URL: feedURL, Error: resp.Status, Time: time.Now()})
		return nil
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading feed:", err)
		return nil
	}
	CheckError()
	content := string(body)
	//log.Println("Feed content:", content[:400])
	if strings.Contains(content, "<rss") {
		log.Println("RSS feed detected:", feedURL)
		return ParseRSS(body, category)
	} else if strings.Contains(content, "<feed") {
		log.Println("Atom feed detected", feedURL)
		return ParseAtom(body, category)
	} else {
		log.Println("Unknown feed format", feedURL)
		return nil
	}
}

func CheckError() {
	if len(ErrorURLs) > 0 {
		log.Println("Error URLs:")
		for _, errURL := range ErrorURLs {
			log.Printf("Time: %s, URL: %s, Error: %s\n", errURL.Time.Format(time.RFC3339), errURL.URL, errURL.Error)
		}
	} else {
		log.Println("No errors in URLs")
	}
}
