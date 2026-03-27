package domainsvc

type NameRouteReq struct {
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name"`
}

type NameTargetRouteReq struct {
	Name            string `protobuf:"bytes,1,opt,name=name,proto3" json:"name"`
	TargetPubkeyHex string `protobuf:"bytes,2,opt,name=target_pubkey_hex,json=targetPubkeyHex,proto3" json:"target_pubkey_hex"`
}

type DomainPricingBody struct {
	ResolveFeeSatoshi        uint64 `protobuf:"varint,1,opt,name=resolve_fee_satoshi,json=resolveFeeSatoshi,proto3" json:"resolve_fee_satoshi"`
	QueryFeeSatoshi          uint64 `protobuf:"varint,2,opt,name=query_fee_satoshi,json=queryFeeSatoshi,proto3" json:"query_fee_satoshi"`
	RegisterLockFeeSatoshi   uint64 `protobuf:"varint,3,opt,name=register_lock_fee_satoshi,json=registerLockFeeSatoshi,proto3" json:"register_lock_fee_satoshi"`
	RegisterSubmitFeeSatoshi uint64 `protobuf:"varint,4,opt,name=register_submit_fee_satoshi,json=registerSubmitFeeSatoshi,proto3" json:"register_submit_fee_satoshi"`
	SetTargetFeeSatoshi      uint64 `protobuf:"varint,5,opt,name=set_target_fee_satoshi,json=setTargetFeeSatoshi,proto3" json:"set_target_fee_satoshi"`
	RegisterPriceSatoshi     uint64 `protobuf:"varint,6,opt,name=register_price_satoshi,json=registerPriceSatoshi,proto3" json:"register_price_satoshi"`
}

type ListOwnedReq struct {
	OwnerPubkeyHex string `protobuf:"bytes,1,opt,name=owner_pubkey_hex,json=ownerPubkeyHex,proto3" json:"owner_pubkey_hex,omitempty"`
	Limit          uint32 `protobuf:"varint,2,opt,name=limit,proto3" json:"limit,omitempty"`
}

type OwnedNameItem struct {
	Name            string `protobuf:"bytes,1,opt,name=name,proto3" json:"name"`
	OwnerPubkeyHex  string `protobuf:"bytes,2,opt,name=owner_pubkey_hex,json=ownerPubkeyHex,proto3" json:"owner_pubkey_hex"`
	TargetPubkeyHex string `protobuf:"bytes,3,opt,name=target_pubkey_hex,json=targetPubkeyHex,proto3" json:"target_pubkey_hex"`
	ExpireAtUnix    int64  `protobuf:"varint,4,opt,name=expire_at_unix,json=expireAtUnix,proto3" json:"expire_at_unix"`
	RegisterTxID    string `protobuf:"bytes,5,opt,name=register_txid,json=registerTxid,proto3" json:"register_txid,omitempty"`
	UpdatedAtUnix   int64  `protobuf:"varint,6,opt,name=updated_at_unix,json=updatedAtUnix,proto3" json:"updated_at_unix,omitempty"`
}

type ListOwnedResp struct {
	Success        bool             `protobuf:"varint,1,opt,name=success,proto3" json:"success"`
	Status         string           `protobuf:"bytes,2,opt,name=status,proto3" json:"status"`
	OwnerPubkeyHex string           `protobuf:"bytes,3,opt,name=owner_pubkey_hex,json=ownerPubkeyHex,proto3" json:"owner_pubkey_hex"`
	Items          []*OwnedNameItem `protobuf:"bytes,4,rep,name=items,proto3" json:"items,omitempty"`
	Total          int32            `protobuf:"varint,5,opt,name=total,proto3" json:"total"`
	Error          string           `protobuf:"bytes,6,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
	QueriedAtUnix  int64            `protobuf:"varint,7,opt,name=queried_at_unix,json=queriedAtUnix,proto3" json:"queried_at_unix,omitempty"`
}
