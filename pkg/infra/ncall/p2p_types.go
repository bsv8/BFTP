package ncall

const (
	ContentTypeProto = "application/x-protobuf"
)

type CallReq struct {
	To          string `protobuf:"bytes,1,opt,name=to,proto3" json:"to"`
	Route       string `protobuf:"bytes,2,opt,name=route,proto3" json:"route"`
	ContentType string `protobuf:"bytes,3,opt,name=content_type,json=contentType,proto3" json:"content_type"`
	Body        []byte `protobuf:"bytes,4,opt,name=body,proto3" json:"body,omitempty"`

	PaymentScheme  string `protobuf:"bytes,5,opt,name=payment_scheme,json=paymentScheme,proto3" json:"payment_scheme,omitempty"`
	PaymentPayload []byte `protobuf:"bytes,6,opt,name=payment_payload,json=paymentPayload,proto3" json:"payment_payload,omitempty"`
}

type PaymentOption struct {
	Scheme                   string `protobuf:"bytes,1,opt,name=scheme,proto3" json:"scheme"`
	PaymentDomain            string `protobuf:"bytes,2,opt,name=payment_domain,json=paymentDomain,proto3" json:"payment_domain,omitempty"`
	AmountSatoshi            uint64 `protobuf:"varint,3,opt,name=amount_satoshi,json=amountSatoshi,proto3" json:"amount_satoshi,omitempty"`
	Description              string `protobuf:"bytes,4,opt,name=description,proto3" json:"description,omitempty"`
	MinimumPoolAmountSatoshi uint64 `protobuf:"varint,5,opt,name=minimum_pool_amount_satoshi,json=minimumPoolAmountSatoshi,proto3" json:"minimum_pool_amount_satoshi,omitempty"`
	FeeRateSatPerByteMilli   uint64 `protobuf:"varint,6,opt,name=fee_rate_sat_per_byte_milli,json=feeRateSatPerByteMilli,proto3" json:"fee_rate_sat_per_byte_milli,omitempty"`
	SingleCycleFeeSatoshi    uint64 `protobuf:"varint,7,opt,name=single_cycle_fee_satoshi,json=singleCycleFeeSatoshi,proto3" json:"single_cycle_fee_satoshi,omitempty"`
	SingleQueryFeeSatoshi    uint64 `protobuf:"varint,8,opt,name=single_query_fee_satoshi,json=singleQueryFeeSatoshi,proto3" json:"single_query_fee_satoshi,omitempty"`
	SinglePublishFeeSatoshi  uint64 `protobuf:"varint,9,opt,name=single_publish_fee_satoshi,json=singlePublishFeeSatoshi,proto3" json:"single_publish_fee_satoshi,omitempty"`
	LockBlocks               uint32 `protobuf:"varint,10,opt,name=lock_blocks,json=lockBlocks,proto3" json:"lock_blocks,omitempty"`
	PricingMode              string `protobuf:"bytes,11,opt,name=pricing_mode,json=pricingMode,proto3" json:"pricing_mode,omitempty"`
	ServiceQuantity          uint64 `protobuf:"varint,12,opt,name=service_quantity,json=serviceQuantity,proto3" json:"service_quantity,omitempty"`
	ServiceQuantityUnit      string `protobuf:"bytes,13,opt,name=service_quantity_unit,json=serviceQuantityUnit,proto3" json:"service_quantity_unit,omitempty"`
	QuoteStatus              string `protobuf:"bytes,14,opt,name=quote_status,json=quoteStatus,proto3" json:"quote_status,omitempty"`
}

type CallResp struct {
	Ok          bool   `protobuf:"varint,1,opt,name=ok,proto3" json:"ok"`
	Code        string `protobuf:"bytes,2,opt,name=code,proto3" json:"code,omitempty"`
	Message     string `protobuf:"bytes,3,opt,name=message,proto3" json:"message,omitempty"`
	ContentType string `protobuf:"bytes,4,opt,name=content_type,json=contentType,proto3" json:"content_type,omitempty"`
	Body        []byte `protobuf:"bytes,5,opt,name=body,proto3" json:"body,omitempty"`

	PaymentSchemes       []*PaymentOption `protobuf:"bytes,6,rep,name=payment_schemes,json=paymentSchemes,proto3" json:"payment_schemes,omitempty"`
	PaymentReceiptScheme string           `protobuf:"bytes,7,opt,name=payment_receipt_scheme,json=paymentReceiptScheme,proto3" json:"payment_receipt_scheme,omitempty"`
	PaymentReceipt       []byte           `protobuf:"bytes,8,opt,name=payment_receipt,json=paymentReceipt,proto3" json:"payment_receipt,omitempty"`
	ServiceQuote         []byte           `protobuf:"bytes,9,opt,name=service_quote,json=serviceQuote,proto3" json:"service_quote,omitempty"`
	ServiceReceipt       []byte           `protobuf:"bytes,10,opt,name=service_receipt,json=serviceReceipt,proto3" json:"service_receipt,omitempty"`
}

