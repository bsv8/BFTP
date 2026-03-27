package proof

import (
	"encoding/json"
	"fmt"
)

// signedArrayEnvelope 统一签名协议的线格式。
// 设计说明：
// - 线格式固定为 [[字段...], "signature_hex"]；
// - 外层只保留“原文数组 + 签名”，避免以后再手工拼最后一个签名字段。
type signedArrayEnvelope struct {
	Fields       []any
	SignatureHex string
}

func marshalSignedArrayEnvelope(fields []any, signatureHex string) ([]byte, error) {
	return json.Marshal([]any{append([]any(nil), fields...), signatureHex})
}

func unmarshalSignedArrayEnvelope(raw []byte) ([]json.RawMessage, string, error) {
	var parts []json.RawMessage
	if err := json.Unmarshal(raw, &parts); err != nil {
		return nil, "", err
	}
	if len(parts) != 2 {
		return nil, "", fmt.Errorf("signed envelope fields mismatch")
	}
	var fields []json.RawMessage
	if err := json.Unmarshal(parts[0], &fields); err != nil {
		return nil, "", err
	}
	var signatureHex string
	if err := json.Unmarshal(parts[1], &signatureHex); err != nil {
		return nil, "", err
	}
	return fields, signatureHex, nil
}
