package domainsvc

type ListOwnedReq struct {
	OwnerPubkeyHex string `json:"owner_pubkey_hex,omitempty"`
	Limit          uint32 `json:"limit,omitempty"`
}

type OwnedNameItem struct {
	Name            string `json:"name"`
	OwnerPubkeyHex  string `json:"owner_pubkey_hex"`
	TargetPubkeyHex string `json:"target_pubkey_hex"`
	ExpireAtUnix    int64  `json:"expire_at_unix"`
	RegisterTxID    string `json:"register_txid,omitempty"`
	UpdatedAtUnix   int64  `json:"updated_at_unix,omitempty"`
}

type ListOwnedResp struct {
	Success        bool            `json:"success"`
	Status         string          `json:"status"`
	OwnerPubkeyHex string          `json:"owner_pubkey_hex"`
	Items          []OwnedNameItem `json:"items,omitempty"`
	Total          int             `json:"total"`
	Error          string          `json:"error_message,omitempty"`
	QueriedAtUnix  int64           `json:"queried_at_unix,omitempty"`
}

type RegisterSubmitRouteReq struct {
	RegisterTxHex string `json:"register_tx_hex"`
}

type RegisterSubmitRouteResp struct {
	Success           bool   `json:"success"`
	Status            string `json:"status"`
	Name              string `json:"name,omitempty"`
	OwnerPubkeyHex    string `json:"owner_pubkey_hex,omitempty"`
	TargetPubkeyHex   string `json:"target_pubkey_hex,omitempty"`
	ExpireAtUnix      int64  `json:"expire_at_unix,omitempty"`
	RegisterTxID      string `json:"register_txid,omitempty"`
	SignedReceiptJSON []byte `json:"signed_receipt_json,omitempty"`
	Error             string `json:"error_message,omitempty"`
}
