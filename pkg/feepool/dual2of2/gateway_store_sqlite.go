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

type ListenerTargetRow struct {
	ClientID string
	PeerID   string
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
			billing_cycle_seconds INTEGER NOT NULL DEFAULT 0,
			effective_until_unix INTEGER NOT NULL DEFAULT 0,
			pool_balance_after_satoshi INTEGER NOT NULL DEFAULT 0,
			created_at_unix INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_fee_pool_charge_client_reason ON fee_pool_charge_events(client_id, charge_reason)`,
		`CREATE TABLE IF NOT EXISTS fee_pool_client_presence (
			client_id TEXT PRIMARY KEY,
			peer_id TEXT NOT NULL,
			online INTEGER NOT NULL,
			last_online_at_unix INTEGER NOT NULL,
			last_offline_at_unix INTEGER NOT NULL,
			updated_at_unix INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_fee_pool_client_presence_online ON fee_pool_client_presence(online, updated_at_unix DESC)`,
		`CREATE TABLE IF NOT EXISTS fee_pool_service_state (
			client_id TEXT PRIMARY KEY,
			state TEXT NOT NULL,
			online INTEGER NOT NULL,
			coverage_active INTEGER NOT NULL,
			active_session_count INTEGER NOT NULL,
			current_span_id INTEGER NOT NULL,
			current_pause_id INTEGER NOT NULL,
			current_interrupt_id INTEGER NOT NULL,
			last_transition_at_unix INTEGER NOT NULL,
			updated_at_unix INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_fee_pool_service_state_state ON fee_pool_service_state(state, updated_at_unix DESC)`,
		`CREATE TABLE IF NOT EXISTS fee_pool_service_spans (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			client_id TEXT NOT NULL,
			start_at_unix INTEGER NOT NULL,
			end_at_unix INTEGER NOT NULL,
			status TEXT NOT NULL,
			start_spend_txid TEXT NOT NULL,
			end_spend_txid TEXT NOT NULL,
			end_reason TEXT NOT NULL,
			created_at_unix INTEGER NOT NULL,
			updated_at_unix INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_fee_pool_service_spans_client ON fee_pool_service_spans(client_id, id DESC)`,
		`CREATE TABLE IF NOT EXISTS fee_pool_service_pauses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			client_id TEXT NOT NULL,
			span_id INTEGER NOT NULL,
			start_at_unix INTEGER NOT NULL,
			end_at_unix INTEGER NOT NULL,
			status TEXT NOT NULL,
			reason TEXT NOT NULL,
			created_at_unix INTEGER NOT NULL,
			updated_at_unix INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_fee_pool_service_pauses_client ON fee_pool_service_pauses(client_id, id DESC)`,
		`CREATE TABLE IF NOT EXISTS fee_pool_service_interruptions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			client_id TEXT NOT NULL,
			start_at_unix INTEGER NOT NULL,
			end_at_unix INTEGER NOT NULL,
			status TEXT NOT NULL,
			reason TEXT NOT NULL,
			created_at_unix INTEGER NOT NULL,
			updated_at_unix INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_fee_pool_service_interruptions_client ON fee_pool_service_interruptions(client_id, id DESC)`,
		`CREATE TABLE IF NOT EXISTS fee_pool_service_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			client_id TEXT NOT NULL,
			event_name TEXT NOT NULL,
			prev_state TEXT NOT NULL,
			next_state TEXT NOT NULL,
			reason TEXT NOT NULL,
			spend_txid TEXT NOT NULL,
			peer_id TEXT NOT NULL,
			created_at_unix INTEGER NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_fee_pool_service_events_client ON fee_pool_service_events(client_id, id DESC)`,
		`CREATE TABLE IF NOT EXISTS gateway_admin_config (
			config_key TEXT PRIMARY KEY,
			config_value TEXT NOT NULL,
			updated_at_unix INTEGER NOT NULL
		)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	// 旧库升级：历史版本的 charge_events 不包含计费补充字段，这里自动补齐列定义。
	if err := ensureChargeEventColumns(db); err != nil {
		return err
	}
	return nil
}

func ensureChargeEventColumns(db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}
	stmts := []string{
		`ALTER TABLE fee_pool_charge_events ADD COLUMN billing_cycle_seconds INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE fee_pool_charge_events ADD COLUMN effective_until_unix INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE fee_pool_charge_events ADD COLUMN pool_balance_after_satoshi INTEGER NOT NULL DEFAULT 0`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			lower := strings.ToLower(strings.TrimSpace(err.Error()))
			if strings.Contains(lower, "duplicate column name") {
				continue
			}
			return err
		}
	}
	return nil
}

func SetGatewayAdminConfig(db *sql.DB, key string, value string) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf("config key is required")
	}
	_, err := db.Exec(
		`INSERT INTO gateway_admin_config(config_key,config_value,updated_at_unix)
		 VALUES(?,?,?)
		 ON CONFLICT(config_key) DO UPDATE SET
			config_value=excluded.config_value,
			updated_at_unix=excluded.updated_at_unix`,
		key, value, time.Now().Unix(),
	)
	return err
}

