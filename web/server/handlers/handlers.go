package handlers

import (
	"bytes"
	"encoding/json"
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
)

func UpdateNews() {
	for {
		feeds := []string{
			"https://cointelegraph.com/rss",
			"https://bitcoinmagazine.com/feed",
			"https://feeds.bloomberg.com/markets/news.rss",
			"https://ir.thomsonreuters.com/rss/news-releases.xml?items=15",
			"https://www.reddit.com/r/birding.rss",
			//"https://feeds.bbci.co.uk/news/world/rss.xml",
		}

		var newItems []models.NewsItem
		for _, feed := range feeds {
			news := fetcher.FetchNews(feed)
			newItems = append(newItems, news...)
		}

		mu.Lock()
		newsItems = newItems

		timeFilter = 24 * time.Hour
		sortFilter = "desc"

		filterItems = newsItems
		filterItems = utils.FilterNewsByTime(filterItems, timeFilter, sortFilter)
		filterItems = utils.SortNewsByDate(filterItems, timeFilter, sortFilter)

		mu.Unlock()

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

	uniqueLinks := make(map[string]models.NewsItem)
	for _, item := range filterItems {
		if _, exists := uniqueLinks[item.ChannelLink]; !exists {
			uniqueLinks[item.ChannelLink] = item
		}
	}

	var uniqueItems []models.NewsItem
	for _, item := range uniqueLinks {
		uniqueItems = append(uniqueItems, item)
	}

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

func HandleLoadNews(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("query")

	mu.Lock()
	var filteredItems []models.NewsItem
	for _, item := range filterItems {
		if strings.Contains(strings.ToLower(item.Title), strings.ToLower(query)) {
			filteredItems = append(filteredItems, item)
		}
	}
	mu.Unlock()

	var tmpl *template.Template
	if len(filteredItems) == 0 {
		tmpl = template.Must(template.New("no-news").Parse(`
		<div class="feed-item">
			<h3 >No news. Try changing the filter.</h3>
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
                {{ range .NewsItems }}
                    <div class="feed-item">
                    <div class="feed-info">
                        <h3>{{.Title}}</h3>
                        <p>{{ truncate .Description 150 }}</p>
                        <span><a href="{{.ChannelLink}}" target="_blank">{{.ChannelTitle}}</a> · {{ formatDate .PubDate }}</span>
                    </div>
                </div>
                {{ end }}
		`))
	}

	err := tmpl.Execute(w, filteredItems)
	if err != nil {
		log.Println("Error rendering filtered news:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func HandleSortNews(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	hoursStr := r.URL.Query().Get("hours")
	if hoursStr != "" {
		hours, err := strconv.Atoi(hoursStr)
		if err != nil {
			hours = 24
		}
		timeFilter = time.Duration(hours) * time.Hour
	}

	sortOrder := r.Header.Get("Sort-Order")
	if sortOrder == "asc" || sortOrder == "desc" {
		sortFilter = sortOrder
	}

	filteredItems := utils.FilterNewsByTime(newsItems, timeFilter, sortFilter)
	filteredItems = utils.SortNewsByDate(filteredItems, timeFilter, sortFilter)
	filterItems = filteredItems

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
                <span><a href="{{.ChannelLink}}" target="_blank">{{.ChannelTitle}}</a> · {{formatDate .PubDate}}</span>
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
