package model

import (
	"encoding/json"

	"github.com/siddontang/chaos/pkg/core"
	"github.com/siddontang/chaos/pkg/history"
)

// CasOp is an operation.
type CasOp int

// cas operation
const (
	CasRegisterRead CasOp = iota
	CasRegisterWrite
	CasRegisterCAS
)

// CasRegisterRequest is the request that is issued to a cas register.
type CasRegisterRequest struct {
	Op   CasOp
	Arg1 int // used for write, or for CAS from argument
	Arg2 int // used for CAS to argument
}

// CasRegisterResponse is the response returned by a cas register.
type CasRegisterResponse struct {
	Ok      bool // used for CAS
	Exists  bool // used for read
	Value   int  // used for read
	Unknown bool // used when operation times out
}

var _ core.UnknownResponse = (*CasRegisterResponse)(nil)

// IsUnknown implements UnknownResponse interface
func (r CasRegisterResponse) IsUnknown() bool {
	return r.Unknown
}

type casRegister struct {
	perparedState *int
}

func (c *casRegister) Prepare(state interface{}) {
	s := state.(int)
	c.perparedState = &s
}

func (c *casRegister) Init() interface{} {
	if c.perparedState != nil {
		return *c.perparedState
	}
	return -1
}

func (*casRegister) Step(state interface{}, input interface{}, output interface{}) (bool, interface{}) {
	st := state.(int)
	inp := input.(CasRegisterRequest)
	out := output.(CasRegisterResponse)
	if inp.Op == CasRegisterRead {
		// read
		ok := (out.Exists == false && st == -1) || (out.Exists == true && st == out.Value) || out.Unknown
		return ok, state
	} else if inp.Op == CasRegisterWrite {
		// write
		return true, inp.Arg1
	}

	// cas
	ok := (inp.Arg1 == st && out.Ok) || (inp.Arg1 != st && !out.Ok) || out.Unknown
	result := st
	if inp.Arg1 == st {
		result = inp.Arg2
	}
	return ok, result
}

func (*casRegister) Equal(state1, state2 interface{}) bool {
	st1 := state1.(int)
	st2 := state2.(int)
	return st1 == st2
}

func (*casRegister) Name() string {
	return "cas_register"
}

// CasRegisterModel returns a cas register model
func CasRegisterModel() core.Model {
	return &casRegister{}
}

type casRegisterParser struct {
}

func (p casRegisterParser) OnRequest(data json.RawMessage) (interface{}, error) {
	r := CasRegisterRequest{}
	err := json.Unmarshal(data, &r)
	return r, err
}

func (p casRegisterParser) OnResponse(data json.RawMessage) (interface{}, error) {
	r := CasRegisterResponse{}
	err := json.Unmarshal(data, &r)
	if r.Unknown {
		return nil, err
	}
	return r, nil
}

func (p casRegisterParser) OnNoopResponse() interface{} {
	return CasRegisterResponse{Unknown: true}
}

func (p casRegisterParser) OnSummarize(data json.RawMessage) (interface{}, error) {
	var state int
	err := json.Unmarshal(data, &state)
	if err != nil {
		return nil, err
	}
	return state, nil
}

// CasRegisterParser parses CasRegister history.
func CasRegisterParser() history.RecordParser {
	return casRegisterParser{}
}