func GetGatewayAdminConfig(db *sql.DB, key string) (string, bool, error) {
	if db == nil {
		return "", false, fmt.Errorf("db is nil")
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return "", false, fmt.Errorf("config key is required")
	}
	var value string
	err := db.QueryRow(`SELECT config_value FROM gateway_admin_config WHERE config_key=?`, key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", false, nil
		}
		return "", false, err
	}
	return value, true, nil
}

func UpsertClientPresence(db *sql.DB, clientID string, peerID string, online bool) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}
	clientID = NormalizeClientIDLoose(clientID)
	peerID = strings.TrimSpace(peerID)
	if clientID == "" {
		return fmt.Errorf("client_id required")
	}
	now := time.Now().Unix()
	onlineInt := 0
	lastOnline := int64(0)
	lastOffline := int64(0)
	if online {
		onlineInt = 1
		lastOnline = now
	} else {
		lastOffline = now
	}
	_, err := db.Exec(
		`INSERT INTO fee_pool_client_presence(client_id,peer_id,online,last_online_at_unix,last_offline_at_unix,updated_at_unix)
		 VALUES(?,?,?,?,?,?)
		 ON CONFLICT(client_id) DO UPDATE SET
			peer_id=excluded.peer_id,
			online=excluded.online,
			last_online_at_unix=CASE WHEN excluded.last_online_at_unix>0 THEN excluded.last_online_at_unix ELSE fee_pool_client_presence.last_online_at_unix END,
			last_offline_at_unix=CASE WHEN excluded.last_offline_at_unix>0 THEN excluded.last_offline_at_unix ELSE fee_pool_client_presence.last_offline_at_unix END,
			updated_at_unix=excluded.updated_at_unix`,
		clientID, peerID, onlineInt, lastOnline, lastOffline, now,
	)
	return err
}

func ListServingListenerTargetsByChargeReason(db *sql.DB, reason string) ([]ListenerTargetRow, error) {
	if db == nil {
		return nil, fmt.Errorf("db is nil")
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil, fmt.Errorf("reason is required")
	}
	rows, err := db.Query(
		`SELECT DISTINCT s.client_id, p.peer_id
		 FROM fee_pool_service_state s
		 JOIN fee_pool_client_presence p ON p.client_id=s.client_id
		 JOIN fee_pool_charge_events c ON c.client_id=s.client_id
		 WHERE c.charge_reason=? AND s.state='serving' AND p.online=1`,
		reason,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]ListenerTargetRow, 0, 16)
	for rows.Next() {
		var row ListenerTargetRow
		if err := rows.Scan(&row.ClientID, &row.PeerID); err != nil {
			return nil, err
		}
		row.ClientID = strings.ToLower(strings.TrimSpace(row.ClientID))
		row.PeerID = strings.TrimSpace(row.PeerID)
		if row.ClientID == "" || row.PeerID == "" {
			continue
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func InsertChargeEvent(db *sql.DB, clientID string, spendTxID string, sequence uint32, reason string, amount uint64, billingCycleSeconds uint32, effectiveUntilUnix int64, poolBalanceAfterSatoshi uint64) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}
	clientID = NormalizeClientIDLoose(clientID)
	spendTxID = strings.TrimSpace(spendTxID)
	reason = strings.TrimSpace(reason)
	if clientID == "" || spendTxID == "" || reason == "" {
		return fmt.Errorf("invalid charge event")
	}
	if effectiveUntilUnix <= 0 {
		return fmt.Errorf("effective_until_unix must be positive")
	}
	_, err := db.Exec(
		`INSERT INTO fee_pool_charge_events(client_id,spend_txid,sequence_num,charge_reason,charge_amount_satoshi,billing_cycle_seconds,effective_until_unix,pool_balance_after_satoshi,created_at_unix) VALUES(?,?,?,?,?,?,?,?,?)`,
		clientID, spendTxID, sequence, reason, amount, billingCycleSeconds, effectiveUntilUnix, poolBalanceAfterSatoshi, time.Now().Unix(),
	)
	return err
}

