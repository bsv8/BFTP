package woc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bsv8/BFTP/internal/wocraw"
)

type GuardServer struct {
	chain *guardUpstream
}

const (
	GuardServiceName    = "woc-guard"
	GuardServiceVersion = "1.0.0"
)

func NewGuardServer(network string, baseURLOverride string, protectInterval time.Duration) *GuardServer {
	raw := wocraw.NewForNetwork(network, baseURLOverride)
	return &GuardServer{
		chain: newGuardUpstream(raw, protectInterval),
	}
}

func (s *GuardServer) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/v1/tip", s.handleTip)
	mux.HandleFunc("/v1/utxos/", s.handleUTXOs)
	mux.HandleFunc("/v1/broadcast", s.handleBroadcast)
	mux.HandleFunc("/v1/address-history/", s.handleAddressHistory)
	mux.HandleFunc("/v1/tx/", s.handleTx)
	return mux
}

func (s *GuardServer) Stats() GuardStats {
	return s.chain.SnapshotStats()
}

func (s *GuardServer) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"service": GuardServiceName,
		"version": GuardServiceVersion,
	})
}

func (s *GuardServer) handleTip(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	tip, err := s.chain.GetTipHeightContext(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tip_height": tip})
}

func (s *GuardServer) handleUTXOs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	addr := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/v1/utxos/"))
	if addr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing address"})
		return
	}
	utxos, err := s.chain.GetUTXOsContext(r.Context(), addr)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"utxos": utxos})
}

func (s *GuardServer) handleBroadcast(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	var req struct {
		TxHex string `json:"tx_hex"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("invalid json: %v", err)})
		return
	}
	req.TxHex = strings.TrimSpace(req.TxHex)
	if req.TxHex == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "tx_hex is required"})
		return
	}
	txid, err := s.chain.BroadcastContext(r.Context(), req.TxHex)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"txid": txid})
}

func (s *GuardServer) handleAddressHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	addr := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/v1/address-history/"))
	if addr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing address"})
		return
	}
	items, err := s.chain.GetAddressHistoryContext(r.Context(), addr)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *GuardServer) handleTx(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	txid := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/v1/tx/"))
	if txid == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing txid"})
		return
	}
	txj, err := s.chain.GetTxDetailContext(r.Context(), txid)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tx": txj})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
