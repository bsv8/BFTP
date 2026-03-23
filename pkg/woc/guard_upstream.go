package woc

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bsv8/BFTP/internal/wocraw"
)

type GuardStats struct {
	Calls       uint64        `json:"calls"`
	WaitedCalls uint64        `json:"waited_calls"`
	WaitTotal   time.Duration `json:"wait_total"`
	WaitAvg     time.Duration `json:"wait_avg"`
}

// guardUpstream 只在 guard server 内使用：
// 所有上游 WOC 调用统一串行+间隔保护。
type guardUpstream struct {
	raw *wocraw.Client
	// interval 是两次上游调用的最小间隔。
	interval time.Duration

	mu          sync.Mutex
	nextAllowed time.Time

	callCount      atomic.Uint64
	waitedCall     atomic.Uint64
	waitTotalNanos atomic.Int64
}

func newGuardUpstream(raw *wocraw.Client, interval time.Duration) *guardUpstream {
	if raw == nil {
		raw = wocraw.New(wocraw.TestnetBaseURL)
	}
	if interval <= 0 {
		interval = 1 * time.Second
	}
	return &guardUpstream{raw: raw, interval: interval}
}

func (g *guardUpstream) GetAddressConfirmedUnspent(ctx context.Context, address string) ([]UTXO, error) {
	if err := g.waitTurn(ctx); err != nil {
		return nil, err
	}
	rawUTXOs, err := g.raw.GetAddressConfirmedUnspent(ctx, address)
	if err != nil {
		return nil, err
	}
	out := make([]UTXO, 0, len(rawUTXOs))
	for _, u := range rawUTXOs {
		out = append(out, UTXO{TxID: u.TxID, Vout: u.Vout, Value: u.Value})
	}
	return out, nil
}

func (g *guardUpstream) GetChainInfo(ctx context.Context) (uint32, error) {
	if err := g.waitTurn(ctx); err != nil {
		return 0, err
	}
	return g.raw.GetChainInfo(ctx)
}

func (g *guardUpstream) PostTxRaw(ctx context.Context, txHex string) (string, error) {
	if err := g.waitTurn(ctx); err != nil {
		return "", err
	}
	return g.raw.PostTxRaw(ctx, txHex)
}

func (g *guardUpstream) GetAddressConfirmedHistory(ctx context.Context, address string) ([]AddressHistoryItem, error) {
	if err := g.waitTurn(ctx); err != nil {
		return nil, err
	}
	rawItems, err := g.raw.GetAddressConfirmedHistory(ctx, address)
	if err != nil {
		return nil, err
	}
	out := make([]AddressHistoryItem, 0, len(rawItems))
	for _, it := range rawItems {
		out = append(out, AddressHistoryItem{TxID: it.TxID, Height: it.Height})
	}
	return out, nil
}

func (g *guardUpstream) GetAddressConfirmedHistoryPage(ctx context.Context, address string, q ConfirmedHistoryQuery) (ConfirmedHistoryPage, error) {
	if err := g.waitTurn(ctx); err != nil {
		return ConfirmedHistoryPage{}, err
	}
	rawPage, err := g.raw.GetAddressConfirmedHistoryPage(ctx, address, wocraw.ConfirmedHistoryQuery{
		Order:  q.Order,
		Limit:  q.Limit,
		Height: q.Height,
		Token:  q.Token,
	})
	if err != nil {
		return ConfirmedHistoryPage{}, err
	}
	out := ConfirmedHistoryPage{
		Items:         make([]AddressHistoryItem, 0, len(rawPage.Items)),
		NextPageToken: strings.TrimSpace(rawPage.NextPageToken),
	}
	for _, item := range rawPage.Items {
		out.Items = append(out.Items, AddressHistoryItem{TxID: item.TxID, Height: item.Height})
	}
	return out, nil
}

func (g *guardUpstream) GetAddressUnconfirmedHistory(ctx context.Context, address string) ([]string, error) {
	if err := g.waitTurn(ctx); err != nil {
		return nil, err
	}
	return g.raw.GetAddressUnconfirmedHistory(ctx, address)
}

func (g *guardUpstream) GetTxHash(ctx context.Context, txid string) (TxDetail, error) {
	if err := g.waitTurn(ctx); err != nil {
		return TxDetail{}, err
	}
	rawTx, err := g.raw.GetTxHash(ctx, txid)
	if err != nil {
		return TxDetail{}, err
	}
	out := TxDetail{
		TxID: strings.TrimSpace(rawTx.TxID),
		Vin:  make([]TxInput, 0, len(rawTx.Vin)),
		Vout: make([]TxOutput, 0, len(rawTx.Vout)),
	}
	for _, in := range rawTx.Vin {
		out.Vin = append(out.Vin, TxInput{TxID: in.TxID, Vout: in.Vout})
	}
	for _, vout := range rawTx.Vout {
		out.Vout = append(out.Vout, TxOutput{
			N:     vout.N,
			Value: vout.Value,
			ScriptPubKey: ScriptPubKey{
				Hex: vout.ScriptPubKey.Hex,
			},
		})
	}
	return out, nil
}

func (g *guardUpstream) SnapshotStats() GuardStats {
	calls := g.callCount.Load()
	waited := g.waitedCall.Load()
	total := time.Duration(g.waitTotalNanos.Load())
	avg := time.Duration(0)
	if calls > 0 {
		avg = total / time.Duration(calls)
	}
	return GuardStats{
		Calls:       calls,
		WaitedCalls: waited,
		WaitTotal:   total,
		WaitAvg:     avg,
	}
}

func (g *guardUpstream) waitTurn(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	g.callCount.Add(1)

	waitUntil := g.reserveSlot()
	waitDur := time.Until(waitUntil)
	if waitDur <= 0 {
		return nil
	}
	g.waitedCall.Add(1)
	g.waitTotalNanos.Add(waitDur.Nanoseconds())

	timer := time.NewTimer(waitDur)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// reserveSlot 只在临界区中预约时间片，不在锁内阻塞等待。
func (g *guardUpstream) reserveSlot() time.Time {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now()
	slot := now
	if !g.nextAllowed.IsZero() && g.nextAllowed.After(now) {
		slot = g.nextAllowed
	}
	g.nextAllowed = slot.Add(g.interval)
	return slot
}
