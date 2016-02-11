package collections

import "sync"

func NewConcurrentQueue(maxSize int) *ConcurrentQueue {
	return &ConcurrentQueue{MaxSize: maxSize, storage: make(chan interface{}, maxSize), latch: sync.Mutex{}}
}

type ConcurrentQueue struct {
	MaxSize int
	storage chan interface{}
	latch   sync.Mutex
}

func (cq *ConcurrentQueue) Length() int {
	return len(cq.storage)
}

func (cq *ConcurrentQueue) Push(item interface{}) {
	cq.storage <- item
}

func (cq *ConcurrentQueue) Dequeue() interface{} {
	if len(cq.storage) != 0 {
		return <-cq.storage
	}
	return nil
}

func (cq *ConcurrentQueue) ToArray() []interface{} {
	cq.latch.Lock()
	defer cq.latch.Unlock()

	values := []interface{}{}
	for len(cq.storage) != 0 {
		v := <-cq.storage
		values = append(values, v)
	}
	for _, v := range values {
		cq.storage <- v
	}
	return values
}
