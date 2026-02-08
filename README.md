# StreamCoach AI 🎥🚀

**Elevate Your Live Stream Sales with Multimodal AI Audits.**

StreamCoach AI is a powerful web application designed to automatically analyze the performance quality of live streams. Powered by **Google Gemini 3 Flash Preview**, it acts as a virtual consultant, auditing both the visual presentation and audio delivery of your stream to provide actionable, timestamped feedback that helps boost engagement and sales conversion.

![StreamCoach infographics](infographics.jpg)

---

## ✨ Key Features

*   **Multimodal Analysis**: "Sees" your product presentation (lighting, clarity, gestures) and "hears" your sales pitch (tone, enthusiasm, pacing) simultaneously.
*   **Internationalization (i18n)**: Fully localized UI and AI responses in **English, Indonesian, Spanish, Chinese, and Japanese**.
*   **Smart Async Queue**: Protects the server from overload. Implements a robust **Job System** with **Polling** to provide real-time status updates (Queued, Processing, Completed) to the user.
*   **Scalable Concurrency**: Supports **Local (In-Memory)** and **Production (Redis)** queuing modes to strictly limit concurrent heavy tasks.
*   **Large File Support**: Optimized handling for video uploads up to **1GB**.
*   **Resource Efficiency**: Automatically cancels heavy processing (FFmpeg & AI) if the user disconnects, saving server resources and API quota.
*   **Timeline Analysis**: Identifies specific moments (timestamped flags) where issues occurred.
*   **Privacy-First & Secure**:
    *   **BYOK (Bring Your Own Key)**: Your Google Gemini API Key is stored safely in your browser's **LocalStorage**, never on our servers.
    *   **Auto-Cleanup**: Video files and extracted assets are deleted immediately after processing.
*   **PDF Reports**: Export your audit results into a professional PDF format.

---

## 🛠️ Tech Stack

*   **Backend**: Go (Golang)
*   **AI Engine**: Google GenAI SDK (Gemini 3 Flash Preview)
*   **Video Processing**: FFmpeg (Frame extraction & Audio separation)
*   **Concurrency**: 
    *   **Queue**: Native Channels (Local) / Redis (Production)
    *   **Job Management**: Async polling architecture (UUID based)
*   **Frontend**: Vue.js 3 (Composition API)
*   **Styling**: Tailwind CSS
*   **Alerts**: SweetAlert2
*   **PDF Generation**: jsPDF & AutoTable

## 🏗️ Architecture Overview
<img width="2816" height="1536" alt="Gemini_Generated_Image_464juz464juz464j" src="https://github.com/user-attachments/assets/a075972a-5693-4fb6-94c5-b2027747bd17" />
The system follows a modern, decoupled architecture designed for speed and privacy. It leverages **Golang** for high-performance backend processing and **Google Gemini 3 Flash Preview** for multimodal AI analysis.

#### 1. Frontend Layer (User Interface)
* **Tech Stack:** **Vue.js** (Framework) & **Tailwind CSS** (Styling).
* **Function:**
    * Provides a responsive dashboard for users to upload live stream recordings.
    * **Privacy-First Security:** The user's Gemini API Key is stored securely in the browser's **Local Storage**. It is never saved to our database, ensuring a "Bring Your Own Key" (BYOK) architecture.
    * Visualizes the JSON analysis data into interactive charts, timelines, and scorecards.

#### 2. Backend Orchestration
* **Tech Stack:** **Golang (Go)**.
* **Function:**
    * Acts as the central orchestrator, handling API requests from the frontend.
    * Manages **Temporary Storage** to briefly hold uploaded video files during the processing stage.
    * Utilizes Go’s concurrency model (Goroutines) to handle multiple analysis requests efficiently without blocking the server.

