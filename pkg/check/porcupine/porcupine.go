package porcupine

import (
	"log"

	"github.com/pkg/errors"

	"github.com/anishathalye/porcupine"
	"github.com/siddontang/chaos/pkg/core"
)

// Checker is a linearizability checker powered by Porcupine.
type Checker struct{}

// Check checks the history of operations meets liearizability or not with model.
// False means the history is not linearizable.
func (Checker) Check(m core.Model, ops []core.Operation) (bool, error) {
	pModel := porcupine.Model{
		Init:  m.Init,
		Step:  m.Step,
		Equal: m.Equal,
	}
	events, err := ConvertOperationsToEvents(ops)
	if err != nil {
		return false, err
	}
	log.Printf("begin to verify %d events", len(events))
	return porcupine.CheckEvents(pModel, events), nil
}

// Name is the name of porcupine checker
func (Checker) Name() string {
	return "porcupine_checker"
}

// ConvertOperationsToEvents converts core.Operations to porcupine.Event.
func ConvertOperationsToEvents(ops []core.Operation) ([]porcupine.Event, error) {
	if len(ops)%2 != 0 {
		return nil, errors.New("history is not complete")
	}

	procID := map[int64]uint{}
	id := uint(0)
	events := make([]porcupine.Event, 0, len(ops))
	for _, op := range ops {
		if op.Action == core.InvokeOperation {
			event := porcupine.Event{
				Kind:  porcupine.CallEvent,
				Id:    id,
				Value: op.Data,
			}
			events = append(events, event)
			procID[op.Proc] = id
			id++
		} else {
			if op.Data == nil {
				continue
			}

			matchID := procID[op.Proc]
			delete(procID, op.Proc)
			event := porcupine.Event{
				Kind:  porcupine.ReturnEvent,
				Id:    matchID,
				Value: op.Data,
			}
			events = append(events, event)
		}
	}

	if len(procID) != 0 {
		return nil, errors.New("history is not complete")
	}

	return events, nil
}
