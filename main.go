package main

import (
	"log"
	"net/http"
	"streamcoach/handlers"
)

func main() {
	// Serve static files
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	// Routes
	http.HandleFunc("/", handlers.DashboardHandler)
	http.HandleFunc("/app", handlers.AppHandler)
	http.HandleFunc("/api/analyze", handlers.AnalyzeHandler)

	log.Println("Server starting on :8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
