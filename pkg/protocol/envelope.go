package protocol

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
)

type SignedEnvelope struct {
	Version      uint32          `json:"version"`
	Domain       string          `json:"domain"`
	Network      string          `json:"network"`
	MsgType      string          `json:"msg_type"`
	MsgID        string          `json:"msg_id"`
	Timestamp    int64           `json:"timestamp"`
	ExpireAt     int64           `json:"expire_at"`
	SenderPubkey string          `json:"sender_pubkey"`
	Payload      json.RawMessage `json:"payload"`
	SigAlg       string          `json:"sig_alg"`
	Signature    string          `json:"signature"`
}

type unsignedEnvelope struct {
	Version      uint32          `json:"version"`
	Domain       string          `json:"domain"`
	Network      string          `json:"network"`
	MsgType      string          `json:"msg_type"`
	MsgID        string          `json:"msg_id"`
	Timestamp    int64           `json:"timestamp"`
	ExpireAt     int64           `json:"expire_at"`
	SenderPubkey string          `json:"sender_pubkey"`
	Payload      json.RawMessage `json:"payload"`
	SigAlg       string          `json:"sig_alg"`
}

func NewSignedEnvelope(priv crypto.PrivKey, domain, network, msgType, msgID string, ttl time.Duration, payload json.RawMessage) (*SignedEnvelope, error) {
	pub := priv.GetPublic()
	pubBytes, err := crypto.MarshalPublicKey(pub)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	e := &SignedEnvelope{
		Version:      1,
		Domain:       domain,
		Network:      network,
		MsgType:      msgType,
		MsgID:        msgID,
		Timestamp:    now,
		ExpireAt:     now + int64(ttl.Seconds()),
		SenderPubkey: hex.EncodeToString(pubBytes),
		Payload:      payload,
		SigAlg:       "ed25519",
	}
	b, err := signingBytes(e)
	if err != nil {
		return nil, err
	}
	sig, err := priv.Sign(b)
	if err != nil {
		return nil, err
	}
	e.Signature = hex.EncodeToString(sig)
	return e, nil
}

func VerifyEnvelope(e *SignedEnvelope, now time.Time) error {
	if e.Version != 1 {
		return fmt.Errorf("invalid version")
	}
	if e.MsgID == "" || e.MsgType == "" || e.Domain == "" || e.Network == "" {
		return fmt.Errorf("missing envelope fields")
	}
	if e.SigAlg != "ed25519" {
		return fmt.Errorf("unsupported sig_alg")
	}
	if e.ExpireAt <= now.Unix() {
		return fmt.Errorf("envelope expired")
	}
	pubBytes, err := hex.DecodeString(e.SenderPubkey)
	if err != nil {
		return fmt.Errorf("decode sender_pubkey: %w", err)
	}
	pub, err := crypto.UnmarshalPublicKey(pubBytes)
	if err != nil {
		return fmt.Errorf("unmarshal sender_pubkey: %w", err)
	}
	sig, err := hex.DecodeString(e.Signature)
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}
	b, err := signingBytes(e)
	if err != nil {
		return err
	}
	ok, err := pub.Verify(b, sig)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("signature invalid")
	}
	return nil
}

func EnvelopePayloadHash(payload json.RawMessage) string {
	s := sha256.Sum256(payload)
	return hex.EncodeToString(s[:])
}

func signingBytes(e *SignedEnvelope) ([]byte, error) {
	u := unsignedEnvelope{
		Version:      e.Version,
		Domain:       e.Domain,
		Network:      e.Network,
		MsgType:      e.MsgType,
		MsgID:        e.MsgID,
		Timestamp:    e.Timestamp,
		ExpireAt:     e.ExpireAt,
		SenderPubkey: e.SenderPubkey,
		Payload:      e.Payload,
		SigAlg:       e.SigAlg,
	}
	return json.Marshal(u)
}
