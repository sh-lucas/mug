package global

import "sync"

func WaitMany(funcs ...func()) {
	wg := sync.WaitGroup{}
	wg.Add(len(funcs))
	for _, f := range funcs {
		go func() {
			defer wg.Done()
			f()
		}()
	}
	wg.Wait()
}
