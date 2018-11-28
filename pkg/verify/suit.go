package verify

import (
	"log"

	"github.com/siddontang/chaos/pkg/core"
	"github.com/siddontang/chaos/pkg/history"
)

// Suit collects a checker, a model and a parser.
type Suit struct {
	Checker core.Checker
	Model   core.Model
	Parser  history.RecordParser
}

// Verify creates the verifier from model name and verfies the history file.
func (s Suit) Verify(historyFile string) {
	log.Printf("begin to check %s with %s", s.Model.Name(), s.Checker.Name())
	ops, state, err := history.ReadHistory(historyFile, s.Parser)
	if err != nil {
		log.Fatalf("verify failed: %v", err)
	}

	ops, err = history.CompleteOperations(ops, s.Parser)
	if err != nil {
		log.Fatalf("verify failed: %v", err)
	}

	s.Model.Prepare(state)
	ok, err := s.Checker.Check(s.Model, ops)
	if err != nil {
		log.Fatalf("verify history failed %v", err)
	}

	if !ok {
		log.Fatalf("%s: history %s is not linearizable", s.Model.Name(), historyFile)
	} else {
		log.Printf("%s: history %s is linearizable", s.Model.Name(), historyFile)
	}
}
