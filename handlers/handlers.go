package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"streamcoach/internal/ai"
	"streamcoach/internal/job"
	"streamcoach/internal/queue"
	"streamcoach/internal/video"

	"github.com/google/uuid"
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

// StatusHandler returns the job status
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing job ID", http.StatusBadRequest)
		return
	}

	j, ok := job.Get(id)
	if !ok {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(j)
}

// UploadHandler handles the video upload and starts the async job
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Parse Multipart Form
	r.Body = http.MaxBytesReader(w, r.Body, 1<<30) // 1GB
	err := r.ParseMultipartForm(1 << 30)
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

	// 2. Generate Job ID
	jobID := uuid.New().String()
	job.Create(jobID)

	// 3. Save uploaded file with JobID prefix to avoid collisions
	tempPath := filepath.Join("uploads", jobID+"_"+header.Filename)
	dst, err := os.Create(tempPath)
	if err != nil {
		http.Error(w, "Internal server error saving file", http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(dst, file); err != nil {
		dst.Close()
		http.Error(w, "Internal server error saving file", http.StatusInternalServerError)
		return
	}
	dst.Close()

	// 4. Start Async Processing
	// We pass a background context because the HTTP request context will cancel when this handler returns.
	go processJob(context.Background(), jobID, tempPath, apiKey, category, language)

	// 5. Return Job ID
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"jobId": jobID})
}

// processJob handles the heavy lifting in a goroutine
func processJob(ctx context.Context, jobID, tempPath, apiKey, category, language string) {
	// Clean up file if anything goes wrong or finishes
	defer os.Remove(tempPath)

	log.Printf("[JOB %s] Entering queue...", jobID)
	
	// --- QUEUE ACQUISITION ---
	// Note: We use the background context passed in, but we might want a timeout or cancellation mechanism if the user abandons.
	// For now, since we don't have a way to know if user abandoned (polling stops), we run to completion.
	// Implementing "cancel on abandon" with polling requires a heartbeat or timeout logic in the JobStore, which is more complex.
	// We will stick to "Run to completion" for simplicity as per previous conversation, unless explicit context cancellation is wired up.
	
	release, err := queue.Manager.Acquire(ctx)
	if err != nil {
		log.Printf("[JOB %s] Queue error: %v", jobID, err)
		job.Fail(jobID, "Queue error: "+err.Error())
		return
	}
	defer release()

	log.Printf("[JOB %s] Slot acquired. Processing...", jobID)
	job.UpdateStatus(jobID, job.StatusProcessing)

	// FFmpeg
	log.Printf("[JOB %s] Starting FFmpeg...", jobID)
	processResult, err := video.ProcessVideo(ctx, tempPath, "uploads")
	if err != nil {
		log.Printf("[JOB %s] FFmpeg error: %v", jobID, err)
		job.Fail(jobID, "Video processing failed: "+err.Error())
		return
	}
	
	// Cleanup extracted assets later
	jobDir := filepath.Dir(processResult.AudioPath)
	defer os.RemoveAll(jobDir)

	// Gemini
	log.Printf("[JOB %s] Calling Gemini...", jobID)
	analysis, err := ai.AnalyzeStream(ctx, apiKey, category, language, processResult.AudioPath, processResult.FramePaths)
	if err != nil {
		log.Printf("[JOB %s] Gemini error: %v", jobID, err)
		job.Fail(jobID, "AI Analysis failed: "+err.Error())
		return
	}

	log.Printf("[JOB %s] Success! Score: %d", jobID, analysis.OverallScore)
	job.Complete(jobID, analysis)
}