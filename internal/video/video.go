package video

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ProcessResult holds the paths to extracted assets
type ProcessResult struct {
	AudioPath  string
	FramePaths []string
}

// ProcessVideo extracts frames every 10 seconds and full audio from the video
func ProcessVideo(videoPath string, outputDir string) (*ProcessResult, error) {
	// Create unique subfolder for this processing job
	jobID := filepath.Base(videoPath)
	jobDir := filepath.Join(outputDir, jobID+"_processed")
	err := os.MkdirAll(jobDir, 0755)
	if err != nil {
		return nil, err
	}

	// 1. Extract Audio
	audioPath := filepath.Join(jobDir, "audio.mp3")
	cmdAudio := exec.Command("ffmpeg", "-i", videoPath, "-vn", "-ar", "44100", "-ac", "2", "-b:a", "128k", audioPath)
	if err := cmdAudio.Run(); err != nil {
		return nil, fmt.Errorf("failed to extract audio: %v", err)
	}

	// 2. Extract Frames (every 10 seconds)
	// %03d.jpg will create 001.jpg, 002.jpg, etc.
	framePattern := filepath.Join(jobDir, "frame_%03d.jpg")
	cmdFrames := exec.Command("ffmpeg", "-i", videoPath, "-vf", "fps=1/10", framePattern)
	if err := cmdFrames.Run(); err != nil {
		return nil, fmt.Errorf("failed to extract frames: %v", err)
	}

	// Get list of extracted frames
	frames, err := filepath.Glob(filepath.Join(jobDir, "frame_*.jpg"))
	if err != nil {
		return nil, err
	}

	return &ProcessResult{
		AudioPath:  audioPath,
		FramePaths: frames,
	}, nil
}

// Cleanup removes the temporary files
func Cleanup(videoPath string, jobDir string) {
	os.Remove(videoPath)
	os.RemoveAll(jobDir)
}
