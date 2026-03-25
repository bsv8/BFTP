package p2prpc

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

type ReplayStore interface {
	Get(scope, msgID string) (entry ReplayEntry, found bool, err error)
	Put(scope, msgID, payloadHash string, response []byte, expireAt int64) error
}

type ReplayEntry struct {
	PayloadHash string
	Response    []byte
	ExpireAt    int64
}

type SecurityConfig struct {
	Domain  string
	Network string
	TTL     time.Duration
	Replay  ReplayStore
	Trace   TraceSink
}

type contextKey string

const senderPubkeyHexContextKey contextKey = "p2prpc_sender_pubkey_hex"
const messageIDContextKey contextKey = "p2prpc_message_id"

func SenderPubkeyHexFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(senderPubkeyHexContextKey)
	s, ok := v.(string)
	return s, ok && s != ""
}

func MessageIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(messageIDContextKey)
	s, ok := v.(string)
	return s, ok && s != ""
}

func verifyRemotePeerBytes(remote peer.ID, senderPub []byte) error {
	pub, err := crypto.UnmarshalPublicKey(senderPub)
	if err != nil {
		return fmt.Errorf("invalid sender pubkey: %w", err)
	}
	expected, err := peer.IDFromPublicKey(pub)
	if err != nil {
		return err
	}
	if expected != remote {
		return fmt.Errorf("remote peer mismatch")
	}
	return nil
}

func senderPubkeyHex(senderPub []byte) string {
	pub, err := crypto.UnmarshalPublicKey(senderPub)
	if err != nil {
		return strings.ToLower(hex.EncodeToString(senderPub))
	}
	raw, err := pub.Raw()
	if err != nil {
		return strings.ToLower(hex.EncodeToString(senderPub))
	}
	return strings.ToLower(hex.EncodeToString(raw))
}

func randomMsgID() (string, error) {
	b := make([]byte, 16)
	if _, err := cryptorand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
