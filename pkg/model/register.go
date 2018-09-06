package model

import (
	"encoding/json"

	"github.com/anishathalye/porcupine"
	"github.com/siddontang/chaos/pkg/history"
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
	Unknown bool
	Value   int
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
				ok := out.Value == st || out.Unknown
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

type registerParser struct {
}

func (p registerParser) OnRequest(data json.RawMessage) (interface{}, error) {
	r := RegisterRequest{}
	err := json.Unmarshal(data, &r)
	return r, err
}

func (p registerParser) OnResponse(data json.RawMessage) (interface{}, error) {
	r := RegisterResponse{}
	err := json.Unmarshal(data, &r)
	if r.Unknown {
		return nil, err
	}
	return r, nil
}

func (p registerParser) OnNoopResponse() interface{} {
	return RegisterResponse{Unknown: true}
}

// RegisterVerifier can verify a register history.
type RegisterVerifier struct {
}

// Verify verifies a register history.
func (RegisterVerifier) Verify(historyFile string) (bool, error) {
	return history.IsLinearizable(historyFile, RegisterModel(), registerParser{})
}

// Name returns the name of the verifier.
func (RegisterVerifier) Name() string {
	return "register_verifier"
}