type ResolveReq struct {
	To    string `protobuf:"bytes,1,opt,name=to,proto3" json:"to"`
	Route string `protobuf:"bytes,2,opt,name=route,proto3" json:"route,omitempty"`
}

type ResolveResp struct {
	Ok          bool   `protobuf:"varint,1,opt,name=ok,proto3" json:"ok"`
	Code        string `protobuf:"bytes,2,opt,name=code,proto3" json:"code,omitempty"`
	Message     string `protobuf:"bytes,3,opt,name=message,proto3" json:"message,omitempty"`
	ContentType string `protobuf:"bytes,4,opt,name=content_type,json=contentType,proto3" json:"content_type,omitempty"`
	Body        []byte `protobuf:"bytes,5,opt,name=body,proto3" json:"body,omitempty"`
}

type FeePool2of2Payment struct {
	SpendTxID           string `protobuf:"bytes,1,opt,name=spend_txid,json=spendTxid,proto3" json:"spend_txid"`
	SequenceNumber      uint32 `protobuf:"varint,2,opt,name=sequence_number,json=sequenceNumber,proto3" json:"sequence_number"`
	ServerAmount        uint64 `protobuf:"varint,3,opt,name=server_amount,json=serverAmount,proto3" json:"server_amount"`
	ChargeAmountSatoshi uint64 `protobuf:"varint,4,opt,name=charge_amount_satoshi,json=chargeAmountSatoshi,proto3" json:"charge_amount_satoshi"`
	Fee                 uint64 `protobuf:"varint,5,opt,name=fee,proto3" json:"fee"`
	ClientSignature     []byte `protobuf:"bytes,6,opt,name=client_signature,json=clientSignature,proto3" json:"client_signature"`
	ChargeReason        string `protobuf:"bytes,7,opt,name=charge_reason,json=chargeReason,proto3" json:"charge_reason,omitempty"`
	ProofIntent         []byte `protobuf:"bytes,8,opt,name=proof_intent,json=proofIntent,proto3" json:"proof_intent,omitempty"`
	SignedProofCommit   []byte `protobuf:"bytes,9,opt,name=signed_proof_commit,json=signedProofCommit,proto3" json:"signed_proof_commit,omitempty"`
	ServiceQuote        []byte `protobuf:"bytes,11,opt,name=service_quote,json=serviceQuote,proto3" json:"service_quote,omitempty"`
}

type FeePool2of2Receipt struct {
	AcceptedAtUnix        int64  `protobuf:"varint,1,opt,name=accepted_at_unix,json=acceptedAtUnix,proto3" json:"accepted_at_unix,omitempty"`
	ChargedAmountSatoshi  uint64 `protobuf:"varint,2,opt,name=charged_amount_satoshi,json=chargedAmountSatoshi,proto3" json:"charged_amount_satoshi,omitempty"`
	UpdatedTxID           string `protobuf:"bytes,3,opt,name=updated_txid,json=updatedTxid,proto3" json:"updated_txid,omitempty"`
	MergedCurrentTx       []byte `protobuf:"bytes,4,opt,name=merged_current_tx,json=mergedCurrentTx,proto3" json:"merged_current_tx,omitempty"`
	ProofStatePayload     []byte `protobuf:"bytes,5,opt,name=proof_state_payload,json=proofStatePayload,proto3" json:"proof_state_payload,omitempty"`
}

type ChainTxV1Payment struct {
	RawTx []byte `protobuf:"bytes,1,opt,name=raw_tx,json=rawTx,proto3" json:"raw_tx,omitempty"`
}

type ChainTxV1Receipt struct {
	AcceptedAtUnix int64  `protobuf:"varint,1,opt,name=accepted_at_unix,json=acceptedAtUnix,proto3" json:"accepted_at_unix,omitempty"`
	PaymentTxID    string `protobuf:"bytes,2,opt,name=payment_txid,json=paymentTxid,proto3" json:"payment_txid,omitempty"`
}

type CapabilityItem struct {
	// 公开能力快照只负责声明“节点对外提供什么能力”。
	// 价格和支付细节统一交给 offer -> quote -> pay -> receipt 流程。
	ID      string `protobuf:"bytes,1,opt,name=id,proto3" json:"id"`
	Version uint32 `protobuf:"varint,2,opt,name=version,proto3" json:"version"`
}

type CapabilitiesShowBody struct {
	NodePubkeyHex string            `protobuf:"bytes,1,opt,name=node_pubkey_hex,json=nodePubkeyHex,proto3" json:"node_pubkey_hex"`
	Capabilities  []*CapabilityItem `protobuf:"bytes,2,rep,name=capabilities,proto3" json:"capabilities"`
}
