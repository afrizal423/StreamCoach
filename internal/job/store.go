package job

import (
	"sync"
	"time"
)

type Status string

const (
	StatusQueued     Status = "queued"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
)

type Job struct {
	ID        string      `json:"id"`
	Status    Status      `json:"status"`
	Result    interface{} `json:"result,omitempty"`
	Error     string      `json:"error,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
}

var (
	store = make(map[string]*Job)
	mu    sync.RWMutex
)

func Create(id string) *Job {
	mu.Lock()
	defer mu.Unlock()
	j := &Job{
		ID:        id,
		Status:    StatusQueued,
		CreatedAt: time.Now(),
	}
	store[id] = j
	return j
}

func UpdateStatus(id string, status Status) {
	mu.Lock()
	defer mu.Unlock()
	if j, ok := store[id]; ok {
		j.Status = status
	}
}

func Complete(id string, result interface{}) {
	mu.Lock()
	defer mu.Unlock()
	if j, ok := store[id]; ok {
		j.Status = StatusCompleted
		j.Result = result
	}
}

func Fail(id string, err string) {
	mu.Lock()
	defer mu.Unlock()
	if j, ok := store[id]; ok {
		j.Status = StatusFailed
		j.Error = err
	}
}

func Get(id string) (*Job, bool) {
	mu.RLock()
	defer mu.RUnlock()
	j, ok := store[id]
	return j, ok
}
