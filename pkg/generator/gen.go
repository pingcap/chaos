package generator

import (
	"math/rand"
	"time"

	"github.com/pingcap/chaos/pkg/control"
)

type PairedGenerator struct {
	key interface{}
	gen control.Generator
}

// Reserve takes a series of count, generator pairs, and a final default
// generator.     
//     (reserve 5 write 10 cas read)
// The first 5 threads will call the `write` generator, the next 10 will emit
// CAS operations, and the remaining threads will perform reads. This is
// particularly useful when you want to ensure that two classes of operations
// have a chance to proceed concurrently--for instance, if writes begin
// blocking, you might like reads to proceed concurrently without every thread
// getting tied up in a write.
func Reserve(final control.Generator, gens ...PairedGenerator) control.Generator {
	return func(cfg *control.Config, proc int64) interface{} {
		thread := (proc - 1) % len(cfg.Nodes)
		cnt := 0
		for _, pair in gens {
			n := pair.key.(int)
			if thread >= cnt && thread < cnt + n {
				return pair.gen(cfg, proc)
			}
			cnt += n
		}
		return final(cfg, proc)
	}
}

// Stagger introduces uniform random timing noise with a mean delay of
// dt duration for every operation. Delays range from 0 to 2 * dt."
func Stagger(dt time.Duration, gen control.Generator) control.Generator {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return func(cfg *control.Config, proc int64) interface{} {
		time.Sleep(time.Duration(r.Int63n(2 * int64(dt))))
		return gen(cfg, proc)
	}
}
