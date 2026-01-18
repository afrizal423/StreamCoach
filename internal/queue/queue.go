package queue

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type TaskManager interface {
	Acquire(ctx context.Context) (func(), error)
}

// LocalTaskManager uses a Go channel as a semaphore
type LocalTaskManager struct {
	sem chan struct{}
}

func NewLocalTaskManager(maxConcurrent int) *LocalTaskManager {
	return &LocalTaskManager{
		sem: make(chan struct{}, maxConcurrent),
	}
}

func (l *LocalTaskManager) Acquire(ctx context.Context) (func(), error) {
	log.Println("Waiting for local execution slot...")
	select {
	case l.sem <- struct{}{}:
		log.Println("Acquired local execution slot.")
		return func() {
			<-l.sem
			log.Println("Released local execution slot.")
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// RedisTaskManager uses Redis to manage concurrent slots via a token bucket pattern or simple key counting
// For simplicity and blocking behavior, we will use a blocking list (BLPOP) as a semaphore.
// We initialize a list with N tokens. Taking a token = BLPOP. Returning = RPUSH.
type RedisTaskManager struct {
	client    *redis.Client
	queueKey  string
	maxTasks  int
}

func NewRedisTaskManager(addr, password string, db int, maxTasks int) (*RedisTaskManager, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %v", err)
	}

	tm := &RedisTaskManager{
		client:   rdb,
		queueKey: "streamcoach:task_slots",
		maxTasks: maxTasks,
	}

	// Initialize the semaphore tokens if they don't exist
	// We need to ensure the list has exactly maxTasks tokens.
	// This is a bit tricky in distributed envs on restart, but we'll do a basic check.
	// A safer way for persistent queue is just to ensure length.
	ctx := context.Background()
	len, err := rdb.LLen(ctx, tm.queueKey).Result()
	if err != nil {
		return nil, err
	}

	// If list is empty (first run or flushed), fill it. 
	// NOTE: This logic assumes a single initialization or non-persistent redis. 
	// For production robustness, one might use a dedicated "init" script or check carefully.
	if len == 0 {
		log.Printf("Initializing Redis semaphore with %d slots", maxTasks)
		for i := 0; i < maxTasks; i++ {
			rdb.RPush(ctx, tm.queueKey, "token")
		}
	} else if int(len) < maxTasks {
		// replenish missing
		diff := maxTasks - int(len)
		log.Printf("Replenishing Redis semaphore with %d slots", diff)
		for i := 0; i < diff; i++ {
			rdb.RPush(ctx, tm.queueKey, "token")
		}
	}

	return tm, nil
}

func (r *RedisTaskManager) Acquire(ctx context.Context) (func(), error) {
	log.Println("Waiting for Redis execution slot...")
	// BLPOP blocks until a token is available or context timeout (if we implement logic for that, 
	// but redis-go BLPOP takes a timeout. 0 = infinite).
	// We map context cancellation to client closing or manual handling if needed.
	
	// Use a long timeout loop to respect context cancellation? 
	// Or just blocking wait. Let's do blocking wait with 0 (infinite) for simplicity as per requirement.
	
	// However, we should respect the HTTP request context.
	// Since BLPOP blocks the connection, we can't easily interrupt it with just ctx.
	// We'll use a loop with short timeouts to check context.
	
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// Try to pop for 1 second
			res, err := r.client.BLPop(ctx, 1*time.Second, r.queueKey).Result()
			if err != nil {
				if err == redis.Nil {
					// Timeout, loop again
					continue
				}
				return nil, err
			}
			// Success
			_ = res // token value
			log.Println("Acquired Redis execution slot.")
			
			return func() {
				// Return token
				r.client.RPush(context.Background(), r.queueKey, "token")
				log.Println("Released Redis execution slot.")
			}, nil
		}
	}
}

// Global instance
var Manager TaskManager

// InitQueue initializes the global manager based on ENV
func InitQueue() error {
	env := os.Getenv("APP_ENV")
	maxTasksStr := os.Getenv("MAX_CONCURRENT_TASKS")
	maxTasks, err := strconv.Atoi(maxTasksStr)
	if err != nil || maxTasks < 1 {
		maxTasks = 2 // Default
	}

	if env == "production" {
		log.Println("Initializing Queue in PRODUCTION mode (Redis)")
		addr := os.Getenv("REDIS_ADDR")
		pass := os.Getenv("REDIS_PASSWORD")
		dbStr := os.Getenv("REDIS_DB")
		db, _ := strconv.Atoi(dbStr)
		
		rm, err := NewRedisTaskManager(addr, pass, db, maxTasks)
		if err != nil {
			return err
		}
		Manager = rm
	} else {
		log.Println("Initializing Queue in LOCAL mode (In-Memory)")
		Manager = NewLocalTaskManager(maxTasks)
	}
	return nil
}
