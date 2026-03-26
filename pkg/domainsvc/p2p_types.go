package domainsvc

type ResolveNamePaidReq struct {
	ClientID string `protobuf:"bytes,1,opt,name=client_pubkey_hex,json=clientId,proto3" json:"client_pubkey_hex"`

	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name"`

	SpendTxID           string `protobuf:"bytes,3,opt,name=spend_txid,json=spendTxid,proto3" json:"spend_txid"`
	SequenceNumber      uint32 `protobuf:"varint,4,opt,name=sequence_number,json=sequenceNumber,proto3" json:"sequence_number"`
	ServerAmount        uint64 `protobuf:"varint,5,opt,name=server_amount,json=serverAmount,proto3" json:"server_amount"`
	ChargeAmountSatoshi uint64 `protobuf:"varint,6,opt,name=charge_amount_satoshi,json=chargeAmountSatoshi,proto3" json:"charge_amount_satoshi"`
	Fee                 uint64 `protobuf:"varint,7,opt,name=fee,proto3" json:"fee"`
	ClientSignature     []byte `protobuf:"bytes,8,opt,name=client_signature,json=clientSignature,proto3" json:"client_signature"`
	ChargeReason        string `protobuf:"bytes,9,opt,name=charge_reason,json=chargeReason,proto3" json:"charge_reason,omitempty"`
}

type ResolveNamePaidResp struct {
	Success          bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success"`
	Status           string `protobuf:"bytes,2,opt,name=status,proto3" json:"status"`
	Name             string `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	OwnerPubkeyHex   string `protobuf:"bytes,4,opt,name=owner_pubkey_hex,json=ownerPubkeyHex,proto3" json:"owner_pubkey_hex,omitempty"`
	TargetPubkeyHex  string `protobuf:"bytes,5,opt,name=target_pubkey_hex,json=targetPubkeyHex,proto3" json:"target_pubkey_hex,omitempty"`
	ExpireAtUnix     int64  `protobuf:"varint,6,opt,name=expire_at_unix,json=expireAtUnix,proto3" json:"expire_at_unix,omitempty"`
	ChargedAmount    uint64 `protobuf:"varint,7,opt,name=charged_amount_satoshi,json=chargedAmountSatoshi,proto3" json:"charged_amount_satoshi,omitempty"`
	UpdatedTxID      string `protobuf:"bytes,8,opt,name=updated_txid,json=updatedTxid,proto3" json:"updated_txid,omitempty"`
	SignedRecordJSON []byte `protobuf:"bytes,9,opt,name=signed_record_json,json=signedRecordJson,proto3" json:"signed_record_json,omitempty"`
	Error            string `protobuf:"bytes,10,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
}

type QueryNamePaidReq struct {
	ClientID string `protobuf:"bytes,1,opt,name=client_pubkey_hex,json=clientId,proto3" json:"client_pubkey_hex"`

	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name"`

	SpendTxID           string `protobuf:"bytes,3,opt,name=spend_txid,json=spendTxid,proto3" json:"spend_txid"`
	SequenceNumber      uint32 `protobuf:"varint,4,opt,name=sequence_number,json=sequenceNumber,proto3" json:"sequence_number"`
	ServerAmount        uint64 `protobuf:"varint,5,opt,name=server_amount,json=serverAmount,proto3" json:"server_amount"`
	ChargeAmountSatoshi uint64 `protobuf:"varint,6,opt,name=charge_amount_satoshi,json=chargeAmountSatoshi,proto3" json:"charge_amount_satoshi"`
	Fee                 uint64 `protobuf:"varint,7,opt,name=fee,proto3" json:"fee"`
	ClientSignature     []byte `protobuf:"bytes,8,opt,name=client_signature,json=clientSignature,proto3" json:"client_signature"`
	ChargeReason        string `protobuf:"bytes,9,opt,name=charge_reason,json=chargeReason,proto3" json:"charge_reason,omitempty"`
}

