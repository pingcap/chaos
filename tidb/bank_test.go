package tidb

import (
	"testing"

	"github.com/anishathalye/porcupine"
)

func TestBankVerify(t *testing.T) {
	m := getBankModel(2)

	events := []porcupine.Event{
		newBankEvent(bankRequest{Op: 0}, 1),
		newBankEvent(bankResponse{Balances: []int64{1000, 1000}}, 1),
		newBankEvent(bankRequest{Op: 1, From: 0, To: 1, Amount: 500}, 2),
		newBankEvent(bankResponse{Ok: true}, 2),
		newBankEvent(bankRequest{Op: 0}, 3),
		newBankEvent(bankResponse{Balances: []int64{500, 1500}}, 3),
	}

	if !porcupine.CheckEvents(m, events) {
		t.Fatal("must be linearizable")
	}
}

func TestBankVerifyUnknown(t *testing.T) {
	m := getBankModel(2)

	events := []porcupine.Event{
		newBankEvent(bankRequest{Op: 0}, 1),
		newBankEvent(bankResponse{Balances: []int64{1000, 1000}}, 1),
		newBankEvent(bankRequest{Op: 1, From: 0, To: 1, Amount: 500}, 2),
		// write return unknow, so we consider its return time is infinite.
		// newBankEvent(bankResponse{Unknown: true}, 2),

		newBankEvent(bankRequest{Op: 0}, 3),
		newBankEvent(bankResponse{Balances: []int64{500, 1500}}, 3),
		newBankEvent(bankRequest{Op: 0}, 4),

		newBankEvent(bankResponse{Unknown: true}, 2),
		newBankEvent(bankResponse{Unknown: true}, 4),
	}

	if !porcupine.CheckEvents(m, events) {
		t.Fatal("must be linearizable")
	}
}

func TestBankVerifyNoLinerizable(t *testing.T) {
	m := getBankModel(2)

	events := []porcupine.Event{
		newBankEvent(bankRequest{Op: 0}, 1),
		newBankEvent(bankResponse{Balances: []int64{1000, 1000}}, 1),
		newBankEvent(bankRequest{Op: 1, From: 0, To: 1, Amount: 500}, 2),
		newBankEvent(bankResponse{Ok: true}, 2),
		newBankEvent(bankRequest{Op: 0}, 3),
		newBankEvent(bankResponse{Balances: []int64{1000, 1000}}, 3),
	}

	if porcupine.CheckEvents(m, events) {
		t.Fatal("must be not linearizable")
	}
}
