package broadcast

// DemandPublishPaidReq/Resp 是“发布广播 + 扣费”的组合接口。
// 注意：demand 仍落在 dealprod 的 demands 表里；扣费走费用池（spend tx update）。
type DemandPublishPaidReq struct {
	ClientID string `protobuf:"bytes,1,opt,name=client_pubkey_hex,json=clientId,proto3" json:"client_pubkey_hex"`

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
	ChargeReason        string `protobuf:"bytes,12,opt,name=charge_reason,json=chargeReason,proto3" json:"charge_reason,omitempty"` // 默认 QuoteServiceTypeDemandPublish
	ProofIntent         []byte `protobuf:"bytes,13,opt,name=proof_intent,json=proofIntent,proto3" json:"proof_intent,omitempty"`
	SignedProofCommit   []byte `protobuf:"bytes,14,opt,name=signed_proof_commit,json=signedProofCommit,proto3" json:"signed_proof_commit,omitempty"`
	ServiceQuote        []byte `protobuf:"bytes,16,opt,name=service_quote,json=serviceQuote,proto3" json:"service_quote,omitempty"`
}

type DemandPublishPaidResp struct {
	Success           bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success"`
	Status            string `protobuf:"bytes,2,opt,name=status,proto3" json:"status"`
	DemandID          string `protobuf:"bytes,3,opt,name=demand_id,json=demandId,proto3" json:"demand_id,omitempty"`
	Published         bool   `protobuf:"varint,4,opt,name=published,proto3" json:"published"`
	ChargedAmount     uint64 `protobuf:"varint,5,opt,name=charged_amount_satoshi,json=chargedAmountSatoshi,proto3" json:"charged_amount_satoshi,omitempty"`
	UpdatedTxID       string `protobuf:"bytes,6,opt,name=updated_txid,json=updatedTxid,proto3" json:"updated_txid,omitempty"`
	Error             string `protobuf:"bytes,7,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
	MergedCurrentTx   []byte `protobuf:"bytes,8,opt,name=merged_current_tx,json=mergedCurrentTx,proto3" json:"merged_current_tx,omitempty"`
	ProofStatePayload []byte `protobuf:"bytes,9,opt,name=proof_state_payload,json=proofStatePayload,proto3" json:"proof_state_payload,omitempty"`
	ServiceReceipt    []byte `protobuf:"bytes,10,opt,name=service_receipt,json=serviceReceipt,proto3" json:"service_receipt,omitempty"`
}

// DemandPublishBatchPaidReq/Resp 是“批量发布静态需求 + 扣费”的组合接口。
// 设计说明：
// - 浏览器打开 HTML 后会发现很多 hash 子资源，应该一次付费把这批 demand 发出去；
// - 网关只收一次 publish fee，但会为每个资源写出独立 demand_id，卖家仍然按单资源报价；
// - 这样保留了原有 c2c 报价模型，同时把浏览器静态资源启动时延压下去。
type DemandPublishBatchPaidItem struct {
	SeedHash   string `protobuf:"bytes,1,opt,name=seed_hash,json=seedHash,proto3" json:"seed_hash"`
	ChunkCount uint32 `protobuf:"varint,2,opt,name=chunk_count,json=chunkCount,proto3" json:"chunk_count"`
}

type DemandPublishBatchPaidReq struct {
	ClientID string `protobuf:"bytes,1,opt,name=client_pubkey_hex,json=clientId,proto3" json:"client_pubkey_hex"`

	Items      []*DemandPublishBatchPaidItem `protobuf:"bytes,2,rep,name=items,proto3" json:"items,omitempty"`
	BuyerAddrs []string                      `protobuf:"bytes,3,rep,name=buyer_addrs,json=buyerAddrs,proto3" json:"buyer_addrs,omitempty"`
	SpendTxID  string                        `protobuf:"bytes,4,opt,name=spend_txid,json=spendTxid,proto3" json:"spend_txid"`

	SequenceNumber      uint32 `protobuf:"varint,5,opt,name=sequence_number,json=sequenceNumber,proto3" json:"sequence_number"`
	ServerAmount        uint64 `protobuf:"varint,6,opt,name=server_amount,json=serverAmount,proto3" json:"server_amount"`
	ChargeAmountSatoshi uint64 `protobuf:"varint,7,opt,name=charge_amount_satoshi,json=chargeAmountSatoshi,proto3" json:"charge_amount_satoshi"`
	Fee                 uint64 `protobuf:"varint,8,opt,name=fee,proto3" json:"fee"`
	ClientSignature     []byte `protobuf:"bytes,9,opt,name=client_signature,json=clientSignature,proto3" json:"signature"`
	ChargeReason        string `protobuf:"bytes,10,opt,name=charge_reason,json=chargeReason,proto3" json:"charge_reason,omitempty"`
	ProofIntent         []byte `protobuf:"bytes,11,opt,name=proof_intent,json=proofIntent,proto3" json:"proof_intent,omitempty"`
	SignedProofCommit   []byte `protobuf:"bytes,12,opt,name=signed_proof_commit,json=signedProofCommit,proto3" json:"signed_proof_commit,omitempty"`
	ServiceQuote        []byte `protobuf:"bytes,14,opt,name=service_quote,json=serviceQuote,proto3" json:"service_quote,omitempty"`
}

type DemandPublishBatchPaidResult struct {
	SeedHash   string `protobuf:"bytes,1,opt,name=seed_hash,json=seedHash,proto3" json:"seed_hash"`
	ChunkCount uint32 `protobuf:"varint,2,opt,name=chunk_count,json=chunkCount,proto3" json:"chunk_count"`
	DemandID   string `protobuf:"bytes,3,opt,name=demand_id,json=demandId,proto3" json:"demand_id,omitempty"`
	Status     string `protobuf:"bytes,4,opt,name=status,proto3" json:"status"`
}

type DemandPublishBatchPaidResp struct {
	Success           bool                            `protobuf:"varint,1,opt,name=success,proto3" json:"success"`
	Status            string                          `protobuf:"bytes,2,opt,name=status,proto3" json:"status"`
	Items             []*DemandPublishBatchPaidResult `protobuf:"bytes,3,rep,name=items,proto3" json:"items,omitempty"`
	PublishedCount    uint32                          `protobuf:"varint,4,opt,name=published_count,json=publishedCount,proto3" json:"published_count,omitempty"`
	ChargedAmount     uint64                          `protobuf:"varint,5,opt,name=charged_amount_satoshi,json=chargedAmountSatoshi,proto3" json:"charged_amount_satoshi,omitempty"`
	UpdatedTxID       string                          `protobuf:"bytes,6,opt,name=updated_txid,json=updatedTxid,proto3" json:"updated_txid,omitempty"`
	Error             string                          `protobuf:"bytes,7,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
	MergedCurrentTx   []byte                          `protobuf:"bytes,8,opt,name=merged_current_tx,json=mergedCurrentTx,proto3" json:"merged_current_tx,omitempty"`
	ProofStatePayload []byte                          `protobuf:"bytes,9,opt,name=proof_state_payload,json=proofStatePayload,proto3" json:"proof_state_payload,omitempty"`
	ServiceReceipt    []byte                          `protobuf:"bytes,10,opt,name=service_receipt,json=serviceReceipt,proto3" json:"service_receipt,omitempty"`
}

