package util

const (
	MAX_RETRIES = 10
)

// static vars
var _workerQueue chan processQueueWorker
var _processQueue = make(chan processQueueEntry, 1024) //buffered work channel. 1024 better be enough

type ProcessQueueAction func(value interface{}) error

func QueueWorkItem(action ProcessQueueAction, value interface{}) {
	_processQueue <- processQueueEntry{Action: action, Value: value}
}

func StartProcessQueueDispatchers(numWorkers int) {
	_workerQueue = make(chan processQueueWorker, numWorkers)

	for id := 0; id < numWorkers; id++ {
		worker := newProcessQueueWorker(id, _workerQueue)
		worker.Start()
	}

	go func() {
		for {
			select {
			case work := <-_processQueue:
				go func() {
					worker := <-_workerQueue
					worker.Work <- work
				}()
			}
		}
	}()
}

type processQueueEntry struct {
	Action ProcessQueueAction
	Value  interface{}
	Tries  int
}

type processQueueWorker struct {
	Id          int
	Work        chan processQueueEntry
	WorkerQueue chan processQueueWorker
	QuitChan    chan bool
}

func (pqw processQueueWorker) Start() {
	go func() {
		for {
			pqw.WorkerQueue <- pqw
			select {
			case work := <-pqw.Work:
				workErr := work.Action(work.Value)
				if workErr != nil {
					work.Tries = work.Tries + 1
					if work.Tries < MAX_RETRIES {
						_processQueue <- work
					}
				}
			case <-pqw.QuitChan:
				return
			}
		}
	}()
}

func (pqw processQueueWorker) Stop() {
	go func() {
		pqw.QuitChan <- true
	}()
}

func newProcessQueueWorker(id int, workerQueue chan processQueueWorker) processQueueWorker {
	worker := processQueueWorker{
		Id:          id,
		Work:        make(chan processQueueEntry),
		WorkerQueue: workerQueue,
		QuitChan:    make(chan bool),
	}
	return worker
}
