package server

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"news-aggregator/web/server/handlers"
)

func StartWebServer() {
	go handlers.UpdateNews()

	handlers.HandleStaticFiles()
	http.HandleFunc("/", handlers.HandleIndex)
	http.HandleFunc("/sse", handlers.HandleSSE)
	http.HandleFunc("/load-news", handlers.HandleLoadNews)
	http.HandleFunc("/home-view", handlers.HandleHomeView)
	http.HandleFunc("/setting-view", handlers.HandleSettingView)
	http.HandleFunc("/add-feed", handlers.HandleAddFeedView)
	http.HandleFunc("/help-view", handlers.HandleHelpView)
	http.HandleFunc("/filter-by-search", handlers.HandleFilterNewsBySearch)
	http.HandleFunc("/filter-by-link", handlers.HandleFilterNewsByLink)
	http.HandleFunc("/sort-news", handlers.HandleSortNews)

	log.Println("Server is running on http://localhost:8080")
	log.Println("Debug pprof available at http://localhost:8080/debug/pprof/")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
