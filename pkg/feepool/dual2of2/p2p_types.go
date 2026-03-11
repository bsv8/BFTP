package dual2of2

// InfoReq/InfoResp 用于 client 启动时获取网关的费用池握手参数。
// 说明：client_id 规范为 secp256k1 压缩公钥 hex（33 字节，02/03 开头，小写）。
// 网关会兼容旧输入（libp2p MarshalPublicKey hex）并在落库前归一化到上述格式。
type InfoReq struct {
	ClientID string `protobuf:"bytes,1,opt,name=client_id,json=clientId,proto3" json:"client_id"`
}

type InfoResp struct {
	Status string `protobuf:"bytes,1,opt,name=status,proto3" json:"status"`

	MinimumPoolAmountSatoshi uint64  `protobuf:"varint,2,opt,name=minimum_pool_amount_satoshi,json=minimumPoolAmountSatoshi,proto3" json:"minimum_pool_amount_satoshi"`
	LockBlocks               uint32  `protobuf:"varint,3,opt,name=lock_blocks,json=lockBlocks,proto3" json:"lock_blocks"`
	FeeRateSatPerByte        float64 `protobuf:"fixed64,4,opt,name=fee_rate_sat_per_byte,json=feeRateSatPerByte,proto3" json:"fee_rate_sat_per_byte"`

	BillingCycleSeconds      uint32 `protobuf:"varint,5,opt,name=billing_cycle_seconds,json=billingCycleSeconds,proto3" json:"billing_cycle_seconds"`
	SingleCycleFeeSatoshi    uint64 `protobuf:"varint,6,opt,name=single_cycle_fee_satoshi,json=singleCycleFeeSatoshi,proto3" json:"single_cycle_fee_satoshi"`
	SinglePublishFeeSatoshi  uint64 `protobuf:"varint,7,opt,name=single_publish_fee_satoshi,json=singlePublishFeeSatoshi,proto3" json:"single_publish_fee_satoshi"`
	RenewNotifyBeforeSeconds uint32 `protobuf:"varint,8,opt,name=renew_notify_before_seconds,json=renewNotifyBeforeSeconds,proto3" json:"renew_notify_before_seconds"`
}

type CreateReq struct {
	ClientID string `protobuf:"bytes,1,opt,name=client_id,json=clientId,proto3" json:"client_id"`

	SpendTx        []byte `protobuf:"bytes,2,opt,name=spend_tx,json=spendTx,proto3" json:"spend_tx"`
	InputAmount    uint64 `protobuf:"varint,3,opt,name=input_amount,json=inputAmount,proto3" json:"input_amount"`
	SequenceNumber uint32 `protobuf:"varint,4,opt,name=sequence_number,json=sequenceNumber,proto3" json:"sequence_number"`
	ServerAmount   uint64 `protobuf:"varint,5,opt,name=server_amount,json=serverAmount,proto3" json:"server_amount"`
	ClientSig      []byte `protobuf:"bytes,6,opt,name=client_sig,json=clientSig,proto3" json:"client_signature"`
}

