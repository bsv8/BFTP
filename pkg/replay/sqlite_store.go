package replay

import (
	"database/sql"
	"fmt"
	"time"
)

type Entry struct {
	PayloadHash string
	Response    []byte
	ExpireAt    int64
}

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(db *sql.DB) *SQLiteStore { return &SQLiteStore{db: db} }

func InitDB(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS replay_cache (
		scope TEXT NOT NULL,
		msg_id TEXT NOT NULL,
		payload_hash TEXT NOT NULL,
		response_json BLOB NOT NULL,
		expire_at INTEGER NOT NULL,
		created_at INTEGER NOT NULL,
		PRIMARY KEY (scope, msg_id)
	)`)
	return err
}

func (s *SQLiteStore) Get(scope, msgID string) (Entry, bool, error) {
	var e Entry
	err := s.db.QueryRow(`SELECT payload_hash,response_json,expire_at FROM replay_cache WHERE scope=? AND msg_id=?`, scope, msgID).
		Scan(&e.PayloadHash, &e.Response, &e.ExpireAt)
	if err == sql.ErrNoRows {
		return Entry{}, false, nil
	}
	if err != nil {
		return Entry{}, false, err
	}
	if e.ExpireAt <= time.Now().Unix() {
		_, _ = s.db.Exec(`DELETE FROM replay_cache WHERE scope=? AND msg_id=?`, scope, msgID)
		return Entry{}, false, nil
	}
	return e, true, nil
}

func (s *SQLiteStore) Put(scope, msgID, payloadHash string, response []byte, expireAt int64) error {
	if msgID == "" || scope == "" || payloadHash == "" {
		return fmt.Errorf("invalid replay put")
	}
	_, err := s.db.Exec(`INSERT INTO replay_cache(scope,msg_id,payload_hash,response_json,expire_at,created_at)
	VALUES(?,?,?,?,?,?)
	ON CONFLICT(scope,msg_id) DO UPDATE SET payload_hash=excluded.payload_hash,response_json=excluded.response_json,expire_at=excluded.expire_at`,
		scope, msgID, payloadHash, response, expireAt, time.Now().Unix())
	return err
}
