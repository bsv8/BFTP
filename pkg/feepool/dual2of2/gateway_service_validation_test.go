package dual2of2

import (
	"strings"
	"testing"

	"github.com/bsv-blockchain/go-sdk/chainhash"
	"github.com/bsv-blockchain/go-sdk/script"
	tx "github.com/bsv-blockchain/go-sdk/transaction"
	_ "modernc.org/sqlite"
)

type stubChainForValidation struct {
	tip uint32
}

func (s stubChainForValidation) GetUTXOs(address string) ([]UTXO, error) { return nil, nil }
func (s stubChainForValidation) GetTipHeight() (uint32, error)           { return s.tip, nil }
func (s stubChainForValidation) Broadcast(txHex string) (string, error)  { return "txid_stub", nil }

func TestValidateBaseTxMatchesSessionSpend(t *testing.T) {
	lock, err := script.NewFromHex("51")
	if err != nil {
		t.Fatalf("build lock script failed: %v", err)
	}
	baseTx := tx.NewTransaction()
	baseTx.Outputs = append(baseTx.Outputs, &tx.TransactionOutput{
		Satoshis:      1000,
		LockingScript: lock,
	})
	baseID := baseTx.TxID().String()
	h, err := chainhash.NewHashFromHex(baseID)
	if err != nil {
		t.Fatalf("hash base txid failed: %v", err)
	}
	spendTx := tx.NewTransaction()
	spendTx.Inputs = append(spendTx.Inputs, &tx.TransactionInput{
		SourceTXID:       h,
		SourceTxOutIndex: 0,
		SequenceNumber:   1,
	})
	spendTx.Outputs = append(spendTx.Outputs, &tx.TransactionOutput{
		Satoshis:      900,
		LockingScript: lock,
	})

	if err := validateBaseTxMatchesSessionSpend(baseTx, spendTx.Hex()); err != nil {
		t.Fatalf("validate match failed: %v", err)
	}

	spendTx.Inputs[0].SourceTxOutIndex = 1
	if err := validateBaseTxMatchesSessionSpend(baseTx, spendTx.Hex()); err == nil {
		t.Fatalf("expect vout mismatch error")
	}
}

func TestGatewayServicePayConfirmRejectListenFeeMismatch(t *testing.T) {
	db := openTimelineTestDB(t)
	row := GatewaySessionRow{
		SpendTxID:                 "tx_listen_fee_mismatch",
		ClientID:                  "client_a",
		ClientBSVCompressedPubHex: "02aa",
		ServerBSVCompressedPubHex: "03bb",
		InputAmountSat:            1000,
		PoolAmountSat:             1000,
		SpendTxFeeSat:             1,
		Sequence:                  1,
		ServerAmountSat:           100,
		ClientAmountSat:           899,
		BaseTxID:                  "base_a",
		FinalTxID:                 "",
		BaseTxHex:                 "00",
		CurrentTxHex:              "01000000000064000000",
		LifecycleState:            "active",
	}
	if err := InsertSession(db, row); err != nil {
		t.Fatalf("insert session failed: %v", err)
	}

	svc := &GatewayService{
		DB: db,
		Chain: stubChainForValidation{
			tip: 10,
		},
		Params: GatewayParams{
			BillingCycleSeconds:         60,
			SingleCycleFeeSatoshi:       50,
			PayRejectBeforeExpiryBlocks: 1,
		},
	}
	resp, err := svc.PayConfirm(PayConfirmReq{
		SpendTxID:           row.SpendTxID,
		SequenceNumber:      2,
		ServerAmount:        row.ServerAmountSat + 1,
		Fee:                 row.SpendTxFeeSat,
		ChargeReason:        "listen_cycle_fee",
		ChargeAmountSatoshi: 1,
	})
	if err != nil {
		t.Fatalf("pay confirm returned unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatalf("pay confirm should be rejected")
	}
	if !strings.Contains(strings.ToLower(resp.Error), "listen cycle fee amount mismatch") {
		t.Fatalf("unexpected reject error: %s", resp.Error)
	}

	after, found, err := LoadSessionBySpendTxID(db, row.SpendTxID)
	if err != nil {
		t.Fatalf("reload session failed: %v", err)
	}
	if !found {
		t.Fatalf("session should exist")
	}
	if after.Sequence != row.Sequence || after.ServerAmountSat != row.ServerAmountSat {
		t.Fatalf("session should remain unchanged: seq=%d/%d server=%d/%d", after.Sequence, row.Sequence, after.ServerAmountSat, row.ServerAmountSat)
	}
}

func TestLoadPreferredSessionByClientIDPrefersActive(t *testing.T) {
	db := openTimelineTestDB(t)
	if err := InsertSession(db, GatewaySessionRow{
		SpendTxID:                 "tx_closed",
		ClientID:                  "client_pref",
		ClientBSVCompressedPubHex: "02aa",
		ServerBSVCompressedPubHex: "03bb",
		InputAmountSat:            1000,
		PoolAmountSat:             1000,
		SpendTxFeeSat:             1,
		Sequence:                  1,
		ServerAmountSat:           1,
		ClientAmountSat:           998,
		BaseTxID:                  "base_closed",
		FinalTxID:                 "final_closed",
		BaseTxHex:                 "00",
		CurrentTxHex:              "00",
		LifecycleState:            "closed",
	}); err != nil {
		t.Fatalf("insert closed session failed: %v", err)
	}
	if err := InsertSession(db, GatewaySessionRow{
		SpendTxID:                 "tx_active",
		ClientID:                  "client_pref",
		ClientBSVCompressedPubHex: "02aa",
		ServerBSVCompressedPubHex: "03bb",
		InputAmountSat:            1000,
		PoolAmountSat:             1000,
		SpendTxFeeSat:             1,
		Sequence:                  2,
		ServerAmountSat:           2,
		ClientAmountSat:           997,
		BaseTxID:                  "base_active",
		FinalTxID:                 "",
		BaseTxHex:                 "00",
		CurrentTxHex:              "00",
		LifecycleState:            "active",
	}); err != nil {
		t.Fatalf("insert active session failed: %v", err)
	}

	// 模拟旧池在 rotate 后被晚一步 close，更新时间更晚。
	if _, err := db.Exec(`UPDATE fee_pool_sessions SET updated_at_unix=updated_at_unix+100 WHERE spend_txid='tx_closed'`); err != nil {
		t.Fatalf("bump closed updated_at failed: %v", err)
	}

	row, found, err := LoadPreferredSessionByClientID(db, "client_pref")
	if err != nil {
		t.Fatalf("load preferred failed: %v", err)
	}
	if !found {
		t.Fatalf("preferred session should exist")
	}
	if row.SpendTxID != "tx_active" {
		t.Fatalf("preferred session mismatch: got=%s want=tx_active", row.SpendTxID)
	}
}
