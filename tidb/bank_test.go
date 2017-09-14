package tidb

import (
	"testing"

	"github.com/anishathalye/porcupine"
)

func TestBankVerify(t *testing.T) {
	m := getBankModel(2)

	events := []porcupine.Event{
		newBankEvent(bankRequest{op: 0}, 1),
		newBankEvent(bankResponse{ok: true, balances: []int64{1000, 1000}}, 1),
		newBankEvent(bankRequest{op: 1, from: 0, to: 1, amount: 500}, 2),
		newBankEvent(bankResponse{ok: true}, 2),
		newBankEvent(bankRequest{op: 0}, 3),
		newBankEvent(bankResponse{ok: true, balances: []int64{500, 1500}}, 3),
	}

	if !porcupine.CheckEvents(m, events) {
		t.Fatal("must be linearizable")
	}
}

func TestBankVerifyUnknown(t *testing.T) {
	m := getBankModel(2)

	events := []porcupine.Event{
		newBankEvent(bankRequest{op: 0}, 1),
		newBankEvent(bankResponse{balances: []int64{1000, 1000}}, 1),
		newBankEvent(bankRequest{op: 1, from: 0, to: 1, amount: 500}, 2),
		// write return unknow, so we consider its return time is infinite.
		// newBankEvent(bankResponse{unknown: true}, 2),

		newBankEvent(bankRequest{op: 0}, 3),
		newBankEvent(bankResponse{balances: []int64{500, 1500}}, 3),
		newBankEvent(bankRequest{op: 0}, 4),

		newBankEvent(bankResponse{unknown: true}, 2),
		newBankEvent(bankResponse{unknown: true}, 4),
	}

	if !porcupine.CheckEvents(m, events) {
		t.Fatal("must be linearizable")
	}
}

func TestBankVerifyNoLinerizable(t *testing.T) {
	m := getBankModel(2)

	events := []porcupine.Event{
		newBankEvent(bankRequest{op: 0}, 1),
		newBankEvent(bankResponse{balances: []int64{1000, 1000}}, 1),
		newBankEvent(bankRequest{op: 1, from: 0, to: 1, amount: 500}, 2),
		newBankEvent(bankResponse{ok: true}, 2),
		newBankEvent(bankRequest{op: 0}, 3),
		newBankEvent(bankResponse{balances: []int64{1000, 1000}}, 3),
	}

	if porcupine.CheckEvents(m, events) {
		t.Fatal("must be not linearizable")
	}
}