type QueryNamePaidResp struct {
	Success                  bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success"`
	Status                   string `protobuf:"bytes,2,opt,name=status,proto3" json:"status"`
	Name                     string `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	Available                bool   `protobuf:"varint,4,opt,name=available,proto3" json:"available"`
	Locked                   bool   `protobuf:"varint,5,opt,name=locked,proto3" json:"locked"`
	Registered               bool   `protobuf:"varint,6,opt,name=registered,proto3" json:"registered"`
	OwnerPubkeyHex           string `protobuf:"bytes,7,opt,name=owner_pubkey_hex,json=ownerPubkeyHex,proto3" json:"owner_pubkey_hex,omitempty"`
	TargetPubkeyHex          string `protobuf:"bytes,8,opt,name=target_pubkey_hex,json=targetPubkeyHex,proto3" json:"target_pubkey_hex,omitempty"`
	ExpireAtUnix             int64  `protobuf:"varint,9,opt,name=expire_at_unix,json=expireAtUnix,proto3" json:"expire_at_unix,omitempty"`
	LockExpiresAtUnix        int64  `protobuf:"varint,10,opt,name=lock_expires_at_unix,json=lockExpiresAtUnix,proto3" json:"lock_expires_at_unix,omitempty"`
	RegisterPriceSatoshi     uint64 `protobuf:"varint,11,opt,name=register_price_satoshi,json=registerPriceSatoshi,proto3" json:"register_price_satoshi"`
	RegisterSubmitFeeSatoshi uint64 `protobuf:"varint,12,opt,name=register_submit_fee_satoshi,json=registerSubmitFeeSatoshi,proto3" json:"register_submit_fee_satoshi"`
	RegisterLockFeeSatoshi   uint64 `protobuf:"varint,13,opt,name=register_lock_fee_satoshi,json=registerLockFeeSatoshi,proto3" json:"register_lock_fee_satoshi"`
	SetTargetFeeSatoshi      uint64 `protobuf:"varint,14,opt,name=set_target_fee_satoshi,json=setTargetFeeSatoshi,proto3" json:"set_target_fee_satoshi"`
	ResolveFeeSatoshi        uint64 `protobuf:"varint,15,opt,name=resolve_fee_satoshi,json=resolveFeeSatoshi,proto3" json:"resolve_fee_satoshi"`
	QueryFeeSatoshi          uint64 `protobuf:"varint,16,opt,name=query_fee_satoshi,json=queryFeeSatoshi,proto3" json:"query_fee_satoshi"`
	ChargedAmount            uint64 `protobuf:"varint,17,opt,name=charged_amount_satoshi,json=chargedAmountSatoshi,proto3" json:"charged_amount_satoshi,omitempty"`
	UpdatedTxID              string `protobuf:"bytes,18,opt,name=updated_txid,json=updatedTxid,proto3" json:"updated_txid,omitempty"`
	SignedRecordJSON         []byte `protobuf:"bytes,19,opt,name=signed_record_json,json=signedRecordJson,proto3" json:"signed_record_json,omitempty"`
	Error                    string `protobuf:"bytes,20,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
}

type RegisterLockPaidReq struct {
	ClientID string `protobuf:"bytes,1,opt,name=client_pubkey_hex,json=clientId,proto3" json:"client_pubkey_hex"`

	Name            string `protobuf:"bytes,2,opt,name=name,proto3" json:"name"`
	TargetPubkeyHex string `protobuf:"bytes,3,opt,name=target_pubkey_hex,json=targetPubkeyHex,proto3" json:"target_pubkey_hex"`

	SpendTxID           string `protobuf:"bytes,4,opt,name=spend_txid,json=spendTxid,proto3" json:"spend_txid"`
	SequenceNumber      uint32 `protobuf:"varint,5,opt,name=sequence_number,json=sequenceNumber,proto3" json:"sequence_number"`
	ServerAmount        uint64 `protobuf:"varint,6,opt,name=server_amount,json=serverAmount,proto3" json:"server_amount"`
	ChargeAmountSatoshi uint64 `protobuf:"varint,7,opt,name=charge_amount_satoshi,json=chargeAmountSatoshi,proto3" json:"charge_amount_satoshi"`
	Fee                 uint64 `protobuf:"varint,8,opt,name=fee,proto3" json:"fee"`
	ClientSignature     []byte `protobuf:"bytes,9,opt,name=client_signature,json=clientSignature,proto3" json:"client_signature"`
	ChargeReason        string `protobuf:"bytes,10,opt,name=charge_reason,json=chargeReason,proto3" json:"charge_reason,omitempty"`
}

type RegisterLockPaidResp struct {
	Success           bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success"`
	Status            string `protobuf:"bytes,2,opt,name=status,proto3" json:"status"`
	Name              string `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	TargetPubkeyHex   string `protobuf:"bytes,4,opt,name=target_pubkey_hex,json=targetPubkeyHex,proto3" json:"target_pubkey_hex,omitempty"`
	LockExpiresAtUnix int64  `protobuf:"varint,5,opt,name=lock_expires_at_unix,json=lockExpiresAtUnix,proto3" json:"lock_expires_at_unix,omitempty"`
	ChargedAmount     uint64 `protobuf:"varint,6,opt,name=charged_amount_satoshi,json=chargedAmountSatoshi,proto3" json:"charged_amount_satoshi,omitempty"`
	UpdatedTxID       string `protobuf:"bytes,7,opt,name=updated_txid,json=updatedTxid,proto3" json:"updated_txid,omitempty"`
	SignedQuoteJSON   []byte `protobuf:"bytes,8,opt,name=signed_quote_json,json=signedQuoteJson,proto3" json:"signed_quote_json,omitempty"`
	Error             string `protobuf:"bytes,9,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
}

