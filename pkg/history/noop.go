package history

import (
	"encoding/json"
)

// NoopRequest is a noop request.
type NoopRequest struct {
	// 0 for read, 1 for write
	Op    int
	Value int
}

// NoopResponse is a noop response.
type NoopResponse struct {
	Value   int
	Ok      bool
	Unknown bool
}

type action struct {
	proc int64
	op   interface{}
}

// NoopParser is a noop parser.
type NoopParser struct {
	State int
}

// OnRequest impls RecordParser.
func (p NoopParser) OnRequest(data json.RawMessage) (interface{}, error) {
	r := NoopRequest{}
	err := json.Unmarshal(data, &r)
	return r, err
}

// OnResponse impls RecordParser.
func (p NoopParser) OnResponse(data json.RawMessage) (interface{}, error) {
	r := NoopResponse{}
	err := json.Unmarshal(data, &r)
	if r.Unknown {
		return nil, err
	}
	return r, err
}

// OnNoopResponse impls RecordParser.
func (p NoopParser) OnNoopResponse() interface{} {
	return NoopResponse{Unknown: true}
}

// OnState impls RecordParser.
func (p NoopParser) OnState(state json.RawMessage) (interface{}, error) {
	return p.State, nil
}