func CountChargeEventsByClientAndReason(db *sql.DB, clientID string, reason string) (int, error) {
	if db == nil {
		return 0, fmt.Errorf("db is nil")
	}
	clientID = NormalizeClientIDLoose(clientID)
	reason = strings.TrimSpace(reason)
	if clientID == "" || reason == "" {
		return 0, fmt.Errorf("client_id and reason are required")
	}
	aliases := ClientIDAliasesForQuery(clientID)
	if len(aliases) == 0 {
		return 0, fmt.Errorf("client_id and reason are required")
	}
	var n int
	if len(aliases) == 1 {
		if err := db.QueryRow(
			`SELECT COUNT(*) FROM fee_pool_charge_events WHERE client_id=? AND charge_reason=?`,
			aliases[0], reason,
		).Scan(&n); err != nil {
			return 0, err
		}
		return n, nil
	}
	if err := db.QueryRow(
		`SELECT COUNT(*) FROM fee_pool_charge_events WHERE client_id IN (?,?) AND charge_reason=?`,
		aliases[0], aliases[1], reason,
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
	clientID = NormalizeClientIDLoose(clientID)
	if clientID == "" {
		return GatewaySessionRow{}, false, fmt.Errorf("client_id required")
	}
	aliases := ClientIDAliasesForQuery(clientID)
	if len(aliases) == 0 {
		return GatewaySessionRow{}, false, fmt.Errorf("client_id required")
	}
	var row GatewaySessionRow
	query := `SELECT spend_txid,client_id,client_bsv_pubkey_hex,server_bsv_pubkey_hex,input_amount_satoshi,pool_amount_satoshi,spend_tx_fee_satoshi,sequence_num,server_amount_satoshi,client_amount_satoshi,base_txid,final_txid,base_tx_hex,current_tx_hex,status,created_at_unix,updated_at_unix
		 FROM fee_pool_sessions
		 WHERE client_id=?
		 ORDER BY updated_at_unix DESC
		 LIMIT 1`
	args := []any{aliases[0]}
	if len(aliases) > 1 {
		query = strings.ReplaceAll(query, "client_id=?", "client_id IN (?,?)")
		args = []any{aliases[0], aliases[1]}
	}
	err := db.QueryRow(query, args...).Scan(
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
	row.ClientID = NormalizeClientIDLoose(row.ClientID)
	if row.ClientID == "" {
		return fmt.Errorf("client_id required")
	}
	_, err := db.Exec(
		`INSERT INTO fee_pool_sessions(spend_txid,client_id,client_bsv_pubkey_hex,server_bsv_pubkey_hex,input_amount_satoshi,pool_amount_satoshi,spend_tx_fee_satoshi,sequence_num,server_amount_satoshi,client_amount_satoshi,base_txid,final_txid,base_tx_hex,current_tx_hex,status,created_at_unix,updated_at_unix)
		 VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		row.SpendTxID,
		row.ClientID,
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
