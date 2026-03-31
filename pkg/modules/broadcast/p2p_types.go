package broadcast

// DemandPublishReq/Resp 是统一 node.call 路由里的纯业务请求体。
// 设计说明：
// - 这里不再夹带 2of2 或链上支付字段；
// - 支付统一放到 ncall.CallReq 的 payment_scheme/payment_payload 外壳里；
// - 这样 route body 只表达“我要办什么事”。
type DemandPublishReq struct {
	SeedHash   string   `protobuf:"bytes,1,opt,name=seed_hash,json=seedHash,proto3" json:"seed_hash"`
	ChunkCount uint32   `protobuf:"varint,2,opt,name=chunk_count,json=chunkCount,proto3" json:"chunk_count"`
	BuyerAddrs []string `protobuf:"bytes,3,rep,name=buyer_addrs,json=buyerAddrs,proto3" json:"buyer_addrs,omitempty"`
}

type ListenCycleReq struct {
	RequestedDurationSeconds uint32 `protobuf:"varint,1,opt,name=requested_duration_seconds,json=requestedDurationSeconds,proto3" json:"requested_duration_seconds,omitempty"`
	RequestedUntilUnix       int64  `protobuf:"varint,2,opt,name=requested_until_unix,json=requestedUntilUnix,proto3" json:"requested_until_unix,omitempty"`
	ProposedPaymentSatoshi   uint64 `protobuf:"varint,3,opt,name=proposed_payment_satoshi,json=proposedPaymentSatoshi,proto3" json:"proposed_payment_satoshi,omitempty"`
}

type DemandPublishBatchReq struct {
	Items      []*DemandPublishBatchPaidItem `protobuf:"bytes,1,rep,name=items,proto3" json:"items,omitempty"`
	BuyerAddrs []string                      `protobuf:"bytes,2,rep,name=buyer_addrs,json=buyerAddrs,proto3" json:"buyer_addrs,omitempty"`
}

type LiveDemandPublishReq struct {
	StreamID         string   `protobuf:"bytes,1,opt,name=stream_id,json=streamId,proto3" json:"stream_id"`
	HaveSegmentIndex int64    `protobuf:"varint,2,opt,name=have_segment_index,json=haveSegmentIndex,proto3" json:"have_segment_index"`
	Window           uint32   `protobuf:"varint,3,opt,name=window,proto3" json:"window"`
	BuyerAddrs       []string `protobuf:"bytes,4,rep,name=buyer_addrs,json=buyerAddrs,proto3" json:"buyer_addrs,omitempty"`
}

type NodeReachabilityAnnounceReq struct {
	SignedAnnouncement []byte `protobuf:"bytes,1,opt,name=signed_announcement,json=signedAnnouncement,proto3" json:"signed_announcement"`
}

type NodeReachabilityQueryReq struct {
	TargetNodePubkeyHex string `protobuf:"bytes,1,opt,name=target_node_pubkey_hex,json=targetNodePubkeyHex,proto3" json:"target_node_pubkey_hex"`
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

type ListenCyclePaidResp struct {
	Success                bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success"`
	Status                 string `protobuf:"bytes,2,opt,name=status,proto3" json:"status"`
	ChargedAmount          uint64 `protobuf:"varint,3,opt,name=charged_amount_satoshi,json=chargedAmountSatoshi,proto3" json:"charged_amount_satoshi,omitempty"`
	UpdatedTxID            string `protobuf:"bytes,4,opt,name=updated_txid,json=updatedTxid,proto3" json:"updated_txid,omitempty"`
	GrantedDurationSeconds uint32 `protobuf:"varint,5,opt,name=granted_duration_seconds,json=grantedDurationSeconds,proto3" json:"granted_duration_seconds,omitempty"`
	GrantedUntilUnix       int64  `protobuf:"varint,6,opt,name=granted_until_unix,json=grantedUntilUnix,proto3" json:"granted_until_unix,omitempty"`
	Error                  string `protobuf:"bytes,7,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
	MergedCurrentTx        []byte `protobuf:"bytes,8,opt,name=merged_current_tx,json=mergedCurrentTx,proto3" json:"merged_current_tx,omitempty"`
	ProofStatePayload      []byte `protobuf:"bytes,9,opt,name=proof_state_payload,json=proofStatePayload,proto3" json:"proof_state_payload,omitempty"`
	ServiceReceipt         []byte `protobuf:"bytes,10,opt,name=service_receipt,json=serviceReceipt,proto3" json:"service_receipt,omitempty"`
}

type DemandPublishBatchPaidItem struct {
	SeedHash   string `protobuf:"bytes,1,opt,name=seed_hash,json=seedHash,proto3" json:"seed_hash"`
	ChunkCount uint32 `protobuf:"varint,2,opt,name=chunk_count,json=chunkCount,proto3" json:"chunk_count"`
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
