package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"news-aggregator/fetcher"
	"news-aggregator/models"
	"news-aggregator/utils"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	newsItems   []models.NewsItem
	filterItems []models.NewsItem
	timeFilter  time.Duration
	sortFilter  string
	mu          sync.Mutex
	sseClients  map[chan string]bool
)

func init() {
	sseClients = make(map[chan string]bool)
}

func broadcastUpdate() {
	mu.Lock()
	defer mu.Unlock()

	for client := range sseClients {
		select {
		case client <- "update":
		default:
		}
	}
}
func UpdateNews() {
	if timeFilter == 0 {
		timeFilter = 24 * time.Hour
	}
	if sortFilter == "" {
		sortFilter = "desc"
	}

	feeds := []string{
		"https://cointelegraph.com/rss",
		"https://bitcoinmagazine.com/feed",
		"https://feeds.bloomberg.com/markets/news.rss",
		"https://ir.thomsonreuters.com/rss/news-releases.xml?items=15",
		"https://www.reddit.com/r/birding.rss",
	}

	for {
		var newItems []models.NewsItem

		for _, feed := range feeds {
			news := fetcher.FetchNews(feed)
			newItems = append(newItems, news...)
			mu.Lock()
			newsItems = newItems
			filterItems = utils.FilterNewsByTime(newsItems, timeFilter, sortFilter)
			filterItems = utils.SortNewsByDate(filterItems, timeFilter, sortFilter)
			mu.Unlock()
			broadcastUpdate()
			log.Printf("Fetched %d items from %s\n", len(news), feed)
			time.Sleep(1 * time.Second)
		}
		time.Sleep(30 * time.Minute)
	}
}

