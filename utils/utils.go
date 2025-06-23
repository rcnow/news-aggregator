package utils

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"news-aggregator/models"
	"regexp"
	"sort"
	"strings"
	"time"
)

type UniqueItemsResult struct {
	Items       []models.NewsItem
	Counts      map[string]int
	FaviconURLs map[string]string
}

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
		"Mon, 02 Jan 2006 15:04:05 -0700",
		"Mon, 02 Jan 2006 15:04:05 +0000",
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

	// log.Printf("FilterNewsByTime - TimeFilter: %v, SortFilter: %s, Total news items: %d, Filtered items: %d\n",
	// 	timeFilter, sortFilter, len(newsItems), len(filteredItems))

	return filteredItems
}

func SortByDirection(filteredItems []models.NewsItem, timeFilter time.Duration, sortFilter string) []models.NewsItem {
	sort.Slice(filteredItems, func(i, j int) bool {
		if sortFilter == "asc" {
			return filteredItems[i].PubDate.Before(filteredItems[j].PubDate)
		}
		return filteredItems[i].PubDate.After(filteredItems[j].PubDate)
	})

	// log.Printf("SortNewsByDate - TimeFilter: %v, SortFilter: %s, Filtered items: %d\n",
	// 	timeFilter, sortFilter, len(filteredItems))

	return filteredItems
}

func FilterNewsByLink(filteredItems []models.NewsItem, link string) []models.NewsItem {
	if link == "" {
		return filteredItems
	}

	filtered := make([]models.NewsItem, 0, len(filteredItems))
	for _, item := range filteredItems {
		if item.ChannelLink == link {
			filtered = append(filtered, item)
		}
	}

	return filtered
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

func GetUniqueItems(items []models.NewsItem) UniqueItemsResult {
	uniqueLinks := make(map[string]models.NewsItem)
	uniqueCounts := make(map[string]int)
	faviconURLs := make(map[string]string)

	for _, item := range items {
		if _, exists := uniqueLinks[item.ChannelLink]; !exists {
			uniqueLinks[item.ChannelLink] = item
			if item.Favicon != "" {
				faviconURLs[item.ChannelLink] = item.Favicon
			} else {
				faviconURLs[item.ChannelLink] = GetFaviconURL(item.ChannelLink)
			}
		}
		uniqueCounts[item.ChannelLink]++
	}

	uniqueItems := make([]models.NewsItem, 0, len(uniqueLinks))
	for link := range uniqueLinks {
		uniqueItems = append(uniqueItems, uniqueLinks[link])
	}

	sort.Slice(uniqueItems, func(i, j int) bool {
		return uniqueCounts[uniqueItems[i].ChannelLink] > uniqueCounts[uniqueItems[j].ChannelLink]
	})

	return UniqueItemsResult{
		Items:       uniqueItems,
		Counts:      uniqueCounts,
		FaviconURLs: faviconURLs,
	}
}

var faviconCache = make(map[string]string)

func GetFaviconURL(link string) string {
	u, err := url.Parse(link)
	if err != nil {
		return ""
	}
	host := u.Host

	// Кеш
	if cached, ok := faviconCache[host]; ok {
		return cached
	}

	// Попытка достать favicon из <link rel="icon"> в HTML
	home := fmt.Sprintf("%s://%s", u.Scheme, host)
	client := http.Client{Timeout: 3 * time.Second} // добавим таймаут
	resp, err := client.Get(home)
	if err == nil {
		defer resp.Body.Close()
		bodyBytes, err := io.ReadAll(resp.Body)
		if err == nil {
			body := string(bodyBytes)
			re := regexp.MustCompile(`(?i)<link[^>]+rel=["'][^"']*icon[^"']*["'][^>]*href=["']?([^"'>\s]+)["']?`)
			matches := re.FindStringSubmatch(body)
			if len(matches) >= 2 {
				iconHref := matches[1]
				iconURL, err := url.Parse(iconHref)
				if err == nil {
					if !iconURL.IsAbs() {
						iconURL = u.ResolveReference(iconURL)
					}
					faviconCache[host] = iconURL.String()
					return iconURL.String()
				}
			}
		}
	}

	// fallback на стандартный путь
	defaultFavicon := fmt.Sprintf("%s/favicon.ico", home)
	faviconCache[host] = defaultFavicon
	return defaultFavicon
}
