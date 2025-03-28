package utils

import (
	"fmt"
	"html/template"
	"log"
	"news-aggregator/models"
	"sort"
	"strings"
	"time"
)

func StripHTMLTags(html string) string {
	var result strings.Builder
	var inTag bool

	for _, char := range html {
		if char == '<' {
			inTag = true
		} else if char == '>' {
			inTag = false
		} else if !inTag {
			result.WriteRune(char)
		}
	}
	return result.String()
}

func FormatDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("empty date string")
	}

	formats := []string{
		time.RFC3339,
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822,
		time.RFC822Z,
	}

	for _, format := range formats {
		t, err := time.Parse(format, dateStr)
		if err == nil {
			return t.Local(), nil
		}
	}

	return time.Time{}, fmt.Errorf("unrecognized date format: %s", dateStr)
}

func FilterNewsByTime(newsItems []models.NewsItem, timeFilter time.Duration, sortFilter string) []models.NewsItem {
	var filteredItems []models.NewsItem
	now := time.Now().UTC()

	for _, item := range newsItems {
		if now.Sub(item.PubDate) <= timeFilter {
			filteredItems = append(filteredItems, item)
		}
	}

	log.Printf("FilterNewsByTime - TimeFilter: %v, SortFilter: %s, Total news items: %d, Filtered items: %d\n",
		timeFilter, sortFilter, len(newsItems), len(filteredItems))

	return filteredItems
}

func SortNewsByDate(filteredItems []models.NewsItem, timeFilter time.Duration, sortFilter string) []models.NewsItem {
	sort.Slice(filteredItems, func(i, j int) bool {
		if sortFilter == "asc" {
			return filteredItems[i].PubDate.Before(filteredItems[j].PubDate)
		}
		return filteredItems[i].PubDate.After(filteredItems[j].PubDate)
	})

	log.Printf("SortNewsByDate - TimeFilter: %v, SortFilter: %s, Filtered items: %d\n",
		timeFilter, sortFilter, len(filteredItems))

	return filteredItems
}

func TruncateDescription(description template.HTML, maxLen int) template.HTML {
	descStr := string(description)
	if len(descStr) <= maxLen {
		return description
	}

	truncated := descStr[:maxLen]

	if lastSpace := strings.LastIndex(truncated, " "); lastSpace > 0 {
		truncated = truncated[:lastSpace]
	}

	return template.HTML(strings.TrimSpace(truncated) + " …")
}
