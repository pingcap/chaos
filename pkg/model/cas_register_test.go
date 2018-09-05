package model

import (
	"testing"

	"github.com/anishathalye/porcupine"
)

func TestCasRegisterModel(t *testing.T) {
	events := []porcupine.Event{
		{Kind: porcupine.CallEvent, Value: CasRegisterRequest{Op: CasRegisterWrite, Arg1: 100, Arg2: 0}, Id: 0},
		{Kind: porcupine.CallEvent, Value: CasRegisterRequest{Op: CasRegisterRead}, Id: 1},
		{Kind: porcupine.CallEvent, Value: CasRegisterRequest{Op: CasRegisterRead}, Id: 2},
		{Kind: porcupine.CallEvent, Value: CasRegisterRequest{Op: CasRegisterCAS, Arg1: 100, Arg2: 200}, Id: 3},
		{Kind: porcupine.CallEvent, Value: CasRegisterRequest{Op: CasRegisterRead}, Id: 4},
		{Kind: porcupine.ReturnEvent, Value: CasRegisterResponse{}, Id: 2},
		{Kind: porcupine.ReturnEvent, Value: CasRegisterResponse{Value: 100}, Id: 1},
		{Kind: porcupine.ReturnEvent, Value: CasRegisterResponse{}, Id: 0},
		{Kind: porcupine.ReturnEvent, Value: CasRegisterResponse{Ok: true}, Id: 3},
		{Kind: porcupine.ReturnEvent, Value: CasRegisterResponse{Value: 200}, Id: 4},
	}
	res := porcupine.CheckEvents(CasRegisterModel(), events)
	if res != true {
		t.Fatal("expected operations to be linearizable")
	}
}
