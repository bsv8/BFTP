package dual2of2

// InfoReq/InfoResp 用于 client 启动时获取网关的费用池握手参数。
// 说明：这里的“client_id”沿用现有 p2prpc 语义：必须等于 SignedEnvelope.sender_pubkey（hex，libp2p MarshalPublicKey 格式）。
type InfoReq struct {
	ClientID string `json:"client_id"`
}

type InfoResp struct {
	Status string `json:"status"`

	MinimumPoolAmountSatoshi uint64  `json:"minimum_pool_amount_satoshi"`
	LockBlocks               uint32  `json:"lock_blocks"`
	FeeRateSatPerByte        float64 `json:"fee_rate_sat_per_byte"`

	BillingCycleSeconds      uint32 `json:"billing_cycle_seconds"`
	SingleCycleFeeSatoshi    uint64 `json:"single_cycle_fee_satoshi"`
	SinglePublishFeeSatoshi  uint64 `json:"single_publish_fee_satoshi"`
	RenewNotifyBeforeSeconds uint32 `json:"renew_notify_before_seconds"`
}

type CreateReq struct {
	ClientID string `json:"client_id"`

	SpendTxHex     string `json:"spend_tx_hex"`
	InputAmount    uint64 `json:"input_amount"`
	SequenceNumber uint32 `json:"sequence_number"`
	ServerAmount   uint64 `json:"server_amount"`
	ClientSigHex   string `json:"client_signature"`
}

type CreateResp struct {
	SpendTxID     string `json:"spend_txid"`
	ServerSigHex  string `json:"server_signature"`
	ErrorMessage  string `json:"error_message,omitempty"`
	SpendTxFeeSat uint64 `json:"spend_tx_fee_satoshi"`
	PoolAmountSat uint64 `json:"pool_amount_satoshi"`
}

type BaseTxReq struct {
	ClientID string `json:"client_id"`

	SpendTxID    string `json:"spend_txid"`
	BaseTxHex    string `json:"base_tx_hex"`
	ClientSigHex string `json:"client_signature"`
}

type BaseTxResp struct {
	Success  bool   `json:"success"`
	Status   string `json:"status"`
	BaseTxID string `json:"base_txid,omitempty"`
	Error    string `json:"error_message,omitempty"`
}

type PayConfirmReq struct {
	ClientID string `json:"client_id"`

	SpendTxID      string `json:"spend_txid"`
	SequenceNumber uint32 `json:"sequence_number"`
	ServerAmount   uint64 `json:"server_amount"`
	Fee            uint64 `json:"fee"`
	ClientSigHex   string `json:"signature"`

	// 业务字段（不上链，仅用于观测/审计/幂等核对）。
	ChargeReason        string `json:"charge_reason,omitempty"`
	ChargeAmountSatoshi uint64 `json:"charge_amount_satoshi,omitempty"`
	FileHash            string `json:"file_hash,omitempty"`
}

type PayConfirmResp struct {
	Success     bool   `json:"success"`
	Status      string `json:"status"`
	UpdatedTxID string `json:"updated_txid,omitempty"`

	Sequence     uint32 `json:"sequence,omitempty"`
	ServerAmount uint64 `json:"server_amount,omitempty"`
	ClientAmount uint64 `json:"client_amount,omitempty"`

	Error string `json:"error_message,omitempty"`
}

type CloseReq struct {
	ClientID string `json:"client_id"`

	SpendTxID    string `json:"spend_txid"`
	ServerAmount uint64 `json:"server_amount"`
	Fee          uint64 `json:"fee"`
	ClientSigHex string `json:"signature"`
}

type CloseResp struct {
	Success        bool   `json:"success"`
	Status         string `json:"status"`
	Broadcasted    bool   `json:"broadcasted"`
	FinalSpendTxID string `json:"final_spend_txid,omitempty"`
	Error          string `json:"error_message,omitempty"`
}

type StateReq struct {
	ClientID  string `json:"client_id"`
	SpendTxID string `json:"spend_txid,omitempty"`
}

type StateResp struct {
	Status string `json:"status"`

	SpendTxID    string `json:"spend_txid,omitempty"`
	BaseTxID     string `json:"base_txid,omitempty"`
	FinalTxID    string `json:"final_txid,omitempty"`
	CurrentTxHex string `json:"current_tx_hex,omitempty"`

	PoolAmountSat   uint64 `json:"pool_amount_satoshi"`
	SpendTxFeeSat   uint64 `json:"spend_tx_fee_satoshi"`
	Sequence        uint32 `json:"sequence"`
	ServerAmountSat uint64 `json:"server_amount_satoshi"`
	ClientAmountSat uint64 `json:"client_amount_satoshi"`
}

// DemandPublishPaidReq/Resp 是“发布广播 + 扣费”的组合接口。
// 注意：demand 仍落在 dealprod 的 demands 表里；扣费走费用池（spend tx update）。
type DemandPublishPaidReq struct {
	ClientID string `json:"client_id"`

	SeedHash   string   `json:"seed_hash"`
	ChunkCount uint32   `json:"chunk_count"`
	BuyerAddrs []string `json:"buyer_addrs,omitempty"`
	SpendTxID  string   `json:"spend_txid"`
	FileHash   string   `json:"file_hash,omitempty"` // 兼容：若你发布的是 seed_hash，这里可为空

	SequenceNumber      uint32 `json:"sequence_number"`
	ServerAmount        uint64 `json:"server_amount"`
	ChargeAmountSatoshi uint64 `json:"charge_amount_satoshi"`
	Fee                 uint64 `json:"fee"`
	ClientSignatureHex  string `json:"signature"`
	ChargeReason        string `json:"charge_reason,omitempty"` // 默认 demand_publish_fee
}

type DemandPublishPaidResp struct {
	Success       bool   `json:"success"`
	Status        string `json:"status"`
	DemandID      string `json:"demand_id,omitempty"`
	Published     bool   `json:"published"`
	ChargedAmount uint64 `json:"charged_amount_satoshi,omitempty"`
	UpdatedTxID   string `json:"updated_txid,omitempty"`
	Error         string `json:"error_message,omitempty"`
}
