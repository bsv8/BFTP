package wocraw

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetUTXOsContext_NewConfirmedPath(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/address/addr1/confirmed/unspent" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"result": []map[string]any{
				{"tx_hash": "tx1", "tx_pos": 1, "value": 10},
			},
			"error": "",
		})
	}))
	defer srv.Close()

	c := New(srv.URL)
	utxos, err := c.GetUTXOsContext(context.Background(), "addr1")
	if err != nil {
		t.Fatalf("GetUTXOsContext failed: %v", err)
	}
	if len(utxos) != 1 || utxos[0].TxID != "tx1" || utxos[0].Vout != 1 || utxos[0].Value != 10 {
		t.Fatalf("unexpected utxos: %+v", utxos)
	}
}

func TestGetUTXOsContext_FallbackLegacyPath(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/address/addr1/confirmed/unspent":
			http.NotFound(w, r)
		case "/address/addr1/unspent":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"tx_hash": "tx2", "tx_pos": 2, "value": 20},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := New(srv.URL)
	utxos, err := c.GetUTXOsContext(context.Background(), "addr1")
	if err != nil {
		t.Fatalf("GetUTXOsContext failed: %v", err)
	}
	if len(utxos) != 1 || utxos[0].TxID != "tx2" || utxos[0].Vout != 2 || utxos[0].Value != 20 {
		t.Fatalf("unexpected utxos: %+v", utxos)
	}
}

func TestGetAddressHistoryContext_NewConfirmedPath(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/address/addr1/confirmed/history" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"result": []map[string]any{
				{"tx_hash": "tx3", "height": 100},
			},
			"error": "",
		})
	}))
	defer srv.Close()

	c := New(srv.URL)
	items, err := c.GetAddressHistoryContext(context.Background(), "addr1")
	if err != nil {
		t.Fatalf("GetAddressHistoryContext failed: %v", err)
	}
	if len(items) != 1 || items[0].TxID != "tx3" || items[0].Height != 100 {
		t.Fatalf("unexpected history items: %+v", items)
	}
}

func TestGetAddressHistoryContext_FallbackLegacyPath(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/address/addr1/confirmed/history":
			http.NotFound(w, r)
		case "/address/addr1/history":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"tx_hash": "tx4", "height": 200},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := New(srv.URL)
	items, err := c.GetAddressHistoryContext(context.Background(), "addr1")
	if err != nil {
		t.Fatalf("GetAddressHistoryContext failed: %v", err)
	}
	if len(items) != 1 || items[0].TxID != "tx4" || items[0].Height != 200 {
		t.Fatalf("unexpected history items: %+v", items)
	}
}

func TestGetUTXOsContext_FilterSpentInMempool(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/address/addr1/confirmed/unspent" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"result": []map[string]any{
				{"tx_hash": "tx5", "tx_pos": 5, "value": 50, "isSpentInMempoolTx": true},
				{"tx_hash": "tx6", "tx_pos": 6, "value": 60, "isSpentInMempoolTx": false},
			},
			"error": "",
		})
	}))
	defer srv.Close()

	c := New(srv.URL)
	utxos, err := c.GetUTXOsContext(context.Background(), "addr1")
	if err != nil {
		t.Fatalf("GetUTXOsContext failed: %v", err)
	}
	if len(utxos) != 1 || utxos[0].TxID != "tx6" || utxos[0].Vout != 6 || utxos[0].Value != 60 {
		t.Fatalf("unexpected utxos after mempool filter: %+v", utxos)
	}
}
