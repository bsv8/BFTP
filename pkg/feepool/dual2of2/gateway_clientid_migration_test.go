package dual2of2

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"testing"

	"github.com/libp2p/go-libp2p/core/crypto"
	_ "modernc.org/sqlite"
)

func TestMigrateGatewayClientIDs_ReconcileClosesLegacyOpenInterruption(t *testing.T) {
	db := openGatewayMigrationTestDB(t)
	canonicalID, legacyID := mustClientIDPair(t)
	now := int64(1773212000)

	_, err := db.Exec(
		`INSERT INTO fee_pool_sessions(spend_txid,client_id,client_bsv_pubkey_hex,server_bsv_pubkey_hex,input_amount_satoshi,pool_amount_satoshi,spend_tx_fee_satoshi,sequence_num,server_amount_satoshi,client_amount_satoshi,base_txid,final_txid,base_tx_hex,current_tx_hex,lifecycle_state,created_at_unix,updated_at_unix)
		 VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		"tx_active_1", canonicalID, canonicalID, "02aa", 600, 600, 1, 1, 1, 598, "base1", "", "00", "00", "active", now, now,
	)
	if err != nil {
		t.Fatalf("insert session failed: %v", err)
	}
	_, err = db.Exec(
		`INSERT INTO fee_pool_client_presence(client_id,peer_id,online,last_online_at_unix,last_offline_at_unix,updated_at_unix)
		 VALUES(?,?,?,?,?,?)`,
		canonicalID, "peer_canon", 1, now, 0, now,
	)
	if err != nil {
		t.Fatalf("insert canonical presence failed: %v", err)
	}
	_, err = db.Exec(
		`INSERT INTO fee_pool_client_presence(client_id,peer_id,online,last_online_at_unix,last_offline_at_unix,updated_at_unix)
		 VALUES(?,?,?,?,?,?)`,
		legacyID, "peer_legacy", 0, 0, now-10, now-10,
	)
	if err != nil {
		t.Fatalf("insert legacy presence failed: %v", err)
	}
	_, err = db.Exec(
		`INSERT INTO fee_pool_service_state(client_id,state,online,coverage_active,active_session_count,current_span_id,current_pause_id,current_interrupt_id,last_transition_at_unix,updated_at_unix)
		 VALUES(?,?,?,?,?,?,?,?,?,?)`,
		canonicalID, "serving", 1, 1, 1, 2, 0, 0, now, now,
	)
	if err != nil {
		t.Fatalf("insert service_state failed: %v", err)
	}
	_, err = db.Exec(
		`INSERT INTO fee_pool_service_spans(id,client_id,start_at_unix,end_at_unix,status,start_spend_txid,end_spend_txid,end_reason,created_at_unix,updated_at_unix)
		 VALUES(?,?,?,?,?,?,?,?,?,?)`,
		1, legacyID, now-120, now-60, "closed", "tx_old", "tx_old", "coverage_lost", now-120, now-60,
	)
	if err != nil {
		t.Fatalf("insert legacy span failed: %v", err)
	}
	_, err = db.Exec(
		`INSERT INTO fee_pool_service_spans(id,client_id,start_at_unix,end_at_unix,status,start_spend_txid,end_spend_txid,end_reason,created_at_unix,updated_at_unix)
		 VALUES(?,?,?,?,?,?,?,?,?,?)`,
		2, canonicalID, now-59, 0, "open", "tx_active_1", "", "", now-59, now-59,
	)
	if err != nil {
		t.Fatalf("insert canonical span failed: %v", err)
	}
	_, err = db.Exec(
		`INSERT INTO fee_pool_service_interruptions(id,client_id,start_at_unix,end_at_unix,status,reason,created_at_unix,updated_at_unix)
		 VALUES(?,?,?,?,?,?,?,?)`,
		1, legacyID, now-30, 0, "open", "session_passive_closed", now-30, now-30,
	)
	if err != nil {
		t.Fatalf("insert legacy interruption failed: %v", err)
	}

	if err := MigrateGatewayClientIDs(db); err != nil {
		t.Fatalf("migrate client_id failed: %v", err)
	}
	svc := &GatewayService{DB: db}
	if err := svc.ReconcileServiceStates("unit_startup_reconcile"); err != nil {
		t.Fatalf("reconcile states failed: %v", err)
	}

	assertClientIDCount(t, db, "fee_pool_client_presence", legacyID, 0)
	assertClientIDCount(t, db, "fee_pool_service_state", legacyID, 0)
	assertClientIDCount(t, db, "fee_pool_service_spans", legacyID, 0)
	assertClientIDCount(t, db, "fee_pool_service_interruptions", legacyID, 0)

	var state string
	var currentInterruptID int64
	if err := db.QueryRow(`SELECT state,current_interrupt_id FROM fee_pool_service_state WHERE client_id=?`, canonicalID).Scan(&state, &currentInterruptID); err != nil {
		t.Fatalf("query service_state failed: %v", err)
	}
	if state != "serving" {
		t.Fatalf("state mismatch: got=%s want=serving", state)
	}
	if currentInterruptID != 0 {
		t.Fatalf("current_interrupt_id mismatch: got=%d want=0", currentInterruptID)
	}

	var openInterruptions int
	if err := db.QueryRow(`SELECT COUNT(*) FROM fee_pool_service_interruptions WHERE client_id=? AND status='open'`, canonicalID).Scan(&openInterruptions); err != nil {
		t.Fatalf("count open interruptions failed: %v", err)
	}
	if openInterruptions != 0 {
		t.Fatalf("open interruption should be closed after reconcile, got=%d", openInterruptions)
	}
}

func TestLoadLatestSessionByClientID_AcceptsLegacyAlias(t *testing.T) {
	db := openGatewayMigrationTestDB(t)
	canonicalID, legacyID := mustClientIDPair(t)
	if err := InsertSession(db, GatewaySessionRow{
		SpendTxID:                 "tx_alias_1",
		ClientID:                  canonicalID,
		ClientBSVCompressedPubHex: canonicalID,
		ServerBSVCompressedPubHex: "02aa",
		InputAmountSat:            600,
		PoolAmountSat:             600,
		SpendTxFeeSat:             1,
		Sequence:                  1,
		ServerAmountSat:           1,
		ClientAmountSat:           598,
		BaseTxID:                  "",
		FinalTxID:                 "",
		BaseTxHex:                 "00",
		CurrentTxHex:              "00",
		LifecycleState:            "active",
	}); err != nil {
		t.Fatalf("insert session failed: %v", err)
	}
	row, found, err := LoadLatestSessionByClientID(db, legacyID)
	if err != nil {
		t.Fatalf("load by legacy alias failed: %v", err)
	}
	if !found {
		t.Fatalf("session should be found by legacy alias")
	}
	if row.ClientID != canonicalID {
		t.Fatalf("client_id mismatch: got=%s want=%s", row.ClientID, canonicalID)
	}
}

func openGatewayMigrationTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := InitGatewayStore(db); err != nil {
		t.Fatalf("init gateway store failed: %v", err)
	}
	return db
}

func mustClientIDPair(t *testing.T) (canonical string, legacy string) {
	t.Helper()
	priv, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
	if err != nil {
		t.Fatalf("generate key failed: %v", err)
	}
	pub := priv.GetPublic()
	raw, err := pub.Raw()
	if err != nil {
		t.Fatalf("public raw failed: %v", err)
	}
	canonical = hex.EncodeToString(raw)
	marshal, err := crypto.MarshalPublicKey(pub)
	if err != nil {
		t.Fatalf("marshal public key failed: %v", err)
	}
	legacy = hex.EncodeToString(marshal)
	return canonical, legacy
}

func assertClientIDCount(t *testing.T, db *sql.DB, table string, clientID string, want int) {
	t.Helper()
	var got int
	q := `SELECT COUNT(*) FROM ` + table + ` WHERE client_id=?`
	if err := db.QueryRow(q, clientID).Scan(&got); err != nil {
		t.Fatalf("query count failed: table=%s err=%v", table, err)
	}
	if got != want {
		t.Fatalf("count mismatch: table=%s got=%d want=%d", table, got, want)
	}
}