type CreateResp struct {
	SpendTxID     string `protobuf:"bytes,1,opt,name=spend_txid,json=spendTxid,proto3" json:"spend_txid"`
	ServerSig     []byte `protobuf:"bytes,2,opt,name=server_sig,json=serverSig,proto3" json:"server_signature"`
	ErrorMessage  string `protobuf:"bytes,3,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
	SpendTxFeeSat uint64 `protobuf:"varint,4,opt,name=spend_tx_fee_satoshi,json=spendTxFeeSatoshi,proto3" json:"spend_tx_fee_satoshi"`
	PoolAmountSat uint64 `protobuf:"varint,5,opt,name=pool_amount_satoshi,json=poolAmountSatoshi,proto3" json:"pool_amount_satoshi"`
}

type BaseTxReq struct {
	ClientID string `protobuf:"bytes,1,opt,name=client_id,json=clientId,proto3" json:"client_id"`

	SpendTxID string `protobuf:"bytes,2,opt,name=spend_txid,json=spendTxid,proto3" json:"spend_txid"`
	BaseTx    []byte `protobuf:"bytes,3,opt,name=base_tx,json=baseTx,proto3" json:"base_tx"`
	ClientSig []byte `protobuf:"bytes,4,opt,name=client_sig,json=clientSig,proto3" json:"client_signature"`
}

type BaseTxResp struct {
	Success  bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success"`
	Status   string `protobuf:"bytes,2,opt,name=status,proto3" json:"status"`
	BaseTxID string `protobuf:"bytes,3,opt,name=base_txid,json=baseTxid,proto3" json:"base_txid,omitempty"`
	Error    string `protobuf:"bytes,4,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
}

type PayConfirmReq struct {
	ClientID string `protobuf:"bytes,1,opt,name=client_id,json=clientId,proto3" json:"client_id"`

	SpendTxID      string `protobuf:"bytes,2,opt,name=spend_txid,json=spendTxid,proto3" json:"spend_txid"`
	SequenceNumber uint32 `protobuf:"varint,3,opt,name=sequence_number,json=sequenceNumber,proto3" json:"sequence_number"`
	ServerAmount   uint64 `protobuf:"varint,4,opt,name=server_amount,json=serverAmount,proto3" json:"server_amount"`
	Fee            uint64 `protobuf:"varint,5,opt,name=fee,proto3" json:"fee"`
	ClientSig      []byte `protobuf:"bytes,6,opt,name=client_sig,json=clientSig,proto3" json:"signature"`

	// 业务字段（不上链，仅用于观测/审计/幂等核对）。
	ChargeReason        string `protobuf:"bytes,7,opt,name=charge_reason,json=chargeReason,proto3" json:"charge_reason,omitempty"`
	ChargeAmountSatoshi uint64 `protobuf:"varint,8,opt,name=charge_amount_satoshi,json=chargeAmountSatoshi,proto3" json:"charge_amount_satoshi,omitempty"`
	FileHash            string `protobuf:"bytes,9,opt,name=file_hash,json=fileHash,proto3" json:"file_hash,omitempty"`
}

type PayConfirmResp struct {
	Success     bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success"`
	Status      string `protobuf:"bytes,2,opt,name=status,proto3" json:"status"`
	UpdatedTxID string `protobuf:"bytes,3,opt,name=updated_txid,json=updatedTxid,proto3" json:"updated_txid,omitempty"`

	Sequence     uint32 `protobuf:"varint,4,opt,name=sequence,proto3" json:"sequence,omitempty"`
	ServerAmount uint64 `protobuf:"varint,5,opt,name=server_amount,json=serverAmount,proto3" json:"server_amount,omitempty"`
	ClientAmount uint64 `protobuf:"varint,6,opt,name=client_amount,json=clientAmount,proto3" json:"client_amount,omitempty"`

	Error string `protobuf:"bytes,7,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
}

type CloseReq struct {
	ClientID string `protobuf:"bytes,1,opt,name=client_id,json=clientId,proto3" json:"client_id"`

	SpendTxID    string `protobuf:"bytes,2,opt,name=spend_txid,json=spendTxid,proto3" json:"spend_txid"`
	ServerAmount uint64 `protobuf:"varint,3,opt,name=server_amount,json=serverAmount,proto3" json:"server_amount"`
	Fee          uint64 `protobuf:"varint,4,opt,name=fee,proto3" json:"fee"`
	ClientSig    []byte `protobuf:"bytes,5,opt,name=client_sig,json=clientSig,proto3" json:"signature"`
}

type CloseResp struct {
	Success        bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success"`
	Status         string `protobuf:"bytes,2,opt,name=status,proto3" json:"status"`
	Broadcasted    bool   `protobuf:"varint,3,opt,name=broadcasted,proto3" json:"broadcasted"`
	FinalSpendTxID string `protobuf:"bytes,4,opt,name=final_spend_txid,json=finalSpendTxid,proto3" json:"final_spend_txid,omitempty"`
	Error          string `protobuf:"bytes,5,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
}

type StateReq struct {
	ClientID  string `protobuf:"bytes,1,opt,name=client_id,json=clientId,proto3" json:"client_id"`
	SpendTxID string `protobuf:"bytes,2,opt,name=spend_txid,json=spendTxid,proto3" json:"spend_txid,omitempty"`
}

type StateResp struct {
	Status string `protobuf:"bytes,1,opt,name=status,proto3" json:"status"`

	SpendTxID string `protobuf:"bytes,2,opt,name=spend_txid,json=spendTxid,proto3" json:"spend_txid,omitempty"`
	BaseTxID  string `protobuf:"bytes,3,opt,name=base_txid,json=baseTxid,proto3" json:"base_txid,omitempty"`
	FinalTxID string `protobuf:"bytes,4,opt,name=final_txid,json=finalTxid,proto3" json:"final_txid,omitempty"`
	CurrentTx []byte `protobuf:"bytes,5,opt,name=current_tx,json=currentTx,proto3" json:"current_tx,omitempty"`

	PoolAmountSat   uint64 `protobuf:"varint,6,opt,name=pool_amount_satoshi,json=poolAmountSatoshi,proto3" json:"pool_amount_satoshi"`
	SpendTxFeeSat   uint64 `protobuf:"varint,7,opt,name=spend_tx_fee_satoshi,json=spendTxFeeSatoshi,proto3" json:"spend_tx_fee_satoshi"`
	Sequence        uint32 `protobuf:"varint,8,opt,name=sequence,proto3" json:"sequence"`
	ServerAmountSat uint64 `protobuf:"varint,9,opt,name=server_amount_satoshi,json=serverAmountSatoshi,proto3" json:"server_amount_satoshi"`
	ClientAmountSat uint64 `protobuf:"varint,10,opt,name=client_amount_satoshi,json=clientAmountSatoshi,proto3" json:"client_amount_satoshi"`
}

// DemandPublishPaidReq/Resp 是“发布广播 + 扣费”的组合接口。
// 注意：demand 仍落在 dealprod 的 demands 表里；扣费走费用池（spend tx update）。
type DemandPublishPaidReq struct {
	ClientID string `protobuf:"bytes,1,opt,name=client_id,json=clientId,proto3" json:"client_id"`

	SeedHash   string   `protobuf:"bytes,2,opt,name=seed_hash,json=seedHash,proto3" json:"seed_hash"`
	ChunkCount uint32   `protobuf:"varint,3,opt,name=chunk_count,json=chunkCount,proto3" json:"chunk_count"`
	BuyerAddrs []string `protobuf:"bytes,4,rep,name=buyer_addrs,json=buyerAddrs,proto3" json:"buyer_addrs,omitempty"`
	SpendTxID  string   `protobuf:"bytes,5,opt,name=spend_txid,json=spendTxid,proto3" json:"spend_txid"`
	FileHash   string   `protobuf:"bytes,6,opt,name=file_hash,json=fileHash,proto3" json:"file_hash,omitempty"` // 兼容：若你发布的是 seed_hash，这里可为空

	SequenceNumber      uint32 `protobuf:"varint,7,opt,name=sequence_number,json=sequenceNumber,proto3" json:"sequence_number"`
	ServerAmount        uint64 `protobuf:"varint,8,opt,name=server_amount,json=serverAmount,proto3" json:"server_amount"`
	ChargeAmountSatoshi uint64 `protobuf:"varint,9,opt,name=charge_amount_satoshi,json=chargeAmountSatoshi,proto3" json:"charge_amount_satoshi"`
	Fee                 uint64 `protobuf:"varint,10,opt,name=fee,proto3" json:"fee"`
	ClientSignature     []byte `protobuf:"bytes,11,opt,name=client_signature,json=clientSignature,proto3" json:"signature"`
	ChargeReason        string `protobuf:"bytes,12,opt,name=charge_reason,json=chargeReason,proto3" json:"charge_reason,omitempty"` // 默认 demand_publish_fee
}

type DemandPublishPaidResp struct {
	Success       bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success"`
	Status        string `protobuf:"bytes,2,opt,name=status,proto3" json:"status"`
	DemandID      string `protobuf:"bytes,3,opt,name=demand_id,json=demandId,proto3" json:"demand_id,omitempty"`
	Published     bool   `protobuf:"varint,4,opt,name=published,proto3" json:"published"`
	ChargedAmount uint64 `protobuf:"varint,5,opt,name=charged_amount_satoshi,json=chargedAmountSatoshi,proto3" json:"charged_amount_satoshi,omitempty"`
	UpdatedTxID   string `protobuf:"bytes,6,opt,name=updated_txid,json=updatedTxid,proto3" json:"updated_txid,omitempty"`
	Error         string `protobuf:"bytes,7,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
}

// LiveDemandPublishPaidReq/Resp 是“直播需求广播 + 扣费”的组合接口。
// 设计说明：
// - 网关只负责付费发布和广播 live demand；
// - 真正的直播 segment 交易仍走后续 c2c。
type LiveDemandPublishPaidReq struct {
	ClientID string `protobuf:"bytes,1,opt,name=client_id,json=clientId,proto3" json:"client_id"`

	StreamID         string   `protobuf:"bytes,2,opt,name=stream_id,json=streamId,proto3" json:"stream_id"`
	HaveSegmentIndex int64    `protobuf:"varint,3,opt,name=have_segment_index,json=haveSegmentIndex,proto3" json:"have_segment_index"`
	Window           uint32   `protobuf:"varint,4,opt,name=window,proto3" json:"window"`
	BuyerAddrs       []string `protobuf:"bytes,5,rep,name=buyer_addrs,json=buyerAddrs,proto3" json:"buyer_addrs,omitempty"`
	SpendTxID        string   `protobuf:"bytes,6,opt,name=spend_txid,json=spendTxid,proto3" json:"spend_txid"`

	SequenceNumber      uint32 `protobuf:"varint,7,opt,name=sequence_number,json=sequenceNumber,proto3" json:"sequence_number"`
	ServerAmount        uint64 `protobuf:"varint,8,opt,name=server_amount,json=serverAmount,proto3" json:"server_amount"`
	ChargeAmountSatoshi uint64 `protobuf:"varint,9,opt,name=charge_amount_satoshi,json=chargeAmountSatoshi,proto3" json:"charge_amount_satoshi"`
	Fee                 uint64 `protobuf:"varint,10,opt,name=fee,proto3" json:"fee"`
	ClientSignature     []byte `protobuf:"bytes,11,opt,name=client_signature,json=clientSignature,proto3" json:"signature"`
	ChargeReason        string `protobuf:"bytes,12,opt,name=charge_reason,json=chargeReason,proto3" json:"charge_reason,omitempty"` // 默认 live_demand_publish_fee
}

type LiveDemandPublishPaidResp struct {
	Success       bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success"`
	Status        string `protobuf:"bytes,2,opt,name=status,proto3" json:"status"`
	DemandID      string `protobuf:"bytes,3,opt,name=demand_id,json=demandId,proto3" json:"demand_id,omitempty"`
	Published     bool   `protobuf:"varint,4,opt,name=published,proto3" json:"published"`
	ChargedAmount uint64 `protobuf:"varint,5,opt,name=charged_amount_satoshi,json=chargedAmountSatoshi,proto3" json:"charged_amount_satoshi,omitempty"`
	UpdatedTxID   string `protobuf:"bytes,6,opt,name=updated_txid,json=updatedTxid,proto3" json:"updated_txid,omitempty"`
	Error         string `protobuf:"bytes,7,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
}
