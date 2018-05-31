package util

import (
	"testing"
	"time"
)

func TestBlockRunner(t *testing.T) {
	r := &BlockRunner{}

	r.Init(2)

	ch := make(chan int, 2)
	for i := 0; i < 2; i++ {
		go func(i int) {
			// Can only initialize once
			r.Init(3)
			r.Run(func() {
				t.Logf("block run %d", i+1)
				time.Sleep(time.Second)
			})
			ch <- i
		}(i)
	}

	select {
	case <-ch:
		t.Fatal("can't get data at this time")
	case <-time.After(100 * time.Millisecond):
	}

	for i := 0; i < 2; i++ {
		select {
		case <-ch:
		case <-time.After(5 * time.Second):
			t.Fatal("can't wait ")
		}
	}
}
