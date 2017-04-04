package logger

import (
	"sync"
	"time"
)

// DebugPrintAverageLatency prints the average queue latency for an agent.
func DebugPrintAverageLatency(agent *Agent) {
	var (
		debugLatenciesLock sync.Mutex
		debugLatencies     = []time.Duration{}
	)

	agent.AddDebugListener(func(_ Logger, ts TimeSource, _ EventFlag, _ ...interface{}) {
		debugLatenciesLock.Lock()
		debugLatencies = append(debugLatencies, time.Now().UTC().Sub(ts.UTCNow()))
		debugLatenciesLock.Unlock()
	})

	var averageLatency time.Duration
	poll := time.NewTicker(5 * time.Second)
	go func() {
		for {
			select {
			case <-poll.C:
				{
					debugLatenciesLock.Lock()
					averageLatency = MeanOfDuration(debugLatencies)
					debugLatencies = []time.Duration{}
					debugLatenciesLock.Unlock()
					if averageLatency != time.Duration(0) {
						agent.Debugf("average event queue latency (%v)", averageLatency)
					}
				}
			}
		}
	}()
}
