package dual2of2

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
)

// MigrateGatewayClientIDs 在启动阶段做 client_id 归一化迁移：
// - 历史 marshal client_id -> canonical compressed client_id
// - 合并 presence/service_state 的主键冲突行
// - 其余表统一更新 client_id 外键
func MigrateGatewayClientIDs(db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	mappings, err := collectClientIDMappingsTx(tx)
	if err != nil {
		return err
	}
	keys := make([]string, 0, len(mappings))
	for legacy := range mappings {
		keys = append(keys, legacy)
	}
	sort.Strings(keys)
	for _, legacy := range keys {
		canonical := mappings[legacy]
		if legacy == canonical || legacy == "" || canonical == "" {
			continue
		}
		if err = migrateOneClientIDTx(tx, legacy, canonical); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func collectClientIDMappingsTx(tx *sql.Tx) (map[string]string, error) {
	rows, err := tx.Query(
		`SELECT client_id FROM fee_pool_sessions
		 UNION
		 SELECT client_id FROM fee_pool_charge_events
		 UNION
		 SELECT client_id FROM fee_pool_client_presence
		 UNION
		 SELECT client_id FROM fee_pool_service_state
		 UNION
		 SELECT client_id FROM fee_pool_service_spans
		 UNION
		 SELECT client_id FROM fee_pool_service_pauses
		 UNION
		 SELECT client_id FROM fee_pool_service_interruptions
		 UNION
		 SELECT client_id FROM fee_pool_service_events`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]string, 16)
	for rows.Next() {
		var clientID string
		if err := rows.Scan(&clientID); err != nil {
			return nil, err
		}
		legacy := strings.ToLower(strings.TrimSpace(clientID))
		if legacy == "" {
			continue
		}
		canonical, nErr := NormalizeClientIDStrict(legacy)
		if nErr != nil || canonical == legacy {
			continue
		}
		out[legacy] = canonical
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// 消除映射链：A->B, B->C 时最终统一为 A->C。
	for legacy, canonical := range out {
		cur := canonical
		seen := map[string]struct{}{legacy: {}}
		for {
			next, ok := out[cur]
			if !ok || next == "" {
				break
			}
			if _, dup := seen[cur]; dup {
				break
			}
			seen[cur] = struct{}{}
			cur = next
		}
		out[legacy] = cur
	}
	return out, nil
}

func migrateOneClientIDTx(tx *sql.Tx, legacy string, canonical string) error {
	if err := mergePresenceRowsTx(tx, legacy, canonical); err != nil {
		return err
	}
	if err := mergeServiceStateRowsTx(tx, legacy, canonical); err != nil {
		return err
	}
	tables := []string{
		"fee_pool_sessions",
		"fee_pool_charge_events",
		"fee_pool_service_spans",
		"fee_pool_service_pauses",
		"fee_pool_service_interruptions",
		"fee_pool_service_events",
	}
	for _, table := range tables {
		if _, err := tx.Exec(`UPDATE `+table+` SET client_id=? WHERE client_id=?`, canonical, legacy); err != nil {
			return err
		}
	}
	return nil
}

type presenceRow struct {
	PeerID      string
	Online      int
	LastOnline  int64
	LastOffline int64
	UpdatedAt   int64
}

func loadPresenceRowTx(tx *sql.Tx, clientID string) (presenceRow, bool, error) {
	var row presenceRow
	err := tx.QueryRow(
		`SELECT peer_id,online,last_online_at_unix,last_offline_at_unix,updated_at_unix
		 FROM fee_pool_client_presence WHERE client_id=?`,
		clientID,
	).Scan(&row.PeerID, &row.Online, &row.LastOnline, &row.LastOffline, &row.UpdatedAt)
	if err == sql.ErrNoRows {
		return presenceRow{}, false, nil
	}
	if err != nil {
		return presenceRow{}, false, err
	}
	return row, true, nil
}

func mergePresenceRowsTx(tx *sql.Tx, legacy string, canonical string) error {
	legacyRow, hasLegacy, err := loadPresenceRowTx(tx, legacy)
	if err != nil || !hasLegacy {
		return err
	}
	canonRow, hasCanon, err := loadPresenceRowTx(tx, canonical)
	if err != nil {
		return err
	}
	if !hasCanon {
		_, err = tx.Exec(`UPDATE fee_pool_client_presence SET client_id=? WHERE client_id=?`, canonical, legacy)
		return err
	}
	mergedPeerID := strings.TrimSpace(canonRow.PeerID)
	if mergedPeerID == "" {
		mergedPeerID = strings.TrimSpace(legacyRow.PeerID)
	}
	mergedOnline := 0
	if canonRow.Online == 1 || legacyRow.Online == 1 {
		mergedOnline = 1
	}
	_, err = tx.Exec(
		`UPDATE fee_pool_client_presence
		 SET peer_id=?,online=?,last_online_at_unix=?,last_offline_at_unix=?,updated_at_unix=?
		 WHERE client_id=?`,
		mergedPeerID,
		mergedOnline,
		maxInt64(canonRow.LastOnline, legacyRow.LastOnline),
		maxInt64(canonRow.LastOffline, legacyRow.LastOffline),
		maxInt64(canonRow.UpdatedAt, legacyRow.UpdatedAt),
		canonical,
	)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`DELETE FROM fee_pool_client_presence WHERE client_id=?`, legacy)
	return err
}

type serviceStateMergeRow struct {
	State              string
	Online             int
	CoverageActive     int
	ActiveSessionCount int
	CurrentSpanID      int64
	CurrentPauseID     int64
	CurrentInterruptID int64
	LastTransitionAt   int64
	UpdatedAt          int64
}

func loadServiceStateRowTx(tx *sql.Tx, clientID string) (serviceStateMergeRow, bool, error) {
	var row serviceStateMergeRow
	err := tx.QueryRow(
		`SELECT state,online,coverage_active,active_session_count,current_span_id,current_pause_id,current_interrupt_id,last_transition_at_unix,updated_at_unix
		 FROM fee_pool_service_state WHERE client_id=?`,
		clientID,
	).Scan(
		&row.State, &row.Online, &row.CoverageActive, &row.ActiveSessionCount,
		&row.CurrentSpanID, &row.CurrentPauseID, &row.CurrentInterruptID,
		&row.LastTransitionAt, &row.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return serviceStateMergeRow{}, false, nil
	}
	if err != nil {
		return serviceStateMergeRow{}, false, err
	}
	return row, true, nil
}

func mergeServiceStateRowsTx(tx *sql.Tx, legacy string, canonical string) error {
	legacyRow, hasLegacy, err := loadServiceStateRowTx(tx, legacy)
	if err != nil || !hasLegacy {
		return err
	}
	canonRow, hasCanon, err := loadServiceStateRowTx(tx, canonical)
	if err != nil {
		return err
	}
	if !hasCanon {
		_, err = tx.Exec(`UPDATE fee_pool_service_state SET client_id=? WHERE client_id=?`, canonical, legacy)
		return err
	}

	merged := canonRow
	if legacyRow.UpdatedAt > canonRow.UpdatedAt {
		merged.State = legacyRow.State
	}
	if canonRow.Online == 1 || legacyRow.Online == 1 {
		merged.Online = 1
	} else {
		merged.Online = 0
	}
	if canonRow.CoverageActive == 1 || legacyRow.CoverageActive == 1 {
		merged.CoverageActive = 1
	} else {
		merged.CoverageActive = 0
	}
	if legacyRow.ActiveSessionCount > merged.ActiveSessionCount {
		merged.ActiveSessionCount = legacyRow.ActiveSessionCount
	}
	merged.CurrentSpanID = maxInt64(canonRow.CurrentSpanID, legacyRow.CurrentSpanID)
	merged.CurrentPauseID = maxInt64(canonRow.CurrentPauseID, legacyRow.CurrentPauseID)
	merged.CurrentInterruptID = maxInt64(canonRow.CurrentInterruptID, legacyRow.CurrentInterruptID)
	merged.LastTransitionAt = maxInt64(canonRow.LastTransitionAt, legacyRow.LastTransitionAt)
	merged.UpdatedAt = maxInt64(canonRow.UpdatedAt, legacyRow.UpdatedAt)

	_, err = tx.Exec(
		`UPDATE fee_pool_service_state
		 SET state=?,online=?,coverage_active=?,active_session_count=?,current_span_id=?,current_pause_id=?,current_interrupt_id=?,last_transition_at_unix=?,updated_at_unix=?
		 WHERE client_id=?`,
		strings.TrimSpace(merged.State),
		merged.Online,
		merged.CoverageActive,
		merged.ActiveSessionCount,
		merged.CurrentSpanID,
		merged.CurrentPauseID,
		merged.CurrentInterruptID,
		merged.LastTransitionAt,
		merged.UpdatedAt,
		canonical,
	)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`DELETE FROM fee_pool_service_state WHERE client_id=?`, legacy)
	return err
}

func maxInt64(a int64, b int64) int64 {
	if a >= b {
		return a
	}
	return b
}
