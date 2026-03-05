package dual2of2

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type GatewaySessionRow struct {
	SpendTxID string

	ClientID                  string
	ClientBSVCompressedPubHex string
	ServerBSVCompressedPubHex string

	InputAmountSat  uint64
	PoolAmountSat   uint64
	SpendTxFeeSat   uint64
	Sequence        uint32
	ServerAmountSat uint64
	ClientAmountSat uint64

	BaseTxID  string
	FinalTxID string

	BaseTxHex    string
	CurrentTxHex string

	Status    string
	CreatedAt int64
	UpdatedAt int64
}

func InitGatewayStore(db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS fee_pool_sessions (
			spend_txid TEXT PRIMARY KEY,

			client_id TEXT NOT NULL,
			client_bsv_pubkey_hex TEXT NOT NULL,
			server_bsv_pubkey_hex TEXT NOT NULL,

			input_amount_satoshi INTEGER NOT NULL,
			pool_amount_satoshi INTEGER NOT NULL,
			spend_tx_fee_satoshi INTEGER NOT NULL,
			sequence_num INTEGER NOT NULL,
			server_amount_satoshi INTEGER NOT NULL,
			client_amount_satoshi INTEGER NOT NULL,

			base_txid TEXT NOT NULL,
			final_txid TEXT NOT NULL,

			base_tx_hex TEXT NOT NULL,
			current_tx_hex TEXT NOT NULL,

			status TEXT NOT NULL,
			created_at_unix INTEGER NOT NULL,
			updated_at_unix INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_fee_pool_client_id ON fee_pool_sessions(client_id)`,
		`CREATE INDEX IF NOT EXISTS idx_fee_pool_status ON fee_pool_sessions(status)`,
		`CREATE INDEX IF NOT EXISTS idx_fee_pool_updated_at ON fee_pool_sessions(updated_at_unix DESC)`,
		`CREATE TABLE IF NOT EXISTS fee_pool_charge_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			client_id TEXT NOT NULL,
			spend_txid TEXT NOT NULL,
			sequence_num INTEGER NOT NULL,
			charge_reason TEXT NOT NULL,
			charge_amount_satoshi INTEGER NOT NULL,
			created_at_unix INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_fee_pool_charge_client_reason ON fee_pool_charge_events(client_id, charge_reason)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	return nil
}

func InsertChargeEvent(db *sql.DB, clientID string, spendTxID string, sequence uint32, reason string, amount uint64) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}
	clientID = strings.ToLower(strings.TrimSpace(clientID))
	spendTxID = strings.TrimSpace(spendTxID)
	reason = strings.TrimSpace(reason)
	if clientID == "" || spendTxID == "" || reason == "" {
		return fmt.Errorf("invalid charge event")
	}
	_, err := db.Exec(
		`INSERT INTO fee_pool_charge_events(client_id,spend_txid,sequence_num,charge_reason,charge_amount_satoshi,created_at_unix) VALUES(?,?,?,?,?,?)`,
		clientID, spendTxID, sequence, reason, amount, time.Now().Unix(),
	)
	return err
}

func CountChargeEventsByClientAndReason(db *sql.DB, clientID string, reason string) (int, error) {
	if db == nil {
		return 0, fmt.Errorf("db is nil")
	}
	clientID = strings.ToLower(strings.TrimSpace(clientID))
	reason = strings.TrimSpace(reason)
	if clientID == "" || reason == "" {
		return 0, fmt.Errorf("client_id and reason are required")
	}
	var n int
	if err := db.QueryRow(
		`SELECT COUNT(*) FROM fee_pool_charge_events WHERE client_id=? AND charge_reason=?`,
		clientID, reason,
	).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func ListClientsByChargeReason(db *sql.DB, reason string) ([]string, error) {
	if db == nil {
		return nil, fmt.Errorf("db is nil")
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil, fmt.Errorf("reason is required")
	}
	rows, err := db.Query(`SELECT DISTINCT client_id FROM fee_pool_charge_events WHERE charge_reason=?`, reason)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]string, 0, 8)
	for rows.Next() {
		var clientID string
		if err := rows.Scan(&clientID); err != nil {
			return nil, err
		}
		clientID = strings.ToLower(strings.TrimSpace(clientID))
		if clientID == "" {
			continue
		}
		out = append(out, clientID)
	}
	return out, nil
}

func LoadLatestSessionByClientID(db *sql.DB, clientID string) (GatewaySessionRow, bool, error) {
	clientID = strings.ToLower(strings.TrimSpace(clientID))
	if clientID == "" {
		return GatewaySessionRow{}, false, fmt.Errorf("client_id required")
	}
	var row GatewaySessionRow
	err := db.QueryRow(
		`SELECT spend_txid,client_id,client_bsv_pubkey_hex,server_bsv_pubkey_hex,input_amount_satoshi,pool_amount_satoshi,spend_tx_fee_satoshi,sequence_num,server_amount_satoshi,client_amount_satoshi,base_txid,final_txid,base_tx_hex,current_tx_hex,status,created_at_unix,updated_at_unix
		 FROM fee_pool_sessions
		 WHERE client_id=?
		 ORDER BY updated_at_unix DESC
		 LIMIT 1`, clientID,
	).Scan(
		&row.SpendTxID,
		&row.ClientID, &row.ClientBSVCompressedPubHex, &row.ServerBSVCompressedPubHex,
		&row.InputAmountSat, &row.PoolAmountSat, &row.SpendTxFeeSat,
		&row.Sequence, &row.ServerAmountSat, &row.ClientAmountSat,
		&row.BaseTxID, &row.FinalTxID,
		&row.BaseTxHex, &row.CurrentTxHex,
		&row.Status, &row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return GatewaySessionRow{}, false, nil
		}
		return GatewaySessionRow{}, false, err
	}
	return row, true, nil
}

