package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"streamcoach/internal/ai"
	"streamcoach/internal/queue"
	"streamcoach/internal/video"
)

// DashboardHandler serves the landing page
func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, filepath.Join("views", "dashboard.html"))
}

// AppHandler serves the main application page
func AppHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("views", "app.html"))
}

// AnalyzeHandler handles the video analysis request
func AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Parse Multipart Form
	// Limit request body size to 1GB (1 << 30)
	r.Body = http.MaxBytesReader(w, r.Body, 1<<30)
	
	// ParseMultipartForm maxMemory determines how much is stored in memory vs disk.
	err := r.ParseMultipartForm(1<<30)
	if err != nil {
		http.Error(w, "File too large (Max 1GB) or invalid form", http.StatusBadRequest)
		return
	}

		file, header, err := r.FormFile("video")
		if err != nil {
			http.Error(w, "Video file is required", http.StatusBadRequest)
			return
		}
		defer file.Close()
	
		log.Printf("[UPLOAD] Received file: %s (%d bytes)", header.Filename, header.Size)
	
		apiKey := r.FormValue("apiKey")
			category := r.FormValue("category")

		language := r.FormValue("language")

		if language == "" {

			language = "English"

		}

	

		if apiKey == "" || category == "" {

			http.Error(w, "API Key and Category are required", http.StatusBadRequest)

			return

		}

	

		// 2. Save uploaded file

		tempPath := filepath.Join("uploads", header.Filename)

		dst, err := os.Create(tempPath)

		if err != nil {

			http.Error(w, "Internal server error saving file", http.StatusInternalServerError)

			return

		}

		if _, err := io.Copy(dst, file); err != nil {

			dst.Close() // Close on error

			http.Error(w, "Internal server error saving file", http.StatusInternalServerError)

			return

		}

				dst.Close() // Close immediately to release file lock for FFmpeg/Deletion

			

				// --- QUEUE ACQUISITION START ---

				log.Println("[QUEUE] User entering queue...")

				release, err := queue.Manager.Acquire(r.Context())

				if err != nil {

					log.Printf("[QUEUE] User left queue or error: %v", err)

					os.Remove(tempPath)

					http.Error(w, "Service busy or request cancelled: "+err.Error(), http.StatusServiceUnavailable)

					return

				}

				defer release()

				log.Println("[QUEUE] Slot acquired. Starting process...")

			

				// 3. Process Video (FFmpeg)

				log.Println("[VIDEO] Starting FFmpeg processing (extracting audio and frames)...")

				processResult, err := video.ProcessVideo(r.Context(), tempPath, "uploads")

				if err != nil {

					log.Printf("[VIDEO] FFmpeg error: %v", err)

					os.Remove(tempPath)

					http.Error(w, "Error processing video: "+err.Error(), http.StatusInternalServerError)

					return

				}

				log.Println("[VIDEO] FFmpeg processing complete.")

				

				os.Remove(tempPath)

			

				jobDir := filepath.Dir(processResult.AudioPath)

				defer os.RemoveAll(jobDir)

			

				// 4. Call Gemini AI

				log.Printf("[AI] Sending request to Gemini 3 Flash Preview (Category: %s, Lang: %s)...", category, language)

				analysis, err := ai.AnalyzeStream(r.Context(), apiKey, category, language, processResult.AudioPath, processResult.FramePaths)

				if err != nil {

					log.Printf("[AI] Gemini error: %v", err)

					http.Error(w, "AI Analysis failed: "+err.Error(), http.StatusInternalServerError)

					return

				}

				log.Printf("[AI] Received response from Gemini. Score: %d", analysis.OverallScore)

			

				// 5. Return JSON Result

			
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analysis)
}
