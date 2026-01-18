# Project Context: StreamCoach AI (Live Stream Performance Analyzer)

## 1. Project Overview
"StreamCoach AI" is a web application designed to analyze the performance quality of live streams automatically using **Google Gemini 3 Flash Preview (Multimodal)**. 

### Core Value Proposition
*   **Speed & Cost:** Utilizes Gemini 3 Flash Preview for low-latency analysis.
*   **Multimodal:** Assesses performance based on Visuals (Video Frames) and Audio (Tone/Vocal).
*   **Global Reach:** Fully localized (i18n) for EN, ID, ES, ZH, and JA.
*   **Scalable Architecture:** Implements a Smart **Async Job Queue** (Local/Redis) to manage server load.
*   **User-Owned API Key:** BYOK model ensures privacy and scalability.

---

## 2. Tech Stack Specification

### Backend
*   **Language:** Go (Golang).
*   **Concurrency:**
    *   **Queue:** `internal/queue` package supporting In-Memory (Channel) or Redis (List/BLPOP).
    *   **Job Management:** `internal/job` for thread-safe state tracking.
    *   **Resource Management:** Context-aware cancellation (kills FFmpeg/AI request on disconnect).
*   **Video Processing:** `FFmpeg` (executed via `os/exec` with context).
*   **AI Integration:** Google GenAI SDK for Go.

### Frontend
*   **Framework:** Vue.js (via CDN).
*   **Styling:** Tailwind CSS (via CDN).
*   **Notifications:** SweetAlert2.
*   **PDF:** jsPDF & AutoTable.
*   **Storage:** `localStorage` for API Key and Language preference.

---

## 3. Application Architecture & Routing

### Route Structure
1.  **GET /** (`DashboardHandler`) - Landing Page with i18n support.
2.  **GET /app** (`AppHandler`) - Main Analysis Interface.
3.  **POST /api/upload** (`UploadHandler`) - Initiates async job, returns `jobId`.
4.  **GET /api/status** (`StatusHandler`) - Polls job status (`?id=uuid`).

---

## 4. User Flow & Features

### A. Dashboard (`/`)
*   **Language Selector:** Users can toggle between 5 supported languages.
*   **API Key Management:** Secure client-side storage.

### B. Application (`/app`)
*   **Inputs:** File Upload (Max 1GB), Category, Language.
*   **Validation:** Client-side check for >1GB files using SweetAlert2.
*   **Process:**
    1.  User clicks Analyze -> File uploaded to `/api/upload`.
    2.  Server returns `jobId`.
    3.  Frontend polls `/api/status?id=...` every 2s.
    4.  **Status "queued":** UI shows "Waiting in queue..."
    5.  **Status "processing":** UI shows "Gemini is analyzing..."
    6.  **Status "completed":** UI shows results.

### C. Backend Processing (Go)
1.  **UploadHandler:** Saves file, creates Job Record, spawns Goroutine.
2.  **Async Worker:**
    *   **Queue Acquisition:** Wait for slot (Redis/Local).
    *   **Update Job Status:** Set to `processing`.
    *   **FFmpeg:** Extract audio/frames (cancellable).
    *   **Gemini AI:** Analyze using Multimodal Prompt.
    *   **Update Job Status:** Set to `completed` with JSON result.
3.  **Cleanup:** Immediate deletion of all temp files.

---

## 5. Prompt Engineering Strategy

The System Instruction is dynamic based on:
1.  **Category:** (Jewelry, Fashion, etc.)
2.  **Language:** (English, Indonesian, etc.)

**Prompt Template:**
```text
Role: You are a professional Live Stream Audit Consultant.
Context: The user is conducting a live stream in the [CATEGORY_INPUT] category.
Data: FULL AUDIO + Sampled IMAGE FRAMES.
Output Language: [LANGUAGE_INPUT]

Task:
1. Analyze AUDIO (tone, enthusiasm).
2. Analyze VISUALS (lighting, focus).
3. Provide Score (0-100).
4. Identify Timeline Flags.

Constraint: Output MUST be in JSON format. All text values MUST be in [LANGUAGE_INPUT].
```

## 6. Implementation Status (Completed)

### ✅ Phase 1: Core Setup
- Go Module & Directory Structure.
- Basic Routes & Static Serving.

### ✅ Phase 2: Frontend
- Vue.js + Tailwind UI (Polished).
- API Key Logic (LocalStorage).
- **Internationalization (i18n)** implemented.
- **PDF Export** implemented.

### ✅ Phase 3: Backend Logic
- **FFmpeg Integration** (Audio/Frame extraction).
- **Gemini 3 Flash Preview** Integration.
- **JSON Parsing Fix** (Handle Array/Object).

### ✅ Phase 4: Reliability & Scale
- **Async Job System:** Polling architecture for better UX.
- **Queue System:** Local & Redis support.
- **Large File Support:** 1GB limit enforced.
- **Resource Conservation:** Context cancellation on disconnect.
- **Logging:** Detailed lifecycle logging.
