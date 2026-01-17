# Project Context: StreamCoach AI (Live Stream Performance Analyzer)

## 1. Project Overview
"StreamCoach AI" is a web application designed to analyze the performance quality of live streams automatically using **Google Gemini 3 Flash (Multimodal)**. This tool is **General Purpose** (applicable for Jewelry, Fashion, Gaming, General Sales, etc.), but the hackathon demo will focus on **Jewelry** and **Fashion** use cases.

### Core Value Proposition
* **Speed & Cost:** Utilizes Gemini 3 Flash for low-latency and cost-effective analysis.
* **Multimodal:** Assesses performance based on Visuals (Video Frames) and Audio (Tone/Vocal).
* **User-Owned API Key:** Users provide their own API Key (BYOK - Bring Your Own Key), stored locally, to ensure scalability and avoid server-side quota limits during the trial.

---

## 2. Tech Stack Specification

### Backend
* **Language:** Go (Golang).
* **Router:** Standard `net/http` or a lightweight framework like `Fiber` or `Gin` (optimized for view routing).
* **Video Processing:** `FFmpeg` (executed via `os/exec` in Go).
    * *Function:* Extract image frames every X seconds (sampling) and extract the **FULL** audio track.
* **AI Integration:** Google GenAI SDK for Go.

### Frontend
* **Framework:** Vue.js (via CDN for simplicity in Go Templates).
* **Styling:** Tailwind CSS (via CDN).
* **Storage:** `localStorage` (Browser) to store the User's Gemini API Key.

---

## 3. Application Architecture & Routing

### Route Structure
1.  **GET /** (`DashboardView`)
    * Landing Page.
    * Explains the app's benefits, how it works, and target audience.
    * "Let's Try / Analyze Now" button.
2.  **GET /app** (`AppView`)
    * Main Application Page (Protected).
    * Accessible only if a valid API Key exists (client-side validation).
    * Video Upload Form & Stream Category Dropdown.
3.  **POST /api/analyze**
    * Endpoint to receive Video File + API Key + Category.
    * Returns the Analysis Result in JSON format.

---

## 4. User Flow & Features

### A. Dashboard (`/`)
* **UI/UX:** Modern, clean, and persuasive design suitable for a global audience.
* **Content:**
    * Headline: "Elevate Your Live Stream Sales with AI Audits."
    * **How to get API Key:** A modal or tooltip guiding users to Google AI Studio to generate their key.
* **Logic:**
    * When the user clicks "Start", check `localStorage.getItem('gemini_api_key')`.
    * **If Null:** Trigger a Popup asking for the API Key. Once entered -> Save to LocalStorage -> Redirect to `/app`.
    * **If Exist:** Redirect immediately to `/app`.

### B. Application (`/app`)
* **Input:**
    * File Upload (Video).
    * Category Dropdown: "Jewelry", "Fashion", "General Sales", "Gaming", "Education".
* **Process:**
    * User clicks "Analyze".
    * Frontend retrieves API Key from LocalStorage.
    * Frontend sends `FormData` (Video + Key + Category) to the Backend.
    * **Loading State:** Display a progress bar with engaging text: *"Analyzing Stream Performance with Gemini..."*.

### C. Backend Processing (Go)
1.  Receive video file.
2.  Execute **FFmpeg**:
    * Snapshot/Frame sampling every 5-10 seconds.
    * Extract **FULL Audio** (do not cut) for complete tonal analysis.
3.  Initialize Gemini Client using the **User's API Key**.
4.  Construct the Multimodal Prompt based on the selected **Category**.
5.  Send Request to Gemini 3 Flash.
6.  Parse response to JSON.

### D. Result Display
Display the following data on the Frontend once loading is complete:
1.  **Overall Score:** (e.g., 78/100). Displayed with a gauge chart or color-coded ring.
2.  **Gemini's Reasoning:** A brief explanation of *why* the score is 78/100.
3.  **Timeline Analysis (Flags):**
    * List of specific timestamps with issues.
    * Example: *02:15 - [RED FLAG] Product Blur/Out of Focus.*
4.  **Actionable Advice:** Specific tips for improvement.
    * Example: *"Lighting is too dim on the ring details at 02:15; consider using an additional ring light."*

---

## 5. Prompt Engineering Strategy (Instruction for Gemini)

The System Instruction must be dynamic based on the user's input category.

**Prompt Template (English):**
```text
Role: You are a professional Live Stream Audit Consultant.
Context: The user is conducting a live stream in the [CATEGORY_INPUT] category.
Data: I am providing the FULL AUDIO from the session and several sampled IMAGE FRAMES from the video.

Task:
1. Analyze the AUDIO for tone, enthusiasm, articulation, and sales persuasiveness.
2. Analyze the VISUALS (Frames) for lighting, product focus/clarity, and host engagement (eye contact/gestures).
3. Provide an objective Overall Score (0-100).
4. Identify specific moments (timestamps) where issues occurred (e.g., blur, dead air, low energy).

Constraint: Output MUST be in JSON format with the following structure:
{
  "overall_score": 78,
  "summary_reasoning": "...",
  "strengths": ["..."],
  "weaknesses": ["..."],
  "timeline_flags": [
    {"time": "00:15", "issue": "Audio noise/Dead air", "severity": "medium"}
  ],
  "improvement_tips": "..."
}
```
## 6. Implementation Steps (For Coding Assistant)

Please assist me in implementing this project in the following order:

### 1. Project Setup
- Initialize Go module
- Create folder structure:(/views, /public, /handlers)

### 2. Backend Routes
- Create `main.go`
- Implement basic routing:
- `/` → Dashboard
- `/app` → Main application
- Enable static file serving for `/public`

### 3. Frontend Dashboard
- Create `views/dashboard.html`
- Use:
- Vue.js
- Tailwind CSS
- Implement LocalStorage logic for storing the API Key

### 4. Frontend App
- Create `views/app.html`
- Implement:
- Video upload logic
- Category selection UI

### 5. Backend Logic (FFmpeg)
- Create Go function to:
- Handle video uploads
- Execute FFmpeg commands
- Extract and organize outputs:
  - Images
  - Audio

### 6. Backend Logic (Gemini)
- Integrate Google GenAI (Gemini) SDK
- Send multimodal prompt (text + media)
- Receive and parse AI response

### 7. Final Integration
- Connect backend response to frontend
- Display results dynamically in the UI

