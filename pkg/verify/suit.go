package verify

import (
	"log"

	"github.com/pingcap/chaos/pkg/core"
	"github.com/pingcap/chaos/pkg/history"
)

// Suit collects a checker, a model and a parser.
type Suit struct {
	Checker core.Checker
	Model   core.Model
	Parser  history.RecordParser
}

// Verify creates the verifier from model name and verfies the history file.
func (s Suit) Verify(historyFile string) {
	if s.Model == nil {
		log.Printf("begin to check %s", s.Checker.Name())
	} else {
		log.Printf("begin to check %s with %s", s.Model.Name(), s.Checker.Name())
	}
	ops, state, err := history.ReadHistory(historyFile, s.Parser)
	if err != nil {
		log.Fatalf("verify failed: %v", err)
	}

	ops, err = history.CompleteOperations(ops, s.Parser)
	if err != nil {
		log.Fatalf("verify failed: %v", err)
	}

	if s.Model != nil {
		s.Model.Prepare(state)
	}
	ok, err := s.Checker.Check(s.Model, ops)
	if err != nil {
		log.Fatalf("verify history failed %v", err)
	}

	if !ok {
		log.Fatalf("history %s is not valid", historyFile)
	} else {
		log.Printf("history %s is valid", historyFile)
	}
}
