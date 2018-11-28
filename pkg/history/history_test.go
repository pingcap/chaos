package history

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/siddontang/chaos/pkg/core"
)

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
		{1, NoopRequest{Op: 0}},
		{1, NoopResponse{Value: 10}},
		{2, NoopRequest{Op: 1, Value: 15}},
		{2, NoopResponse{Value: 15}},
		{3, NoopRequest{Op: 0}},
		{3, NoopResponse{Value: 15}},
	}
	summarizeState := 7

	for _, action := range actions {
		switch v := action.op.(type) {
		case NoopRequest:
			if err = r.RecordRequest(action.proc, v); err != nil {
				t.Fatalf("record request failed %v", err)
			}
		case NoopResponse:
			if err = r.RecordResponse(action.proc, v); err != nil {
				t.Fatalf("record response failed %v", err)
			}
		}
	}
	if err = r.SummarizeState(summarizeState); err != nil {
		t.Fatalf("record summarize failed %v", err)
	}

	ops, state, err := ReadHistory(name, NoopParser{SummarizeState: summarizeState})
	if err != nil {
		t.Fatal(err)
	}

	if state.(int) != summarizeState {
		t.Fatalf("expect state to be %v, got %v", summarizeState, state)
	}

	if len(ops) != len(actions) {
		t.Fatalf("actions %v mismatchs ops %v", actions, ops)
	}

	for idx, ac := range actions {
		switch v := ac.op.(type) {
		case NoopRequest:
			a, ok := ops[idx].Data.(NoopRequest)
			if !ok {
				t.Fatalf("unexpected: %#v", ops[idx])
			}
			if a != v {
				t.Fatalf("actions %#v mismatchs ops %#v", a, ops[idx])
			}
		case NoopResponse:
			a, ok := ops[idx].Data.(NoopResponse)
			if !ok {
				t.Fatalf("unexpected: %#v", ops[idx])
			}
			if a != v {
				t.Fatalf("actions %#v mismatchs ops %#v", a, ops[idx])
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
				{core.InvokeOperation, 1, NoopRequest{Op: 0}},
				{core.ReturnOperation, 1, NoopResponse{Value: 10}},
				{core.InvokeOperation, 2, NoopRequest{Op: 1, Value: 15}},
				{core.ReturnOperation, 2, NoopResponse{Value: 15}},
			},
			compOps: []core.Operation{
				{core.InvokeOperation, 1, NoopRequest{Op: 0}},
				{core.ReturnOperation, 1, NoopResponse{Value: 10}},
				{core.InvokeOperation, 2, NoopRequest{Op: 1, Value: 15}},
				{core.ReturnOperation, 2, NoopResponse{Value: 15}},
			},
		},
		// A complete but repeated proc operations.
		{
			ops: []core.Operation{
				{core.InvokeOperation, 1, NoopRequest{Op: 0}},
				{core.ReturnOperation, 1, NoopResponse{Value: 10}},
				{core.InvokeOperation, 2, NoopRequest{Op: 1, Value: 15}},
				{core.ReturnOperation, 2, NoopResponse{Value: 15}},
				{core.InvokeOperation, 1, NoopRequest{Op: 0}},
				{core.ReturnOperation, 1, NoopResponse{Value: 15}},
			},
			compOps: []core.Operation{
				{core.InvokeOperation, 1, NoopRequest{Op: 0}},
				{core.ReturnOperation, 1, NoopResponse{Value: 10}},
				{core.InvokeOperation, 2, NoopRequest{Op: 1, Value: 15}},
				{core.ReturnOperation, 2, NoopResponse{Value: 15}},
				{core.InvokeOperation, 1, NoopRequest{Op: 0}},
				{core.ReturnOperation, 1, NoopResponse{Value: 15}},
			},
		},

		// Pending requests.
		{
			ops: []core.Operation{
				{core.InvokeOperation, 1, NoopRequest{Op: 0}},
				{core.ReturnOperation, 1, nil},
			},
			compOps: []core.Operation{
				{core.InvokeOperation, 1, NoopRequest{Op: 0}},
				{core.ReturnOperation, 1, NoopResponse{Unknown: true}},
			},
		},

		// Missing a response
		{
			ops: []core.Operation{
				{core.InvokeOperation, 1, NoopRequest{Op: 0}},
			},
			compOps: []core.Operation{
				{core.InvokeOperation, 1, NoopRequest{Op: 0}},
				{core.ReturnOperation, 1, NoopResponse{Unknown: true}},
			},
		},

		// A complex out of order history.
		{
			ops: []core.Operation{
				{core.InvokeOperation, 1, NoopRequest{Op: 0}},
				{core.InvokeOperation, 3, NoopRequest{Op: 0}},
				{core.InvokeOperation, 2, NoopRequest{Op: 1, Value: 15}},
				{core.ReturnOperation, 2, nil},
				{core.InvokeOperation, 4, NoopRequest{Op: 1, Value: 16}},
				{core.ReturnOperation, 3, nil},
			},
			compOps: []core.Operation{
				{core.InvokeOperation, 1, NoopRequest{Op: 0}},
				{core.InvokeOperation, 3, NoopRequest{Op: 0}},
				{core.InvokeOperation, 2, NoopRequest{Op: 1, Value: 15}},
				{core.InvokeOperation, 4, NoopRequest{Op: 1, Value: 16}},
				{core.ReturnOperation, 1, NoopResponse{Unknown: true}},
				{core.ReturnOperation, 2, NoopResponse{Unknown: true}},
				{core.ReturnOperation, 3, NoopResponse{Unknown: true}},
				{core.ReturnOperation, 4, NoopResponse{Unknown: true}},
			},
		},
	}

	for i, cs := range cases {
		compOps, err := CompleteOperations(cs.ops, NoopParser{})
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
