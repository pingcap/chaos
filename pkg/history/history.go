package history

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"path"
	"sync"

	"github.com/anishathalye/porcupine"
)

// Operation action
const (
	InvokeOperation = "call"
	ReturnOperation = "return"
)

type operation struct {
	Action string          `json:"action"`
	Proc   int64           `json:"proc"`
	Data   json.RawMessage `json:"data"`
}

// Recorder records operation histogry.
type Recorder struct {
	sync.Mutex
	f *os.File
}

// NewRecorder creates a recorder to log the history to the file.
func NewRecorder(name string) (*Recorder, error) {
	os.MkdirAll(path.Dir(name), 0755)

	f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	return &Recorder{f: f}, nil
}

// Close closes the recorder.
func (r *Recorder) Close() {
	r.f.Close()
}

// RecordRequest records the request.
func (r *Recorder) RecordRequest(proc int64, op interface{}) error {
	return r.record(proc, InvokeOperation, op)
}

// RecordResponse records the response.
func (r *Recorder) RecordResponse(proc int64, op interface{}) error {
	return r.record(proc, ReturnOperation, op)
}

func (r *Recorder) record(proc int64, action string, op interface{}) error {
	data, err := json.Marshal(op)
	if err != nil {
		return err
	}

	v := operation{
		Action: action,
		Proc:   proc,
		Data:   json.RawMessage(data),
	}

	data, err = json.Marshal(v)
	if err != nil {
		return err
	}

	r.Lock()
	defer r.Unlock()

	if _, err = r.f.Write(data); err != nil {
		return err
	}

	if _, err = r.f.WriteString("\n"); err != nil {
		return err
	}

	return nil
}

// RecordParser is to parses the operation data.
type RecordParser interface {
	// OnRequest parses the request record.
	OnRequest(data json.RawMessage) (interface{}, error)
	// OnResponse parses the response record. Return nil means
	// the operation has an infinite end time.
	// E.g, we meet timeout for a operation.
	OnResponse(data json.RawMessage) (interface{}, error)
	// If we have some infinite operations, we should return a
	// noop response to complete the operation.
	OnNoopResponse() interface{}
}

// Verifier verifies the history.
type Verifier interface {
	Verify(historyFile string) (bool, error)
}

// VerifyHistory checks the history file with model.
// False means the history is not linearizable.
func VerifyHistory(historyFile string, m porcupine.Model, p RecordParser) (bool, error) {
	events, err := ParseEvents(historyFile, p)
	if err != nil {
		return false, err
	}
	log.Printf("begin to verify %d events", len(events))
	return porcupine.CheckEvents(m, events), nil
}

// ParseEvents parses the history and returns a procupine Event list.
func ParseEvents(historyFile string, p RecordParser) ([]porcupine.Event, error) {
	f, err := os.Open(historyFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	procID := map[int64]uint{}
	id := uint(0)

	events := make([]porcupine.Event, 0, 1024)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var op operation
		if err = json.Unmarshal(scanner.Bytes(), &op); err != nil {
			return nil, err
		}

		var value interface{}
		if op.Action == InvokeOperation {
			if value, err = p.OnRequest(op.Data); err != nil {
				return nil, err
			}

			event := porcupine.Event{
				Kind:  porcupine.CallEvent,
				Id:    id,
				Value: value,
			}
			events = append(events, event)
			procID[op.Proc] = id
			id++
		} else {
			if value, err = p.OnResponse(op.Data); err != nil {
				return nil, err
			}

			if value == nil {
				continue
			}

			matchID := procID[op.Proc]
			delete(procID, op.Proc)
			event := porcupine.Event{
				Kind:  porcupine.ReturnEvent,
				Id:    matchID,
				Value: value,
			}
			events = append(events, event)
		}
	}

	if err = scanner.Err(); err != nil {
		return nil, err
	}

	for _, id := range procID {
		response := p.OnNoopResponse()
		event := porcupine.Event{
			Kind:  porcupine.ReturnEvent,
			Id:    id,
			Value: response,
		}
		events = append(events, event)
	}

	return events, nil
}
