package p2prpc

import (
	"bufio"
	"context"
	cryptorand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	appprotocol "github.com/bsv8/BFTP/pkg/protocol"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	coreprotocol "github.com/libp2p/go-libp2p/core/protocol"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

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

func HandleJSON[TReq any, TResp any](h host.Host, protoID coreprotocol.ID, cfg SecurityConfig, fn func(context.Context, TReq) (TResp, error)) {
	h.SetStreamHandler(protoID, func(s network.Stream) {
		defer s.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		var env appprotocol.SignedEnvelope
		if err := json.NewDecoder(bufio.NewReader(s)).Decode(&env); err != nil {
			writeErr(s, "decode envelope: "+err.Error())
			return
		}
		if err := appprotocol.VerifyEnvelope(&env, time.Now()); err != nil {
			writeErr(s, err.Error())
			return
		}
		if env.Domain != cfg.Domain || env.Network != cfg.Network {
			writeErr(s, "domain/network mismatch")
			return
		}
		if env.MsgType != string(protoID) {
			writeErr(s, "msg_type mismatch")
			return
		}
		if err := verifyRemotePeer(s.Conn().RemotePeer(), env.SenderPubkey); err != nil {
			writeErr(s, err.Error())
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
		var traceReq any
		if trace != nil {
			traceReq = NormalizeJSONForTrace(env.Payload, NormalizeOptions{})
		}
		ctx = context.WithValue(ctx, senderPubkeyHexContextKey, strings.ToLower(env.SenderPubkey))

		payloadHash := appprotocol.EnvelopePayloadHash(env.Payload)
		if cfg.Replay != nil {
			entry, found, err := cfg.Replay.Get(string(protoID), env.MsgID)
			if err != nil {
				writeErr(s, err.Error())
				if trace != nil {
					trace.Handle(TraceEvent{
						TS:             nowRFC3339(),
						Direction:      "recv",
						Domain:         cfg.Domain,
						Network:        cfg.Network,
						ProtoID:        string(protoID),
						MsgID:          env.MsgID,
						LocalPeerID:    localPID,
						RemotePeerID:   remotePID,
						SenderPubkeyHex: strings.ToLower(env.SenderPubkey),
						ExpireAt:       env.ExpireAt,
						Request:        traceReq,
						Error:          err.Error(),
					})
				}
				return
			}
			if found {
				if entry.PayloadHash != payloadHash {
					writeErr(s, "ERR_IDEMPOTENCY_REPLAY")
					if trace != nil {
						trace.Handle(TraceEvent{
							TS:             nowRFC3339(),
							Direction:      "recv",
							Domain:         cfg.Domain,
							Network:        cfg.Network,
							ProtoID:        string(protoID),
							MsgID:          env.MsgID,
							LocalPeerID:    localPID,
							RemotePeerID:   remotePID,
							SenderPubkeyHex: strings.ToLower(env.SenderPubkey),
							ExpireAt:       env.ExpireAt,
							ReplayHit:      true,
							Request:        traceReq,
							Error:          "ERR_IDEMPOTENCY_REPLAY",
						})
					}
					return
				}
				_, _ = s.Write(entry.Response)
				if trace != nil {
					trace.Handle(TraceEvent{
						TS:             nowRFC3339(),
						Direction:      "recv",
						Domain:         cfg.Domain,
						Network:        cfg.Network,
						ProtoID:        string(protoID),
						MsgID:          env.MsgID,
						LocalPeerID:    localPID,
						RemotePeerID:   remotePID,
						SenderPubkeyHex: strings.ToLower(env.SenderPubkey),
						ExpireAt:       env.ExpireAt,
						ReplayHit:      true,
						Request:        traceReq,
						Response:       NormalizeJSONForTrace(entry.Response, NormalizeOptions{}),
					})
				}
				return
			}
		}

		var req TReq
		if err := json.Unmarshal(env.Payload, &req); err != nil {
			writeErr(s, "decode payload: "+err.Error())
			if trace != nil {
				trace.Handle(TraceEvent{
					TS:             nowRFC3339(),
					Direction:      "recv",
					Domain:         cfg.Domain,
					Network:        cfg.Network,
					ProtoID:        string(protoID),
					MsgID:          env.MsgID,
					LocalPeerID:    localPID,
					RemotePeerID:   remotePID,
					SenderPubkeyHex: strings.ToLower(env.SenderPubkey),
					ExpireAt:       env.ExpireAt,
					Request:        traceReq,
					Error:          "decode payload: " + err.Error(),
				})
			}
			return
		}
		resp, err := fn(ctx, req)
		if err != nil {
			writeErr(s, err.Error())
			if trace != nil {
				trace.Handle(TraceEvent{
					TS:             nowRFC3339(),
					Direction:      "recv",
					Domain:         cfg.Domain,
					Network:        cfg.Network,
					ProtoID:        string(protoID),
					MsgID:          env.MsgID,
					LocalPeerID:    localPID,
					RemotePeerID:   remotePID,
					SenderPubkeyHex: strings.ToLower(env.SenderPubkey),
					ExpireAt:       env.ExpireAt,
					Request:        traceReq,
					Error:          err.Error(),
				})
			}
			return
		}
		respBody, err := json.Marshal(resp)
		if err != nil {
			writeErr(s, err.Error())
			if trace != nil {
				trace.Handle(TraceEvent{
					TS:             nowRFC3339(),
					Direction:      "recv",
					Domain:         cfg.Domain,
					Network:        cfg.Network,
					ProtoID:        string(protoID),
					MsgID:          env.MsgID,
					LocalPeerID:    localPID,
					RemotePeerID:   remotePID,
					SenderPubkeyHex: strings.ToLower(env.SenderPubkey),
					ExpireAt:       env.ExpireAt,
					Request:        traceReq,
					Error:          err.Error(),
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
				TS:             nowRFC3339(),
				Direction:      "recv",
				Domain:         cfg.Domain,
				Network:        cfg.Network,
				ProtoID:        string(protoID),
				MsgID:          env.MsgID,
				LocalPeerID:    localPID,
				RemotePeerID:   remotePID,
				SenderPubkeyHex: strings.ToLower(env.SenderPubkey),
				ExpireAt:       env.ExpireAt,
				Request:        traceReq,
				Response:       NormalizeJSONForTrace(respBody, NormalizeOptions{}),
			})
		}
	})
}

func SenderPubkeyHexFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(senderPubkeyHexContextKey)
	s, ok := v.(string)
	return s, ok && s != ""
}

func CallJSON[TReq any, TResp any](ctx context.Context, h host.Host, pid peer.ID, protoID coreprotocol.ID, cfg SecurityConfig, req TReq, out *TResp) error {
	start := time.Now()
	s, err := h.NewStream(ctx, pid, protoID)
	if err != nil {
		return err
	}
	defer s.Close()

	payload, err := json.Marshal(req)
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
	env, err := appprotocol.NewSignedEnvelope(priv, cfg.Domain, cfg.Network, string(protoID), msgID, ttl, payload)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(s).Encode(env); err != nil {
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
				Request:      NormalizeJSONForTrace(payload, NormalizeOptions{}),
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
				Request:      NormalizeJSONForTrace(payload, NormalizeOptions{}),
				Error:        "empty response",
				DurationMS:   time.Since(start).Milliseconds(),
			})
		}
		return fmt.Errorf("empty response")
	}
	var e ErrorResponse
	if err := json.Unmarshal(body, &e); err == nil && e.Error != "" {
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
				Request:      NormalizeJSONForTrace(payload, NormalizeOptions{}),
				Response:     NormalizeJSONForTrace(body, NormalizeOptions{}),
				Error:        e.Error,
				DurationMS:   time.Since(start).Milliseconds(),
			})
		}
		return errors.New(e.Error)
	}
	if err := json.Unmarshal(body, out); err != nil {
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
				Request:      NormalizeJSONForTrace(payload, NormalizeOptions{}),
				Response:     NormalizeJSONForTrace(body, NormalizeOptions{}),
				Error:        err.Error(),
				DurationMS:   time.Since(start).Milliseconds(),
			})
		}
		return err
	}
	if cfg.Trace != nil {
		respBody, _ := json.Marshal(out)
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
			Request:      NormalizeJSONForTrace(payload, NormalizeOptions{}),
			Response:     NormalizeJSONForTrace(respBody, NormalizeOptions{}),
			DurationMS:   time.Since(start).Milliseconds(),
		})
	}
	return nil
}

func verifyRemotePeer(remote peer.ID, senderPubHex string) error {
	pubBytes, err := hex.DecodeString(senderPubHex)
	if err != nil {
		return fmt.Errorf("decode sender pubkey: %w", err)
	}
	pub, err := crypto.UnmarshalPublicKey(pubBytes)
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

func writeErr(w io.Writer, msg string) {
	b, _ := json.Marshal(ErrorResponse{Error: msg})
	_, _ = w.Write(b)
}

func randomMsgID() (string, error) {
	b := make([]byte, 16)
	if _, err := cryptorand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