type RegisterSubmitReq struct {
	ClientID string `protobuf:"bytes,1,opt,name=client_pubkey_hex,json=clientId,proto3" json:"client_pubkey_hex"`

	RegisterTx []byte `protobuf:"bytes,2,opt,name=register_tx,json=registerTx,proto3" json:"register_tx"`
}

type RegisterSubmitResp struct {
	Success           bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success"`
	Status            string `protobuf:"bytes,2,opt,name=status,proto3" json:"status"`
	Name              string `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	OwnerPubkeyHex    string `protobuf:"bytes,4,opt,name=owner_pubkey_hex,json=ownerPubkeyHex,proto3" json:"owner_pubkey_hex,omitempty"`
	TargetPubkeyHex   string `protobuf:"bytes,5,opt,name=target_pubkey_hex,json=targetPubkeyHex,proto3" json:"target_pubkey_hex,omitempty"`
	ExpireAtUnix      int64  `protobuf:"varint,6,opt,name=expire_at_unix,json=expireAtUnix,proto3" json:"expire_at_unix,omitempty"`
	RegisterTxID      string `protobuf:"bytes,7,opt,name=register_txid,json=registerTxid,proto3" json:"register_txid,omitempty"`
	SignedReceiptJSON []byte `protobuf:"bytes,8,opt,name=signed_receipt_json,json=signedReceiptJson,proto3" json:"signed_receipt_json,omitempty"`
	Error             string `protobuf:"bytes,9,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
}

type SetTargetPaidReq struct {
	ClientID string `protobuf:"bytes,1,opt,name=client_pubkey_hex,json=clientId,proto3" json:"client_pubkey_hex"`

	Name            string `protobuf:"bytes,2,opt,name=name,proto3" json:"name"`
	TargetPubkeyHex string `protobuf:"bytes,3,opt,name=target_pubkey_hex,json=targetPubkeyHex,proto3" json:"target_pubkey_hex"`

	SpendTxID           string `protobuf:"bytes,4,opt,name=spend_txid,json=spendTxid,proto3" json:"spend_txid"`
	SequenceNumber      uint32 `protobuf:"varint,5,opt,name=sequence_number,json=sequenceNumber,proto3" json:"sequence_number"`
	ServerAmount        uint64 `protobuf:"varint,6,opt,name=server_amount,json=serverAmount,proto3" json:"server_amount"`
	ChargeAmountSatoshi uint64 `protobuf:"varint,7,opt,name=charge_amount_satoshi,json=chargeAmountSatoshi,proto3" json:"charge_amount_satoshi"`
	Fee                 uint64 `protobuf:"varint,8,opt,name=fee,proto3" json:"fee"`
	ClientSignature     []byte `protobuf:"bytes,9,opt,name=client_signature,json=clientSignature,proto3" json:"client_signature"`
	ChargeReason        string `protobuf:"bytes,10,opt,name=charge_reason,json=chargeReason,proto3" json:"charge_reason,omitempty"`
}

type SetTargetPaidResp struct {
	Success          bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success"`
	Status           string `protobuf:"bytes,2,opt,name=status,proto3" json:"status"`
	Name             string `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	OwnerPubkeyHex   string `protobuf:"bytes,4,opt,name=owner_pubkey_hex,json=ownerPubkeyHex,proto3" json:"owner_pubkey_hex,omitempty"`
	TargetPubkeyHex  string `protobuf:"bytes,5,opt,name=target_pubkey_hex,json=targetPubkeyHex,proto3" json:"target_pubkey_hex,omitempty"`
	ExpireAtUnix     int64  `protobuf:"varint,6,opt,name=expire_at_unix,json=expireAtUnix,proto3" json:"expire_at_unix,omitempty"`
	ChargedAmount    uint64 `protobuf:"varint,7,opt,name=charged_amount_satoshi,json=chargedAmountSatoshi,proto3" json:"charged_amount_satoshi,omitempty"`
	UpdatedTxID      string `protobuf:"bytes,8,opt,name=updated_txid,json=updatedTxid,proto3" json:"updated_txid,omitempty"`
	SignedRecordJSON []byte `protobuf:"bytes,9,opt,name=signed_record_json,json=signedRecordJson,proto3" json:"signed_record_json,omitempty"`
	Error            string `protobuf:"bytes,10,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
}
