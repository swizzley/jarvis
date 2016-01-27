package collections

import (
	"time"
)

func NewRateLimiter(numberOfActions int, quantum time.Duration) *RateLimiter {
	rl := RateLimiter{}
	rl.NumberOfActions = numberOfActions
	rl.Quantum = quantum
	rl.Limits = map[string]*Queue{}
	return &rl
}

type RateLimiter struct {
	NumberOfActions int
	Quantum         time.Duration
	Limits          map[string]*Queue
}

func (rl *RateLimiter) Check(id string) bool {
	queue, has_queue := rl.Limits[id]
	if !has_queue {
		queue = &Queue{}
		rl.Limits[id] = queue
	}

	current_time := time.Now().UTC()
	queue.Push(current_time)
	if queue.Length() < rl.NumberOfActions {
		return false
	}

	oldest, _ := queue.Dequeue().(time.Time)
	return current_time.Sub(oldest) < rl.Quantum
}