func HandleStaticFiles() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/templates/static"))))
}

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("index.html").
		Funcs(template.FuncMap{
			"truncate": func(html template.HTML, length int) string {
				return string(utils.TruncateDescription(html, length))
			},
			"formatDate": func(t time.Time) string {
				return t.Format("02.01.2006 15:04:05")
			},
		}).
		ParseFiles("web/templates/index.html")
	if err != nil {
		log.Println("Error parsing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	todayDate := time.Now().Format("02.01.2006")

	mu.Lock()
	defer mu.Unlock()

	uniqueItems, uniqueСounts := utils.GetUniqueItems(filterItems)

	if len(filterItems) == 0 {
		err = tmpl.Execute(w, map[string]any{
			"newsItems":       []models.NewsItem{},
			"uniqueItems":     []models.NewsItem{},
			"todayDate":       todayDate,
			"totalCount":      0,
			"loading":         true,
			"timeFilterValue": timeFilter,
			"sortFilter":      sortFilter,
		})
	} else {
		err = tmpl.Execute(w, map[string]any{
			"newsItems":       filterItems,
			"uniqueItems":     uniqueItems,
			"uniqueCounts":    uniqueСounts,
			"todayDate":       todayDate,
			"totalCount":      len(filterItems),
			"loading":         false,
			"timeFilterValue": int(timeFilter.Hours()),
			"sortFilter":      sortFilter,
		})
	}

	if err != nil {
		log.Println("Error rendering template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func HandleFilterNewsBySearch(w http.ResponseWriter, r *http.Request) {
	query := r.Header.Get("Search-Query")

	mu.Lock()
	var filteredItems []models.NewsItem
	for _, item := range filterItems {
		if strings.Contains(strings.ToLower(item.Title), strings.ToLower(query)) {
			filteredItems = append(filteredItems, item)
		}
	}
	mu.Unlock()
	uniqueItems, uniqueСounts := utils.GetUniqueItems(filteredItems)

	var tmpl *template.Template
	var err error
	if len(filteredItems) == 0 {
		tmpl = template.Must(template.New("no-news").Parse(`
        <div class="feed-item">
            <h3>No news. Try changing the filter.</h3>
        </div>
        `))
	} else {
		tmpl = template.Must(template.New("news").
			Funcs(template.FuncMap{
				"truncate": func(html template.HTML, length int) string {
					return string(utils.TruncateDescription(html, length))
				},
				"formatDate": func(t time.Time) string {
					return t.Format("02.01.2006 15:04:05")
				},
			}).Parse(`
                {{ range . }}
                    <div class="feed-item">
                    <div class="feed-info">
                        <h3>{{.Title}}</h3>
                        <p>{{ truncate .Description 150 }}</p>
                        <span><a href="{{.ChannelLink}}" target="_blank">{{.ChannelTitle}}</a> &#8226 {{ formatDate .PubDate }}</span>
                    </div>
                </div>
                {{ end }}
        `))
	}

	var feedViewHTML bytes.Buffer
	err = tmpl.Execute(&feedViewHTML, filteredItems)
	if err != nil {
		log.Println("Error rendering filtered news:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"feedViewHTML": feedViewHTML.String(),
		"totalCount":   len(filteredItems),
		"uniqueItems":  uniqueItems,
		"uniqueCounts": uniqueСounts,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func HandleSortNews(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	hoursStr := r.Header.Get("timeFilter")
	if hoursStr != "" {
		hours, err := strconv.Atoi(hoursStr)
		if err != nil {
			hours = 24
		}
		timeFilter = time.Duration(hours) * time.Hour
	}

	sortOrder := r.Header.Get("sortFilter")
	if sortOrder == "asc" || sortOrder == "desc" {
		sortFilter = sortOrder
	}

	filteredItems := utils.FilterNewsByTime(newsItems, timeFilter, sortFilter)
	filteredItems = utils.SortNewsByDate(filteredItems, timeFilter, sortFilter)
	filterItems = filteredItems
	uniqueItems, uniqueСounts := utils.GetUniqueItems(filterItems)
	log.Printf("TimeFilter: %v, SortFilter: %s, Total news items: %d, Filtered items: %d\n",
		timeFilter, sortFilter, len(newsItems), len(filterItems))

	tmpl := template.Must(template.New("news").
		Funcs(template.FuncMap{
			"truncate": func(html template.HTML, length int) string {
				return string(utils.TruncateDescription(html, length))
			},
			"formatDate": func(t time.Time) string {
				return t.Format("02.01.2006 15:04:05")
			},
		}).Parse(`
        {{ range .newsItems }}
        <div class="feed-item">
            <div class="feed-info">
                <h3>{{.Title}}</h3>
                <p>{{ truncate .Description 150 }}</p>
                <span><a href="{{.ChannelLink}}" target="_blank">{{.ChannelTitle}}</a> &#8226 {{formatDate .PubDate}}</span>
            </div>
        </div>
        {{ end }}
    `))

	var feedViewHTML bytes.Buffer
	err := tmpl.Execute(&feedViewHTML, map[string]interface{}{
		"newsItems": filterItems,
	})
	if err != nil {
		log.Println("Error rendering feed-view template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"feedViewHTML":    feedViewHTML.String(),
		"totalCount":      len(filterItems),
		"timeFilterValue": int(timeFilter.Hours()),
		"sortFilter":      sortFilter,
		"uniqueItems":     uniqueItems,
		"uniqueCounts":    uniqueСounts,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func HandleAddFeedForm(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("add-feed-form").Parse(`
	            <h2>Add new feed </h2>
            <div class="main-info">
                <p>add feed</p>
				<input type="text">
                <a href="#" class="button">add feed</a>
            </div>
`))
	err := tmpl.Execute(w, nil)
	if err != nil {
		log.Println("Error rendering add feed form:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func HandleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	lastEventID := r.Header.Get("Last-Event-ID")
	log.Printf("Last-Event-ID: %s", lastEventID)
	messageChan := make(chan string)

	mu.Lock()
	sseClients[messageChan] = true
	mu.Unlock()

	defer func() {
		mu.Lock()
		delete(sseClients, messageChan)
		close(messageChan)
		mu.Unlock()
	}()

	fmt.Fprintf(w, "event: init\ndata: connected\n\n")
	flusher.Flush()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	var eventID int64 = time.Now().UnixNano()
	for {
		select {
		case msg := <-messageChan:
			eventID++
			fmt.Fprintf(w, "id: %d\nevent: update\ndata: %s\n\n", eventID, msg)

			flusher.Flush()
		case <-ticker.C:
			fmt.Fprintf(w, "id: %d\nevent: ping\ndata: keep-alive\n\n", eventID)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func HandleLoadNews(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	if timeFilter == 0 {
		timeFilter = 24 * time.Hour
	}
	if sortFilter == "" {
		sortFilter = "desc"
	}

	loading := len(filterItems) == 0
	uniqueItems, uniqueСounts := utils.GetUniqueItems(filterItems)

	tmpl := template.Must(template.New("news").Funcs(template.FuncMap{
		"truncate": func(html template.HTML, length int) string {
			return string(utils.TruncateDescription(html, length))
		},
		"formatDate": func(t time.Time) string {
			return t.Format("02.01.2006 15:04:05")
		},
	}).Parse(`
        {{ if .loading }}
        <div id="loading" class="loading">
            <h3>Loading...</h3>
        </div>
        {{ else }}
        {{ range .newsItems }}
        <div class="feed-item">
            <div class="feed-info">
                <h3>{{.Title}}</h3>
                <p>{{ truncate .Description 150 }}</p>
                <span><a href="{{.ChannelLink}}" target="_blank">{{.ChannelTitle}}</a> &#8226 {{ formatDate .PubDate }}</span>
            </div>
        </div>
        {{ end }}
        {{ end }}
    `))

	var feedViewHTML bytes.Buffer
	err := tmpl.Execute(&feedViewHTML, map[string]interface{}{
		"newsItems": filterItems,
		"loading":   loading,
	})
	if err != nil {
		log.Println("Error rendering news template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"feedViewHTML":    feedViewHTML.String(),
		"totalCount":      len(filterItems),
		"timeFilterValue": int(timeFilter.Hours()),
		"sortFilter":      sortFilter,
		"uniqueItems":     uniqueItems,
		"uniqueCounts":    uniqueСounts,
	}
	log.Printf("HandleLoadNews - Filtered: items %d, uniqueCounts: %d", len(filterItems), len(uniqueСounts))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func HandleFilterNewsByLink(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	linkQuery := r.Header.Get("Link")
	if linkQuery == "" {
		http.Error(w, "Link header is required", http.StatusBadRequest)
		return
	}

	filteredByLink := utils.FilterNewsByLink(filterItems, linkQuery)

	uniqueItems, uniqueCounts := utils.GetUniqueItems(filterItems)

	tmpl := template.Must(template.New("news").
		Funcs(template.FuncMap{
			"truncate": func(html template.HTML, length int) string {
				return string(utils.TruncateDescription(html, length))
			},
			"formatDate": func(t time.Time) string {
				return t.Format("02.01.2006 15:04:05")
			},
		}).Parse(`
        {{ range . }}
        <div class="feed-item">
            <div class="feed-info">
                <h3>{{.Title}}</h3>
                <p>{{ truncate .Description 150 }}</p>
                <span><a href="{{.ChannelLink}}" target="_blank">{{.ChannelTitle}}</a> • {{ formatDate .PubDate }}</span>
            </div>
        </div>
        {{ end }}
    `))

	var feedViewHTML bytes.Buffer
	if err := tmpl.Execute(&feedViewHTML, filteredByLink); err != nil {
		log.Println("Error rendering template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"feedViewHTML": feedViewHTML.String(),
		"totalCount":   len(filterItems),
		"uniqueItems":  uniqueItems,
		"uniqueCounts": uniqueCounts,
	}
	log.Printf("HandleSortNewsByLink - Filtered: items %d, uniqueCounts: %d", len(filterItems), len(uniqueCounts))
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Println("Error encoding JSON:", err)
	}
}
