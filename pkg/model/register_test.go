package model

import (
	"testing"

	"github.com/anishathalye/porcupine"
)

func TestRegisterModel(t *testing.T) {
	// Test taken from github.com/anishathalye/porcupine
	// examples taken from http://nil.csail.mit.edu/6.824/2017/quizzes/q2-17-ans.pdf
	// section VII

	ops := []porcupine.Operation{
		{RegisterRequest{RegisterWrite, 100}, 0, RegisterResponse{nil, 0}, 100},
		{RegisterRequest{RegisterRead, 0}, 25, RegisterResponse{nil, 100}, 75},
		{RegisterRequest{RegisterRead, 0}, 30, RegisterResponse{nil, 0}, 60},
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
		{porcupine.ReturnEvent, RegisterResponse{nil, 0}, 2},
		{porcupine.ReturnEvent, RegisterResponse{nil, 100}, 1},
		{porcupine.ReturnEvent, RegisterResponse{nil, 0}, 0},
	}
	res = porcupine.CheckEvents(RegisterModel(), events)
	if res != true {
		t.Fatal("expected operations to be linearizable")
	}

	ops = []porcupine.Operation{
		{RegisterRequest{RegisterWrite, 200}, 0, RegisterResponse{nil, 0}, 100},
		{RegisterRequest{RegisterRead, 0}, 10, RegisterResponse{nil, 200}, 30},
		{RegisterRequest{RegisterRead, 0}, 40, RegisterResponse{nil, 0}, 90},
	}
	res = porcupine.CheckOperations(RegisterModel(), ops)
	if res != false {
		t.Fatal("expected operations to not be linearizable")
	}

	// same example as above, but with Event
	events = []porcupine.Event{
		{porcupine.CallEvent, RegisterRequest{RegisterWrite, 200}, 0},
		{porcupine.CallEvent, RegisterRequest{RegisterRead, 0}, 1},
		{porcupine.ReturnEvent, RegisterResponse{nil, 200}, 1},
		{porcupine.CallEvent, RegisterRequest{RegisterRead, 0}, 2},
		{porcupine.ReturnEvent, RegisterResponse{nil, 0}, 2},
		{porcupine.ReturnEvent, RegisterResponse{nil, 0}, 0},
	}
	res = porcupine.CheckEvents(RegisterModel(), events)
	if res != false {
		t.Fatal("expected operations to not be linearizable")
	}
}