#### 3. Media Processing Engine
* **Tool:** **FFmpeg**.
* **Workflow:**
    * Once the video reaches the backend, Golang executes FFmpeg commands to split the media into two modalities:
        1.  **Visual Sampling:** Extracts image frames at specific intervals (e.g., every 5-10 seconds) to reduce payload size while retaining visual context (lighting, gestures, product focus).
        2.  **Audio Extraction:** Separates the **full audio track** to ensure the AI can analyze vocal intonation, pitch, and energy continuity without interruption.

#### 4. AI Analysis Layer (The Brain)
* **Model:** **Google Gemini 3 Flash Preview**.
* **Workflow:**
    * The Backend constructs a **Multimodal Prompt** containing the sampled image frames, the full audio file, and the user-selected context (e.g., "Jewelry Sales" or "Fashion").
    * This payload is sent to the Gemini API.
    * Gemini processes the inputs simultaneously ("watching" the frames and "listening" to the audio) to generate a holistic audit.

#### 5. Data Output & Visualization
* **Result:** Gemini returns a structured **JSON** response containing:
    * Overall Performance Score (0-100).
    * Timestamped flags for issues (e.g., "Blurry product at 02:15").
    * Actionable coaching tips.
* The Golang backend forwards this JSON to the Frontend, which renders it into the user-friendly "Stream Health Report."

---

## ⚙️ Prerequisites

Before running the application, ensure you have the following installed:

1.  **Go** (version 1.21 or higher) - [Download Go](https://go.dev/dl/)
2.  **FFmpeg** - [Download FFmpeg](https://ffmpeg.org/download.html)
    *   *Crucial*: Ensure `ffmpeg` is added to your system's PATH variable.
3.  **Google Gemini API Key** - [Get a free key](https://aistudio.google.com/app/apikey)
4.  **(Optional) Redis**: Required only if running in `production` mode for distributed queuing.

---

## 🚀 Installation & Setup

1.  **Clone the Repository**
    ```bash
    git clone https://github.com/yourusername/streamcoach-ai.git
    cd streamcoach-ai
    ```

2.  **Install Dependencies**
    ```bash
    go mod tidy
    ```

3.  **Configure Environment**
    Copy the example environment file:
    ```bash
    cp .env.example .env
    ```
    *   By default, `APP_ENV=local` uses in-memory queuing (no Redis required).
    *   To use Redis, set `APP_ENV=production` and configure `REDIS_ADDR`.

4.  **Verify FFmpeg**
    ```bash
    ffmpeg -version
    ```

5.  **Build and Run**
    ```bash
    go build -o streamcoach.exe
    ./streamcoach.exe
    ```

6.  **Access the App**
    Open `http://localhost:8080`

---

## 🔧 Configuration (.env)

| Variable | Default | Description |
| :--- | :--- | :--- |
| `APP_ENV` | `local` | Set to `production` to enable Redis queue. |
| `MAX_CONCURRENT_TASKS` | `2` | Maximum number of simultaneous analysis tasks. |
| `REDIS_ADDR` | `localhost:6379` | Address of your Redis server (Prod only). |
| `REDIS_PASSWORD` | - | Redis password (if any). |
| `REDIS_DB` | `0` | Redis Database index. |

---

## 🔒 Security & Privacy

*   **No Persistent Storage**: We do not store your API keys or video files on the backend.
*   **Ephemeral Processing**: Videos are uploaded to a temporary folder, processed, and **immediately deleted**.
*   **Client-Side Keys**: API keys remain in your browser's LocalStorage.

---

## 🤝 Contribution

This project was built for the **Gemini Hackathon**. Feedback and contributions are welcome!

1.  Fork the repository.
2.  Create your feature branch (`git checkout -b feature/AmazingFeature`).
3.  Commit your changes (`git commit -m 'Add some AmazingFeature'`).
4.  Push to the branch (`git push origin feature/AmazingFeature`).
5.  Open a Pull Request.

---

## 📄 License

Distributed under the MIT License. See `LICENSE` for more information.
