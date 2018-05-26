package verify

import (
	"context"
	"log"
	"strings"
	"sync"

	"github.com/siddontang/chaos/db/tidb"
	"github.com/siddontang/chaos/pkg/history"
)

// Verify creates the verifier from verifer_names and verfies the history file.
func Verify(ctx context.Context, historyFile string, verfier_names string) {
	var verifieres []history.Verifier

	for _, name := range strings.Split(verfier_names, ",") {
		var verifier history.Verifier
		switch name {
		case "tidb_bank":
			verifier = tidb.BankVerifier{}
		case "tidb_bank_tso":
			verifier = tidb.BankTsoVerifier{}
		case "":
			continue
		default:
			log.Printf("%s is not supported", name)
			continue
		}

		verifieres = append(verifieres, verifier)
	}

	childCtx, cancel := context.WithCancel(ctx)

	var wg sync.WaitGroup
	wg.Add(len(verifieres))
	go func() {
		wg.Wait()
		cancel()
	}()

	for _, verifier := range verifieres {
		// Verify may take a long time, we should quit ASAP if receive signal.
		go func(verifier history.Verifier) {
			defer wg.Done()
			ok, err := verifier.Verify(historyFile)
			if err != nil {
				log.Fatalf("verify history failed %v", err)
			}

			if !ok {
				log.Fatalf("history %s is not linearizable", historyFile)
			} else {
				log.Printf("history %s is linearizable", historyFile)
			}
		}(verifier)
	}
	<-childCtx.Done()
}
