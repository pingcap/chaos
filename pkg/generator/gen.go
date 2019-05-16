package generator

import (
	"math/rand"
	"time"
)

// Generator generates a series of operations
type Generator = func() interface{}

// Stagger introduces uniform random timing noise with a mean delay of
// dt duration for every operation. Delays range from 0 to 2 * dt."
func Stagger(dt time.Duration, gen Generator) Generator {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return func() interface{} {
		time.Sleep(time.Duration(r.Int63n(2 * int64(dt))))
		return gen()
	}
}
