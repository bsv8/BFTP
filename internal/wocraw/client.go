package wocraw

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	// TestnetBaseURL 是 Whatsonchain 的 BSV 测试网 HTTP API 地址。
	TestnetBaseURL = "https://api.whatsonchain.com/v1/bsv/test"
	// MainnetBaseURL 是 Whatsonchain 的 BSV 主网 HTTP API 地址。
	MainnetBaseURL = "https://api.whatsonchain.com/v1/bsv/main"
)

// BaseURLForNetwork 返回指定网络的官方 WOC base url。
// network 仅支持 "test" / "main"；其它值会回退到 test。
func BaseURLForNetwork(network string) string {
	switch strings.ToLower(strings.TrimSpace(network)) {
	case "main", "mainnet":
		return MainnetBaseURL
	default:
		return TestnetBaseURL
	}
}

// NewForNetwork 创建一个 WOC 原始上游客户端：
// - 如果 baseURLOverride 非空，则使用它；
// - 否则根据 network 选择官方 test/main 地址。
func NewForNetwork(network string, baseURLOverride string) *Client {
	if s := strings.TrimSpace(baseURLOverride); s != "" {
		return New(s)
	}
	return New(BaseURLForNetwork(network))
}

type Client struct {
	baseURL string
	http    *http.Client
}

type UTXO struct {
	TxID  string `json:"tx_hash"`
	Vout  uint32 `json:"tx_pos"`
	Value uint64 `json:"value"`
}

type AddressHistoryItem struct {
	TxID   string `json:"tx_hash"`
	Height int64  `json:"height"`
}

type TxDetail struct {
	TxID string     `json:"txid"`
	Vin  []TxInput  `json:"vin"`
	Vout []TxOutput `json:"vout"`
}

type TxInput struct {
	TxID string `json:"txid"`
	Vout uint32 `json:"vout"`
}

type TxOutput struct {
	N            uint32       `json:"n"`
	Value        float64      `json:"value"`
	ScriptPubKey ScriptPubKey `json:"scriptPubKey"`
}

type ScriptPubKey struct {
	Hex string `json:"hex"`
}

type chainInfo struct {
	Blocks uint32 `json:"blocks"`
}

type HTTPError struct {
	StatusCode int
	Body       string
	RetryAfter time.Duration
}

func (e *HTTPError) Error() string {
	if e == nil {
		return "http error"
	}
	return fmt.Sprintf("http %d: %s", e.StatusCode, e.Body)
}

func (e *HTTPError) IsRetryable() bool {
	if e == nil {
		return false
	}
	return e.StatusCode == http.StatusTooManyRequests || e.StatusCode >= 500
}

func New(baseURL string) *Client {
	url := strings.TrimSpace(baseURL)
	if url == "" {
		url = TestnetBaseURL
	}
	return &Client{
		baseURL: strings.TrimRight(url, "/"),
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) BaseURL() string { return c.baseURL }

func (c *Client) GetUTXOs(address string) ([]UTXO, error) {
	return c.GetUTXOsContext(context.Background(), address)
}

func (c *Client) GetUTXOsContext(ctx context.Context, address string) ([]UTXO, error) {
	body, err := c.get(ctx, fmt.Sprintf("%s/address/%s/unspent", c.baseURL, address))
	if err != nil {
		return nil, err
	}
	var utxos []UTXO
	if err := json.Unmarshal(body, &utxos); err != nil {
		return nil, fmt.Errorf("decode utxos: %w", err)
	}
	return utxos, nil
}

func (c *Client) GetTipHeight() (uint32, error) {
	return c.GetTipHeightContext(context.Background())
}

func (c *Client) GetTipHeightContext(ctx context.Context) (uint32, error) {
	body, err := c.get(ctx, c.baseURL+"/chain/info")
	if err != nil {
		return 0, err
	}
	var info chainInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return 0, fmt.Errorf("decode chain info: %w", err)
	}
	return info.Blocks, nil
}

func (c *Client) Broadcast(txHex string) (string, error) {
	return c.BroadcastContext(context.Background(), txHex)
}

func (c *Client) BroadcastContext(ctx context.Context, txHex string) (string, error) {
	body, err := c.postJSON(ctx, c.baseURL+"/tx/raw", map[string]string{"txhex": txHex})
	if err != nil {
		return "", err
	}
	var txid string
	if err := json.Unmarshal(body, &txid); err == nil && txid != "" {
		return txid, nil
	}
	var obj map[string]any
	if err := json.Unmarshal(body, &obj); err == nil {
		if v, ok := obj["txid"].(string); ok && v != "" {
			return v, nil
		}
		if v, ok := obj["data"].(string); ok && v != "" {
			return v, nil
		}
	}
	return "", fmt.Errorf("unexpected broadcast response: %s", string(body))
}

func (c *Client) GetAddressHistoryContext(ctx context.Context, address string) ([]AddressHistoryItem, error) {
	body, err := c.get(ctx, fmt.Sprintf("%s/address/%s/history", c.baseURL, strings.TrimSpace(address)))
	if err != nil {
		return nil, err
	}
	var out []AddressHistoryItem
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode address history: %w", err)
	}
	return out, nil
}

func (c *Client) GetTxDetailContext(ctx context.Context, txid string) (TxDetail, error) {
	body, err := c.get(ctx, fmt.Sprintf("%s/tx/hash/%s", c.baseURL, strings.TrimSpace(txid)))
	if err != nil {
		return TxDetail{}, err
	}
	var out TxDetail
	if err := json.Unmarshal(body, &out); err != nil {
		return TxDetail{}, fmt.Errorf("decode tx detail: %w", err)
	}
	return out, nil
}

func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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
		return nil, &HTTPError{
			StatusCode: resp.StatusCode,
			Body:       string(body),
			RetryAfter: parseRetryAfterHeader(resp.Header.Get("Retry-After")),
		}
	}
	return body, nil
}

func (c *Client) postJSON(ctx context.Context, url string, payload any) ([]byte, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
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
		return nil, &HTTPError{
			StatusCode: resp.StatusCode,
			Body:       string(body),
			RetryAfter: parseRetryAfterHeader(resp.Header.Get("Retry-After")),
		}
	}
	return body, nil
}

func parseRetryAfterHeader(v string) time.Duration {
	s := strings.TrimSpace(v)
	if s == "" {
		return 0
	}
	if sec, err := strconv.Atoi(s); err == nil && sec > 0 {
		return time.Duration(sec) * time.Second
	}
	if t, err := http.ParseTime(s); err == nil {
		d := time.Until(t)
		if d > 0 {
			return d
		}
	}
	return 0
}
