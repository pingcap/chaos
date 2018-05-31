package util

import "sync"

// BlockRunner provides a simple way to run tasks,
// block until all the tasks are finished.
type BlockRunner struct {
	once sync.Once
	wg   sync.WaitGroup
}

// Init initializes how many tasks we want to run synchronously.
func (r *BlockRunner) Init(n int) {
	r.once.Do(func() {
		r.wg.Add(n)
	})
}

// Run runs the task in different goroutines and
// block until all the tasks are finished.
func (r *BlockRunner) Run(f func()) {
	f()
	r.wg.Done()
	r.wg.Wait()
}
