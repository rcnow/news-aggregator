package utils

import (
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

func FormatDate(dateStr string) string {
	formats := []string{
		time.RFC3339,
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822,
		time.RFC822Z,
		"Mon, 02 Jan 2006 15:04:05 MST",
		"Mon, 02 Jan 2006 15:04:05 -0700",
		"02 Jan 2006 15:04:05 MST",
		"02 Jan 2006 15:04:05 -0700",
	}

	var t time.Time
	var err error

	for _, format := range formats {
		t, err = time.Parse(format, dateStr)
		if err == nil {
			break
		}
	}
	if err != nil {
		return dateStr
	}
	return t.UTC().Format("02.01.2006 15:04")
}

func FilterNewsByTime(newsItems []models.NewsItem, timeFilter time.Duration, sortFilter string) []models.NewsItem {
	var filteredItems []models.NewsItem
	now := time.Now().UTC()

	for _, item := range newsItems {
		pubDate, err := time.Parse("02.01.2006 15:04", item.PubDate)
		if err != nil {
			log.Println("Error parsing date:", err)
			continue
		}
		if now.Sub(pubDate) <= timeFilter {
			filteredItems = append(filteredItems, item)
		}
	}
	log.Printf("FilterNewsByTime - TimeFilter: %v, SortFilter: %s, Total news items: %d, Filtered items: %d\n",
		timeFilter, sortFilter, len(newsItems), len(filteredItems))

	return filteredItems
}

func SortNewsByDate(filteredItems []models.NewsItem, timeFilter time.Duration, sortFilter string) []models.NewsItem {
	sort.Slice(filteredItems, func(i, j int) bool {
		dateI, errI := time.Parse("02.01.2006 15:04", filteredItems[i].PubDate)
		dateJ, errJ := time.Parse("02.01.2006 15:04", filteredItems[j].PubDate)
		if errI != nil || errJ != nil {
			log.Println("Error parsing date for sorting:", errI, errJ)
			return false
		}
		if sortFilter == "asc" {
			return dateI.Before(dateJ)
		} else {
			return dateI.After(dateJ)
		}
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

	return template.HTML(strings.TrimSpace(truncated) + "…")
}
