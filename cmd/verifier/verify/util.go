package verify

import (
	"context"
	"log"
	"strings"

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

	go func() {
		for _, verifier := range verifieres {
			log.Printf("begin to check with %s", verifier.Name())
			ok, err := verifier.Verify(historyFile)
			if err != nil {
				log.Fatalf("verify history failed %v", err)
			}

			if !ok {
				log.Fatalf("%s: history %s is not linearizable", verifier.Name(), historyFile)
			} else {
				log.Printf("%s: history %s is linearizable", verifier.Name(), historyFile)
			}
		}

		cancel()
	}()

	<-childCtx.Done()
}
