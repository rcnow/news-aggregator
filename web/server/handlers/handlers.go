package handlers

import (
	"html/template"
	"log"
	"net/http"
	"news-aggregator/fetcher"
	"news-aggregator/models"
	"sort"
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

		now := time.Now().UTC()
		filterItems = nil

		for _, item := range newsItems {
			pubDate, err := time.Parse("02.01.2006 15:04", item.PubDate)
			if err != nil {
				log.Println("Error parsing date:", err)
				continue
			}
			if now.Sub(pubDate) <= timeFilter {
				filterItems = append(filterItems, item)
			}
		}

		sort.Slice(filterItems, func(i, j int) bool {
			dateI, errI := time.Parse("02.01.2006 15:04", filterItems[i].PubDate)
			dateJ, errJ := time.Parse("02.01.2006 15:04", filterItems[j].PubDate)
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

		mu.Unlock()

		log.Printf("TimeFilter: %v, SortFilter: %s, Total news items: %d, Filtered items: %d\n",
			timeFilter, sortFilter, len(newsItems), len(filterItems))

		time.Sleep(30 * time.Minute)
	}
}

func HandleStaticFiles() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/templates/static"))))
}

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("index.html").Funcs(template.FuncMap{}).ParseFiles("web/templates/index.html")
	if err != nil {
		log.Println("Error parsing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	todayDate := time.Now().Format("02.01.2006")

	mu.Lock()
	defer mu.Unlock()

	if len(filterItems) == 0 {
		err = tmpl.Execute(w, map[string]interface{}{
			"newsItems":  []models.NewsItem{},
			"todayDate":  todayDate,
			"totalCount": 0,
			"loading":    true,
		})
	} else {
		err = tmpl.Execute(w, map[string]interface{}{
			"newsItems":  filterItems,
			"todayDate":  todayDate,
			"totalCount": len(filterItems),
			"loading":    false,
		})
	}

	// log.Printf("TimeFilter: %v, SortFilter: %s, Total news items: %d, Filtered items: %d\n",
	// 	timeFilter, sortFilter, len(newsItems), len(filterItems))
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
		tmpl = template.Must(template.New("news").Parse(`
                {{ range . }}
                <div class="feed-item">
                    <div class="feed-info">
                        <h3>{{.Title}}</h3>
                        <p>{{.Description}}</p>
                        <span><a href="{{.ChannelLink}}" target="_blank">{{.ChannelTitle}}</a> · {{.PubDate}}</span>
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
		now := time.Now().UTC()
		filterItems = nil
		for _, item := range newsItems {
			pubDate, err := time.Parse("02.01.2006 15:04", item.PubDate)
			if err != nil {
				log.Println("Error parsing date:", err)
				continue
			}
			if now.Sub(pubDate) <= timeFilter {
				filterItems = append(filterItems, item)
			}
		}
	}

	sortOrder := r.Header.Get("Sort-Order")
	if sortOrder == "asc" || sortOrder == "desc" {
		sortFilter = sortOrder
	}

	if len(filterItems) > 1 {
		sort.Slice(filterItems, func(i, k int) bool {
			dateI, errI := time.Parse("02.01.2006 15:04", filterItems[i].PubDate)
			dateK, errK := time.Parse("02.01.2006 15:04", filterItems[k].PubDate)
			if errI != nil || errK != nil {
				log.Println("Error parsing date:", errI, errK)
				return false
			}
			if sortFilter == "asc" {
				return dateI.Before(dateK)
			} else {
				return dateI.After(dateK)
			}
		})
	}

	log.Printf("TimeFilter: %v, SortFilter: %s, Total news items: %d, Filtered items: %d\n",
		timeFilter, sortFilter, len(newsItems), len(filterItems))

	tmpl := template.Must(template.New("news").Parse(`
        {{ range . }}
        <div class="feed-item">
            <div class="feed-info">
                <h3>{{.Title}}</h3>
                <p>{{.Description}}</p>
                <span><a href="{{.ChannelLink}}" target="_blank">{{.ChannelTitle}}</a> · {{.PubDate}}</span>
            </div>
        </div>
        {{ end }}
    `))
	if err := tmpl.Execute(w, filterItems); err != nil {
		log.Println("Error rendering filtered news:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
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
