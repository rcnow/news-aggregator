package main

import (
	"log"
	"news-aggregator/web/server"
)

func main() {
	log.Println("Starting web server...")
	server.StartWebServer()
}
