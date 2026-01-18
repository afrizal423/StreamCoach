package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// AnalysisResult matches the JSON structure requested in GEMINI.md
type AnalysisResult struct {
	OverallScore     int      `json:"overall_score"`
	SummaryReasoning string   `json:"summary_reasoning"`
	Strengths        []string `json:"strengths"`
	Weaknesses       []string `json:"weaknesses"`
	TimelineFlags    []Flag   `json:"timeline_flags"`
	ImprovementTips  string   `json:"improvement_tips"`
}

type Flag struct {
	Time     string `json:"time"`
	Issue    string `json:"issue"`
	Severity string `json:"severity"`
}

// AnalyzeStream calls Gemini 3 Flash Preview with the multimodal data
func AnalyzeStream(apiKey string, category string, language string, audioPath string, framePaths []string) (*AnalysisResult, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-3-flash-preview")
	
	// Set response MIME type to application/json
	model.ResponseMIMEType = "application/json"

	var prompt []genai.Part

	// Add System Instruction / Role
	instruction := fmt.Sprintf(`Role: You are a professional Live Stream Audit Consultant.
Context: The user is conducting a live stream in the %s category.
Data: I am providing the FULL AUDIO from the session and several sampled IMAGE FRAMES from the video (sampled every 10 seconds).
Output Language: %s.

Task:
1. Analyze the AUDIO for tone, enthusiasm, articulation, and sales persuasiveness.
2. Analyze the VISUALS (Frames) for lighting, product focus/clarity, and host engagement (eye contact/gestures).
3. Provide an objective Overall Score (0-100).
4. Identify specific moments (timestamps) where issues occurred (e.g., blur, dead air, low energy). Use the frame sequence to estimate time (Frame 1 is ~0s, Frame 2 is ~10s, etc.).

Constraint: Output MUST be in JSON format with the following structure. 
IMPORTANT: All text values within the JSON (summary, strengths, weaknesses, tips, issues) MUST be written in %s language:
{
  "overall_score": 78,
  "summary_reasoning": "...",
  "strengths": ["..."],
  "weaknesses": ["..."],
  "timeline_flags": [
    {"time": "00:15", "issue": "Audio noise/Dead air", "severity": "medium"}
  ],
  "improvement_tips": "..."
}`, category, language, language)

	prompt = append(prompt, genai.Text(instruction))

	// Add Audio
	audioData, err := ioutil.ReadFile(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio file: %v", err)
	}
	prompt = append(prompt, genai.Blob{
		MIMEType: "audio/mpeg",
		Data:     audioData,
	})

	// Add Image Frames (limit to avoid hitting token/size limits if many)
	maxFrames := 15 // Roughly 2.5 minutes of video at 1 frame per 10s
	for i, path := range framePaths {
		if i >= maxFrames {
			break
		}
		imgData, err := ioutil.ReadFile(path)
		if err != nil {
			log.Printf("Warning: failed to read frame %s: %v", path, err)
			continue
		}
		prompt = append(prompt, genai.Blob{
			MIMEType: "image/jpeg",
			Data:     imgData,
		})
	}

	resp, err := model.GenerateContent(ctx, prompt...)
	if err != nil {
		return nil, fmt.Errorf("gemini generation failed: %v", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return nil, fmt.Errorf("no response from gemini")
	}

	// Extract JSON from response
	var result AnalysisResult
	
	// Combine all text parts
	var sb strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			sb.WriteString(string(text))
		}
	}

	jsonStr := sb.String()
	// Basic sanitization in case Gemini wraps in ```json ... ```
	jsonStr = strings.TrimPrefix(jsonStr, "```json")
	jsonStr = strings.TrimSuffix(jsonStr, "```")
	jsonStr = strings.TrimSpace(jsonStr)

	// Handle case where Gemini returns an array [ { ... } ]
	if strings.HasPrefix(jsonStr, "[") && strings.HasSuffix(jsonStr, "]") {
		var results []AnalysisResult
		if err := json.Unmarshal([]byte(jsonStr), &results); err != nil {
			return nil, fmt.Errorf("failed to parse gemini response array: %v\nResponse: %s", err, jsonStr)
		}
		if len(results) > 0 {
			result = results[0]
		} else {
			return nil, fmt.Errorf("gemini returned empty array")
		}
	} else {
		// Handle case where Gemini returns a single object { ... }
		if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
			return nil, fmt.Errorf("failed to parse gemini response: %v\nResponse: %s", err, jsonStr)
		}
	}

	return &result, nil
}