// LiveDemandPublishPaidReq/Resp 是“直播需求广播 + 扣费”的组合接口。
// 设计说明：
// - 网关只负责付费发布和广播 live demand；
// - 真正的直播 segment 交易仍走后续 c2c。
type LiveDemandPublishPaidReq struct {
	ClientID string `protobuf:"bytes,1,opt,name=client_pubkey_hex,json=clientId,proto3" json:"client_pubkey_hex"`

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
	ChargeReason        string `protobuf:"bytes,12,opt,name=charge_reason,json=chargeReason,proto3" json:"charge_reason,omitempty"` // 默认 QuoteServiceTypeLiveDemandPublish
	ProofIntent         []byte `protobuf:"bytes,13,opt,name=proof_intent,json=proofIntent,proto3" json:"proof_intent,omitempty"`
	SignedProofCommit   []byte `protobuf:"bytes,14,opt,name=signed_proof_commit,json=signedProofCommit,proto3" json:"signed_proof_commit,omitempty"`
	ServiceQuote        []byte `protobuf:"bytes,16,opt,name=service_quote,json=serviceQuote,proto3" json:"service_quote,omitempty"`
}

type LiveDemandPublishPaidResp struct {
	Success           bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success"`
	Status            string `protobuf:"bytes,2,opt,name=status,proto3" json:"status"`
	DemandID          string `protobuf:"bytes,3,opt,name=demand_id,json=demandId,proto3" json:"demand_id,omitempty"`
	Published         bool   `protobuf:"varint,4,opt,name=published,proto3" json:"published"`
	ChargedAmount     uint64 `protobuf:"varint,5,opt,name=charged_amount_satoshi,json=chargedAmountSatoshi,proto3" json:"charged_amount_satoshi,omitempty"`
	UpdatedTxID       string `protobuf:"bytes,6,opt,name=updated_txid,json=updatedTxid,proto3" json:"updated_txid,omitempty"`
	Error             string `protobuf:"bytes,7,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
	MergedCurrentTx   []byte `protobuf:"bytes,8,opt,name=merged_current_tx,json=mergedCurrentTx,proto3" json:"merged_current_tx,omitempty"`
	ProofStatePayload []byte `protobuf:"bytes,9,opt,name=proof_state_payload,json=proofStatePayload,proto3" json:"proof_state_payload,omitempty"`
	ServiceReceipt    []byte `protobuf:"bytes,10,opt,name=service_receipt,json=serviceReceipt,proto3" json:"service_receipt,omitempty"`
}

// NodeReachabilityAnnouncePaidReq/Resp 是“地址声明发布 + 扣费”的组合接口。
// 设计说明：
// - 地址声明是节点自己对“我当前在哪些 libp2p 地址上可达”的签名声明；
// - gateway 只负责收费、缓存、转发，不能伪造节点主体；
// - 地址声明版本只看 head_height + seq，避免网关自行发明“最佳地址”语义。
type NodeReachabilityAnnouncePaidReq struct {
	ClientID string `protobuf:"bytes,1,opt,name=client_pubkey_hex,json=clientId,proto3" json:"client_pubkey_hex"`

	SignedAnnouncement []byte `protobuf:"bytes,2,opt,name=signed_announcement,json=signedAnnouncement,proto3" json:"signed_announcement"`

	SpendTxID           string `protobuf:"bytes,3,opt,name=spend_txid,json=spendTxid,proto3" json:"spend_txid"`
	SequenceNumber      uint32 `protobuf:"varint,4,opt,name=sequence_number,json=sequenceNumber,proto3" json:"sequence_number"`
	ServerAmount        uint64 `protobuf:"varint,5,opt,name=server_amount,json=serverAmount,proto3" json:"server_amount"`
	ChargeAmountSatoshi uint64 `protobuf:"varint,6,opt,name=charge_amount_satoshi,json=chargeAmountSatoshi,proto3" json:"charge_amount_satoshi"`
	Fee                 uint64 `protobuf:"varint,7,opt,name=fee,proto3" json:"fee"`
	ClientSignature     []byte `protobuf:"bytes,8,opt,name=client_signature,json=clientSignature,proto3" json:"client_signature"`
	ChargeReason        string `protobuf:"bytes,9,opt,name=charge_reason,json=chargeReason,proto3" json:"charge_reason,omitempty"`
	ProofIntent         []byte `protobuf:"bytes,10,opt,name=proof_intent,json=proofIntent,proto3" json:"proof_intent,omitempty"`
	SignedProofCommit   []byte `protobuf:"bytes,11,opt,name=signed_proof_commit,json=signedProofCommit,proto3" json:"signed_proof_commit,omitempty"`
	ServiceQuote        []byte `protobuf:"bytes,13,opt,name=service_quote,json=serviceQuote,proto3" json:"service_quote,omitempty"`
}

type NodeReachabilityAnnouncePaidResp struct {
	Success           bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success"`
	Status            string `protobuf:"bytes,2,opt,name=status,proto3" json:"status"`
	Published         bool   `protobuf:"varint,3,opt,name=published,proto3" json:"published"`
	ChargedAmount     uint64 `protobuf:"varint,4,opt,name=charged_amount_satoshi,json=chargedAmountSatoshi,proto3" json:"charged_amount_satoshi,omitempty"`
	UpdatedTxID       string `protobuf:"bytes,5,opt,name=updated_txid,json=updatedTxid,proto3" json:"updated_txid,omitempty"`
	Error             string `protobuf:"bytes,6,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
	MergedCurrentTx   []byte `protobuf:"bytes,7,opt,name=merged_current_tx,json=mergedCurrentTx,proto3" json:"merged_current_tx,omitempty"`
	ProofStatePayload []byte `protobuf:"bytes,8,opt,name=proof_state_payload,json=proofStatePayload,proto3" json:"proof_state_payload,omitempty"`
	ServiceReceipt    []byte `protobuf:"bytes,9,opt,name=service_receipt,json=serviceReceipt,proto3" json:"service_receipt,omitempty"`
}

// NodeReachabilityQueryPaidReq/Resp 是“地址目录查询 + 扣费”的组合接口。
// 设计说明：
// - 查询扣费看“是否执行了目录查询动作”，与命中与否无关；
// - 返回只给“最新有效声明”，避免 gateway 用主观标准挑“最佳地址”。
type NodeReachabilityQueryPaidReq struct {
	ClientID string `protobuf:"bytes,1,opt,name=client_pubkey_hex,json=clientId,proto3" json:"client_pubkey_hex"`

	TargetNodePubkeyHex string `protobuf:"bytes,2,opt,name=target_node_pubkey_hex,json=targetNodePubkeyHex,proto3" json:"target_node_pubkey_hex"`

	SpendTxID           string `protobuf:"bytes,3,opt,name=spend_txid,json=spendTxid,proto3" json:"spend_txid"`
	SequenceNumber      uint32 `protobuf:"varint,4,opt,name=sequence_number,json=sequenceNumber,proto3" json:"sequence_number"`
	ServerAmount        uint64 `protobuf:"varint,5,opt,name=server_amount,json=serverAmount,proto3" json:"server_amount"`
	ChargeAmountSatoshi uint64 `protobuf:"varint,6,opt,name=charge_amount_satoshi,json=chargeAmountSatoshi,proto3" json:"charge_amount_satoshi"`
	Fee                 uint64 `protobuf:"varint,7,opt,name=fee,proto3" json:"fee"`
	ClientSignature     []byte `protobuf:"bytes,8,opt,name=client_signature,json=clientSignature,proto3" json:"client_signature"`
	ChargeReason        string `protobuf:"bytes,9,opt,name=charge_reason,json=chargeReason,proto3" json:"charge_reason,omitempty"`
	ProofIntent         []byte `protobuf:"bytes,10,opt,name=proof_intent,json=proofIntent,proto3" json:"proof_intent,omitempty"`
	SignedProofCommit   []byte `protobuf:"bytes,11,opt,name=signed_proof_commit,json=signedProofCommit,proto3" json:"signed_proof_commit,omitempty"`
	ServiceQuote        []byte `protobuf:"bytes,13,opt,name=service_quote,json=serviceQuote,proto3" json:"service_quote,omitempty"`
}

type NodeReachabilityQueryPaidResp struct {
	Success       bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success"`
	Status        string `protobuf:"bytes,2,opt,name=status,proto3" json:"status"`
	Found         bool   `protobuf:"varint,3,opt,name=found,proto3" json:"found"`
	ChargedAmount uint64 `protobuf:"varint,4,opt,name=charged_amount_satoshi,json=chargedAmountSatoshi,proto3" json:"charged_amount_satoshi,omitempty"`
	UpdatedTxID   string `protobuf:"bytes,5,opt,name=updated_txid,json=updatedTxid,proto3" json:"updated_txid,omitempty"`
	Error         string `protobuf:"bytes,6,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`

	TargetNodePubkeyHex string `protobuf:"bytes,7,opt,name=target_node_pubkey_hex,json=targetNodePubkeyHex,proto3" json:"target_node_pubkey_hex,omitempty"`
	SignedAnnouncement  []byte `protobuf:"bytes,8,opt,name=signed_announcement,json=signedAnnouncement,proto3" json:"signed_announcement,omitempty"`
	MergedCurrentTx     []byte `protobuf:"bytes,9,opt,name=merged_current_tx,json=mergedCurrentTx,proto3" json:"merged_current_tx,omitempty"`
	ProofStatePayload   []byte `protobuf:"bytes,10,opt,name=proof_state_payload,json=proofStatePayload,proto3" json:"proof_state_payload,omitempty"`
	ServiceReceipt      []byte `protobuf:"bytes,11,opt,name=service_receipt,json=serviceReceipt,proto3" json:"service_receipt,omitempty"`
}
