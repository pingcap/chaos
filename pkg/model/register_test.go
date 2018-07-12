package model

import (
	"encoding/json"
	"testing"

	"github.com/anishathalye/porcupine"
)

func TestRegisterModel(t *testing.T) {
	// Test taken from github.com/anishathalye/porcupine
	// examples taken from http://nil.csail.mit.edu/6.824/2017/quizzes/q2-17-ans.pdf
	// section VII

	ops := []porcupine.Operation{
		{RegisterRequest{RegisterWrite, 100}, 0, RegisterResponse{false, 0}, 100},
		{RegisterRequest{RegisterRead, 0}, 25, RegisterResponse{false, 100}, 75},
		{RegisterRequest{RegisterRead, 0}, 30, RegisterResponse{false, 0}, 60},
	}
	res := porcupine.CheckOperations(RegisterModel(), ops)
	if res != true {
		t.Fatal("expected operations to be linearizable")
	}

	// same example as above, but with Event
	events := []porcupine.Event{
		{porcupine.CallEvent, RegisterRequest{RegisterWrite, 100}, 0},
		{porcupine.CallEvent, RegisterRequest{RegisterRead, 0}, 1},
		{porcupine.CallEvent, RegisterRequest{RegisterRead, 0}, 2},
		{porcupine.ReturnEvent, RegisterResponse{false, 0}, 2},
		{porcupine.ReturnEvent, RegisterResponse{false, 100}, 1},
		{porcupine.ReturnEvent, RegisterResponse{false, 0}, 0},
	}
	res = porcupine.CheckEvents(RegisterModel(), events)
	if res != true {
		t.Fatal("expected operations to be linearizable")
	}

	ops = []porcupine.Operation{
		{RegisterRequest{RegisterWrite, 200}, 0, RegisterResponse{false, 0}, 100},
		{RegisterRequest{RegisterRead, 0}, 10, RegisterResponse{false, 200}, 30},
		{RegisterRequest{RegisterRead, 0}, 40, RegisterResponse{false, 0}, 90},
	}
	res = porcupine.CheckOperations(RegisterModel(), ops)
	if res != false {
		t.Fatal("expected operations to not be linearizable")
	}

	// same example as above, but with Event
	events = []porcupine.Event{
		{porcupine.CallEvent, RegisterRequest{RegisterWrite, 200}, 0},
		{porcupine.CallEvent, RegisterRequest{RegisterRead, 0}, 1},
		{porcupine.ReturnEvent, RegisterResponse{false, 200}, 1},
		{porcupine.CallEvent, RegisterRequest{RegisterRead, 0}, 2},
		{porcupine.ReturnEvent, RegisterResponse{false, 0}, 2},
		{porcupine.ReturnEvent, RegisterResponse{false, 0}, 0},
	}
	res = porcupine.CheckEvents(RegisterModel(), events)
	if res != false {
		t.Fatal("expected operations to not be linearizable")
	}
}

func TestRegisterParser(t *testing.T) {
	parser := registerParser{}

	for _, c := range []struct {
		data []byte
		req  RegisterRequest
	}{
		{[]byte(`{"Op":false,"Value":0}`),
			RegisterRequest{RegisterRead, 0}},
		{[]byte(`{"Op":true,"Value":5135627519135374257}`),
			RegisterRequest{RegisterWrite, 5135627519135374257}},
	} {
		raw := json.RawMessage{}
		if err := json.Unmarshal(c.data, &raw); err != nil {
			t.Fatalf("unexpected err: %v, req: %v", err, c.req)
		}
		r, err := parser.OnRequest(raw)
		if err != nil {
			t.Fatalf("unexpected err: %v, req: %v", err, c.req)
		}
		req := r.(RegisterRequest)
		if req.Op != c.req.Op || req.Value != c.req.Value {
			t.Fatalf("expected equal: %v, req: %v", req, c.req)
		}
	}

	for _, c := range []struct {
		data []byte
		resp RegisterResponse
	}{
		{[]byte(`{"Unknown":false,"Value":0}`),
			RegisterResponse{false, 0}},
		{[]byte(`{"Unknown":false,"Value":5135627519135374257}`),
			RegisterResponse{false, 5135627519135374257}},
		{[]byte(`{"Unknown":true,"Value":7}`),
			RegisterResponse{true, 7}},
	} {
		raw := json.RawMessage{}
		if err := json.Unmarshal(c.data, &raw); err != nil {
			t.Fatalf("unexpected err: %v, resp: %v", err, c.resp)
		}
		r, err := parser.OnResponse(raw)
		if err != nil {
			t.Fatalf("unexpected err: %v, resp: %v", err, c.resp)
		}
		if r != nil {
			resp := r.(RegisterResponse)
			if resp.Value != c.resp.Value || resp.Unknown != c.resp.Unknown {
				t.Fatalf("expected equal: %v != %v", resp, c.resp)
			}
		}
	}
}
