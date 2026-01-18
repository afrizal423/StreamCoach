package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"streamcoach/internal/ai"
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

	

		// 3. Process Video (FFmpeg)

		processResult, err := video.ProcessVideo(tempPath, "uploads")

		if err != nil {

			os.Remove(tempPath) // Clean up the original video on error

			http.Error(w, "Error processing video: "+err.Error(), http.StatusInternalServerError)

			return

		}

		

		// Delete original video immediately after processing as it's no longer needed

		os.Remove(tempPath)

	

		// Prepare job directory for cleanup after AI analysis

		jobDir := filepath.Dir(processResult.AudioPath)

		defer os.RemoveAll(jobDir) // Ensures audio and frames are deleted when handler returns

	

		// 4. Call Gemini AI

		analysis, err := ai.AnalyzeStream(apiKey, category, language, processResult.AudioPath, processResult.FramePaths)

	
	if err != nil {
		http.Error(w, "AI Analysis failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. Return JSON Result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analysis)
}
