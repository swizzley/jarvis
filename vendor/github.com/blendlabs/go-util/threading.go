package util

import "sync"

type AwaitAction func()

func AwaitAll(actions ...AwaitAction) {
	wg := sync.WaitGroup{}
	wg.Add(len(actions))
	for i := 0; i < len(actions); i++ {
		action := actions[i]
		go func() {
			action()
			wg.Done()
		}()
	}

	wg.Wait()
}
