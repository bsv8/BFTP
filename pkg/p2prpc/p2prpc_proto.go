package p2prpc

import (
	"bufio"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	appprotocol "github.com/bsv8/BFTP/pkg/protocol"
	oldproto "github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	coreprotocol "github.com/libp2p/go-libp2p/core/protocol"
)

type ErrorResponsePB struct {
	// IsError 作为错误响应标记位，避免普通响应字段误判为错误。
	IsError bool `protobuf:"varint,1,opt,name=is_error,json=isError,proto3" json:"is_error,omitempty"`
	Error   string `protobuf:"bytes,127,opt,name=error,proto3" json:"error,omitempty"`
}

func (m *ErrorResponsePB) Reset()         { *m = ErrorResponsePB{} }
func (m *ErrorResponsePB) String() string { return oldproto.CompactTextString(m) }
func (*ErrorResponsePB) ProtoMessage()    {}

// legacyErrorResponsePB 兼容旧版错误响应（field #1 直接承载 error 字符串）。
type legacyErrorResponsePB struct {
	Error string `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
}

func (m *legacyErrorResponsePB) Reset()         { *m = legacyErrorResponsePB{} }
func (m *legacyErrorResponsePB) String() string { return oldproto.CompactTextString(m) }
func (*legacyErrorResponsePB) ProtoMessage()    {}

// HandleProto 使用 protobuf 二进制编解码处理 p2p RPC。
// 说明：TReq/TResp 需要在对应包内实现 protobuf Message（手写 tag + ProtoMessage）。
func HandleProto[TReq any, TResp any](h host.Host, protoID coreprotocol.ID, cfg SecurityConfig, fn func(context.Context, TReq) (TResp, error)) {
	h.SetStreamHandler(protoID, func(s network.Stream) {
		defer s.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		body, err := io.ReadAll(bufio.NewReader(s))
		if err != nil {
			writeErrProto(s, "decode envelope: "+err.Error())
			return
		}
		var env appprotocol.SignedEnvelopePB
		if err := oldproto.Unmarshal(body, &env); err != nil {
			writeErrProto(s, "decode envelope: "+err.Error())
			return
		}
		if err := appprotocol.VerifyEnvelopePB(&env, time.Now()); err != nil {
			writeErrProto(s, err.Error())
			return
		}
		if env.Domain != cfg.Domain || env.Network != cfg.Network {
			writeErrProto(s, "domain/network mismatch")
			return
		}
		if env.MsgType != string(protoID) {
			writeErrProto(s, "msg_type mismatch")
			return
		}
		if err := verifyRemotePeerBytes(s.Conn().RemotePeer(), env.SenderPubkey); err != nil {
			writeErrProto(s, err.Error())
			return
		}

		trace := cfg.Trace
		localPID := ""
		if h != nil {
			localPID = h.ID().String()
		}
		remotePID := ""
		if s != nil && s.Conn() != nil {
			remotePID = s.Conn().RemotePeer().String()
		}
		senderPubHex := strings.ToLower(hex.EncodeToString(env.SenderPubkey))
		if trace != nil {
			ctx = context.WithValue(ctx, senderPubkeyHexContextKey, senderPubHex)
		} else {
			ctx = context.WithValue(ctx, senderPubkeyHexContextKey, senderPubHex)
		}
		traceReq := normalizeProtoTracePayload(env.Payload)

		payloadHash := appprotocol.EnvelopePayloadHashBytes(env.Payload)
		if cfg.Replay != nil {
			entry, found, err := cfg.Replay.Get(string(protoID), env.MsgID)
			if err != nil {
				writeErrProto(s, err.Error())
				if trace != nil {
					trace.Handle(TraceEvent{
						TS:              nowRFC3339(),
						Direction:       "recv",
						Domain:          cfg.Domain,
						Network:         cfg.Network,
						ProtoID:         string(protoID),
						MsgID:           env.MsgID,
						LocalPeerID:     localPID,
						RemotePeerID:    remotePID,
						SenderPubkeyHex: senderPubHex,
						ExpireAt:        env.ExpireAt,
						Request:         traceReq,
						Error:           err.Error(),
					})
				}
				return
			}
			if found {
				if entry.PayloadHash != payloadHash {
					writeErrProto(s, "ERR_IDEMPOTENCY_REPLAY")
					if trace != nil {
						trace.Handle(TraceEvent{
							TS:              nowRFC3339(),
							Direction:       "recv",
							Domain:          cfg.Domain,
							Network:         cfg.Network,
							ProtoID:         string(protoID),
							MsgID:           env.MsgID,
							LocalPeerID:     localPID,
							RemotePeerID:    remotePID,
							SenderPubkeyHex: senderPubHex,
							ExpireAt:        env.ExpireAt,
							ReplayHit:       true,
							Request:         traceReq,
							Error:           "ERR_IDEMPOTENCY_REPLAY",
						})
					}
					return
				}
				_, _ = s.Write(entry.Response)
				if trace != nil {
					trace.Handle(TraceEvent{
						TS:              nowRFC3339(),
						Direction:       "recv",
						Domain:          cfg.Domain,
						Network:         cfg.Network,
						ProtoID:         string(protoID),
						MsgID:           env.MsgID,
						LocalPeerID:     localPID,
						RemotePeerID:    remotePID,
						SenderPubkeyHex: senderPubHex,
						ExpireAt:        env.ExpireAt,
						ReplayHit:       true,
						Request:         traceReq,
						Response:        normalizeProtoTracePayload(entry.Response),
					})
				}
				return
			}
		}

		req, err := decodeProtoValue[TReq](env.Payload)
		if err != nil {
			writeErrProto(s, "decode payload: "+err.Error())
			if trace != nil {
				trace.Handle(TraceEvent{
					TS:              nowRFC3339(),
					Direction:       "recv",
					Domain:          cfg.Domain,
					Network:         cfg.Network,
					ProtoID:         string(protoID),
					MsgID:           env.MsgID,
					LocalPeerID:     localPID,
					RemotePeerID:    remotePID,
					SenderPubkeyHex: senderPubHex,
					ExpireAt:        env.ExpireAt,
					Request:         traceReq,
					Error:           "decode payload: " + err.Error(),
				})
			}
			return
		}
		resp, err := fn(ctx, req)
		if err != nil {
			writeErrProto(s, err.Error())
			if trace != nil {
				trace.Handle(TraceEvent{
					TS:              nowRFC3339(),
					Direction:       "recv",
					Domain:          cfg.Domain,
					Network:         cfg.Network,
					ProtoID:         string(protoID),
					MsgID:           env.MsgID,
					LocalPeerID:     localPID,
					RemotePeerID:    remotePID,
					SenderPubkeyHex: senderPubHex,
					ExpireAt:        env.ExpireAt,
					Request:         traceReq,
					Error:           err.Error(),
				})
			}
			return
		}
		respBody, err := encodeProtoValue(resp)
		if err != nil {
			writeErrProto(s, err.Error())
			if trace != nil {
				trace.Handle(TraceEvent{
					TS:              nowRFC3339(),
					Direction:       "recv",
					Domain:          cfg.Domain,
					Network:         cfg.Network,
					ProtoID:         string(protoID),
					MsgID:           env.MsgID,
					LocalPeerID:     localPID,
					RemotePeerID:    remotePID,
					SenderPubkeyHex: senderPubHex,
					ExpireAt:        env.ExpireAt,
					Request:         traceReq,
					Error:           err.Error(),
				})
			}
			return
		}
		if cfg.Replay != nil {
			_ = cfg.Replay.Put(string(protoID), env.MsgID, payloadHash, respBody, env.ExpireAt)
		}
		_, _ = s.Write(respBody)
		if trace != nil {
			trace.Handle(TraceEvent{
				TS:              nowRFC3339(),
				Direction:       "recv",
				Domain:          cfg.Domain,
				Network:         cfg.Network,
				ProtoID:         string(protoID),
				MsgID:           env.MsgID,
				LocalPeerID:     localPID,
				RemotePeerID:    remotePID,
				SenderPubkeyHex: senderPubHex,
				ExpireAt:        env.ExpireAt,
				Request:         traceReq,
				Response:        normalizeProtoTracePayload(respBody),
			})
		}
	})
}

// CallProto 使用 protobuf 二进制编解码发起 p2p RPC。
func CallProto[TReq any, TResp any](ctx context.Context, h host.Host, pid peer.ID, protoID coreprotocol.ID, cfg SecurityConfig, req TReq, out *TResp) error {
	start := time.Now()
	s, err := h.NewStream(ctx, pid, protoID)
	if err != nil {
		return err
	}
	defer s.Close()

	payload, err := encodeProtoValue(req)
	if err != nil {
		return err
	}
	priv := h.Peerstore().PrivKey(h.ID())
	if priv == nil {
		return fmt.Errorf("missing host private key")
	}
	msgID, err := randomMsgID()
	if err != nil {
		return err
	}
	ttl := cfg.TTL
	if ttl <= 0 {
		ttl = 20 * time.Second
	}
	env, err := appprotocol.NewSignedEnvelopePB(priv, cfg.Domain, cfg.Network, string(protoID), msgID, ttl, payload)
	if err != nil {
		return err
	}
	wire, err := marshalProtoDeterministic(env)
	if err != nil {
		return err
	}
	if _, err := s.Write(wire); err != nil {
		return err
	}
	if cw, ok := s.(interface{ CloseWrite() error }); ok {
		_ = cw.CloseWrite()
	}

	body, err := io.ReadAll(s)
	if err != nil {
		if cfg.Trace != nil {
			cfg.Trace.Handle(TraceEvent{
				TS:           nowRFC3339(),
				Direction:    "send",
				Domain:       cfg.Domain,
				Network:      cfg.Network,
				ProtoID:      string(protoID),
				MsgID:        env.MsgID,
				LocalPeerID:  h.ID().String(),
				RemotePeerID: pid.String(),
				ExpireAt:     env.ExpireAt,
				Request:      normalizeProtoTracePayload(payload),
				Error:        err.Error(),
				DurationMS:   time.Since(start).Milliseconds(),
			})
		}
		return err
	}
	if len(body) == 0 {
		if cfg.Trace != nil {
			cfg.Trace.Handle(TraceEvent{
				TS:           nowRFC3339(),
				Direction:    "send",
				Domain:       cfg.Domain,
				Network:      cfg.Network,
				ProtoID:      string(protoID),
				MsgID:        env.MsgID,
				LocalPeerID:  h.ID().String(),
				RemotePeerID: pid.String(),
				ExpireAt:     env.ExpireAt,
				Request:      normalizeProtoTracePayload(payload),
				Error:        "empty response",
				DurationMS:   time.Since(start).Milliseconds(),
			})
		}
		return fmt.Errorf("empty response")
	}
	var e ErrorResponsePB
	if err := oldproto.Unmarshal(body, &e); err == nil && e.IsError && strings.TrimSpace(e.Error) != "" {
		if cfg.Trace != nil {
			cfg.Trace.Handle(TraceEvent{
				TS:           nowRFC3339(),
				Direction:    "send",
				Domain:       cfg.Domain,
				Network:      cfg.Network,
				ProtoID:      string(protoID),
				MsgID:        env.MsgID,
				LocalPeerID:  h.ID().String(),
				RemotePeerID: pid.String(),
				ExpireAt:     env.ExpireAt,
				Request:      normalizeProtoTracePayload(payload),
				Response:     normalizeProtoTracePayload(body),
				Error:        e.Error,
				DurationMS:   time.Since(start).Milliseconds(),
			})
		}
		return errors.New(e.Error)
	}
	if err := decodeProtoInto(body, out); err != nil {
		var legacy legacyErrorResponsePB
		if uerr := oldproto.Unmarshal(body, &legacy); uerr == nil && strings.TrimSpace(legacy.Error) != "" {
			if cfg.Trace != nil {
				cfg.Trace.Handle(TraceEvent{
					TS:           nowRFC3339(),
					Direction:    "send",
					Domain:       cfg.Domain,
					Network:      cfg.Network,
					ProtoID:      string(protoID),
					MsgID:        env.MsgID,
					LocalPeerID:  h.ID().String(),
					RemotePeerID: pid.String(),
					ExpireAt:     env.ExpireAt,
					Request:      normalizeProtoTracePayload(payload),
					Response:     normalizeProtoTracePayload(body),
					Error:        legacy.Error,
					DurationMS:   time.Since(start).Milliseconds(),
				})
			}
			return errors.New(legacy.Error)
		}
		if cfg.Trace != nil {
			cfg.Trace.Handle(TraceEvent{
				TS:           nowRFC3339(),
				Direction:    "send",
				Domain:       cfg.Domain,
				Network:      cfg.Network,
				ProtoID:      string(protoID),
				MsgID:        env.MsgID,
				LocalPeerID:  h.ID().String(),
				RemotePeerID: pid.String(),
				ExpireAt:     env.ExpireAt,
				Request:      normalizeProtoTracePayload(payload),
				Response:     normalizeProtoTracePayload(body),
				Error:        err.Error(),
				DurationMS:   time.Since(start).Milliseconds(),
			})
		}
		return err
	}
	if cfg.Trace != nil {
		respBody, _ := encodeProtoValue(*out)
		cfg.Trace.Handle(TraceEvent{
			TS:           nowRFC3339(),
			Direction:    "send",
			Domain:       cfg.Domain,
			Network:      cfg.Network,
			ProtoID:      string(protoID),
			MsgID:        env.MsgID,
			LocalPeerID:  h.ID().String(),
			RemotePeerID: pid.String(),
			ExpireAt:     env.ExpireAt,
			Request:      normalizeProtoTracePayload(payload),
			Response:     normalizeProtoTracePayload(respBody),
			DurationMS:   time.Since(start).Milliseconds(),
		})
	}
	return nil
}

func writeErrProto(w io.Writer, msg string) {
	b, err := marshalProtoDeterministic(&ErrorResponsePB{IsError: true, Error: msg})
	if err != nil {
		// 兜底：编码失败时直接写空错误串的 protobuf 消息。
		b, _ = marshalProtoDeterministic(&ErrorResponsePB{IsError: true, Error: "internal proto encode failed"})
	}
	_, _ = w.Write(b)
}

func marshalProtoDeterministic(m oldproto.Message) ([]byte, error) {
	buf := oldproto.NewBuffer(nil)
	buf.SetDeterministic(true)
	if err := buf.Marshal(m); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func encodeProtoValue[T any](v T) ([]byte, error) {
	if m, ok := any(v).(oldproto.Message); ok {
		return marshalProtoDeterministic(m)
	}
	if m, ok := any(&v).(oldproto.Message); ok {
		return marshalProtoDeterministic(m)
	}
	return nil, fmt.Errorf("payload type %T does not implement protobuf message", v)
}

func decodeProtoValue[T any](payload []byte) (T, error) {
	var out T
	if m, ok := any(&out).(oldproto.Message); ok {
		if err := oldproto.Unmarshal(payload, m); err != nil {
			return out, err
		}
		return out, nil
	}
	if m, ok := any(out).(oldproto.Message); ok {
		if err := oldproto.Unmarshal(payload, m); err != nil {
			return out, err
		}
		return out, nil
	}
	return out, fmt.Errorf("payload type %T does not implement protobuf message", out)
}

func decodeProtoInto[T any](payload []byte, out *T) error {
	if out == nil {
		return fmt.Errorf("response out is nil")
	}
	if m, ok := any(out).(oldproto.Message); ok {
		return oldproto.Unmarshal(payload, m)
	}
	return fmt.Errorf("response type %T does not implement protobuf message", *out)
}

func normalizeProtoTracePayload(b []byte) any {
	if len(b) == 0 {
		return map[string]any{"payload_bytes": 0}
	}
	preview := hex.EncodeToString(b)
	if len(preview) > 64 {
		preview = preview[:64]
	}
	return map[string]any{
		"payload_bytes":      len(b),
		"payload_hex_prefix": preview,
	}
}
