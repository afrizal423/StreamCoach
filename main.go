package main

import (
	"log"
	"net/http"
	"streamcoach/handlers"
	"streamcoach/internal/queue"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, relying on system environment variables")
	}

	// Initialize Queue
	if err := queue.InitQueue(); err != nil {
		log.Fatalf("Failed to initialize queue: %v", err)
	}

	// Serve static files
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	// Routes
	http.HandleFunc("/", handlers.DashboardHandler)
	http.HandleFunc("/app", handlers.AppHandler)
	http.HandleFunc("/api/analyze", handlers.AnalyzeHandler)

	log.Println("Server starting on :8080...")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
