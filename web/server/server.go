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
	http.HandleFunc("/filter-news", handlers.HandleFilterNews)
	http.HandleFunc("/add-feed", handlers.HandleAddFeedForm)
	http.HandleFunc("/sort-news", handlers.HandleSortNews)

	log.Println("Server is running on http://localhost:8080")
	log.Println("Debug pprof available at http://localhost:8080/debug/pprof/")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
