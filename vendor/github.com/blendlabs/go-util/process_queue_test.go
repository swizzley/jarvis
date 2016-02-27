package util

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/blendlabs/go-assert"
)

func TestProcessQueue(t *testing.T) {
	a := assert.New(t)

	StartProcessQueueDispatchers(1)

	wg := sync.WaitGroup{}
	wg.Add(1)

	output := EMPTY
	QueueWorkItem(func(workData interface{}) error {
		output = workData.(string)
		wg.Done()
		return nil
	}, "Hello")

	wg.Wait()
	a.Equal("Hello", output)
}

func TestProcessQueueReqeue(t *testing.T) {
	a := assert.New(t)
	a.StartTimeout(1*time.Second, "This should take < 1 second")
	defer func() {
		a.EndTimeout()
	}()

	StartProcessQueueDispatchers(2)

	maxErrors := MAX_RETRIES - 1
	numErrors := 0

	wg := sync.WaitGroup{}
	wg.Add(1)

	output := EMPTY

	QueueWorkItem(func(workData interface{}) error {
		if numErrors < maxErrors {
			numErrors = numErrors + 1
			return errors.New("Requeue")
		}

		output = workData.(string)
		wg.Done()
		return nil
	}, "Hello")

	wg.Wait()
	a.Equal(maxErrors, numErrors)
	a.Equal("Hello", output)
}

func TestProcessQueueParallel(t *testing.T) {
	a := assert.New(t)
	a.StartTimeout(1*time.Second, "This should take < 1 second")
	defer func() {
		a.EndTimeout()
	}()

	StartProcessQueueDispatchers(2)

	wg := sync.WaitGroup{}
	wg.Add(4)

	QueueWorkItem(func(workData interface{}) error {
		wg.Done()
		return nil
	}, nil)

	QueueWorkItem(func(workData interface{}) error {
		wg.Done()
		return nil
	}, nil)

	QueueWorkItem(func(workData interface{}) error {
		wg.Done()
		return nil
	}, nil)

	QueueWorkItem(func(workData interface{}) error {
		wg.Done()
		return nil
	}, nil)

	wg.Wait()
}
