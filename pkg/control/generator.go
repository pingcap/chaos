package control

// Generator generates a series of operations
type Generator = func(*Config, int64) interface{}
