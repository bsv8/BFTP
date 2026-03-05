package woc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type GuardClient struct {
	baseURL string
	http    *http.Client
}

func NewGuardClient(baseURL string) *GuardClient {
	u := normalizeGuardBaseURL(baseURL)
	return &GuardClient{
		baseURL: u,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *GuardClient) BaseURL() string { return c.baseURL }

func (c *GuardClient) GetUTXOs(address string) ([]UTXO, error) {
	return c.GetUTXOsContext(context.Background(), address)
}

func (c *GuardClient) GetUTXOsContext(ctx context.Context, address string) ([]UTXO, error) {
	body, err := c.get(ctx, "/v1/utxos/"+strings.TrimSpace(address))
	if err != nil {
		return nil, err
	}
	var out struct {
		UTXOs []UTXO `json:"utxos"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode guard utxos response: %w", err)
	}
	return out.UTXOs, nil
}

func (c *GuardClient) GetTipHeight() (uint32, error) {
	return c.GetTipHeightContext(context.Background())
}

func (c *GuardClient) GetTipHeightContext(ctx context.Context) (uint32, error) {
	body, err := c.get(ctx, "/v1/tip")
	if err != nil {
		return 0, err
	}
	var out struct {
		TipHeight uint32 `json:"tip_height"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return 0, fmt.Errorf("decode guard tip response: %w", err)
	}
	return out.TipHeight, nil
}

func (c *GuardClient) Broadcast(txHex string) (string, error) {
	return c.BroadcastContext(context.Background(), txHex)
}

func (c *GuardClient) BroadcastContext(ctx context.Context, txHex string) (string, error) {
	body, err := c.postJSON(ctx, "/v1/broadcast", map[string]string{"tx_hex": txHex})
	if err != nil {
		return "", err
	}
	var out struct {
		TxID string `json:"txid"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", fmt.Errorf("decode guard broadcast response: %w", err)
	}
	if strings.TrimSpace(out.TxID) == "" {
		return "", fmt.Errorf("guard broadcast response missing txid")
	}
	return strings.TrimSpace(out.TxID), nil
}

func (c *GuardClient) GetAddressHistory(address string) ([]AddressHistoryItem, error) {
	return c.GetAddressHistoryContext(context.Background(), address)
}

func (c *GuardClient) GetAddressHistoryContext(ctx context.Context, address string) ([]AddressHistoryItem, error) {
	body, err := c.get(ctx, "/v1/address-history/"+strings.TrimSpace(address))
	if err != nil {
		return nil, err
	}
	var out struct {
		Items []AddressHistoryItem `json:"items"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode guard address history response: %w", err)
	}
	return out.Items, nil
}

func (c *GuardClient) GetTxDetail(txid string) (TxDetail, error) {
	return c.GetTxDetailContext(context.Background(), txid)
}

func (c *GuardClient) GetTxDetailContext(ctx context.Context, txid string) (TxDetail, error) {
	body, err := c.get(ctx, "/v1/tx/"+strings.TrimSpace(txid))
	if err != nil {
		return TxDetail{}, err
	}
	var out struct {
		Tx TxDetail `json:"tx"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return TxDetail{}, fmt.Errorf("decode guard tx response: %w", err)
	}
	if strings.TrimSpace(out.Tx.TxID) == "" {
		return TxDetail{}, fmt.Errorf("guard tx response missing txid")
	}
	return out.Tx, nil
}

func (c *GuardClient) get(ctx context.Context, path string) ([]byte, error) {
	path = normalizeGuardPath(path)
	if status, body, handled, err := tryInprocGuardRequest(ctx, c.baseURL, http.MethodGet, path, nil, ""); handled {
		if err != nil {
			return nil, err
		}
		if status < 200 || status >= 300 {
			return nil, fmt.Errorf("http %d: %s", status, string(body))
		}
		return body, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

func (c *GuardClient) postJSON(ctx context.Context, path string, payload any) ([]byte, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	path = normalizeGuardPath(path)
	if status, body, handled, err := tryInprocGuardRequest(ctx, c.baseURL, http.MethodPost, path, raw, "application/json"); handled {
		if err != nil {
			return nil, err
		}
		if status < 200 || status >= 300 {
			return nil, fmt.Errorf("http %d: %s", status, string(body))
		}
		return body, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

func normalizeGuardPath(path string) string {
	p := strings.TrimSpace(path)
	if p == "" {
		return "/"
	}
	if strings.HasPrefix(p, "/") {
		return p
	}
	return "/" + p
}
