package model

import (
	"github.com/anishathalye/porcupine"
)

// Op is an operation.
type Op bool

const (
	// RegisterRead a register
	RegisterRead Op = false
	// RegisterWrite a register
	RegisterWrite Op = true
)

// RegisterRequest is the request that is issued to a register.
type RegisterRequest struct {
	Op    Op
	Value int
}

// RegisterResponse is the response returned by a register.
type RegisterResponse struct {
	Err   error
	Value int
}

// RegisterModel returns a read/write register model
func RegisterModel() porcupine.Model {
	return porcupine.Model{
		Init: func() interface{} {
			state := 0
			return state
		},
		Step: func(state interface{}, input interface{}, output interface{}) (bool, interface{}) {
			st := state.(int)
			inp := input.(RegisterRequest)
			out := output.(RegisterResponse)

			// read
			if inp.Op == RegisterRead {
				ok := out.Value == st || out.Err != nil
				return ok, st
			}

			// write
			return true, inp.Value
		},

		Equal: func(state1, state2 interface{}) bool {
			st1 := state1.(int)
			st2 := state2.(int)
			return st1 == st2
		},
	}
}
