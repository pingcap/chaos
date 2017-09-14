package history

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/anishathalye/porcupine"
)

type noopRequest struct {
	// 0 for read, 1 for write
	Op    int
	Value int
}

type noopResponse struct {
	Value   int
	Ok      bool
	Unknown bool
}

func getNoopModel() porcupine.Model {
	return porcupine.Model{
		Init: func() interface{} {
			return 10
		},
		Step: func(state interface{}, input interface{}, output interface{}) (bool, interface{}) {
			st := state.(int)
			inp := input.(noopRequest)
			out := output.(noopResponse)

			if inp.Op == 0 {
				// read
				ok := out.Unknown || st == out.Value
				return ok, state
			}

			// for write
			return out.Ok || out.Unknown, inp.Value
		},
	}
}

type noopParser struct {
}

func (p noopParser) OnRequest(data json.RawMessage) (interface{}, error) {
	r := noopRequest{}
	err := json.Unmarshal(data, &r)
	return r, err
}

func (p noopParser) OnResponse(data json.RawMessage) (interface{}, error) {
	r := noopResponse{}
	err := json.Unmarshal(data, &r)
	if r.Unknown {
		return nil, err
	}
	return r, err
}

func (p noopParser) OnNoopResponse() interface{} {
	return noopResponse{Unknown: true}
}

func TestHistory(t *testing.T) {
	tmpDir, err := ioutil.TempDir(".", "var")
	if err != nil {
		t.Fatalf("create temp dir failed %v", err)
	}

	defer os.RemoveAll(tmpDir)

	var r *Recorder
	name := path.Join(tmpDir, "history.log")
	r, err = NewRecorder(name)
	if err != nil {
		t.Fatalf("create recorder failed %v", err)
	}

	defer r.Close()

	actions := []struct {
		proc int
		op   interface{}
	}{
		{1, noopRequest{Op: 0}},
		{1, noopResponse{Value: 10}},
		{2, noopRequest{Op: 1, Value: 15}},
		{2, noopResponse{Unknown: true}},
		{3, noopRequest{Op: 0}},
		{3, noopResponse{Value: 15}},
	}

	for _, action := range actions {
		switch v := action.op.(type) {
		case noopRequest:
			if err = r.RecordRequest(action.proc, v); err != nil {
				t.Fatalf("record request failed %v", err)
			}
		case noopResponse:
			if err = r.RecordResponse(action.proc, v); err != nil {
				t.Fatalf("record response failed %v", err)
			}
		}
	}

	r.Close()

	m := getNoopModel()
	var ok bool
	if ok, err = VerifyHistory(name, m, noopParser{}); err != nil {
		t.Fatalf("verify history failed %v", err)
	}

	if !ok {
		t.Fatal("must be linearizable")
	}
}
