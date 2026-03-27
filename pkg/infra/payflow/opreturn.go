package payflow

import (
	"fmt"
	"strings"

	"github.com/bsv-blockchain/go-sdk/script"
	txsdk "github.com/bsv-blockchain/go-sdk/transaction"
)

func ExtractProofStatePayloadFromScript(lockingScript *script.Script) ([]byte, error) {
	if lockingScript == nil || !lockingScript.IsData() {
		return nil, fmt.Errorf("op_return output required")
	}
	ops, err := lockingScript.ParseOps()
	if err != nil {
		return nil, err
	}
	start := 0
	if len(ops) == 0 {
		return nil, fmt.Errorf("op_return payload missing")
	}
	if ops[0].Op == script.OpFALSE {
		if len(ops) < 2 || ops[1].Op != script.OpRETURN {
			return nil, fmt.Errorf("op_return script invalid")
		}
		start = 2
	} else if ops[0].Op == script.OpRETURN {
		start = 1
	} else {
		return nil, fmt.Errorf("op_return script invalid")
	}
	payload := make([]byte, 0, len(lockingScript.Bytes()))
	for _, op := range ops[start:] {
		if len(op.Data) == 0 {
			continue
		}
		payload = append(payload, op.Data...)
	}
	if len(payload) == 0 {
		return nil, fmt.Errorf("op_return payload missing")
	}
	return payload, nil
}

func ExtractProofStateFromTxHex(txHex string) (ProofState, bool, error) {
	txHex = strings.TrimSpace(txHex)
	if txHex == "" {
		return ProofState{}, false, nil
	}
	parsed, err := txsdk.NewTransactionFromHex(txHex)
	if err != nil {
		return ProofState{}, false, err
	}
	for _, out := range parsed.Outputs {
		if out == nil || out.LockingScript == nil || !out.LockingScript.IsData() {
			continue
		}
		payload, err := ExtractProofStatePayloadFromScript(out.LockingScript)
		if err != nil {
			return ProofState{}, false, err
		}
		state, err := UnmarshalProofState(payload)
		if err != nil {
			return ProofState{}, false, err
		}
		return state, true, nil
	}
	return ProofState{}, false, nil
}
