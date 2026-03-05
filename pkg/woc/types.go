package woc

// 设计约束：
// - 生产组件与工具只应通过 GuardClient 调用链上；
// - 直接访问 WOC 上游的实现被下沉到 internal/wocraw，仅供 guard 服务内部使用。

const DefaultGuardBaseURL = "http://127.0.0.1:18222"

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
