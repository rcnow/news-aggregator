package handlers

import (
	"html/template"
	"log"
	"net/http"
	"news-aggregator/fetcher"
	"news-aggregator/models"
	"news-aggregator/utils"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	newsItems []models.NewsItem
	mu        sync.Mutex
)

func UpdateNews() {
	for {
		feeds := []string{
			"https://cointelegraph.com/rss",
			"https://bitcoinmagazine.com/feed",
		}
		var items []models.NewsItem
		for _, feed := range feeds {
			news, _ := fetcher.FetchNews(feed)
			//log.Printf("Loaded %d news items from %s\n", len(news), feed)
			// log.Printf("Loaded %d news items from %s (ChannelLink: %s)\n", len(news), feed, channelLink)
			items = append(items, news...)
		}
		mu.Lock()
		newsItems = items
		mu.Unlock()
		log.Printf("Total news items: %d\n", len(newsItems))
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
	todayDate := utils.GetCurrentDate()

	mu.Lock()
	defer mu.Unlock()

	if len(newsItems) == 0 {
		err = tmpl.Execute(w, map[string]interface{}{
			"newsItems":  []models.NewsItem{},
			"todayDate":  todayDate,
			"totalCount": 0,
			"loading":    true,
		})
	} else {
		err = tmpl.Execute(w, map[string]interface{}{
			"newsItems":  newsItems,
			"todayDate":  todayDate,
			"totalCount": len(newsItems),
			"loading":    false,
		})
	}

	if err != nil {
		log.Println("Error rendering template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func HandleFilterNews(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("query")

	mu.Lock()
	var filteredItems []models.NewsItem
	for _, item := range newsItems {
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

	sortOrder := r.Header.Get("Sort-Order")
	if len(newsItems) > 1 {
		sort.Slice(newsItems, func(i, k int) bool {

			dateI, errI := time.Parse("02.01.2006 15:04", newsItems[i].PubDate)
			dateK, errK := time.Parse("02.01.2006 15:04", newsItems[k].PubDate)
			if errI != nil || errK != nil {
				log.Println("Error parsing date:", errI, errK)
				return false
			}

			if sortOrder == "asc" {
				return dateI.Before(dateK)
			} else {
				return dateI.After(dateK)
			}
		})
	}
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
	err := tmpl.Execute(w, newsItems)
	if err != nil {
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