func LoadSessionBySpendTxID(db *sql.DB, spendTxID string) (GatewaySessionRow, bool, error) {
	spendTxID = strings.TrimSpace(spendTxID)
	if spendTxID == "" {
		return GatewaySessionRow{}, false, fmt.Errorf("spend_txid required")
	}
	var row GatewaySessionRow
	err := db.QueryRow(
		`SELECT spend_txid,client_id,client_bsv_pubkey_hex,server_bsv_pubkey_hex,input_amount_satoshi,pool_amount_satoshi,spend_tx_fee_satoshi,sequence_num,server_amount_satoshi,client_amount_satoshi,base_txid,final_txid,base_tx_hex,current_tx_hex,status,created_at_unix,updated_at_unix
		 FROM fee_pool_sessions WHERE spend_txid=?`, spendTxID,
	).Scan(
		&row.SpendTxID,
		&row.ClientID, &row.ClientBSVCompressedPubHex, &row.ServerBSVCompressedPubHex,
		&row.InputAmountSat, &row.PoolAmountSat, &row.SpendTxFeeSat,
		&row.Sequence, &row.ServerAmountSat, &row.ClientAmountSat,
		&row.BaseTxID, &row.FinalTxID,
		&row.BaseTxHex, &row.CurrentTxHex,
		&row.Status, &row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return GatewaySessionRow{}, false, nil
		}
		return GatewaySessionRow{}, false, err
	}
	return row, true, nil
}

func InsertSession(db *sql.DB, row GatewaySessionRow) error {
	now := time.Now().Unix()
	row.CreatedAt = now
	row.UpdatedAt = now
	_, err := db.Exec(
		`INSERT INTO fee_pool_sessions(spend_txid,client_id,client_bsv_pubkey_hex,server_bsv_pubkey_hex,input_amount_satoshi,pool_amount_satoshi,spend_tx_fee_satoshi,sequence_num,server_amount_satoshi,client_amount_satoshi,base_txid,final_txid,base_tx_hex,current_tx_hex,status,created_at_unix,updated_at_unix)
		 VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		row.SpendTxID,
		strings.ToLower(strings.TrimSpace(row.ClientID)),
		strings.ToLower(strings.TrimSpace(row.ClientBSVCompressedPubHex)),
		strings.ToLower(strings.TrimSpace(row.ServerBSVCompressedPubHex)),
		row.InputAmountSat, row.PoolAmountSat, row.SpendTxFeeSat,
		row.Sequence, row.ServerAmountSat, row.ClientAmountSat,
		strings.TrimSpace(row.BaseTxID), strings.TrimSpace(row.FinalTxID),
		row.BaseTxHex, row.CurrentTxHex,
		strings.TrimSpace(row.Status),
		row.CreatedAt, row.UpdatedAt,
	)
	return err
}

func UpdateSession(db *sql.DB, row GatewaySessionRow) error {
	row.UpdatedAt = time.Now().Unix()
	_, err := db.Exec(
		`UPDATE fee_pool_sessions
		 SET sequence_num=?,server_amount_satoshi=?,client_amount_satoshi=?,base_txid=?,final_txid=?,base_tx_hex=?,current_tx_hex=?,status=?,updated_at_unix=?
		 WHERE spend_txid=?`,
		row.Sequence, row.ServerAmountSat, row.ClientAmountSat,
		strings.TrimSpace(row.BaseTxID), strings.TrimSpace(row.FinalTxID),
		row.BaseTxHex, row.CurrentTxHex,
		strings.TrimSpace(row.Status), row.UpdatedAt,
		strings.TrimSpace(row.SpendTxID),
	)
	return err
}

func ListActiveSessions(db *sql.DB) ([]GatewaySessionRow, error) {
	if db == nil {
		return nil, fmt.Errorf("db is nil")
	}
	rows, err := db.Query(
		`SELECT spend_txid,client_id,client_bsv_pubkey_hex,server_bsv_pubkey_hex,input_amount_satoshi,pool_amount_satoshi,spend_tx_fee_satoshi,sequence_num,server_amount_satoshi,client_amount_satoshi,base_txid,final_txid,base_tx_hex,current_tx_hex,status,created_at_unix,updated_at_unix
		 FROM fee_pool_sessions
		 WHERE status='active'`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]GatewaySessionRow, 0, 16)
	for rows.Next() {
		var row GatewaySessionRow
		if err := rows.Scan(
			&row.SpendTxID,
			&row.ClientID, &row.ClientBSVCompressedPubHex, &row.ServerBSVCompressedPubHex,
			&row.InputAmountSat, &row.PoolAmountSat, &row.SpendTxFeeSat,
			&row.Sequence, &row.ServerAmountSat, &row.ClientAmountSat,
			&row.BaseTxID, &row.FinalTxID,
			&row.BaseTxHex, &row.CurrentTxHex,
			&row.Status, &row.CreatedAt, &row.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
