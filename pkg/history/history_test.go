package history

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/siddontang/chaos/pkg/core"
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

type action struct {
	proc int64
	op   interface{}
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

func TestRecordAndReadHistory(t *testing.T) {
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

	actions := []action{
		{1, noopRequest{Op: 0}},
		{1, noopResponse{Value: 10}},
		{2, noopRequest{Op: 1, Value: 15}},
		{2, noopResponse{Value: 15}},
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

	historyOps, err := ReadHistory(name, noopParser{})
	if err != nil {
		t.Fatal(err)
	}

	tbl := [][]core.Operation{historyOps, r.Operations()}

	for _, ops := range tbl {
		if len(ops) != len(actions) {
			t.Fatalf("actions %v mismatchs ops %v", actions, ops)
		}

		for idx, ac := range actions {
			switch v := ac.op.(type) {
			case noopRequest:
				a, ok := ops[idx].Data.(noopRequest)
				if !ok {
					t.Fatalf("unexpected: %#v", ops[idx])
				}
				if a != v {
					t.Fatalf("actions %#v mismatchs ops %#v", a, ops[idx])
				}
			case noopResponse:
				a, ok := ops[idx].Data.(noopResponse)
				if !ok {
					t.Fatalf("unexpected: %#v", ops[idx])
				}
				if a != v {
					t.Fatalf("actions %#v mismatchs ops %#v", a, ops[idx])
				}
			}
		}
	}
}

func TestCompleteOperation(t *testing.T) {
	cases := []struct {
		ops     []core.Operation
		compOps []core.Operation
	}{
		// A complete history of operations.
		{
			ops: []core.Operation{
				{core.InvokeOperation, 1, noopRequest{Op: 0}},
				{core.ReturnOperation, 1, noopResponse{Value: 10}},
				{core.InvokeOperation, 2, noopRequest{Op: 1, Value: 15}},
				{core.ReturnOperation, 2, noopResponse{Value: 15}},
			},
			compOps: []core.Operation{
				{core.InvokeOperation, 1, noopRequest{Op: 0}},
				{core.ReturnOperation, 1, noopResponse{Value: 10}},
				{core.InvokeOperation, 2, noopRequest{Op: 1, Value: 15}},
				{core.ReturnOperation, 2, noopResponse{Value: 15}},
			},
		},
		// A complete but repeated proc operations.
		{
			ops: []core.Operation{
				{core.InvokeOperation, 1, noopRequest{Op: 0}},
				{core.ReturnOperation, 1, noopResponse{Value: 10}},
				{core.InvokeOperation, 2, noopRequest{Op: 1, Value: 15}},
				{core.ReturnOperation, 2, noopResponse{Value: 15}},
				{core.InvokeOperation, 1, noopRequest{Op: 0}},
				{core.ReturnOperation, 1, noopResponse{Value: 15}},
			},
			compOps: []core.Operation{
				{core.InvokeOperation, 1, noopRequest{Op: 0}},
				{core.ReturnOperation, 1, noopResponse{Value: 10}},
				{core.InvokeOperation, 2, noopRequest{Op: 1, Value: 15}},
				{core.ReturnOperation, 2, noopResponse{Value: 15}},
				{core.InvokeOperation, 1, noopRequest{Op: 0}},
				{core.ReturnOperation, 1, noopResponse{Value: 15}},
			},
		},

		// Pending requests.
		{
			ops: []core.Operation{
				{core.InvokeOperation, 1, noopRequest{Op: 0}},
				{core.ReturnOperation, 1, nil},
			},
			compOps: []core.Operation{
				{core.InvokeOperation, 1, noopRequest{Op: 0}},
				{core.ReturnOperation, 1, noopResponse{Unknown: true}},
			},
		},

		// Missing a response
		{
			ops: []core.Operation{
				{core.InvokeOperation, 1, noopRequest{Op: 0}},
			},
			compOps: []core.Operation{
				{core.InvokeOperation, 1, noopRequest{Op: 0}},
				{core.ReturnOperation, 1, noopResponse{Unknown: true}},
			},
		},

		// A complex out of order history.
		{
			ops: []core.Operation{
				{core.InvokeOperation, 1, noopRequest{Op: 0}},
				{core.InvokeOperation, 3, noopRequest{Op: 0}},
				{core.InvokeOperation, 2, noopRequest{Op: 1, Value: 15}},
				{core.ReturnOperation, 2, nil},
				{core.InvokeOperation, 4, noopRequest{Op: 1, Value: 16}},
				{core.ReturnOperation, 3, nil},
			},
			compOps: []core.Operation{
				{core.InvokeOperation, 1, noopRequest{Op: 0}},
				{core.InvokeOperation, 3, noopRequest{Op: 0}},
				{core.InvokeOperation, 2, noopRequest{Op: 1, Value: 15}},
				{core.InvokeOperation, 4, noopRequest{Op: 1, Value: 16}},
				{core.ReturnOperation, 1, noopResponse{Unknown: true}},
				{core.ReturnOperation, 2, noopResponse{Unknown: true}},
				{core.ReturnOperation, 3, noopResponse{Unknown: true}},
				{core.ReturnOperation, 4, noopResponse{Unknown: true}},
			},
		},
	}

	for i, cs := range cases {
		compOps, err := CompleteOperations(cs.ops, noopParser{})
		if err != nil {
			t.Fatalf("err: %s, case %#v", err, cs)
		}
		for idx, op := range compOps {
			if op != cs.compOps[idx] {
				t.Fatalf("op %#v, compOps %#v, case %d", op, cs.compOps[idx], i)
			}
		}
	}
}
