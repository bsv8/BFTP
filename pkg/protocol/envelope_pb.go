package protocol

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	oldproto "github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p/core/crypto"
)

// SignedEnvelopePB 是 p2prpc 的二进制签名信封（protobuf）。
// 设计说明：
// - 外层信封统一 protobuf 编码，避免 JSON/hex 的冗余开销；
// - sender_pubkey/signature 直接使用 bytes，业务层不再承担 hex 编解码成本。
type SignedEnvelopePB struct {
	Version      uint32 `protobuf:"varint,1,opt,name=version,proto3" json:"version,omitempty"`
	Domain       string `protobuf:"bytes,2,opt,name=domain,proto3" json:"domain,omitempty"`
	Network      string `protobuf:"bytes,3,opt,name=network,proto3" json:"network,omitempty"`
	MsgType      string `protobuf:"bytes,4,opt,name=msg_type,json=msgType,proto3" json:"msg_type,omitempty"`
	MsgID        string `protobuf:"bytes,5,opt,name=msg_id,json=msgId,proto3" json:"msg_id,omitempty"`
	Timestamp    int64  `protobuf:"varint,6,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	ExpireAt     int64  `protobuf:"varint,7,opt,name=expire_at,json=expireAt,proto3" json:"expire_at,omitempty"`
	SenderPubkey []byte `protobuf:"bytes,8,opt,name=sender_pubkey,json=senderPubkey,proto3" json:"sender_pubkey,omitempty"`
	Payload      []byte `protobuf:"bytes,9,opt,name=payload,proto3" json:"payload,omitempty"`
	SigAlg       string `protobuf:"bytes,10,opt,name=sig_alg,json=sigAlg,proto3" json:"sig_alg,omitempty"`
	Signature    []byte `protobuf:"bytes,11,opt,name=signature,proto3" json:"signature,omitempty"`
}

type unsignedEnvelopePB struct {
	Version      uint32 `protobuf:"varint,1,opt,name=version,proto3" json:"version,omitempty"`
	Domain       string `protobuf:"bytes,2,opt,name=domain,proto3" json:"domain,omitempty"`
	Network      string `protobuf:"bytes,3,opt,name=network,proto3" json:"network,omitempty"`
	MsgType      string `protobuf:"bytes,4,opt,name=msg_type,json=msgType,proto3" json:"msg_type,omitempty"`
	MsgID        string `protobuf:"bytes,5,opt,name=msg_id,json=msgId,proto3" json:"msg_id,omitempty"`
	Timestamp    int64  `protobuf:"varint,6,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	ExpireAt     int64  `protobuf:"varint,7,opt,name=expire_at,json=expireAt,proto3" json:"expire_at,omitempty"`
	SenderPubkey []byte `protobuf:"bytes,8,opt,name=sender_pubkey,json=senderPubkey,proto3" json:"sender_pubkey,omitempty"`
	Payload      []byte `protobuf:"bytes,9,opt,name=payload,proto3" json:"payload,omitempty"`
	SigAlg       string `protobuf:"bytes,10,opt,name=sig_alg,json=sigAlg,proto3" json:"sig_alg,omitempty"`
}

func (m *SignedEnvelopePB) Reset()         { *m = SignedEnvelopePB{} }
func (m *SignedEnvelopePB) String() string { return oldproto.CompactTextString(m) }
func (*SignedEnvelopePB) ProtoMessage()    {}

func (m *unsignedEnvelopePB) Reset()         { *m = unsignedEnvelopePB{} }
func (m *unsignedEnvelopePB) String() string { return oldproto.CompactTextString(m) }
func (*unsignedEnvelopePB) ProtoMessage()    {}

func NewSignedEnvelopePB(priv crypto.PrivKey, domain, network, msgType, msgID string, ttl time.Duration, payload []byte) (*SignedEnvelopePB, error) {
	pub := priv.GetPublic()
	pubBytes, err := crypto.MarshalPublicKey(pub)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	e := &SignedEnvelopePB{
		Version:      1,
		Domain:       domain,
		Network:      network,
		MsgType:      msgType,
		MsgID:        msgID,
		Timestamp:    now,
		ExpireAt:     now + int64(ttl.Seconds()),
		SenderPubkey: append([]byte(nil), pubBytes...),
		Payload:      append([]byte(nil), payload...),
		SigAlg:       "ed25519",
	}
	b, err := signingBytesPB(e)
	if err != nil {
		return nil, err
	}
	sig, err := priv.Sign(b)
	if err != nil {
		return nil, err
	}
	e.Signature = append([]byte(nil), sig...)
	return e, nil
}

func VerifyEnvelopePB(e *SignedEnvelopePB, now time.Time) error {
	if e == nil {
		return fmt.Errorf("envelope is nil")
	}
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
	if len(e.SenderPubkey) == 0 {
		return fmt.Errorf("sender_pubkey required")
	}
	if len(e.Signature) == 0 {
		return fmt.Errorf("signature required")
	}
	if err := VerifyEnvelopePBSignature(e); err != nil {
		return err
	}
	return nil
}

// VerifyEnvelopePBSignature 仅校验签名与必要字段，不校验过期时间。
// 解析工具复盘历史抓包时需要这条能力，避免历史消息天然过期后无法核查签名。
func VerifyEnvelopePBSignature(e *SignedEnvelopePB) error {
	if e == nil {
		return fmt.Errorf("envelope is nil")
	}
	if e.Version != 1 {
		return fmt.Errorf("invalid version")
	}
	if e.MsgID == "" || e.MsgType == "" || e.Domain == "" || e.Network == "" {
		return fmt.Errorf("missing envelope fields")
	}
	if e.SigAlg != "ed25519" {
		return fmt.Errorf("unsupported sig_alg")
	}
	if len(e.SenderPubkey) == 0 {
		return fmt.Errorf("sender_pubkey required")
	}
	if len(e.Signature) == 0 {
		return fmt.Errorf("signature required")
	}
	pub, err := crypto.UnmarshalPublicKey(e.SenderPubkey)
	if err != nil {
		return fmt.Errorf("unmarshal sender_pubkey: %w", err)
	}
	b, err := signingBytesPB(e)
	if err != nil {
		return err
	}
	ok, err := pub.Verify(b, e.Signature)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("signature invalid")
	}
	return nil
}

func EnvelopePayloadHashBytes(payload []byte) string {
	s := sha256.Sum256(payload)
	return hex.EncodeToString(s[:])
}

func signingBytesPB(e *SignedEnvelopePB) ([]byte, error) {
	u := &unsignedEnvelopePB{
		Version:      e.Version,
		Domain:       e.Domain,
		Network:      e.Network,
		MsgType:      e.MsgType,
		MsgID:        e.MsgID,
		Timestamp:    e.Timestamp,
		ExpireAt:     e.ExpireAt,
		SenderPubkey: append([]byte(nil), e.SenderPubkey...),
		Payload:      append([]byte(nil), e.Payload...),
		SigAlg:       e.SigAlg,
	}
	buf := oldproto.NewBuffer(nil)
	buf.SetDeterministic(true)
	if err := buf.Marshal(u); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
