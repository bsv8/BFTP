package wocraw

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAddressConfirmedUnspent_NewConfirmedPath(t *testing.T) {
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
	utxos, err := c.GetAddressConfirmedUnspent(context.Background(), "addr1")
	if err != nil {
		t.Fatalf("GetAddressConfirmedUnspent failed: %v", err)
	}
	if len(utxos) != 1 || utxos[0].TxID != "tx1" || utxos[0].Vout != 1 || utxos[0].Value != 10 {
		t.Fatalf("unexpected utxos: %+v", utxos)
	}
}

func TestGetAddressConfirmedUnspent_FallbackLegacyPath(t *testing.T) {
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
	utxos, err := c.GetAddressConfirmedUnspent(context.Background(), "addr1")
	if err != nil {
		t.Fatalf("GetAddressConfirmedUnspent failed: %v", err)
	}
	if len(utxos) != 1 || utxos[0].TxID != "tx2" || utxos[0].Vout != 2 || utxos[0].Value != 20 {
		t.Fatalf("unexpected utxos: %+v", utxos)
	}
}

func TestGetAddressConfirmedHistory_NewConfirmedPath(t *testing.T) {
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
	items, err := c.GetAddressConfirmedHistory(context.Background(), "addr1")
	if err != nil {
		t.Fatalf("GetAddressConfirmedHistory failed: %v", err)
	}
	if len(items) != 1 || items[0].TxID != "tx3" || items[0].Height != 100 {
		t.Fatalf("unexpected history items: %+v", items)
	}
}

func TestGetAddressConfirmedHistory_FallbackLegacyPath(t *testing.T) {
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
	items, err := c.GetAddressConfirmedHistory(context.Background(), "addr1")
	if err != nil {
		t.Fatalf("GetAddressConfirmedHistory failed: %v", err)
	}
	if len(items) != 1 || items[0].TxID != "tx4" || items[0].Height != 200 {
		t.Fatalf("unexpected history items: %+v", items)
	}
}

func TestGetAddressConfirmedHistoryPage_UsesQuery(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/address/addr1/confirmed/history" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("order") != "asc" {
			t.Fatalf("unexpected order: %s", q.Get("order"))
		}
		if q.Get("limit") != "50" {
			t.Fatalf("unexpected limit: %s", q.Get("limit"))
		}
		if q.Get("height") != "123" {
			t.Fatalf("unexpected height: %s", q.Get("height"))
		}
		if q.Get("token") != "next-token" {
			t.Fatalf("unexpected token: %s", q.Get("token"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"result": []map[string]any{
				{"tx_hash": "tx7", "height": 123},
			},
			"nextPageToken": "after-7",
			"error":         "",
		})
	}))
	defer srv.Close()

	c := New(srv.URL)
	page, err := c.GetAddressConfirmedHistoryPage(context.Background(), "addr1", ConfirmedHistoryQuery{
		Order:  "asc",
		Limit:  50,
		Height: 123,
		Token:  "next-token",
	})
	if err != nil {
		t.Fatalf("GetAddressConfirmedHistoryPage failed: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].TxID != "tx7" || page.Items[0].Height != 123 {
		t.Fatalf("unexpected confirmed history page: %+v", page)
	}
	if page.NextPageToken != "after-7" {
		t.Fatalf("unexpected next_page_token: %s", page.NextPageToken)
	}
}

func TestGetAddressUnconfirmedHistory(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/address/addr1/unconfirmed/history" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"result": []map[string]any{
				{"tx_hash": "TX8"},
				{"tx_hash": "tx9"},
			},
			"error": "",
		})
	}))
	defer srv.Close()

	c := New(srv.URL)
	items, err := c.GetAddressUnconfirmedHistory(context.Background(), "addr1")
	if err != nil {
		t.Fatalf("GetAddressUnconfirmedHistory failed: %v", err)
	}
	if len(items) != 2 || items[0] != "tx8" || items[1] != "tx9" {
		t.Fatalf("unexpected unconfirmed history: %+v", items)
	}
}

func TestGetAddressConfirmedUnspent_FilterSpentInMempool(t *testing.T) {
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
	utxos, err := c.GetAddressConfirmedUnspent(context.Background(), "addr1")
	if err != nil {
		t.Fatalf("GetAddressConfirmedUnspent failed: %v", err)
	}
	if len(utxos) != 1 || utxos[0].TxID != "tx6" || utxos[0].Vout != 6 || utxos[0].Value != 60 {
		t.Fatalf("unexpected utxos after mempool filter: %+v", utxos)
	}
}
